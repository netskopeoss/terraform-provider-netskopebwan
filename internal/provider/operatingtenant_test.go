package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"

	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/genresource"
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

// TestOperatingTenantIsInvisibleByDefault is the property the whole feature turns
// on. The attribute is undocumented not by leaving it out of the documentation —
// which is generated, and would put it back — but by leaving it out of the
// schema, which is what the documentation, the language server and the registry
// are all built from.
func TestOperatingTenantIsInvisibleByDefault(t *testing.T) {
	schema := providerSchema(t)

	require.NotEmpty(t, schema.ResourceSchemas)
	require.NotEmpty(t, schema.DataSourceSchemas)

	for name, resourceSchema := range schema.ResourceSchemas {
		require.NotContains(t, attributeNames(resourceSchema), genresource.OperatingTenantAttribute, name)
	}

	for name, dataSourceSchema := range schema.DataSourceSchemas {
		require.NotContains(t, attributeNames(dataSourceSchema), genresource.OperatingTenantAttribute, name)
	}

	require.NotContains(t, attributeNames(schema.Provider), genresource.OperatingTenantAttribute,
		"the escape hatch is not a provider argument either")
}

// TestOperatingTenantAppearsOnEveryObjectWhenEnabled also counts the attributes,
// because the escape hatch is written into a generated schema: an object that
// already had an attribute of that name would have it silently replaced rather
// than gaining one, and the count is what notices.
func TestOperatingTenantAppearsOnEveryObjectWhenEnabled(t *testing.T) {
	before := providerSchema(t)

	t.Setenv(genresource.OperatingTenantEnv, "1")

	after := providerSchema(t)

	for name, resourceSchema := range after.ResourceSchemas {
		require.Contains(t, attributeNames(resourceSchema), genresource.OperatingTenantAttribute, name)
		require.Len(t, attributeNames(resourceSchema), len(attributeNames(before.ResourceSchemas[name]))+1, name)
	}

	for name, dataSourceSchema := range after.DataSourceSchemas {
		require.Contains(t, attributeNames(dataSourceSchema), genresource.OperatingTenantAttribute, name)
		require.Len(t, attributeNames(dataSourceSchema), len(attributeNames(before.DataSourceSchemas[name]))+1, name)
	}
}

// TestOperatingTenantClientsAreOnlyBuiltWhenEnabled covers the other half of the
// gate: with it off the provider has no way to reach another tenant at all, so
// state carried over from a run that had it on cannot be applied by accident.
func TestOperatingTenantClientsAreOnlyBuiltWhenEnabled(t *testing.T) {
	off := configure(t, "1.0.0", "")
	require.False(t, off.Diagnostics.HasError(), "%v", off.Diagnostics)

	meta, ok := off.ResourceData.(*genresource.Meta)
	require.True(t, ok)
	require.Nil(t, meta.Tenant)

	t.Setenv(genresource.OperatingTenantEnv, "1")

	on := configure(t, "1.0.0", "")
	require.False(t, on.Diagnostics.HasError(), "%v", on.Diagnostics)

	meta, ok = on.ResourceData.(*genresource.Meta)
	require.True(t, ok)
	require.NotNil(t, meta.Tenant)
}
