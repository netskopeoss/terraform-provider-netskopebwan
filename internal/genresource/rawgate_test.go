package genresource

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// gatedDefinition mirrors a gateway: its configuration is an opaque document, so
// it is only usable once the practitioner has opted in.
func gatedDefinition() Definition {
	def := thingDefinition()
	def.Name = "thing_raw"
	def.RawFeature = "thing"

	return def
}

// TestGatedResourceRefusesEveryOperationWithoutTheOptIn expects no API calls at
// all, which the mock enforces: a gate that reported an error but still reached
// the API would fail here rather than pass quietly.
func TestGatedResourceRefusesEveryOperationWithoutTheOptIn(t *testing.T) {
	_, meta := newAPI(t)

	res, schema := newResource(t, gatedDefinition(), meta)

	state := objectValue(t, schema, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, "thing-1"),
		"name": tftypes.NewValue(tftypes.String, "first"),
	})

	requireGateError(t, createResource(t, res, schema, state).Diagnostics.Errors())
	requireGateError(t, readResource(t, res, schema, state).Diagnostics.Errors())
	requireGateError(t, updateResource(t, res, schema, state, state).Diagnostics.Errors())

	deleteResp := &resource.DeleteResponse{State: tfsdk.State{Schema: schema, Raw: state}}
	res.Delete(context.Background(), resource.DeleteRequest{State: tfsdk.State{Schema: schema, Raw: state}}, deleteResp)
	requireGateError(t, deleteResp.Diagnostics.Errors())
}

// TestGatedResourceRefusesAtPlanTime covers the case that matters most: a
// practitioner should be told before anything is applied.
func TestGatedResourceRefusesAtPlanTime(t *testing.T) {
	_, meta := newAPI(t)

	res, schema := newResource(t, gatedDefinition(), meta)

	planModifier, ok := res.(resource.ResourceWithModifyPlan)
	require.True(t, ok)

	plan := objectValue(t, schema, map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, "first"),
	})

	resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: schema, Raw: plan}}
	planModifier.ModifyPlan(context.Background(), resource.ModifyPlanRequest{
		State: tfsdk.State{Schema: schema, Raw: nullValue(t, schema)},
		Plan:  tfsdk.Plan{Schema: schema, Raw: plan},
	}, resp)

	errs := resp.Diagnostics.Errors()
	requireGateError(t, errs)

	require.Equal(t, "Resource bwan_thing_raw is not enabled", errs[0].Summary())
	require.Contains(t, errs[0].Detail(), "enable_raw_thing = true")
	require.Contains(t, errs[0].Detail(), "NOT covered by the provider's backward-compatibility")
	require.Contains(t, errs[0].Detail(), "WILL be removed once a typed replacement ships")
}

func TestGatedResourceWorksOnceEnabled(t *testing.T) {
	api, meta := newAPI(t)

	meta.RawEnabled["thing"] = true

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodPost, "/things")).
		Return(json.RawMessage(`{"id": "thing-1", "name": "first"}`), nil)

	res, schema := newResource(t, gatedDefinition(), meta)

	plan := objectValue(t, schema, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name": tftypes.NewValue(tftypes.String, "first"),
	})

	resp := createResource(t, res, schema, plan)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, "thing-1", attributeString(t, resp.State.Raw, "id"))
}

func TestGatedDataSourceRefusesWithoutTheOptIn(t *testing.T) {
	_, meta := newAPI(t)

	def := DataSourceDefinition{
		Name: "thing_raw",
		Schema: func(_ context.Context) dschema.Schema {
			return dschema.Schema{Attributes: map[string]dschema.Attribute{
				"id": dschema.StringAttribute{Required: true},
			}}
		},
		Read:       Operation{Method: http.MethodGet, Path: "/things/{id}"},
		RawFeature: "thing",
	}

	source, schema := newDataSource(t, def, meta)

	config := configFor(t, schema, map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, "thing-1"),
	})

	resp := readDataSource(t, source, schema, config)

	errs := resp.Diagnostics.Errors()
	requireGateError(t, errs)
	require.Equal(t, "Data source bwan_thing_raw is not enabled", errs[0].Summary())
}

func TestUngatedObjectsNeedNoOptIn(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodPost, "/things")).
		Return(json.RawMessage(`{"id": "thing-1", "name": "first"}`), nil)

	res, schema := newResource(t, thingDefinition(), meta)

	plan := objectValue(t, schema, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name": tftypes.NewValue(tftypes.String, "first"),
	})

	resp := createResource(t, res, schema, plan)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
}

func TestRawOptInArgumentName(t *testing.T) {
	require.Equal(t, "enable_raw_gateway_template", RawOptInArgument("gateway_template"))
}

func TestRawOptInDescriptionSpellsOutTheConsequences(t *testing.T) {
	description := RawOptInDescription("gateway_template")

	require.Contains(t, description, "gateway template")
	require.Contains(t, description, "NOT covered by the provider's backward-compatibility guarantees")
	require.Contains(t, description, "WILL change without a major release")
	require.Contains(t, description, "WILL be removed once typed replacements ship")
}

func requireGateError(t *testing.T, errs diag.Diagnostics) {
	t.Helper()

	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Summary(), "is not enabled")
}
