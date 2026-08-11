package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// write drops content into a temporary file and returns its path.
func write(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

// specWith builds a provider code specification holding one resource with the
// given attributes.
func specWith(t *testing.T, name, attributes string) string {
	t.Helper()

	return write(t, "spec.json", `{
		"provider": {"name": "bwan"},
		"resources": [
			{"name": "`+name+`", "schema": {"attributes": [`+attributes+`]}}
		]
	}`)
}

const nameAttribute = `{"name": "name", "string": {"computed_optional_required": "required"}}`

func opaqueAttribute(name string) string {
	return `{"name": "` + name + `", "string": {"computed_optional_required": "required", "description": "` + RawJSONDescriptionSuffix + `"}}`
}

func crud(name, path string) string {
	return `  ` + name + `:
    create: {path: "` + path + `", method: POST}
    read: {path: "` + path + `/{id}", method: GET}
    delete: {path: "` + path + `/{id}", method: DELETE}
`
}

func TestRegistryGeneratesAnEntryPerObject(t *testing.T) {
	config := write(t, "generator_config.yml", "resources:\n"+crud("segment", "/segments"))
	spec := specWith(t, "segment", nameAttribute)

	source, err := generateRegistry(config, spec, "", "example.com/mod")

	require.NoError(t, err)
	require.Contains(t, string(source), `resSegment "example.com/mod/internal/gen/resources/resource_segment"`)
	require.Contains(t, string(source), `Name:   "segment",`)
	require.Contains(t, string(source), `Schema: resSegment.SegmentResourceSchema,`)
	require.Contains(t, string(source), `Create: genresource.Operation{Method: "POST", Path: "/segments"}`)
	require.Contains(t, string(source), `func RawFeatures() []string {`)
	require.Contains(t, string(source), `return []string{}`)
}

func TestRegistryRecordsEmbeddedJSONAttributes(t *testing.T) {
	config := write(t, "generator_config.yml", "resources:\n"+crud("cloud_account", "/cloud-accounts"))
	spec := specWith(t, "cloud_account", `
		`+nameAttribute+`,
		{"name": "config", "single_nested": {"computed_optional_required": "required", "attributes": [
			`+opaqueAttribute("creds")+`
		]}}
	`)

	source, err := generateRegistry(config, spec, "", "example.com/mod")

	require.NoError(t, err)
	require.Contains(t, string(source), `RawJSONAttributes: []string{"config.creds"}`)

	// An opaque credentials blob is not an opaque *configuration*: there is no
	// typed replacement to roll out, so it needs no gate.
	require.NotContains(t, string(source), "RawFeature:")
}

func TestRegistryRefusesUngatedOpaqueConfiguration(t *testing.T) {
	config := write(t, "generator_config.yml", "resources:\n"+crud("gateway", "/gateways"))
	spec := specWith(t, "gateway", nameAttribute+`, `+opaqueAttribute("device_config_raw"))

	_, err := generateRegistry(config, spec, "", "example.com/mod")

	require.ErrorContains(t, err, "resource gateway carries its configuration as an opaque document (device_config_raw)")
	require.ErrorContains(t, err, "rename it to gateway_raw and set x_terraform.raw_feature")
}

func TestRegistryRefusesAGatedObjectWithoutTheRawSuffix(t *testing.T) {
	config := write(t, "generator_config.yml", "resources:\n"+crud("gateway", "/gateways")+
		"    x_terraform:\n      raw_feature: gateway\n")
	spec := specWith(t, "gateway", nameAttribute+`, `+opaqueAttribute("device_config_raw"))

	_, err := generateRegistry(config, spec, "", "example.com/mod")

	require.ErrorContains(t, err, `is not named with a "_raw" suffix`)
}

// TestRegistryAcceptsAGateWithNoOpaqueAttribute covers an object gated for a
// reason other than an opaque document — a shorthand the provider cannot express,
// say. The gate is still worth having; what matters is that it leaves the plain
// name free.
func TestRegistryAcceptsAGateWithNoOpaqueAttribute(t *testing.T) {
	config := write(t, "generator_config.yml", "resources:\n"+crud("client_template_raw", "/client-templates")+
		"    x_terraform:\n      raw_feature: client_template\n")
	spec := specWith(t, "client_template_raw", nameAttribute)

	source, err := generateRegistry(config, spec, "", "example.com/mod")

	require.NoError(t, err)
	require.Contains(t, string(source), `RawFeature: "client_template",`)
}

func TestRegistryRefusesAGateOnAPlainName(t *testing.T) {
	config := write(t, "generator_config.yml", "resources:\n"+crud("client_template", "/client-templates")+
		"    x_terraform:\n      raw_feature: client_template\n")
	spec := specWith(t, "client_template", nameAttribute)

	_, err := generateRegistry(config, spec, "", "example.com/mod")

	require.ErrorContains(t, err, "has to be named client_template_raw")
}

func TestRegistryAcceptsACorrectlyGatedObject(t *testing.T) {
	config := write(t, "generator_config.yml", "resources:\n"+crud("gateway_raw", "/gateways")+
		"    x_terraform:\n      raw_feature: gateway\n")
	spec := specWith(t, "gateway_raw", nameAttribute+`, `+opaqueAttribute("device_config_raw"))

	source, err := generateRegistry(config, spec, "", "example.com/mod")

	require.NoError(t, err)
	require.Contains(t, string(source), `RawFeature:        "gateway",`)
	require.Contains(t, string(source), `return []string{"gateway"}`)
	require.Contains(t, string(source), `resGatewayRaw.GatewayRawResourceSchema,`)
}

func TestRegistryReportsWhatTheSchemaGeneratorSkipped(t *testing.T) {
	config := write(t, "generator_config.yml", "resources:\n"+crud("segment", "/segments")+crud("dropped", "/dropped"))
	spec := specWith(t, "segment", nameAttribute)

	_, err := generateRegistry(config, spec, "", "example.com/mod")

	require.ErrorContains(t, err, "the schema generator produced nothing for: resource dropped")
}

func TestRegistryRequiresTheOperationsTheRuntimeNeeds(t *testing.T) {
	config := write(t, "generator_config.yml", `resources:
  segment:
    create: {path: "/segments", method: POST}
    read: {path: "/segments/{id}", method: GET}
`)
	spec := specWith(t, "segment", nameAttribute)

	_, err := generateRegistry(config, spec, "", "example.com/mod")

	require.ErrorContains(t, err, "resource segment: a delete operation with a path and a method is required")
}

func TestRegistryOffersASearchOnlyWhereTheCollectionCanBeFiltered(t *testing.T) {
	config := write(t, "generator_config.yml", `data_sources:
  segment:
    read: {path: "/segments/{id}", method: GET}
  tag:
    read: {path: "/tags/{id}", method: GET}
  service_tenant:
    read: {path: "/tenants/{id}/service-tenant", method: GET}
`)

	spec := write(t, "spec.json", `{
		"provider": {"name": "bwan"},
		"datasources": [
			{"name": "segment", "schema": {"attributes": [`+nameAttribute+`]}},
			{"name": "tag", "schema": {"attributes": [`+nameAttribute+`]}},
			{"name": "service_tenant", "schema": {"attributes": [`+nameAttribute+`]}}
		]
	}`)

	// /segments can be filtered, /tags cannot, and a service tenant is not the
	// last segment of a collection at all.
	openapi := write(t, "openapi.yaml", `
paths:
  /segments:
    get:
      parameters:
        - {name: filter, in: query, schema: {type: string}}
  /tags:
    get:
      parameters:
        - {name: after, in: query, schema: {type: string}}
  /tenants/{id}:
    get:
      parameters:
        - {name: filter, in: query, schema: {type: string}}
`)

	source, err := generateRegistry(config, spec, openapi, "example.com/mod")

	require.NoError(t, err)
	require.Contains(t, string(source), `Search: genresource.Operation{Method: "GET", Path: "/segments"}`)
	require.NotContains(t, string(source), `Path: "/tags"}`)
	require.NotContains(t, string(source), `Path: "/tenants/{id}"}`)
}

func TestPascalMatchesTheFrameworkGenerator(t *testing.T) {
	for name, expected := range map[string]string{
		"segment":              "Segment",
		"ntp_config":           "NtpConfig",
		"ca_certificates":      "CaCertificates",
		"gateway_template_raw": "GatewayTemplateRaw",
		"vpnpeer":              "Vpnpeer",
	} {
		require.Equal(t, expected, pascal(name), name)
	}
}
