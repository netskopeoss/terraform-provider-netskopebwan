package main

import (
	"maps"
	"slices"
	"strings"
)

// ElementSuffix names the synthetic path prep emits for one element of a
// collection, for an object the API only ever lists. Like the paths emitted for
// variants it is one only the generator ever sees — it carries the separator for
// exactly that reason, since no API path can contain it — and the runtime goes on
// calling the collection.
const ElementSuffix = VariantSeparator + "element"

// The properties of the envelope every BWAN collection endpoint answers with.
const (
	dataProperty       = "data"
	pageInfoProperty   = "page_info"
	totalCountProperty = "total_count"
)

// The cursor parameters a collection endpoint takes. Neither reaches Terraform:
// the runtime walks every page itself.
const (
	afterParameter = "after"
	firstParameter = "first"
)

// totalCountDescription documents the one part of the pagination envelope that
// still means something once every page has been read.
const totalCountDescription = "Number of objects the API reports for this collection."

// reshapeCollections rewrites every collection endpoint into the shape a list
// data source exposes: `filter` and `sort` to narrow the list with, and `data`
// beside the count the API reports for it.
//
// The cursor pair is deliberately dropped rather than mapped. The runtime walks
// every page of a collection, so `first` and `after` could only ask for a slice
// of a list the provider has already read in full, and a data source quietly
// holding one page of several is the thing that walk exists to avoid. `page_info`
// goes the same way: `end_cursor` and `has_next` describe a walk that has
// finished by the time state is written, which leaves `total_count` as the only
// field worth keeping, one level up.
func (p *Prep) reshapeCollections() {
	paths, ok := p.doc["paths"].(map[string]any)
	if !ok {
		return
	}

	for _, path := range slices.Sorted(maps.Keys(paths)) {
		item, ok := paths[path].(map[string]any)
		if !ok {
			continue
		}

		operation, ok := item["get"].(map[string]any)
		if !ok {
			continue
		}

		envelope := p.collectionEnvelope(operation)
		if envelope == nil {
			continue
		}

		p.hoistTotalCount(envelope, path)
		dropQueryParameters(operation, afterParameter, firstParameter)
	}
}

// hoistTotalCount replaces the pagination object with the one field of it a
// practitioner can still use, keeping whatever the spec said about that field.
func (p *Prep) hoistTotalCount(envelope map[string]any, loc string) {
	properties, _ := envelope["properties"].(map[string]any)

	total := p.totalCountSchema(properties[pageInfoProperty], loc)
	delete(properties, pageInfoProperty)

	if _, taken := properties[totalCountProperty]; taken {
		p.warnf("%s already has a %q of its own beside %q", loc, totalCountProperty, pageInfoProperty)
	} else {
		properties[totalCountProperty] = total
	}

	required := removedFromList(stringList(envelope["required"]), pageInfoProperty)
	setRequired(envelope, union(required, []string{totalCountProperty}))
}

// totalCountSchema takes the count out of the pagination object, so the attribute
// keeps the type and description the spec gave it, and falls back to a plain
// integer where the spec describes the envelope less fully than the API fills it.
func (p *Prep) totalCountSchema(pageInfo any, loc string) map[string]any {
	fallback := map[string]any{
		"type":        "integer",
		"readOnly":    true,
		"description": totalCountDescription,
	}

	asMap, ok := pageInfo.(map[string]any)
	if !ok {
		return fallback
	}

	target := p.dereference(asMap)
	if target == nil {
		return fallback
	}

	properties, _ := target["properties"].(map[string]any)

	total, ok := deepCopy(properties[totalCountProperty]).(map[string]any)
	if !ok {
		p.warnf("%s: %s declares no %q, exposing a plain integer instead", loc, pageInfoProperty, totalCountProperty)

		return fallback
	}

	if stringOr(total["description"]) == "" {
		total["description"] = totalCountDescription
	}

	return total
}

// ClaimElement records a collection a data source stands for one element of,
// which is how an object the API only ever lists still gets a data source of its
// own.
func (p *Prep) ClaimElement(collection string) {
	if !slices.Contains(p.elements, collection) {
		p.elements = append(p.elements, collection)
	}
}

// emitElementPaths emits a synthetic single-object GET beside every collection a
// data source claims an element of.
//
// Not every collection has a single-object endpoint: app categories, audit events
// and the software catalogue are only ever listed. Without one there is no schema
// for the generator to map a single object from, so those objects could only be
// reached through the list data source — a configuration wanting one of them had
// to read the whole collection and index into it.
//
// The element's schema is right there in the collection's envelope, though, and
// the runtime already knows how to find one object in a collection: it is how a
// resource whose read path has no `{id}` is refreshed. So the schema is lifted
// onto a path of its own for the generator to map, the same trick and the same
// separator as the paths emitted per variant. Nothing in the configuration and
// nothing at runtime names that path: the configuration names the collection,
// because the collection is what the API has.
func (p *Prep) emitElementPaths() {
	paths, ok := p.doc["paths"].(map[string]any)
	if !ok {
		return
	}

	for _, collection := range slices.Sorted(slices.Values(p.elements)) {
		item, ok := paths[collection].(map[string]any)
		if !ok {
			p.warnf("generator_config claims an element of %s, which the spec does not have", collection)

			continue
		}

		operation, ok := item["get"].(map[string]any)
		if !ok {
			p.warnf("generator_config claims an element of %s, which the spec does not list", collection)

			continue
		}

		items := p.collectionItems(operation)
		if items == nil {
			p.warnf("generator_config claims an element of %s, which does not answer with a collection", collection)

			continue
		}

		// An element is addressed by its id, so an element schema without one is a
		// data source nothing could ask for.
		if !p.declaresID(items) {
			p.warnf("generator_config claims an element of %s, whose elements have no %q to address one by",
				collection, idAttribute)
		}

		paths[collection+ElementSuffix] = map[string]any{
			"get": map[string]any{
				"summary": "Get one element of " + collection,
				// Whatever the collection lives under is still needed to reach the
				// element; the parameters that narrow a list are not, and the id is not
				// a parameter of a request that is never made.
				"parameters": pathParameters(operation),
				"responses": map[string]any{
					"200": map[string]any{
						"description": "One element of " + collection,
						"content": map[string]any{
							"application/json": map[string]any{"schema": deepCopy(items)},
						},
					},
				},
			},
		}
	}
}

// declaresID reports whether the elements of a collection carry an id.
func (p *Prep) declaresID(items map[string]any) bool {
	target := p.dereference(items)
	if target == nil {
		return false
	}

	properties, _ := target["properties"].(map[string]any)
	_, ok := properties[idAttribute]

	return ok
}

// elementCollection returns the collection a single-object read path addresses an
// object in.
func elementCollection(path string) (string, bool) {
	const suffix = "/{" + idAttribute + "}"

	if !strings.HasSuffix(path, suffix) {
		return "", false
	}

	collection := strings.TrimSuffix(path, suffix)
	if collection == "" {
		return "", false
	}

	return collection, true
}

// collectionItems returns the schema of one element of the collection an
// operation lists.
func (p *Prep) collectionItems(operation map[string]any) map[string]any {
	envelope := p.responseSchema(operation)
	if envelope == nil {
		return nil
	}

	properties, ok := envelope["properties"].(map[string]any)
	if !ok {
		return nil
	}

	data, ok := properties[dataProperty].(map[string]any)
	if !ok {
		return nil
	}

	items, ok := data["items"].(map[string]any)
	if !ok {
		return nil
	}

	return items
}

// collectionEnvelope returns the response schema of a list operation, or nil when
// the operation answers with something else. A collection is recognised by the
// envelope the API is consistent about: a `data` array beside the pagination
// object.
func (p *Prep) collectionEnvelope(operation map[string]any) map[string]any {
	schema := p.responseSchema(operation)
	if schema == nil {
		return nil
	}

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return nil
	}

	if _, ok := properties[dataProperty]; !ok {
		return nil
	}

	if _, ok := properties[pageInfoProperty]; !ok {
		return nil
	}

	return schema
}

// responseSchema returns the schema of an operation's successful JSON response.
func (p *Prep) responseSchema(operation map[string]any) map[string]any {
	media, ok := nestedMap(operation, "responses", "200", "content", "application/json")
	if !ok {
		return nil
	}

	schema, ok := media["schema"].(map[string]any)
	if !ok {
		return nil
	}

	return p.dereference(schema)
}

// dereference follows a `$ref` to the component it names, so an envelope the spec
// declares once is reshaped where it is declared rather than per path.
func (p *Prep) dereference(schema map[string]any) map[string]any {
	ref, ok := schema["$ref"].(string)
	if !ok {
		return schema
	}

	name, ok := componentName(ref)
	if !ok {
		return nil
	}

	target, _ := p.schemas[name].(map[string]any)

	return target
}

// pathParameters returns the parameters an operation takes in its path, which for
// a collection is whatever that collection lives under.
func pathParameters(operation map[string]any) []any {
	var out []any

	for _, parameter := range parametersOf(operation) {
		if parameter["in"] == "path" {
			out = append(out, deepCopy(parameter))
		}
	}

	return out
}

// dropQueryParameters removes named query parameters from an operation, which is
// how a parameter the provider handles itself stops reaching Terraform as an
// attribute.
func dropQueryParameters(operation map[string]any, names ...string) {
	parameters, ok := operation["parameters"].([]any)
	if !ok {
		return
	}

	kept := make([]any, 0, len(parameters))

	for _, parameter := range parameters {
		asMap, ok := parameter.(map[string]any)
		if ok && asMap["in"] == "query" && slices.Contains(names, stringOr(asMap["name"])) {
			continue
		}

		kept = append(kept, parameter)
	}

	if len(kept) == 0 {
		delete(operation, "parameters")

		return
	}

	operation["parameters"] = kept
}

// parametersOf lists an operation's parameters.
func parametersOf(operation map[string]any) []map[string]any {
	raw, _ := operation["parameters"].([]any)
	out := make([]map[string]any, 0, len(raw))

	for _, parameter := range raw {
		if asMap, ok := parameter.(map[string]any); ok {
			out = append(out, asMap)
		}
	}

	return out
}
