package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/genresource"
	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/provider"
	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/registry"
)

func thingDefinition(schema rschema.Schema) genresource.Definition {
	return genresource.Definition{
		Name:   "thing",
		Schema: func(context.Context) rschema.Schema { return schema },
		Create: genresource.Operation{Method: "POST", Path: "/things"},
		Read:   genresource.Operation{Method: "GET", Path: "/things/{id}"},
		Update: genresource.Operation{Method: "PATCH", Path: "/things/{id}"},
		Delete: genresource.Operation{Method: "DELETE", Path: "/things/{id}"},
	}
}

// TestResourceExampleSetsWhatTheObjectNeeds covers the whole of what a generated
// example is: the required arguments, at the values the schema allows, aligned the
// way terraform fmt would align them.
func TestResourceExampleSetsWhatTheObjectNeeds(t *testing.T) {
	ctx := context.Background()

	definition := thingDefinition(rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"name":     rschema.StringAttribute{Required: true},
			"kind":     rschema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("first", "second")}},
			"enabled":  rschema.BoolAttribute{Required: true},
			"servers":  rschema.ListAttribute{Required: true, ElementType: types.StringType},
			"optional": rschema.StringAttribute{Optional: true},
			"computed": rschema.StringAttribute{Computed: true},
			"nested": rschema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]rschema.Attribute{
					"port":  rschema.Int64Attribute{Required: true},
					"extra": rschema.StringAttribute{Optional: true},
				},
			},
		},
	})

	schema, err := resourceSchema(ctx, definition)
	require.NoError(t, err)

	example := resourceExample(ctx, "netskopebwan_thing", definition, schema)

	require.Equal(t, `resource "netskopebwan_thing" "example" {
  enabled = false
  kind    = "first"
  name    = "example"
  nested = {
    port = 1
  }
  servers = ["<server>"]
}
`, example)

	// An optional argument is a guess, and a computed one cannot be set at all.
	require.NotContains(t, example, "optional")
	require.NotContains(t, example, "computed")
}

// TestResourceExampleSaysWhatTheObjectCosts covers the two things about a resource
// a practitioner wants to know before writing any of it.
func TestResourceExampleSaysWhatTheObjectCosts(t *testing.T) {
	ctx := context.Background()

	schema := rschema.Schema{Attributes: map[string]rschema.Attribute{
		"config_raw": rschema.StringAttribute{Required: true},
	}}

	definition := thingDefinition(schema)
	definition.RawFeature = "thing"
	definition.RawJSONAttributes = []string{"config_raw"}
	definition.Update = genresource.Operation{}

	prepared, err := resourceSchema(ctx, definition)
	require.NoError(t, err)

	example := resourceExample(ctx, "netskopebwan_thing_raw", definition, prepared)

	require.Contains(t, example, "off until enable_raw_thing is set")
	require.Contains(t, example, "cannot update one of these, so any change replaces it")

	// An opaque document is written with jsonencode rather than an escaped string.
	require.Contains(t, example, "config_raw = jsonencode({})")
}

func TestImportExampleTakesTheParentAndTheID(t *testing.T) {
	nested := genresource.Definition{
		Name:   "nested",
		Create: genresource.Operation{Method: "POST", Path: "/parents/{parent_id}/nested"},
		Read:   genresource.Operation{Method: "GET", Path: "/parents/{parent_id}/nested/{id}"},
	}

	example := importExample("netskopebwan_nested", nested)

	require.Contains(t, example, "parent_id, id")
	require.Contains(t, example, "terraform import netskopebwan_nested.example <parent_id>/<id>")

	// Nothing to explain where the id is the whole of it.
	plain := importExample("netskopebwan_thing", thingDefinition(rschema.Schema{}))
	require.Equal(t, "terraform import netskopebwan_thing.example <id>\n", plain)
}

// TestDataSourceExampleShowsEveryWayToAddressTheObject covers the three shapes a
// data source comes in: one object by id or filter, a collection, and an object
// that hangs off a parent and cannot be found without it.
func TestDataSourceExampleShowsEveryWayToAddressTheObject(t *testing.T) {
	ctx := context.Background()

	single := dataSourceExample(ctx, "netskopebwan_thing", genresource.DataSourceDefinition{}, map[string]exampleAttribute{
		"id":     {name: "id", attribute: dschema.StringAttribute{Optional: true}},
		"filter": {name: "filter", attribute: dschema.StringAttribute{Optional: true}},
	})

	require.Contains(t, single, `data "netskopebwan_thing" "by_id" {`)
	require.Contains(t, single, `id = "`+idExample+`"`)
	require.Contains(t, single, `data "netskopebwan_thing" "by_filter" {`)
	require.Contains(t, single, `filter = "name eq \"example\""`)

	collection := dataSourceExample(ctx, "netskopebwan_things", genresource.DataSourceDefinition{}, map[string]exampleAttribute{
		"filter": {name: "filter", attribute: dschema.StringAttribute{Optional: true}},
	})

	require.Contains(t, collection, "Every page is walked")
	require.NotContains(t, collection, "by_id")

	child := dataSourceExample(ctx, "netskopebwan_nested", genresource.DataSourceDefinition{}, map[string]exampleAttribute{
		"parent_id": {name: "parent_id", attribute: dschema.StringAttribute{Required: true}},
		"filter":    {name: "filter", attribute: dschema.StringAttribute{Optional: true}},
	})

	require.Contains(t, child, `parent_id = "`+idExample+`"`)
	require.NotContains(t, child, "by_filter", "there is only one way to address an object under a parent")
}

// TestVariantBlocksGetOneFormSet covers the objects whose configuration is a
// choice between shapes: leaving all of them out is the one thing the schema
// rejects, so an example cannot simply set what is required.
func TestVariantBlocksGetOneFormSet(t *testing.T) {
	ctx := context.Background()

	attributes := map[string]exampleAttribute{
		"config": {name: "config", attribute: rschema.SingleNestedAttribute{
			Required: true,
			Attributes: map[string]rschema.Attribute{
				"aws":   rschema.SingleNestedAttribute{Optional: true, Attributes: map[string]rschema.Attribute{"key_id": rschema.StringAttribute{Optional: true}}},
				"azure": rschema.SingleNestedAttribute{Optional: true, Attributes: map[string]rschema.Attribute{"tenant": rschema.StringAttribute{Optional: true}}},
			},
		}},
	}

	rendered := arguments(ctx, attributes, "", nil, []string{"config.aws", "config.azure"})

	require.Contains(t, rendered, "aws = {")
	require.NotContains(t, rendered, "azure", "exactly one form is set, not both")
}

// TestEveryObjectHasAnExample is what keeps the committed examples in step with
// the provider: a new object arrives with a page of its own, and a page with no
// example is what this generator exists to prevent.
func TestEveryObjectHasAnExample(t *testing.T) {
	for _, definition := range registry.Resources() {
		dir := filepath.Join("..", "..", "examples", "resources", provider.TypeName+"_"+definition.Name)

		for _, base := range []string{"resource.tf", "import.sh"} {
			_, err := os.Stat(filepath.Join(dir, base))
			require.NoError(t, err, "%s has no %s; run make examples", definition.Name, base)
		}
	}

	for _, definition := range registry.DataSources() {
		path := filepath.Join("..", "..", "examples", "data-sources", provider.TypeName+"_"+definition.Name, "data-source.tf")

		_, err := os.Stat(path)
		require.NoError(t, err, "%s has no data-source.tf; run make examples", definition.Name)
	}
}
