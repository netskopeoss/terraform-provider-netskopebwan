package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"

	"infiot.com/infiot/mgmt/tf-provider/internal/genresource"
	"infiot.com/infiot/mgmt/tf-provider/internal/registry"
	"infiot.com/infiot/mgmt/tf-provider/internal/tfschema"
)

func newProvider(t *testing.T) provider.Provider {
	t.Helper()

	return New("test")()
}

// TestProviderSchemaPassesTheFrameworksOwnValidation is the check that matters
// most, because it is the one Terraform itself runs before anything else.
//
// Asking a resource for its schema directly does not validate it. The framework
// only does that when the whole provider is asked over the plugin protocol, and it
// rejects the lot on the first problem: a single attribute named after something
// Terraform reserves — `provider`, `count`, `lifecycle` — makes every resource in
// the provider unusable, not just its own.
func TestProviderSchemaPassesTheFrameworkOwnValidation(t *testing.T) {
	ctx := context.Background()

	server := providerserver.NewProtocol6(New("test")())()

	schema, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)

	for _, diagnostic := range schema.Diagnostics {
		require.NotEqual(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity,
			"%s: %s", diagnostic.Summary, diagnostic.Detail)
	}

	require.Len(t, schema.ResourceSchemas, len(registry.Resources()))
	require.Len(t, schema.DataSourceSchemas, len(registry.DataSources()))
}

// TestEveryResourceSchemaIsUsable walks every generated resource: its schema has
// to build, name itself uniquely, carry the id the runtime tracks it by, and have
// a shape the value mapping supports. A spec change that breaks any of those
// fails here rather than in a practitioner's plan.
func TestEveryResourceSchemaIsUsable(t *testing.T) {
	ctx := context.Background()
	prov := newProvider(t)

	factories := prov.Resources(ctx)
	require.NotEmpty(t, factories)
	require.Len(t, factories, len(registry.Resources()))

	seen := map[string]bool{}

	for _, factory := range factories {
		res := factory()

		metadata := &resource.MetadataResponse{}
		res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: TypeName}, metadata)

		require.True(t, strings.HasPrefix(metadata.TypeName, TypeName+"_"), metadata.TypeName)
		require.False(t, seen[metadata.TypeName], "duplicate resource type %s", metadata.TypeName)
		seen[metadata.TypeName] = true

		schemaResp := &resource.SchemaResponse{}
		res.Schema(ctx, resource.SchemaRequest{}, schemaResp)
		require.False(t, schemaResp.Diagnostics.HasError(), "%s: %v", metadata.TypeName, schemaResp.Diagnostics)

		id, ok := schemaResp.Schema.Attributes["id"]
		require.True(t, ok, "%s has no id attribute to track it by", metadata.TypeName)
		require.True(t, id.IsComputed(), "%s: id is assigned by the API", metadata.TypeName)

		model := tfschema.FromResource(ctx, schemaResp.Schema, nil, nil, nil)
		require.NoError(t, model.Validate(), metadata.TypeName)
	}
}

func TestEveryDataSourceSchemaIsUsable(t *testing.T) {
	ctx := context.Background()
	prov := newProvider(t)

	factories := prov.DataSources(ctx)
	require.NotEmpty(t, factories)
	require.Len(t, factories, len(registry.DataSources()))

	seen := map[string]bool{}

	for _, factory := range factories {
		source := factory()

		metadata := &datasource.MetadataResponse{}
		source.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: TypeName}, metadata)

		require.True(t, strings.HasPrefix(metadata.TypeName, TypeName+"_"), metadata.TypeName)
		require.False(t, seen[metadata.TypeName], "duplicate data source type %s", metadata.TypeName)
		seen[metadata.TypeName] = true

		schemaResp := &datasource.SchemaResponse{}
		source.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
		require.False(t, schemaResp.Diagnostics.HasError(), "%s: %v", metadata.TypeName, schemaResp.Diagnostics)

		model := tfschema.FromDataSource(ctx, schemaResp.Schema, nil, nil, nil)
		require.NoError(t, model.Validate(), metadata.TypeName)
	}
}

// TestSingularDataSourcesCanBeFound checks that every data source addressing one
// object can also look it up by filter, since every BWAN collection is
// filterable.
func TestSingularDataSourcesCanBeFound(t *testing.T) {
	ctx := context.Background()

	searchable := 0

	for _, definition := range registry.DataSources() {
		if !strings.Contains(definition.Read.Path, "{id}") {
			continue
		}

		schema := definition.Schema(ctx)

		if definition.Search.Path == "" {
			// A service tenant hangs off a tenant rather than sitting in a
			// collection of its own, so there is nothing to filter.
			require.True(t, schema.Attributes["id"].IsRequired(), definition.Name)

			continue
		}

		searchable++

		source := genresource.NewDataSource(definition)()

		schemaResp := &datasource.SchemaResponse{}
		source.Schema(ctx, datasource.SchemaRequest{}, schemaResp)

		id, ok := schemaResp.Schema.Attributes["id"]
		require.True(t, ok, definition.Name)
		require.True(t, id.IsOptional(), "%s: an object can be found by filter instead of by id", definition.Name)

		filter, ok := schemaResp.Schema.Attributes["filter"]
		require.True(t, ok, "%s has a filterable collection but offers no filter", definition.Name)
		require.True(t, filter.IsOptional(), definition.Name)
	}

	require.NotEmpty(t, searchable)
}

// TestEveryRawObjectIsNamedAndGated re-checks at runtime what the registry
// generator enforces at build time, so the two cannot disagree.
func TestEveryRawObjectIsNamedAndGated(t *testing.T) {
	features := map[string]bool{}

	for _, feature := range registry.RawFeatures() {
		features[feature] = true
	}

	require.NotEmpty(t, features, "the API still exposes objects configured through an opaque document")

	for _, definition := range registry.Resources() {
		if definition.RawFeature == "" {
			continue
		}

		require.True(t, features[definition.RawFeature], "resource %s is gated behind an unadvertised feature", definition.Name)
		require.True(t, strings.HasSuffix(definition.Name, "_raw"),
			"resource %s has to keep the plain name free for a typed replacement", definition.Name)
	}

	for _, definition := range registry.DataSources() {
		if definition.RawFeature == "" {
			continue
		}

		require.True(t, features[definition.RawFeature], "data source %s is gated behind an unadvertised feature", definition.Name)
		require.True(t, strings.HasSuffix(definition.Name, "_raw"),
			"data source %s has to keep the plain name free for a typed replacement", definition.Name)
	}
}

func TestProviderSchemaOffersAnOptInPerRawFeature(t *testing.T) {
	ctx := context.Background()

	resp := &provider.SchemaResponse{}
	newProvider(t).Schema(ctx, provider.SchemaRequest{}, resp)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	for _, feature := range registry.RawFeatures() {
		argument := genresource.RawOptInArgument(feature)

		attribute, ok := resp.Schema.Attributes[argument]
		require.True(t, ok, "the provider offers no %s argument", argument)
		require.True(t, attribute.IsOptional())
		require.Contains(t, attribute.GetDescription(), "NOT covered by the provider's backward-compatibility guarantees")
	}

	for _, argument := range []string{"endpoint", "token", "request_timeout", "insecure_skip_verify"} {
		require.Contains(t, resp.Schema.Attributes, argument)
	}

	token, ok := resp.Schema.Attributes["token"]
	require.True(t, ok)
	require.True(t, token.IsSensitive())
}

// TestGatewayAndPolicyAreExposedAsRawObjects pins the objects this provider is
// expected to manage through an opaque document, and which of them a typed
// resource has since taken the plain name for.
func TestGatewayAndPolicyAreExposedAsRawObjects(t *testing.T) {
	names := map[string]string{}

	for _, definition := range registry.Resources() {
		names[definition.Name] = definition.RawFeature
	}

	require.Equal(t, "gateway", names["gateway_raw"])
	require.Equal(t, "gateway_template", names["gateway_template_raw"])
	require.Equal(t, "policy", names["policy_raw"])
	require.Equal(t, "client_template", names["client_template_raw"])

	// A gateway's create body has a typed form as well as the opaque one, so the
	// plain name is taken and needs no opt-in.
	require.Contains(t, names, "gateway")
	require.Empty(t, names["gateway"])

	// The rest have no typed form yet, so their plain names stay free.
	require.NotContains(t, names, "policy")
	require.NotContains(t, names, "gateway_template")
	require.NotContains(t, names, "client_template")
}

// TestTagKindsAreSeparateResources covers the endpoint that serves four kinds of
// object: each kind is its own resource, pinned to the variant that recognises it.
func TestTagKindsAreSeparateResources(t *testing.T) {
	variants := map[string]string{}

	for _, definition := range registry.Resources() {
		if definition.Variant != nil {
			variants[definition.Name] = definition.Variant.Value
		}
	}

	require.Equal(t, map[string]string{
		"tag_wanlink":  "wanlink",
		"tag_overlay":  "overlay",
		"tag_topology": "topology",
		"tag_gateway":  "gateway",
	}, variants)

	// The umbrella name is gone: there is no one resource that manages any tag.
	for _, definition := range registry.Resources() {
		require.NotEqual(t, "tag", definition.Name)
	}
}
