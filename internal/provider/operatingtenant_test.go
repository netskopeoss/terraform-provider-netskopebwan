package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"

	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/genresource"
	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/registry"
)

// attributeNames lists what a schema offers over the plugin protocol, which is
// the only description of the provider anything downstream has: tfplugindocs
// reads it, the language server reads it, the registry reads it.
func attributeNames(schema *tfprotov6.Schema) []string {
	if schema == nil || schema.Block == nil {
		return nil
	}

	out := make([]string, 0, len(schema.Block.Attributes))

	for _, attribute := range schema.Block.Attributes {
		out = append(out, attribute.Name)
	}

	return out
}

func providerSchema(t *testing.T) *tfprotov6.GetProviderSchemaResponse {
	t.Helper()

	schema, err := providerserver.NewProtocol6(New("test")())().
		GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)

	// The framework validates every schema here and rejects the whole provider on
	// the first problem, so an attribute added to every object has to keep the
	// lot valid, not just its own.
	for _, diagnostic := range schema.Diagnostics {
		require.NotEqual(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity,
			"%s: %s", diagnostic.Summary, diagnostic.Detail)
	}

	return schema
}

// TestOperatingTenantIsOnEveryObject is the half of the feature the provider
// owns: the attribute has to be in the schema, because Terraform validates a
// configuration against the schema and an argument that is not in it cannot be
// written. Keeping it out of the documentation is tools/tfdocs' job.
func TestOperatingTenantIsOnEveryObject(t *testing.T) {
	schema := providerSchema(t)

	require.NotEmpty(t, schema.ResourceSchemas)
	require.NotEmpty(t, schema.DataSourceSchemas)

	for name, resourceSchema := range schema.ResourceSchemas {
		require.Contains(t, attributeNames(resourceSchema), genresource.OperatingTenantAttribute, name)
	}

	for name, dataSourceSchema := range schema.DataSourceSchemas {
		require.Contains(t, attributeNames(dataSourceSchema), genresource.OperatingTenantAttribute, name)
	}

	require.NotContains(t, attributeNames(schema.Provider), genresource.OperatingTenantAttribute,
		"the tenant is chosen per object, not once for the provider")
}

// TestOperatingTenantDoesNotDisplaceAnAttribute guards the way it is added: it is
// written into a generated schema, so an object that already had an attribute of
// that name would have it silently replaced rather than gaining one. Nothing in
// the spec is called this today, and this is what notices if that changes.
func TestOperatingTenantDoesNotDisplaceAnAttribute(t *testing.T) {
	ctx := context.Background()

	for _, definition := range registry.Resources() {
		require.NotContains(t, definition.Schema(ctx).Attributes, genresource.OperatingTenantAttribute, definition.Name)
	}

	for _, definition := range registry.DataSources() {
		require.NotContains(t, definition.Schema(ctx).Attributes, genresource.OperatingTenantAttribute, definition.Name)
	}
}

// TestOperatingTenantClientsAreConfigured covers the wiring: every configured
// provider can reach another tenant, because the attribute is always there to
// ask it to.
func TestOperatingTenantClientsAreConfigured(t *testing.T) {
	configured := configure(t, "1.0.0", "")
	require.False(t, configured.Diagnostics.HasError(), "%v", configured.Diagnostics)

	meta, ok := configured.ResourceData.(*genresource.Meta)
	require.True(t, ok)
	require.NotNil(t, meta.Tenant)
	require.Equal(t, meta, configured.DataSourceData)
}
