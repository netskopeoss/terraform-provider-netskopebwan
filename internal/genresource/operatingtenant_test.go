package genresource

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/netskopeoss/terraform-provider-netskopebwan/internal/bwanclient"
	mock_bwanclient "github.com/netskopeoss/terraform-provider-netskopebwan/internal/bwanclient/mock"
)

// newTenantAPI returns the provider's own client, the client standing for tenant
// "42", and the Meta wiring the two together. Both are strict mocks, so a
// request sent to the wrong tenant fails the test by arriving somewhere nothing
// expects it.
func newTenantAPI(t *testing.T) (own, tenant *mock_bwanclient.MockAPI, meta *Meta) {
	t.Helper()

	controller := gomock.NewController(t)
	own = mock_bwanclient.NewMockAPI(controller)
	tenant = mock_bwanclient.NewMockAPI(controller)

	meta = &Meta{
		Client:     own,
		TypeName:   "bwan",
		RawEnabled: map[string]bool{},
		Tenant: func(id string) (bwanclient.API, error) {
			require.Equal(t, "42", id)

			return tenant, nil
		},
	}

	return own, tenant, meta
}

// TestOperatingTenantIsOnEveryKindOfObject covers the two decorations an object
// can go through, including the data source that has no search endpoint and so
// takes the early way out of decorateDataSource.
func TestOperatingTenantIsOnEveryKindOfObject(t *testing.T) {
	_, meta := newAPI(t)

	_, resourceSchema := newResource(t, thingDefinition(), meta)
	require.Contains(t, resourceSchema.Attributes, OperatingTenantAttribute)

	_, dataSourceSchema := newDataSource(t, singularDefinition(), meta)
	require.Contains(t, dataSourceSchema.Attributes, OperatingTenantAttribute)

	_, listSchema := newDataSource(t, thingsDefinition(), meta)
	require.Contains(t, listSchema.Attributes, OperatingTenantAttribute)
}

func TestOperatingTenantIsOptionalAndForcesReplacement(t *testing.T) {
	_, meta := newAPI(t)

	_, resourceSchema := newResource(t, thingDefinition(), meta)

	attribute, ok := resourceSchema.Attributes[OperatingTenantAttribute]
	require.True(t, ok)
	require.True(t, attribute.IsOptional())
	require.False(t, attribute.IsRequired())
	require.False(t, attribute.IsComputed())

	// Moving an object to another tenant is not something the API can be asked
	// for: it has to be created there and destroyed here.
	plan := objectValue(t, resourceSchema, map[string]tftypes.Value{
		"id":                     tftypes.NewValue(tftypes.String, "thing-1"),
		"name":                   tftypes.NewValue(tftypes.String, "first"),
		OperatingTenantAttribute: tftypes.NewValue(tftypes.String, "42"),
	})

	state := objectValue(t, resourceSchema, map[string]tftypes.Value{
		"id":                     tftypes.NewValue(tftypes.String, "thing-1"),
		"name":                   tftypes.NewValue(tftypes.String, "first"),
		OperatingTenantAttribute: tftypes.NewValue(tftypes.String, "7"),
	})

	typed, ok := attribute.(rschema.StringAttribute)
	require.True(t, ok)
	require.Len(t, typed.PlanModifiers, 1)

	modified := &planmodifier.StringResponse{}
	typed.PlanModifiers[0].PlanModifyString(context.Background(), planmodifier.StringRequest{
		Path:        path.Root(OperatingTenantAttribute),
		State:       tfsdk.State{Schema: resourceSchema, Raw: state},
		StateValue:  types.StringValue("7"),
		Plan:        tfsdk.Plan{Schema: resourceSchema, Raw: plan},
		PlanValue:   types.StringValue("42"),
		ConfigValue: types.StringValue("42"),
	}, modified)

	require.False(t, modified.Diagnostics.HasError(), "%v", modified.Diagnostics)
	require.True(t, modified.RequiresReplace)
}

// TestOperatingTenantRejectsAnIdentifierThatIsNotOne is the same guard as the one
// in TenantClients, one step earlier: a practitioner hears about it at validate
// time rather than at apply time.
func TestOperatingTenantRejectsAnIdentifierThatIsNotOne(t *testing.T) {
	_, meta := newAPI(t)

	_, resourceSchema := newResource(t, thingDefinition(), meta)

	typed, ok := resourceSchema.Attributes[OperatingTenantAttribute].(rschema.StringAttribute)
	require.True(t, ok)
	require.Len(t, typed.Validators, 1)

	for _, id := range []string{"evil.example.net", "42/../7", "42:8443"} {
		resp := &validator.StringResponse{}
		typed.Validators[0].ValidateString(context.Background(), validator.StringRequest{
			Path:        path.Root(OperatingTenantAttribute),
			ConfigValue: types.StringValue(id),
		}, resp)

		require.True(t, resp.Diagnostics.HasError(), "%q is not an identifier", id)
	}

	accepted := &validator.StringResponse{}
	typed.Validators[0].ValidateString(context.Background(), validator.StringRequest{
		Path:        path.Root(OperatingTenantAttribute),
		ConfigValue: types.StringValue("42"),
	}, accepted)

	require.False(t, accepted.Diagnostics.HasError(), "%v", accepted.Diagnostics)
}

// TestResourceCreateAddressesTheOperatingTenant is the whole point of the
// attribute: the request goes to the tenant the object names, and the name
// itself is not part of the object.
func TestResourceCreateAddressesTheOperatingTenant(t *testing.T) {
	_, tenant, meta := newTenantAPI(t)

	tenant.EXPECT().
		Do(gomock.Any(), request(http.MethodPost, "/things").withBody(`{"name": "first"}`)).
		Return(json.RawMessage(`{"id": "thing-1", "name": "first"}`), nil)

	res, schema := newResource(t, thingDefinition(), meta)

	plan := objectValue(t, schema, map[string]tftypes.Value{
		"id":                     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":                   tftypes.NewValue(tftypes.String, "first"),
		OperatingTenantAttribute: tftypes.NewValue(tftypes.String, "42"),
	})

	resp := createResource(t, res, schema, plan)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, "thing-1", attributeString(t, resp.State.Raw, "id"))
	require.Equal(t, "42", attributeString(t, resp.State.Raw, OperatingTenantAttribute),
		"the tenant an object was created in stays in its state")
}

// TestResourceRefreshAndDeleteAddressTheOperatingTenant covers the operations
// that have only state to go on. An object created in another tenant is only
// findable there, so a read that fell back to the provider's own tenant would
// report it as deleted and a delete would leave it behind.
func TestResourceRefreshAndDeleteAddressTheOperatingTenant(t *testing.T) {
	_, tenant, meta := newTenantAPI(t)

	tenant.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things/thing-1")).
		Return(json.RawMessage(`{"id": "thing-1", "name": "first"}`), nil)
	tenant.EXPECT().
		Do(gomock.Any(), request(http.MethodDelete, "/things/thing-1")).
		Return(nil, nil)

	res, schema := newResource(t, thingDefinition(), meta)

	state := objectValue(t, schema, map[string]tftypes.Value{
		"id":                     tftypes.NewValue(tftypes.String, "thing-1"),
		"name":                   tftypes.NewValue(tftypes.String, "first"),
		OperatingTenantAttribute: tftypes.NewValue(tftypes.String, "42"),
	})

	read := readResource(t, res, schema, state)
	require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
	require.Equal(t, "42", attributeString(t, read.State.Raw, OperatingTenantAttribute),
		"a refresh does not lose the tenant, which the API says nothing about")

	deleted := &resource.DeleteResponse{State: tfsdk.State{Schema: schema, Raw: state}}
	res.Delete(context.Background(), resource.DeleteRequest{State: tfsdk.State{Schema: schema, Raw: state}}, deleted)
	require.False(t, deleted.Diagnostics.HasError(), "%v", deleted.Diagnostics)
}

func TestDataSourceAddressesTheOperatingTenant(t *testing.T) {
	_, tenant, meta := newTenantAPI(t)

	// The tenant is how the request is addressed, so it is not also a filter.
	tenant.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things/thing-1").withQuery("")).
		Return(json.RawMessage(`{"id": "thing-1", "name": "first"}`), nil)

	source, schema := newDataSource(t, singularDefinition(), meta)

	config := configFor(t, schema, map[string]tftypes.Value{
		"id":                     tftypes.NewValue(tftypes.String, "thing-1"),
		OperatingTenantAttribute: tftypes.NewValue(tftypes.String, "42"),
	})

	resp := readDataSource(t, source, schema, config)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Equal(t, "42", attributeString(t, resp.State.Raw, OperatingTenantAttribute))
}

// TestOperatingTenantWithoutAFactoryIsRefused pins the failure mode of a Meta
// built without one: nothing may be sent to the provider's own tenant instead,
// because that is the wrong tenant and the object would be created in it.
func TestOperatingTenantWithoutAFactoryIsRefused(t *testing.T) {
	_, meta := newAPI(t)
	meta.Tenant = nil

	res, schema := newResource(t, thingDefinition(), meta)

	plan := objectValue(t, schema, map[string]tftypes.Value{
		"id":                     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":                   tftypes.NewValue(tftypes.String, "first"),
		OperatingTenantAttribute: tftypes.NewValue(tftypes.String, "42"),
	})

	resp := createResource(t, res, schema, plan)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), OperatingTenantAttribute)
}

func TestTenantClientsCachesOneClientPerTenant(t *testing.T) {
	clients := TenantClients(bwanclient.Config{Endpoint: "https://acme.api.example.net", Token: "t"})

	first, err := clients("42")
	require.NoError(t, err)

	again, err := clients("42")
	require.NoError(t, err)
	require.Same(t, first, again, "a client is a connection pool, not a per-operation value")

	other, err := clients("7")
	require.NoError(t, err)
	require.NotSame(t, first, other)
}

// TestTenantClientsRefusesAnIdentifierThatIsNotOne guards the endpoint: the value
// becomes the leading label of the host the provider sends its bearer token to.
func TestTenantClientsRefusesAnIdentifierThatIsNotOne(t *testing.T) {
	clients := TenantClients(bwanclient.Config{Endpoint: "https://acme.api.example.net", Token: "t"})

	for _, id := range []string{"", "evil.example.net", "42/../7", "42:8443"} {
		_, err := clients(id)
		require.Error(t, err, "%q is not an identifier", id)
	}
}
