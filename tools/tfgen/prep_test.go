package main

import (
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func parse(t *testing.T, document string) map[string]any {
	t.Helper()

	var doc map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(document), &doc))

	return doc
}

func runPrepOn(t *testing.T, document string) (map[string]any, []string) {
	t.Helper()

	doc := parse(t, document)
	prep := NewPrep(doc)
	prep.Run()

	return doc, prep.Warnings
}

func schema(t *testing.T, doc map[string]any, name string) map[string]any {
	t.Helper()

	schemas, ok := nestedMap(doc, "components", "schemas")
	require.True(t, ok, "document has no component schemas")

	out, ok := schemas[name].(map[string]any)
	require.True(t, ok, "no component schema named %q", name)

	return out
}

func TestPrepCollapsesOneOfIntoObjectUnion(t *testing.T) {
	doc, warnings := runPrepOn(t, `
components:
  schemas:
    Ipv4:
      type: object
      required: [ipv4, name]
      additionalProperties: false
      properties:
        ipv4: {type: string}
        name: {type: string}
    Fqdn:
      type: object
      required: [fqdn, name]
      properties:
        fqdn: {type: string}
        name: {type: string}
    Monitor:
      type: object
      required: [id]
      properties:
        id: {type: string}
      oneOf:
        - $ref: '#/components/schemas/Ipv4'
        - $ref: '#/components/schemas/Fqdn'
`)

	require.Empty(t, warnings)

	monitor := schema(t, doc, "Monitor")
	require.Equal(t, "object", monitor["type"])
	require.NotContains(t, monitor, "oneOf")
	require.NotContains(t, monitor, "additionalProperties")

	properties, _ := monitor["properties"].(map[string]any)
	require.NotNil(t, properties)

	// The object keeps what it declared beside the composition, and gains a block
	// per form named after the component the form came from.
	require.Contains(t, properties, "id")
	require.ElementsMatch(t, []string{"id", "ipv4", "fqdn"}, slices.Sorted(maps.Keys(properties)))
	require.ElementsMatch(t, []string{"id"}, stringList(monitor["required"]),
		"a form cannot be required: which one is set is the choice being made")

	ipv4, _ := properties["ipv4"].(map[string]any)
	require.Equal(t, "object", ipv4["type"])
	require.ElementsMatch(t, []string{"ipv4", "name"}, slices.Sorted(maps.Keys(ipv4["properties"].(map[string]any))))
	require.ElementsMatch(t, []string{"ipv4", "name"}, stringList(ipv4["required"]),
		"inside a form, the API's own requirements still hold")

	// The marker is what lets the registry find the blocks again in the generated
	// code specification.
	require.Contains(t, stringOr(ipv4["description"]), VariantBlockDescriptionSuffix)

	fqdn, _ := properties["fqdn"].(map[string]any)
	require.ElementsMatch(t, []string{"fqdn", "name"}, slices.Sorted(maps.Keys(fqdn["properties"].(map[string]any))))
}

// TestPrepNamesBlocksFromTheDiscriminatorMapping covers the case the spec names
// itself: a discriminator says which value selects which form, so the blocks take
// those values as their names.
func TestPrepNamesBlocksFromTheDiscriminatorMapping(t *testing.T) {
	doc, warnings := runPrepOn(t, `
components:
  schemas:
    Wanlink:
      type: object
      required: [type]
      properties:
        type: {type: string, enum: [wanlink]}
        frequency: {type: integer}
    Overlay:
      type: object
      required: [type]
      properties:
        type: {type: string, enum: [overlay]}
        private: {type: boolean}
    Tag:
      oneOf:
        - $ref: '#/components/schemas/Wanlink'
        - $ref: '#/components/schemas/Overlay'
      discriminator:
        propertyName: type
        mapping:
          wanlink: '#/components/schemas/Wanlink'
          overlay: '#/components/schemas/Overlay'
`)

	require.Empty(t, warnings)

	tag := schema(t, doc, "Tag")
	require.NotContains(t, tag, "discriminator")

	properties, _ := tag["properties"].(map[string]any)
	require.ElementsMatch(t, []string{"wanlink", "overlay"}, slices.Sorted(maps.Keys(properties)))

	// Each form keeps its own fields, so a frequency is only settable on the form
	// that has one.
	wanlinkBlock, _ := properties["wanlink"].(map[string]any)
	require.Contains(t, wanlinkBlock["properties"], "frequency")
	require.NotContains(t, wanlinkBlock["properties"], "private")

	overlayBlock, _ := properties["overlay"].(map[string]any)
	require.Contains(t, overlayBlock["properties"], "private")
	require.NotContains(t, overlayBlock["properties"], "frequency")

	// Splitting must not rewrite the variants themselves: they are still referenced
	// elsewhere in the document.
	wanlink := schema(t, doc, "Wanlink")
	wanlinkProperties, _ := wanlink["properties"].(map[string]any)
	wanlinkType, _ := wanlinkProperties["type"].(map[string]any)
	require.Equal(t, []string{"wanlink"}, anyStrings(wanlinkType["enum"]))
	require.NotContains(t, stringOr(wanlink["description"]), VariantBlockDescriptionSuffix)
}

func TestPrepDropsScalarVariantOfMixedUnion(t *testing.T) {
	doc, warnings := runPrepOn(t, `
components:
  schemas:
    Probe:
      type: object
      properties:
        http:
          type: array
          items:
            oneOf:
              - type: object
                properties:
                  ip: {type: string}
              - type: string
`)

	require.Len(t, warnings, 1)
	require.Contains(t, warnings[0], "only the object form is exposed")

	probe := schema(t, doc, "Probe")
	properties, _ := probe["properties"].(map[string]any)
	http, _ := properties["http"].(map[string]any)
	items, _ := http["items"].(map[string]any)

	require.Equal(t, "object", items["type"])
	require.NotContains(t, items, "oneOf")
	require.Contains(t, items["properties"], "ip")
}

func TestPrepKeepsNullableAnyOfForTheGenerator(t *testing.T) {
	doc, warnings := runPrepOn(t, `
components:
  schemas:
    Thing:
      type: object
      properties:
        name:
          anyOf:
            - type: string
            - type: "null"
`)

	require.Empty(t, warnings)

	thing := schema(t, doc, "Thing")
	properties, _ := thing["properties"].(map[string]any)
	name, _ := properties["name"].(map[string]any)

	require.Contains(t, name, "anyOf", "a single variant next to null is how the spec spells nullable")
}

func TestPrepTypesUntypedSchemaAsJSONString(t *testing.T) {
	doc, warnings := runPrepOn(t, `
components:
  schemas:
    Template:
      type: object
      properties:
        device_config_raw: {}
        described:
          description: The device configuration.
`)

	require.Empty(t, warnings)

	template := schema(t, doc, "Template")
	properties, _ := template["properties"].(map[string]any)

	wantMarker := RawJSONDescriptionSuffix

	raw, _ := properties["device_config_raw"].(map[string]any)
	require.Equal(t, "string", raw["type"])
	require.Equal(t, wantMarker, stringOr(raw["description"]))

	// An existing description is kept, with the marker appended.
	described, _ := properties["described"].(map[string]any)
	require.Equal(t, "string", described["type"])
	require.Equal(t, "The device configuration. "+wantMarker, stringOr(described["description"]))
}

func TestPrepDocumentsEnumValues(t *testing.T) {
	doc, warnings := runPrepOn(t, `
components:
  schemas:
    Gateway:
      type: object
      properties:
        model:
          type: string
          enum: [NSGVirtual, NSG100W]
        protocol:
          type: string
          description: Protocol Enum
          enum: [TCP, UDP]
`)

	require.Empty(t, warnings)

	gateway := schema(t, doc, "Gateway")
	properties, _ := gateway["properties"].(map[string]any)

	model, _ := properties["model"].(map[string]any)
	require.Equal(t, "Must be one of: `NSGVirtual`, `NSG100W`.", stringOr(model["description"]))

	// An existing description is kept, with the allowed values appended.
	protocol, _ := properties["protocol"].(map[string]any)
	require.Equal(t, "Protocol Enum Must be one of: `TCP`, `UDP`.", stringOr(protocol["description"]))
}

func TestPrepRenamesDashedPathParameters(t *testing.T) {
	doc, warnings := runPrepOn(t, `
paths:
  /address-groups/{group-id}/address-objects/{id}:
    get:
      parameters:
        - {name: group-id, in: path, schema: {type: string}}
        - {name: id, in: path, schema: {type: string}}
        - {name: sort-by, in: query, schema: {type: string}}
`)

	require.Empty(t, warnings)

	paths, _ := doc["paths"].(map[string]any)
	require.Contains(t, paths, "/address-groups/{group_id}/address-objects/{id}")
	require.NotContains(t, paths, "/address-groups/{group-id}/address-objects/{id}")

	item, _ := paths["/address-groups/{group_id}/address-objects/{id}"].(map[string]any)
	operation, _ := item["get"].(map[string]any)
	parameters, _ := operation["parameters"].([]any)

	var names []string

	for _, parameter := range parameters {
		asMap, _ := parameter.(map[string]any)
		name, _ := asMap["name"].(string)
		names = append(names, name)
	}

	// Only path parameters become attribute names, so a query parameter keeps
	// the name the API expects on the wire.
	require.Equal(t, []string{"group_id", "id", "sort-by"}, names)
}

func TestPrepDropClearsTheWayForARename(t *testing.T) {
	doc := parse(t, `
components:
  schemas:
    CustomApp:
      type: object
      required: [name]
      properties:
        name: {type: string}
        definitions: {type: array, items: {type: string}}
        definitions_v2: {type: array, items: {type: string}}
`)

	prep := NewPrep(doc)
	prep.Drop("definitions")
	prep.Rename("definitions_v2", "definitions")
	prep.Run()

	require.Empty(t, prep.Warnings)

	customApp := schema(t, doc, "CustomApp")
	properties, _ := customApp["properties"].(map[string]any)

	require.ElementsMatch(t, []string{"name", "definitions"}, slices.Sorted(maps.Keys(properties)))
}

func TestPrepDropRemovesTheFieldFromRequired(t *testing.T) {
	doc := parse(t, `
components:
  schemas:
    Thing:
      type: object
      required: [name, legacy]
      properties:
        name: {type: string}
        legacy: {type: string}
`)

	prep := NewPrep(doc)
	prep.Drop("legacy")
	prep.Run()

	thing := schema(t, doc, "Thing")
	require.Equal(t, []string{"name"}, stringList(thing["required"]))
}

func TestPrepReportsAllOf(t *testing.T) {
	_, warnings := runPrepOn(t, `
components:
  schemas:
    Thing:
      allOf:
        - type: object
          properties:
            name: {type: string}
`)

	require.Len(t, warnings, 1)
	require.Contains(t, warnings[0], "allOf is not collapsed")
}

func TestPrepSurvivesRecursiveRef(t *testing.T) {
	_, warnings := runPrepOn(t, `
components:
  schemas:
    Node:
      type: object
      properties:
        child: {$ref: '#/components/schemas/Node'}
      oneOf:
        - $ref: '#/components/schemas/Node'
        - type: object
          properties:
            leaf: {type: string}
`)

	require.NotEmpty(t, warnings)
	require.Contains(t, warnings[0], "recursive $ref")
}

// runPrepClaiming is prep as `tfgen prep` runs it for a configuration that binds
// variants of a path to Terraform types of their own.
func runPrepClaiming(t *testing.T, document string, claims map[string][]string) (map[string]any, []string) {
	t.Helper()

	doc := parse(t, document)
	prep := NewPrep(doc)

	for _, path := range slices.Sorted(maps.Keys(claims)) {
		for _, name := range claims[path] {
			prep.Claim(path, name)
		}
	}

	prep.Run()

	return doc, prep.Warnings
}

func variantMetadataOf(t *testing.T, doc map[string]any, path string) map[string]any {
	t.Helper()

	paths, _ := doc["paths"].(map[string]any)

	item, ok := paths[path].(map[string]any)
	require.True(t, ok, "no path %q; the document has %v", path, slices.Sorted(maps.Keys(paths)))

	out, ok := item[VariantExtension].(map[string]any)
	require.True(t, ok, "path %q carries no %s", path, VariantExtension)

	return out
}

// The BWAN spec nests the choice between the four kinds of tag inside the tag's
// `config`, and names the kinds through a discriminator mapping without any
// branch declaring the `type` the mapping selects on. Both halves have to reach
// the runtime: where the kind is written on an object it reads back, and how to
// write the kind on an object it creates.
func TestPrepTakesAVariantSelectedByANestedDiscriminator(t *testing.T) {
	doc, warnings := runPrepClaiming(t, `
paths:
  /overlay-tags:
    get:
      responses:
        "200":
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items: {$ref: '#/components/schemas/OverlayTag'}
    post:
      requestBody:
        content:
          application/json:
            schema: {$ref: '#/components/schemas/OverlayTagCreate'}
      responses:
        "201":
          content:
            application/json:
              schema: {$ref: '#/components/schemas/OverlayTag'}
components:
  schemas:
    OverlayTagWanlinkConfig:
      type: object
      properties:
        wan_link_frequency: {type: integer}
    OverlayTagOverlayConfig:
      type: object
      properties:
        overlay_private: {type: boolean}
    OverlayTagConfig:
      oneOf:
        - $ref: '#/components/schemas/OverlayTagWanlinkConfig'
        - $ref: '#/components/schemas/OverlayTagOverlayConfig'
      discriminator:
        propertyName: type
        mapping:
          wanlink: '#/components/schemas/OverlayTagWanlinkConfig'
          overlay: '#/components/schemas/OverlayTagOverlayConfig'
    OverlayTag:
      type: object
      required: [id, name, config]
      properties:
        id: {type: string, readOnly: true}
        name: {type: string}
        config: {$ref: '#/components/schemas/OverlayTagConfig'}
    OverlayTagCreate:
      type: object
      required: [name, config]
      properties:
        name: {type: string}
        config: {$ref: '#/components/schemas/OverlayTagConfig'}
`, map[string][]string{"/overlay-tags": {"wanlink", "overlay"}})

	require.Empty(t, warnings)

	// The discriminator is reported where the runtime will look for it: on the
	// object it recognises, which is the tag rather than the tag's config, and
	// element by element for the collection the same path lists.
	require.Equal(t, map[string]any{
		"name":          "wanlink",
		"discriminator": "config.type",
		"value":         "wanlink",
		"match":         []any{"wan_link_frequency"},
		"wrapper":       "config",
		"wrapped":       []any{"wan_link_frequency"},
	}, variantMetadataOf(t, doc, "/overlay-tags@wanlink"))

	require.Equal(t, "config.type", variantMetadataOf(t, doc, "/overlay-tags@overlay")["discriminator"])
	require.Equal(t, "overlay", variantMetadataOf(t, doc, "/overlay-tags@overlay")["value"])

	// The branch is lifted out of the config it arrived in, and the discriminator
	// goes with it: this Terraform type is the wanlink kind, so a practitioner
	// writing `config = { type = "wanlink" }` would only be repeating the resource
	// name back. The runtime writes both from the metadata above.
	for _, name := range []string{"OverlayTagWanlink", "OverlayTagCreateWanlink"} {
		object := schema(t, doc, name)
		properties, _ := object["properties"].(map[string]any)

		require.Contains(t, properties, "wan_link_frequency", name)
		require.NotContains(t, properties, "config", name)
		require.NotContains(t, properties, "type", name)
		require.NotContains(t, properties, "overlay_private", name)

		require.Contains(t, stringList(object["required"]), "name", name)
		require.NotContains(t, stringList(object["required"]), "config", name)
		require.NotContains(t, stringList(object["required"]), "type", name)
	}

	// The other kind is hoisted into its own copy, not into this one.
	overlay, _ := schema(t, doc, "OverlayTagOverlay")["properties"].(map[string]any)
	require.Contains(t, overlay, "overlay_private")
	require.NotContains(t, overlay, "wan_link_frequency")
}

// A wrapper is only hoisted where what comes out of it can be told from what is
// already there. Anything else keeps the shape the API declared, because a schema
// that has quietly merged two fields into one is worse than a schema with a
// redundant level in it.
func TestPrepLeavesACollidingWrapperAlone(t *testing.T) {
	doc, warnings := runPrepClaiming(t, `
paths:
  /overlay-tags:
    get:
      responses:
        "200":
          content:
            application/json:
              schema: {$ref: '#/components/schemas/OverlayTag'}
components:
  schemas:
    OverlayTagWanlinkConfig:
      type: object
      properties:
        name: {type: string}
    OverlayTagOverlayConfig:
      type: object
      properties:
        overlay_private: {type: boolean}
    OverlayTagConfig:
      oneOf:
        - $ref: '#/components/schemas/OverlayTagWanlinkConfig'
        - $ref: '#/components/schemas/OverlayTagOverlayConfig'
      discriminator:
        propertyName: type
        mapping:
          wanlink: '#/components/schemas/OverlayTagWanlinkConfig'
          overlay: '#/components/schemas/OverlayTagOverlayConfig'
    OverlayTag:
      type: object
      required: [name, config]
      properties:
        name: {type: string}
        config: {$ref: '#/components/schemas/OverlayTagConfig'}
`, map[string][]string{"/overlay-tags": {"wanlink"}})

	require.Len(t, warnings, 1)
	require.Contains(t, warnings[0], `cannot hoist "config"`)
	require.Contains(t, warnings[0], `both declare "name"`)

	// Wrapped in the schema and no wrapper in the metadata: the two halves have to
	// agree, or the runtime would unwrap something that was never wrapped.
	properties, _ := schema(t, doc, "OverlayTagWanlink")["properties"].(map[string]any)
	require.Contains(t, properties, "config")

	require.Empty(t, variantMetadataOf(t, doc, "/overlay-tags@wanlink")["wrapper"])
}

// A branch named after anything other than a discriminator mapping is left
// alone: the name is not a value the API would recognise.
func TestPrepInventsNoDiscriminatorValueForAnUnmappedBranch(t *testing.T) {
	doc, warnings := runPrepClaiming(t, `
paths:
  /monitors:
    get:
      responses:
        "200":
          content:
            application/json:
              schema: {$ref: '#/components/schemas/Monitor'}
components:
  schemas:
    Fqdn:
      type: object
      required: [fqdn]
      properties:
        fqdn: {type: string}
    Ipv4:
      type: object
      required: [ipv4]
      properties:
        ipv4: {type: string}
    Monitor:
      oneOf:
        - $ref: '#/components/schemas/Fqdn'
        - $ref: '#/components/schemas/Ipv4'
`, map[string][]string{"/monitors": {"fqdn"}})

	require.Empty(t, warnings)

	metadata := variantMetadataOf(t, doc, "/monitors@fqdn")
	require.Empty(t, metadata["discriminator"])
	require.Empty(t, metadata["value"])
	require.Equal(t, []any{"fqdn"}, metadata["match"])

	// Nothing was hoisted either: with no value to write in its place, a
	// discriminator the runtime removed from a request could not be put back.
	require.Empty(t, metadata["wrapper"])

	properties, _ := schema(t, doc, "MonitorFqdn")["properties"].(map[string]any)
	require.NotContains(t, properties, "type")
}
