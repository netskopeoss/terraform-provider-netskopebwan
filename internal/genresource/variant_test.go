package genresource

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestVariantMatchesOnTheDiscriminator(t *testing.T) {
	wanlink := &Variant{Name: "wanlink", Discriminator: "type", Value: "wanlink", Match: []string{"frequency"}}

	require.True(t, wanlink.Matches(map[string]any{"type": "wanlink", "name": "probe"}))
	require.False(t, wanlink.Matches(map[string]any{"type": "overlay", "name": "probe"}))

	// The discriminator decides on its own: a field the kind happens to share is
	// not evidence, and a missing discriminator is not a match.
	require.False(t, wanlink.Matches(map[string]any{"frequency": 60}))
	require.False(t, wanlink.Matches(map[string]any{}))
}

// TestVariantMatchesOnANestedDiscriminator covers the tags the API serves from
// /overlay-tags, whose kind is written inside the object rather than on it.
func TestVariantMatchesOnANestedDiscriminator(t *testing.T) {
	wanlink := &Variant{Name: "wanlink", Discriminator: "config.type", Value: "wanlink", Match: []string{"wan_link_frequency"}}

	require.True(t, wanlink.Matches(map[string]any{
		"name":   "probe",
		"config": map[string]any{"type": "wanlink", "wan_link_frequency": 60},
	}))
	require.False(t, wanlink.Matches(map[string]any{
		"name":   "probe",
		"config": map[string]any{"type": "overlay"},
	}))

	// Nothing on the way to the discriminator can be assumed: an object without
	// the field, without the object holding it, or holding something else there,
	// is not this kind.
	require.False(t, wanlink.Matches(map[string]any{"config": map[string]any{}}))
	require.False(t, wanlink.Matches(map[string]any{"type": "wanlink"}))
	require.False(t, wanlink.Matches(map[string]any{"config": "wanlink"}))
	require.False(t, wanlink.Matches(map[string]any{}))

	detail := wanlink.Mismatch("bwan_tag_wanlink", map[string]any{"config": map[string]any{"type": "overlay"}})
	require.Contains(t, detail, `its config.type is "overlay"`)
	require.Contains(t, wanlink.Mismatch("bwan_tag_wanlink", map[string]any{}), `is "unset"`)
}

// TestVariantMatchesOnDistinctiveFieldsWithoutADiscriminator covers the forms the
// API declares with nothing to select them by, where the fields present are all
// there is to go on.
func TestVariantMatchesOnDistinctiveFieldsWithoutADiscriminator(t *testing.T) {
	typed := &Variant{Name: "model_name", Match: []string{"model", "name"}}

	require.True(t, typed.Matches(map[string]any{"model": "n300"}))
	require.True(t, typed.Matches(map[string]any{"name": "branch-1"}))
	require.False(t, typed.Matches(map[string]any{"device_config_raw": "{}"}))

	// A field present but null says the API did not send it.
	require.False(t, typed.Matches(map[string]any{"model": nil}))

	// Nothing to match on cannot match, which is safer than matching everything.
	require.False(t, (&Variant{Name: "empty"}).Matches(map[string]any{"anything": 1}))
}

func TestNilVariantMatchesEverything(t *testing.T) {
	var none *Variant

	require.True(t, none.Matches(map[string]any{"anything": 1}))
	require.True(t, none.Matches(nil))
}

func TestVariantMatchesNothingButAnObject(t *testing.T) {
	wanlink := &Variant{Name: "wanlink", Discriminator: "type", Value: "wanlink"}

	require.False(t, wanlink.Matches([]any{}))
	require.False(t, wanlink.Matches("wanlink"))
	require.False(t, wanlink.Matches(nil))
}

func TestVariantKeepNarrowsACollectionToItsOwnKind(t *testing.T) {
	overlay := &Variant{Name: "overlay", Discriminator: "type", Value: "overlay"}

	kept := overlay.Keep([]any{
		map[string]any{"id": "1", "type": "wanlink"},
		map[string]any{"id": "2", "type": "overlay"},
		map[string]any{"id": "3", "type": "overlay"},
		"not an object",
	})

	require.Equal(t, []any{
		map[string]any{"id": "2", "type": "overlay"},
		map[string]any{"id": "3", "type": "overlay"},
	}, kept)

	var none *Variant
	require.Len(t, none.Keep([]any{1, 2, 3}), 3, "an endpoint serving one kind filters nothing")
}

func TestVariantMismatchNamesWhatItFound(t *testing.T) {
	wanlink := &Variant{Name: "wanlink", Discriminator: "type", Value: "wanlink"}

	detail := wanlink.Mismatch("bwan_tag_wanlink", map[string]any{"type": "overlay"})
	require.Contains(t, detail, `its type is "overlay"`)
	require.Contains(t, detail, "bwan_tag_wanlink")
	require.Contains(t, detail, `Use the Terraform type for "overlay"`)

	require.Contains(t, wanlink.Mismatch("bwan_tag_wanlink", map[string]any{}), `is "unset"`)

	// With no discriminator there is no name to report, so the fields are.
	typed := &Variant{Name: "model_name", Match: []string{"model", "name"}}
	require.Contains(t, typed.Mismatch("bwan_gateway", map[string]any{}), "recognised by [model name]")
}

// tagDefinition mirrors one kind of tag: its own resource over an endpoint that
// serves four.
func tagDefinition() Definition {
	return Definition{
		Name: "tag_wanlink",
		Schema: func(_ context.Context) rschema.Schema {
			return rschema.Schema{
				Attributes: map[string]rschema.Attribute{
					"id":        rschema.StringAttribute{Computed: true},
					"name":      rschema.StringAttribute{Required: true},
					"type":      rschema.StringAttribute{Required: true},
					"frequency": rschema.Int64Attribute{Optional: true},
				},
			}
		},
		Create:  Operation{Method: http.MethodPost, Path: "/tags"},
		Read:    Operation{Method: http.MethodGet, Path: "/tags/{id}"},
		Update:  Operation{Method: http.MethodPatch, Path: "/tags/{id}"},
		Delete:  Operation{Method: http.MethodDelete, Path: "/tags/{id}"},
		Variant: &Variant{Name: "wanlink", Discriminator: "type", Value: "wanlink", Match: []string{"frequency"}},
	}
}

// TestResourceReadRefusesAnObjectOfAnotherKind covers giving one Terraform type the
// id of an object belonging to another: adopting it would have Terraform manage an
// object whose fields it cannot express.
func TestResourceReadRefusesAnObjectOfAnotherKind(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/tags/tag-1")).
		Return(json.RawMessage(`{"id": "tag-1", "name": "corp", "type": "overlay"}`), nil)

	res, schema := newResource(t, tagDefinition(), meta)

	state := objectValue(t, schema, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, "tag-1"),
		"name": tftypes.NewValue(tftypes.String, "corp"),
		"type": tftypes.NewValue(tftypes.String, "wanlink"),
	})

	resp := readResource(t, res, schema, state)

	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "Wrong kind of object", resp.Diagnostics.Errors()[0].Summary())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), `its type is "overlay"`)
	require.False(t, resp.State.Raw.IsNull(), "a wrong-kind object is an error, not a resource to recreate")
}

func TestResourceReadAcceptsItsOwnKind(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/tags/tag-1")).
		Return(json.RawMessage(`{"id": "tag-1", "name": "renamed", "type": "wanlink", "frequency": 60}`), nil)

	res, schema := newResource(t, tagDefinition(), meta)

	state := objectValue(t, schema, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, "tag-1"),
		"name": tftypes.NewValue(tftypes.String, "corp"),
		"type": tftypes.NewValue(tftypes.String, "wanlink"),
	})

	resp := readResource(t, res, schema, state)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, "renamed", attributeString(t, resp.State.Raw, "name"))
}

// TestListDataSourceDropsTheOtherKinds is what stops a data source standing for one
// kind from reporting every object the endpoint serves.
func TestListDataSourceDropsTheOtherKinds(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/tags")).
		Return(page("", false,
			object("1", "a", map[string]any{"type": "wanlink"}),
			object("2", "b", map[string]any{"type": "overlay"}),
			object("3", "c", map[string]any{"type": "wanlink"}),
		), nil)

	def := DataSourceDefinition{
		Name: "tags_wanlink",
		Schema: func(_ context.Context) dschema.Schema {
			return dschema.Schema{
				Attributes: map[string]dschema.Attribute{
					"after":  dschema.StringAttribute{Optional: true, Computed: true},
					"first":  dschema.Int64Attribute{Optional: true, Computed: true},
					"filter": dschema.StringAttribute{Optional: true, Computed: true},
					"sort":   dschema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
					"data": dschema.ListNestedAttribute{
						Computed: true,
						NestedObject: dschema.NestedAttributeObject{
							Attributes: map[string]dschema.Attribute{
								"id":   dschema.StringAttribute{Computed: true},
								"name": dschema.StringAttribute{Computed: true},
								"type": dschema.StringAttribute{Computed: true},
							},
						},
					},
					"page_info": dschema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]dschema.Attribute{
							"end_cursor":  dschema.StringAttribute{Computed: true},
							"has_next":    dschema.BoolAttribute{Computed: true},
							"total_count": dschema.Int64Attribute{Computed: true},
						},
					},
				},
			}
		},
		Read:    Operation{Method: http.MethodGet, Path: "/tags"},
		Variant: &Variant{Name: "wanlink", Discriminator: "type", Value: "wanlink"},
	}

	source, schema := newDataSource(t, def, meta)

	resp := readDataSource(t, source, schema, configFor(t, schema, nil))

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Len(t, dataElements(t, resp.State.Raw), 2, "only the wanlink tags belong to this data source")
}

// monitorDefinition mirrors a link monitor: one resource whose target takes one of
// three forms, each a block of its own.
func monitorDefinition() Definition {
	target := func(field string) rschema.SingleNestedAttribute {
		return rschema.SingleNestedAttribute{
			Optional:   true,
			Attributes: map[string]rschema.Attribute{field: rschema.StringAttribute{Required: true}},
		}
	}

	return Definition{
		Name: "link_monitor",
		Schema: func(_ context.Context) rschema.Schema {
			return rschema.Schema{
				Attributes: map[string]rschema.Attribute{
					"id":   rschema.StringAttribute{Computed: true},
					"fqdn": target("fqdn"),
					"ipv4": target("ipv4"),
					"ipv6": target("ipv6"),
				},
			}
		},
		Create:        Operation{Method: http.MethodPost, Path: "/link-monitors"},
		Read:          Operation{Method: http.MethodGet, Path: "/link-monitors/{id}"},
		Delete:        Operation{Method: http.MethodDelete, Path: "/link-monitors/{id}"},
		VariantBlocks: []string{"fqdn", "ipv4", "ipv6"},
	}
}

func validateResource(t *testing.T, res resource.Resource, schema rschema.Schema, config tftypes.Value) *resource.ValidateConfigResponse {
	t.Helper()

	validator, ok := res.(resource.ResourceWithValidateConfig)
	require.True(t, ok)

	resp := &resource.ValidateConfigResponse{}
	validator.ValidateConfig(context.Background(), resource.ValidateConfigRequest{
		Config: tfsdk.Config{Schema: schema, Raw: config},
	}, resp)

	return resp
}

func TestValidateConfigRequiresExactlyOneForm(t *testing.T) {
	_, meta := newAPI(t)

	res, schema := newResource(t, monitorDefinition(), meta)

	fqdnType, _ := schema.Type().TerraformType(context.Background()).(tftypes.Object).AttributeTypes["fqdn"].(tftypes.Object)
	fqdn := tftypes.NewValue(fqdnType, map[string]tftypes.Value{
		"fqdn": tftypes.NewValue(tftypes.String, "probe.example.com"),
	})
	ipv4Type, _ := schema.Type().TerraformType(context.Background()).(tftypes.Object).AttributeTypes["ipv4"].(tftypes.Object)
	ipv4 := tftypes.NewValue(ipv4Type, map[string]tftypes.Value{
		"ipv4": tftypes.NewValue(tftypes.String, "10.0.0.1"),
	})

	one := validateResource(t, res, schema, objectValue(t, schema, map[string]tftypes.Value{"fqdn": fqdn}))
	require.False(t, one.Diagnostics.HasError(), "%v", one.Diagnostics)

	none := validateResource(t, res, schema, objectValue(t, schema, nil))
	require.True(t, none.Diagnostics.HasError())
	require.Contains(t, none.Diagnostics.Errors()[0].Summary(), "Missing one of fqdn, ipv4, ipv6")

	both := validateResource(t, res, schema, objectValue(t, schema, map[string]tftypes.Value{"fqdn": fqdn, "ipv4": ipv4}))
	require.True(t, both.Diagnostics.HasError())
	require.Contains(t, both.Diagnostics.Errors()[0].Summary(), "Too many of fqdn, ipv4, ipv6")
	require.Contains(t, both.Diagnostics.Errors()[0].Detail(), "These are set: fqdn, ipv4")
}

// TestFormReachesTheAPIFlat is the end-to-end shape check: the practitioner writes
// a block, the API is sent what it declared.
func TestFormReachesTheAPIFlat(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodPost, "/link-monitors").withBody(`{"ipv4": "10.0.0.1"}`)).
		Return(json.RawMessage(`{"id": "monitor-1", "ipv4": "10.0.0.1"}`), nil)

	res, schema := newResource(t, monitorDefinition(), meta)

	ipv4Type, _ := schema.Type().TerraformType(context.Background()).(tftypes.Object).AttributeTypes["ipv4"].(tftypes.Object)

	plan := objectValue(t, schema, map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"ipv4": tftypes.NewValue(ipv4Type, map[string]tftypes.Value{
			"ipv4": tftypes.NewValue(tftypes.String, "10.0.0.1"),
		}),
	})

	resp := createResource(t, res, schema, plan)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, "monitor-1", attributeString(t, resp.State.Raw, "id"))
}
