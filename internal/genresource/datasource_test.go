package genresource

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// fill completes an object with nulls for every attribute the caller did not
// mention, the way Terraform always hands one over.
func fill(t *testing.T, typ tftypes.Type, members map[string]tftypes.Value) tftypes.Value {
	t.Helper()

	asObject, ok := typ.(tftypes.Object)
	require.True(t, ok, "%s is not an object", typ)

	full := make(map[string]tftypes.Value, len(asObject.AttributeTypes))

	for name, attrType := range asObject.AttributeTypes {
		if member, ok := members[name]; ok {
			full[name] = member

			continue
		}

		full[name] = tftypes.NewValue(attrType, nil)
	}

	return tftypes.NewValue(asObject, full)
}

// thingsSchema stands in for a generated list data source: the collection's
// query parameters plus the paginated envelope the API answers with.
func thingsSchema(_ context.Context) dschema.Schema {
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
}

func newDataSource(t *testing.T, def DataSourceDefinition, meta *Meta) (datasource.DataSource, dschema.Schema) {
	t.Helper()

	source := NewDataSource(def)()

	configure, ok := source.(datasource.DataSourceWithConfigure)
	require.True(t, ok)

	configureResp := &datasource.ConfigureResponse{}
	configure.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: meta}, configureResp)
	require.False(t, configureResp.Diagnostics.HasError(), "%v", configureResp.Diagnostics)

	schemaResp := &datasource.SchemaResponse{}
	source.Schema(context.Background(), datasource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), "%v", schemaResp.Diagnostics)

	return source, schemaResp.Schema
}

func readDataSource(t *testing.T, source datasource.DataSource, schema dschema.Schema, config tftypes.Value) *datasource.ReadResponse {
	t.Helper()

	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schema, Raw: tftypes.NewValue(schema.Type().TerraformType(context.Background()), nil)}}
	source.Read(context.Background(), datasource.ReadRequest{Config: tfsdk.Config{Schema: schema, Raw: config}}, resp)

	return resp
}

// configFor builds the configuration Terraform would hand a data source, with
// everything the caller did not set left null.
func configFor(t *testing.T, schema dschema.Schema, members map[string]tftypes.Value) tftypes.Value {
	t.Helper()

	return fill(t, schema.Type().TerraformType(context.Background()), members)
}

func dataElements(t *testing.T, state tftypes.Value) []tftypes.Value {
	t.Helper()

	var members map[string]tftypes.Value
	require.NoError(t, state.As(&members))

	var elements []tftypes.Value
	require.NoError(t, members[dataField].As(&elements))

	return elements
}

func thingsDefinition() DataSourceDefinition {
	return DataSourceDefinition{
		Name:   "things",
		Schema: thingsSchema,
		Read:   Operation{Method: http.MethodGet, Path: "/things"},
	}
}

func TestListDataSourceWalksEveryPageByDefault(t *testing.T) {
	api, meta := newAPI(t)

	// The first request must not carry a cursor; the second has to carry the one the
	// first answered with.
	gomock.InOrder(
		api.EXPECT().
			Do(gomock.Any(), request(http.MethodGet, "/things").withQuery("")).
			Return(page("c1", true, object("1", "a"), object("2", "b")), nil),
		api.EXPECT().
			Do(gomock.Any(), request(http.MethodGet, "/things").withQuery("after=c1")).
			Return(page("c2", false, object("3", "c")), nil),
	)

	source, schema := newDataSource(t, thingsDefinition(), meta)

	resp := readDataSource(t, source, schema, configFor(t, schema, nil))

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Len(t, dataElements(t, resp.State.Raw), 3, "a list data source returns the whole collection, not one page")
}

// TestListDataSourceHonoursAnExplicitPage relies on the mock to make the point:
// asking for a page size means asking for exactly one page, so a second request
// would fail the test.
func TestListDataSourceHonoursAnExplicitPage(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things").withQuery("first=1")).
		Return(page("c1", true, object("1", "a")), nil).
		Times(1)

	source, schema := newDataSource(t, thingsDefinition(), meta)

	config := configFor(t, schema, map[string]tftypes.Value{
		"first": tftypes.NewValue(tftypes.Number, 1),
	})

	resp := readDataSource(t, source, schema, config)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Len(t, dataElements(t, resp.State.Raw), 1)
}

func TestListDataSourcePassesFiltersAsQueryParameters(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things").withQuery("filter=name+eq+a&sort=name&sort=-id")).
		Return(page("", false), nil)

	source, schema := newDataSource(t, thingsDefinition(), meta)

	config := configFor(t, schema, map[string]tftypes.Value{
		"filter": tftypes.NewValue(tftypes.String, "name eq a"),
		"sort": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "name"),
			tftypes.NewValue(tftypes.String, "-id"),
		}),
	})

	resp := readDataSource(t, source, schema, config)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	// The filters the practitioner set have to survive into state unchanged.
	var members map[string]tftypes.Value
	require.NoError(t, resp.State.Raw.As(&members))

	var filter string
	require.NoError(t, members["filter"].As(&filter))
	require.Equal(t, "name eq a", filter)
}

func singularDefinition() DataSourceDefinition {
	return DataSourceDefinition{
		Name: "thing",
		Schema: func(_ context.Context) dschema.Schema {
			return dschema.Schema{
				Attributes: map[string]dschema.Attribute{
					"id":   dschema.StringAttribute{Required: true},
					"name": dschema.StringAttribute{Computed: true},
				},
			}
		},
		Read: Operation{Method: http.MethodGet, Path: "/things/{id}"},
	}
}

func TestSingularDataSourceReadsByID(t *testing.T) {
	api, meta := newAPI(t)

	// An identifier travels in the path, not the query string.
	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things/thing-1").withQuery("")).
		Return(json.RawMessage(`{"id": "thing-1", "name": "first"}`), nil)

	source, schema := newDataSource(t, singularDefinition(), meta)

	config := configFor(t, schema, map[string]tftypes.Value{"id": tftypes.NewValue(tftypes.String, "thing-1")})

	resp := readDataSource(t, source, schema, config)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	var members map[string]tftypes.Value
	require.NoError(t, resp.State.Raw.As(&members))

	var name string
	require.NoError(t, members["name"].As(&name))
	require.Equal(t, "first", name)
}

func TestSingularDataSourceReportsAMissingObject(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things/thing-1")).
		Return(nil, apiError(http.StatusNotFound, "gone"))

	def := singularDefinition()
	def.Schema = func(_ context.Context) dschema.Schema {
		return dschema.Schema{Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{Required: true},
		}}
	}

	source, schema := newDataSource(t, def, meta)

	config := configFor(t, schema, map[string]tftypes.Value{"id": tftypes.NewValue(tftypes.String, "thing-1")})

	resp := readDataSource(t, source, schema, config)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "No thing found")
}

// TestSingularDataSourceReportsAnEmptyResponse covers an endpoint that answers 204:
// there is nothing to put into state, and saying so beats leaving every attribute
// unknown.
func TestSingularDataSourceReportsAnEmptyResponse(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things/thing-1")).
		Return(nil, nil)

	source, schema := newDataSource(t, singularDefinition(), meta)

	config := configFor(t, schema, map[string]tftypes.Value{"id": tftypes.NewValue(tftypes.String, "thing-1")})

	resp := readDataSource(t, source, schema, config)

	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "Unexpected API response", resp.Diagnostics.Errors()[0].Summary())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "returned no document")
}
