package genresource

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// thingSchema stands in for a generated resource schema: an API-assigned id, a
// required name, an optional-and-computed description and an embedded JSON
// document.
func thingSchema(_ context.Context) rschema.Schema {
	return rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"id":                rschema.StringAttribute{Computed: true},
			"name":              rschema.StringAttribute{Required: true},
			"description":       rschema.StringAttribute{Optional: true, Computed: true},
			"device_config_raw": rschema.StringAttribute{Optional: true},
		},
	}
}

func thingDefinition() Definition {
	return Definition{
		Name:              "thing",
		Schema:            thingSchema,
		Create:            Operation{Method: http.MethodPost, Path: "/things"},
		Read:              Operation{Method: http.MethodGet, Path: "/things/{id}"},
		Update:            Operation{Method: http.MethodPatch, Path: "/things/{id}"},
		Delete:            Operation{Method: http.MethodDelete, Path: "/things/{id}"},
		RawJSONAttributes: []string{"device_config_raw"},
	}
}

func newResource(t *testing.T, def Definition, meta *Meta) (resource.Resource, rschema.Schema) {
	t.Helper()

	res := NewResource(def)()

	configure, ok := res.(resource.ResourceWithConfigure)
	require.True(t, ok)

	configureResp := &resource.ConfigureResponse{}
	configure.Configure(context.Background(), resource.ConfigureRequest{ProviderData: meta}, configureResp)
	require.False(t, configureResp.Diagnostics.HasError(), "%v", configureResp.Diagnostics)

	schemaResp := &resource.SchemaResponse{}
	res.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), "%v", schemaResp.Diagnostics)

	return res, schemaResp.Schema
}

func objectValue(t *testing.T, schema rschema.Schema, members map[string]tftypes.Value) tftypes.Value {
	t.Helper()

	return fill(t, schema.Type().TerraformType(context.Background()), members)
}

func nullValue(t *testing.T, schema rschema.Schema) tftypes.Value {
	t.Helper()

	return tftypes.NewValue(schema.Type().TerraformType(context.Background()), nil)
}

func attributeString(t *testing.T, value tftypes.Value, name string) string {
	t.Helper()

	var members map[string]tftypes.Value
	require.NoError(t, value.As(&members))

	member, ok := members[name]
	require.True(t, ok, "no attribute named %q", name)

	if member.IsNull() {
		return ""
	}

	var text string
	require.NoError(t, member.As(&text))

	return text
}

// createResource runs a create against plan and returns the response.
func createResource(t *testing.T, res resource.Resource, schema rschema.Schema, plan tftypes.Value) *resource.CreateResponse {
	t.Helper()

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schema, Raw: nullValue(t, schema)}}
	res.Create(context.Background(), resource.CreateRequest{Plan: tfsdk.Plan{Schema: schema, Raw: plan}}, resp)

	return resp
}

func readResource(t *testing.T, res resource.Resource, schema rschema.Schema, state tftypes.Value) *resource.ReadResponse {
	t.Helper()

	resp := &resource.ReadResponse{State: tfsdk.State{Schema: schema, Raw: state}}
	res.Read(context.Background(), resource.ReadRequest{State: tfsdk.State{Schema: schema, Raw: state}}, resp)

	return resp
}

func updateResource(t *testing.T, res resource.Resource, schema rschema.Schema, plan, state tftypes.Value) *resource.UpdateResponse {
	t.Helper()

	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schema, Raw: state}}
	res.Update(context.Background(), resource.UpdateRequest{
		Plan:  tfsdk.Plan{Schema: schema, Raw: plan},
		State: tfsdk.State{Schema: schema, Raw: state},
	}, resp)

	return resp
}

func TestResourceCreateSendsOnlyWritableAttributes(t *testing.T) {
	api, meta := newAPI(t)

	// The API-assigned id is not offered back to it, the description Terraform does
	// not know yet is left out rather than sent as null, and the embedded document
	// goes out as JSON rather than as the string state holds it in.
	api.EXPECT().
		Do(gomock.Any(), request(http.MethodPost, "/things").withBody(`{"name": "first", "device_config_raw": {"mtu": 1500}}`)).
		Return(json.RawMessage(`{
			"id": "thing-1",
			"name": "first",
			"description": "set by the server",
			"device_config_raw": {"mtu": 1500}
		}`), nil)

	res, schema := newResource(t, thingDefinition(), meta)

	plan := objectValue(t, schema, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":              tftypes.NewValue(tftypes.String, "first"),
		"description":       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"device_config_raw": tftypes.NewValue(tftypes.String, `{"mtu":1500}`),
	})

	resp := createResource(t, res, schema, plan)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, "thing-1", attributeString(t, resp.State.Raw, "id"))
	require.Equal(t, "set by the server", attributeString(t, resp.State.Raw, "description"))
	require.Equal(t, `{"mtu":1500}`, attributeString(t, resp.State.Raw, "device_config_raw"),
		"the document is kept exactly as it was configured")
}

func TestResourceCreateReportsAMissingID(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodPost, "/things")).
		Return(json.RawMessage(`{"name": "first"}`), nil)

	res, schema := newResource(t, thingDefinition(), meta)

	plan := objectValue(t, schema, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name": tftypes.NewValue(tftypes.String, "first"),
	})

	resp := createResource(t, res, schema, plan)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "cannot be tracked")
}

// TestResourceCreateSurfacesATransportFailure covers a request that never reaches
// the API at all, which a practitioner has to be told about verbatim: it is
// usually a wrong endpoint or a proxy in the way, and the message is the only clue.
func TestResourceCreateSurfacesATransportFailure(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodPost, "/things")).
		Return(nil, errors.New("dial tcp 10.0.0.1:443: connect: connection refused"))

	res, schema := newResource(t, thingDefinition(), meta)

	plan := objectValue(t, schema, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name": tftypes.NewValue(tftypes.String, "first"),
	})

	resp := createResource(t, res, schema, plan)

	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "Could not write thing", resp.Diagnostics.Errors()[0].Summary())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "connection refused")
}

func TestResourceCreateReportsAResponseThatIsNotADocument(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodPost, "/things")).
		Return(json.RawMessage(`{"id": "thing-1", `), nil)

	res, schema := newResource(t, thingDefinition(), meta)

	plan := objectValue(t, schema, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name": tftypes.NewValue(tftypes.String, "first"),
	})

	resp := createResource(t, res, schema, plan)

	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "Could not write thing", resp.Diagnostics.Errors()[0].Summary())
}

func TestResourceUpdateSendsOnlyTheBodyAndKeepsThePlan(t *testing.T) {
	api, meta := newAPI(t)

	// The id identifies the object in the path, so it is not repeated in the body.
	api.EXPECT().
		Do(gomock.Any(), request(http.MethodPatch, "/things/thing-1").withBody(`{"name": "renamed", "description": "mine"}`)).
		Return(json.RawMessage(`{
			"id": "thing-1",
			"name": "renamed",
			"description": "server wins nothing"
		}`), nil)

	res, schema := newResource(t, thingDefinition(), meta)

	plan := objectValue(t, schema, map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, "thing-1"),
		"name":        tftypes.NewValue(tftypes.String, "renamed"),
		"description": tftypes.NewValue(tftypes.String, "mine"),
	})

	state := objectValue(t, schema, map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, "thing-1"),
		"name":        tftypes.NewValue(tftypes.String, "first"),
		"description": tftypes.NewValue(tftypes.String, "mine"),
	})

	resp := updateResource(t, res, schema, plan, state)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, "mine", attributeString(t, resp.State.Raw, "description"))
}

// TestResourceUpdateReadsBackWhenTheResponseIsEmpty covers an update the API
// answers with no body: the computed attributes are still unresolved, so the
// object has to be read back before it can be put into state.
func TestResourceUpdateReadsBackWhenTheResponseIsEmpty(t *testing.T) {
	api, meta := newAPI(t)

	gomock.InOrder(
		api.EXPECT().
			Do(gomock.Any(), request(http.MethodPatch, "/things/thing-1")).
			Return(nil, nil),
		api.EXPECT().
			Do(gomock.Any(), request(http.MethodGet, "/things/thing-1").withBody("")).
			Return(json.RawMessage(`{"id": "thing-1", "name": "renamed", "description": "from the server"}`), nil),
	)

	res, schema := newResource(t, thingDefinition(), meta)

	plan := objectValue(t, schema, map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, "thing-1"),
		"name":        tftypes.NewValue(tftypes.String, "renamed"),
		"description": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	})

	resp := updateResource(t, res, schema, plan, plan)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, "from the server", attributeString(t, resp.State.Raw, "description"))
}

// TestResourceUpdateDoesNotReadBackWhenTheResponseCarriesTheObject is the other
// half of the rule above: the mock fails the test if a second request is made, so
// this pins down that an update answering with the object costs one round trip.
func TestResourceUpdateDoesNotReadBackWhenTheResponseCarriesTheObject(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodPatch, "/things/thing-1")).
		Return(json.RawMessage(`{"id": "thing-1", "name": "renamed", "description": "from the server"}`), nil).
		Times(1)

	res, schema := newResource(t, thingDefinition(), meta)

	plan := objectValue(t, schema, map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, "thing-1"),
		"name":        tftypes.NewValue(tftypes.String, "renamed"),
		"description": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	})

	resp := updateResource(t, res, schema, plan, plan)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, "from the server", attributeString(t, resp.State.Raw, "description"))
}

func TestResourceReadDropsAResourceTheAPINoLongerHas(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things/thing-1")).
		Return(nil, apiError(http.StatusNotFound, "gone"))

	res, schema := newResource(t, thingDefinition(), meta)

	state := objectValue(t, schema, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, "thing-1"),
		"name": tftypes.NewValue(tftypes.String, "first"),
	})

	resp := readResource(t, res, schema, state)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.True(t, resp.State.Raw.IsNull(), "the resource has to be dropped from state so Terraform recreates it")
}

// TestResourceReadKeepsAResourceOnAServerError separates "the object is gone" from
// "the API could not say": only the first may drop it from state, or a transient
// failure would have Terraform recreate a resource that still exists.
func TestResourceReadKeepsAResourceOnAServerError(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things/thing-1")).
		Return(nil, apiError(http.StatusBadGateway, "upstream connect error"))

	res, schema := newResource(t, thingDefinition(), meta)

	state := objectValue(t, schema, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, "thing-1"),
		"name": tftypes.NewValue(tftypes.String, "first"),
	})

	resp := readResource(t, res, schema, state)

	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "Could not read thing", resp.Diagnostics.Errors()[0].Summary())
	require.False(t, resp.State.Raw.IsNull())
}

func TestResourceDeleteToleratesAnAlreadyDeletedResource(t *testing.T) {
	api, meta := newAPI(t)

	// One request, and no read to check first: deleting something already gone is
	// the outcome that was wanted.
	api.EXPECT().
		Do(gomock.Any(), request(http.MethodDelete, "/things/thing-1").withBody("")).
		Return(nil, apiError(http.StatusNotFound, "gone")).
		Times(1)

	res, schema := newResource(t, thingDefinition(), meta)

	state := objectValue(t, schema, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, "thing-1"),
		"name": tftypes.NewValue(tftypes.String, "first"),
	})

	resp := &resource.DeleteResponse{State: tfsdk.State{Schema: schema, Raw: state}}
	res.Delete(context.Background(), resource.DeleteRequest{State: tfsdk.State{Schema: schema, Raw: state}}, resp)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
}

func TestResourceDeleteReportsAFailureItCannotIgnore(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodDelete, "/things/thing-1")).
		Return(nil, apiError(http.StatusConflict, "still referenced by a policy"))

	res, schema := newResource(t, thingDefinition(), meta)

	state := objectValue(t, schema, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, "thing-1"),
		"name": tftypes.NewValue(tftypes.String, "first"),
	})

	resp := &resource.DeleteResponse{State: tfsdk.State{Schema: schema, Raw: state}}
	res.Delete(context.Background(), resource.DeleteRequest{State: tfsdk.State{Schema: schema, Raw: state}}, resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "still referenced by a policy")
}

// nestedDefinition mirrors an address object: it lives under a parent the API
// only names in the path, and has no single-object GET.
func nestedDefinition() Definition {
	return Definition{
		Name: "nested",
		Schema: func(_ context.Context) rschema.Schema {
			return rschema.Schema{
				Attributes: map[string]rschema.Attribute{
					"id":   rschema.StringAttribute{Computed: true},
					"name": rschema.StringAttribute{Required: true},
				},
			}
		},
		Create: Operation{Method: http.MethodPost, Path: "/groups/{group_id}/nested"},
		Read:   Operation{Method: http.MethodGet, Path: "/groups/{group_id}/nested"},
		Update: Operation{Method: http.MethodPatch, Path: "/groups/{group_id}/nested/{id}"},
		Delete: Operation{Method: http.MethodDelete, Path: "/groups/{group_id}/nested/{id}"},
	}
}

func TestNestedResourceGainsARequiredParentAttribute(t *testing.T) {
	_, meta := newAPI(t)

	_, schema := newResource(t, nestedDefinition(), meta)

	parent, ok := schema.Attributes["group_id"]
	require.True(t, ok, "the parent identifier has to be part of the schema even though the API only names it in the path")
	require.True(t, parent.IsRequired())
}

func TestNestedResourceCreateKeepsTheParentOutOfTheBody(t *testing.T) {
	api, meta := newAPI(t)

	// group_id identifies the parent in the path, so sending it again in the body
	// would be offering the API a field it never declared.
	api.EXPECT().
		Do(gomock.Any(), request(http.MethodPost, "/groups/group-1/nested").withBody(`{"name": "first"}`)).
		Return(json.RawMessage(`{"id": "nested-1", "name": "first"}`), nil)

	res, schema := newResource(t, nestedDefinition(), meta)

	plan := objectValue(t, schema, map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"group_id": tftypes.NewValue(tftypes.String, "group-1"),
		"name":     tftypes.NewValue(tftypes.String, "first"),
	})

	resp := createResource(t, res, schema, plan)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, "group-1", attributeString(t, resp.State.Raw, "group_id"))
}

func TestNestedResourceReadWalksTheCollection(t *testing.T) {
	api, meta := newAPI(t)

	gomock.InOrder(
		api.EXPECT().
			Do(gomock.Any(), request(http.MethodGet, "/groups/group-1/nested").withQuery("")).
			Return(page("c1", true, object("other", "other")), nil),
		api.EXPECT().
			Do(gomock.Any(), request(http.MethodGet, "/groups/group-1/nested").withQuery("after=c1")).
			Return(page("c2", false, object("nested-1", "found")), nil),
	)

	res, schema := newResource(t, nestedDefinition(), meta)

	state := objectValue(t, schema, map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, "nested-1"),
		"group_id": tftypes.NewValue(tftypes.String, "group-1"),
		"name":     tftypes.NewValue(tftypes.String, "stale"),
	})

	resp := readResource(t, res, schema, state)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, "found", attributeString(t, resp.State.Raw, "name"))
}

func TestNestedResourceReadDropsAnObjectMissingFromTheCollection(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/groups/group-1/nested")).
		Return(page("", false, object("other", "other")), nil)

	res, schema := newResource(t, nestedDefinition(), meta)

	state := objectValue(t, schema, map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, "nested-1"),
		"group_id": tftypes.NewValue(tftypes.String, "group-1"),
		"name":     tftypes.NewValue(tftypes.String, "stale"),
	})

	resp := readResource(t, res, schema, state)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.True(t, resp.State.Raw.IsNull())
}

// TestCollectionWalkGivesUpOnAnEndlessCollection covers the bound on a collection
// walk. A server that keeps reporting another page would otherwise hang Terraform
// rather than fail it.
func TestCollectionWalkGivesUpOnAnEndlessCollection(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/groups/group-1/nested")).
		Return(page("always-more", true, object("other", "other")), nil).
		Times(maxPages)

	res, schema := newResource(t, nestedDefinition(), meta)

	state := objectValue(t, schema, map[string]tftypes.Value{
		"id":       tftypes.NewValue(tftypes.String, "nested-1"),
		"group_id": tftypes.NewValue(tftypes.String, "group-1"),
		"name":     tftypes.NewValue(tftypes.String, "stale"),
	})

	resp := readResource(t, res, schema, state)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "gave up after 1000 pages")
}

func TestImportStateAcceptsTheParentAndTheID(t *testing.T) {
	_, meta := newAPI(t)

	res, schema := newResource(t, nestedDefinition(), meta)

	importer, ok := res.(resource.ResourceWithImportState)
	require.True(t, ok)

	resp := &resource.ImportStateResponse{State: tfsdk.State{Schema: schema, Raw: nullValue(t, schema)}}
	importer.ImportState(context.Background(), resource.ImportStateRequest{ID: "group-1/nested-1"}, resp)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	var groupID, id string
	require.False(t, resp.State.GetAttribute(context.Background(), path.Root("group_id"), &groupID).HasError())
	require.False(t, resp.State.GetAttribute(context.Background(), path.Root("id"), &id).HasError())
	require.Equal(t, "group-1", groupID)
	require.Equal(t, "nested-1", id)

	resp = &resource.ImportStateResponse{State: tfsdk.State{Schema: schema, Raw: nullValue(t, schema)}}
	importer.ImportState(context.Background(), resource.ImportStateRequest{ID: "nested-1"}, resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), `Expected "group_id/id"`)
}

// immutableDefinition mirrors a link monitor: the API offers no update.
func immutableDefinition() Definition {
	def := thingDefinition()
	def.Update = Operation{}

	return def
}

func TestModifyPlanRequiresReplacingAResourceTheAPICannotUpdate(t *testing.T) {
	_, meta := newAPI(t)

	res, schema := newResource(t, immutableDefinition(), meta)

	planModifier, ok := res.(resource.ResourceWithModifyPlan)
	require.True(t, ok)

	state := objectValue(t, schema, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, "thing-1"),
		"name": tftypes.NewValue(tftypes.String, "first"),
	})

	plan := objectValue(t, schema, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, "thing-1"),
		"name": tftypes.NewValue(tftypes.String, "renamed"),
	})

	resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: schema, Raw: plan}}
	planModifier.ModifyPlan(context.Background(), resource.ModifyPlanRequest{
		State: tfsdk.State{Schema: schema, Raw: state},
		Plan:  tfsdk.Plan{Schema: schema, Raw: plan},
	}, resp)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, path.Paths{path.Root("name")}, resp.RequiresReplace)
}

func TestModifyPlanLeavesAnUpdatableResourceAlone(t *testing.T) {
	_, meta := newAPI(t)

	res, schema := newResource(t, thingDefinition(), meta)

	planModifier, ok := res.(resource.ResourceWithModifyPlan)
	require.True(t, ok)

	state := objectValue(t, schema, map[string]tftypes.Value{"name": tftypes.NewValue(tftypes.String, "first")})
	plan := objectValue(t, schema, map[string]tftypes.Value{"name": tftypes.NewValue(tftypes.String, "renamed")})

	resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: schema, Raw: plan}}
	planModifier.ModifyPlan(context.Background(), resource.ModifyPlanRequest{
		State: tfsdk.State{Schema: schema, Raw: state},
		Plan:  tfsdk.Plan{Schema: schema, Raw: plan},
	}, resp)

	require.Empty(t, resp.RequiresReplace)
}

func TestEmbeddedJSONAttributeUsesASemanticStringType(t *testing.T) {
	_, meta := newAPI(t)

	_, schema := newResource(t, thingDefinition(), meta)

	attribute, ok := schema.Attributes["device_config_raw"].(rschema.StringAttribute)
	require.True(t, ok)

	// jsontypes.Normalized compares documents rather than text, so re-serialising
	// one on a refresh cannot show up as a change.
	require.Equal(t, jsontypes.NormalizedType{}, attribute.CustomType)
}

func TestIDIsHeldStableAcrossPlans(t *testing.T) {
	_, meta := newAPI(t)

	_, schema := newResource(t, thingDefinition(), meta)

	attribute, ok := schema.Attributes["id"].(rschema.StringAttribute)
	require.True(t, ok)
	require.Len(t, attribute.PlanModifiers, 1)
}
