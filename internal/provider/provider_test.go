package provider

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"

	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/genresource"
	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/registry"
	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/tfschema"
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
			// An object hanging off another rather than sitting in a collection of
			// its own has nothing to filter, so it can only be addressed by id.
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

// TestProviderWarnsThatItIsAlpha covers the two places a practitioner can meet
// this provider: the registry's home page, which tfplugindocs renders from the
// provider schema, and the README. Both have to carry the same warning, so the
// warning has one wording and the README quotes it.
func TestProviderWarnsThatItIsAlpha(t *testing.T) {
	ctx := context.Background()

	resp := &provider.SchemaResponse{}
	newProvider(t).Schema(ctx, provider.SchemaRequest{}, resp)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	require.Contains(t, resp.Schema.Description, AlphaHeadline)
	require.Contains(t, resp.Schema.Description, AlphaDetail)

	// docs/index.md renders the markdown description, where the warning is a
	// registry callout rather than another paragraph of prose.
	require.Contains(t, resp.Schema.MarkdownDescription, "~> **"+AlphaHeadline+"**")
	require.Contains(t, resp.Schema.MarkdownDescription, AlphaDetail)

	// The notice says what acknowledging a prerelease accepts, not only that one
	// has to be acknowledged.
	require.Contains(t, AlphaDetail, "Setting it accepts that instability")
	require.Contains(t, AlphaDetail, "NOT covered by the provider's backward-compatibility guarantees")

	readme, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	require.NoError(t, err)

	// The README wraps the warning across lines, quotes it as a callout and marks
	// up the arguments it names, so it is the words that have to match rather than
	// the layout. Only the quoting is stripped, not every ">": the warning names a
	// "~>" constraint.
	lines := strings.Split(string(readme), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimPrefix(strings.TrimSpace(line), "> ")
	}

	prose := strings.Join(strings.Fields(strings.ReplaceAll(strings.Join(lines, " "), "`", "")), " ")

	require.Contains(t, prose, AlphaHeadline)
	require.Contains(t, prose, AlphaDetail)
}

// configure runs Configure against a provider stamped with version, with the
// prerelease acknowledgement set to acknowledged (empty meaning unset).
func configure(t *testing.T, version, acknowledged string) *provider.ConfigureResponse {
	t.Helper()

	ctx := context.Background()
	prov := New(version)()

	schemaResp := &provider.SchemaResponse{}
	prov.Schema(ctx, provider.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), "%v", schemaResp.Diagnostics)

	object, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok)

	members := map[string]tftypes.Value{}

	for name, attributeType := range object.AttributeTypes {
		members[name] = tftypes.NewValue(attributeType, nil)
	}

	if acknowledged != "" {
		members[PreReleaseArgument] = tftypes.NewValue(tftypes.String, acknowledged)
	}

	// The endpoint and token come from the environment, so a configuration that
	// gets past the gate reaches the checks after it rather than stopping on
	// arguments this test is not about.
	t.Setenv("BWAN_ENDPOINT", "https://tenant.example.com")
	t.Setenv("BWAN_TOKEN", "token")

	resp := &provider.ConfigureResponse{}
	prov.Configure(ctx, provider.ConfigureRequest{
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: tftypes.NewValue(object, members)},
	}, resp)

	return resp
}

// TestPreReleaseHasToBeAcknowledged covers the gate on a prerelease build: it
// refuses to configure until the configuration names the version it is running,
// which is the version the practitioner is told about in the error rather than in
// the generated documentation, where no tag exists yet to name.
func TestPreReleaseHasToBeAcknowledged(t *testing.T) {
	unacknowledged := configure(t, "1.0.0-alpha.1", "")

	require.True(t, unacknowledged.Diagnostics.HasError())
	require.Equal(t, "This provider is a prerelease", unacknowledged.Diagnostics.Errors()[0].Summary())

	detail := unacknowledged.Diagnostics.Errors()[0].Detail()
	require.Contains(t, detail, "Version 1.0.0-alpha.1 is a prerelease")
	require.Contains(t, detail, PreReleaseArgument+` = "1.0.0-alpha.1"`)
	require.Contains(t, detail, `version = "~> 0.0"`, "the released line is the way out")
	require.Nil(t, unacknowledged.ResourceData, "nothing is configured behind the gate")

	// The gate is an acceptance, not a checkbox: what setting it costs has to be
	// said where it is asked for, the same way the raw opt-ins say it.
	require.Contains(t, detail, "accepts the instability")
	require.Contains(t, detail, "not covered by the provider's backward-compatibility guarantees")

	schemaResp := &provider.SchemaResponse{}
	newProvider(t).Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

	argument, ok := schemaResp.Schema.Attributes[PreReleaseArgument]
	require.True(t, ok)
	require.Contains(t, argument.GetDescription(), "explicit acceptance of the instability")
	require.Contains(t, argument.GetDescription(), "NOT covered by the provider's backward-compatibility guarantees")

	// The right version opens the gate, and only that version.
	acknowledged := configure(t, "1.0.0-alpha.1", "1.0.0-alpha.1")
	require.False(t, acknowledged.Diagnostics.HasError(), "%v", acknowledged.Diagnostics)
	require.NotNil(t, acknowledged.ResourceData)

	// A v prefix and surrounding space are how the version is written elsewhere,
	// so they are not worth an error.
	require.False(t, configure(t, "1.0.0-alpha.1", " v1.0.0-alpha.1 ").Diagnostics.HasError())

	stale := configure(t, "1.0.0-alpha.2", "1.0.0-alpha.1")
	require.True(t, stale.Diagnostics.HasError())
	require.Equal(t, "A different prerelease is acknowledged", stale.Diagnostics.Errors()[0].Summary())
	require.Contains(t, stale.Diagnostics.Errors()[0].Detail(), "the provider being run is 1.0.0-alpha.2")
}

// TestReleasedVersionsAreNotGated keeps the gate off everything that is not a
// prerelease: a released version, and a build from a working tree, which stamps
// no version at all and is what the acceptance tests and dev overrides run.
func TestReleasedVersionsAreNotGated(t *testing.T) {
	released := configure(t, "1.0.0", "")
	require.False(t, released.Diagnostics.HasError(), "%v", released.Diagnostics)
	require.Empty(t, released.Diagnostics.Warnings())

	development := configure(t, "dev", "")
	require.False(t, development.Diagnostics.HasError(), "%v", development.Diagnostics)
	require.Empty(t, development.Diagnostics.Warnings())

	// An acknowledgement left behind after upgrading to a release is not an
	// error, but it is not doing anything either.
	upgraded := configure(t, "1.0.0", "1.0.0-alpha.1")
	require.False(t, upgraded.Diagnostics.HasError(), "%v", upgraded.Diagnostics)
	require.Len(t, upgraded.Diagnostics.Warnings(), 1)
	require.Contains(t, upgraded.Diagnostics.Warnings()[0].Detail(), "does nothing here and can be removed")

	// A development build says nothing about an argument it does not enforce.
	require.Empty(t, configure(t, "dev", "1.0.0-alpha.1").Diagnostics.Warnings())
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

	// A gateway's create body has a typed form as well as the opaque one, so the
	// plain name is taken and needs no opt-in.
	require.Contains(t, names, "gateway")
	require.Empty(t, names["gateway"])

	// client_template took its plain name once its on-premise detection probes
	// were exposed in typed form, so it needs no opt-in either.
	require.Contains(t, names, "client_template")
	require.Empty(t, names["client_template"])

	// The rest have no typed form yet, so their plain names stay free.
	require.NotContains(t, names, "policy")
	require.NotContains(t, names, "gateway_template")
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

	// The API writes the kind inside the tag, so every one of them has to be
	// recognised there rather than on the tag itself.
	for _, definition := range registry.Resources() {
		if definition.Variant != nil && strings.HasPrefix(definition.Name, "tag_") {
			require.Equal(t, "config.type", definition.Variant.Discriminator, definition.Name)

			// And nowhere in the schema, because the resource name has already said
			// it: the config the API wraps a tag's fields in is hoisted away, and the
			// runtime is what puts it back.
			require.Equal(t, "config", definition.Variant.Wrapper, definition.Name)

			attributes := definition.Schema(context.Background()).Attributes
			require.NotContains(t, attributes, "config", definition.Name)
			require.NotContains(t, attributes, "type", definition.Name)
		}
	}

	// The umbrella name is gone: there is no one resource that manages any tag.
	for _, definition := range registry.Resources() {
		require.NotEqual(t, "tag", definition.Name)
	}
}
