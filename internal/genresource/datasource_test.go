package genresource

import (
	"context"
	"encoding/json"
	"math/big"
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

// listSchema stands in for a generated list data source: the two ways a
// collection can be narrowed, the elements, and the count the API reports for
// them. The cursor parameters are absent because the generated schemas no longer
// carry them — the whole collection is read, so there is no page to ask for.
func listSchema(elements map[string]dschema.Attribute) dschema.Schema {
	return dschema.Schema{
		Attributes: map[string]dschema.Attribute{
			"filter":      dschema.StringAttribute{Optional: true, Computed: true},
			"sort":        dschema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"total_count": dschema.Int64Attribute{Computed: true},
			"data": dschema.ListNestedAttribute{
				Computed:     true,
				NestedObject: dschema.NestedAttributeObject{Attributes: elements},
			},
		},
	}
}

func thingsSchema(_ context.Context) dschema.Schema {
	return listSchema(map[string]dschema.Attribute{
		"id":   dschema.StringAttribute{Computed: true},
		"name": dschema.StringAttribute{Computed: true},
	})
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

// count reads a number out of state. Numbers arrive from the API at full
// precision, so they are compared as integers rather than as text.
func count(t *testing.T, value tftypes.Value) int64 {
	t.Helper()

	var number big.Float
	require.NoError(t, value.As(&number))

	out, accuracy := number.Int64()
	require.Equal(t, big.Exact, accuracy, "%s is not a whole number", value)

	return out
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

// TestListDataSourceWalksEveryPage covers the whole of what a list data source
// answers with: every element of the collection, however many pages that took,
// and the count the API reports for it rather than the last page's.
func TestListDataSourceWalksEveryPage(t *testing.T) {
	api, meta := newAPI(t)

	// The first request must not carry a cursor; the second has to carry the one the
	// first answered with.
	gomock.InOrder(
		api.EXPECT().
			Do(gomock.Any(), request(http.MethodGet, "/things").withQuery("")).
			Return(pageOf("c1", true, 3, object("1", "a"), object("2", "b")), nil),
		api.EXPECT().
			Do(gomock.Any(), request(http.MethodGet, "/things").withQuery("after=c1")).
			Return(pageOf("c2", false, 3, object("3", "c")), nil),
	)

	source, schema := newDataSource(t, thingsDefinition(), meta)

	resp := readDataSource(t, source, schema, configFor(t, schema, nil))

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.Len(t, dataElements(t, resp.State.Raw), 3, "a list data source returns the whole collection, not one page")

	var members map[string]tftypes.Value
	require.NoError(t, resp.State.Raw.As(&members))

	require.Equal(t, int64(3), count(t, members[totalCountField]),
		"the count is the API's own, plucked out of the envelope")
}

// TestListDataSourceCountsWhatItReadWhenTheAPIWillNot covers an endpoint
// answering without a pagination envelope: there is still a count to report, and
// it is the elements that arrived.
func TestListDataSourceCountsWhatItReadWhenTheAPIWillNot(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things")).
		Return(json.RawMessage(`{"data": [{"id": "1", "name": "a"}, {"id": "2", "name": "b"}]}`), nil)

	source, schema := newDataSource(t, thingsDefinition(), meta)

	resp := readDataSource(t, source, schema, configFor(t, schema, nil))

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	var members map[string]tftypes.Value
	require.NoError(t, resp.State.Raw.As(&members))

	require.Equal(t, int64(2), count(t, members[totalCountField]))
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

// listedDefinition is a data source for an object the API only ever lists: there
// is no single-object endpoint, so the read is pointed at the collection and the
// schema is the element's.
func listedDefinition() DataSourceDefinition {
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
		Read:   Operation{Method: http.MethodGet, Path: "/things"},
		Search: Operation{Method: http.MethodGet, Path: "/things"},
	}
}

// TestListedObjectIsFoundInItsCollection covers the object the API has no by-id
// endpoint for: the id is what the walk looks for, so it must not also be sent as
// a filter, and the walk goes on across pages until the object turns up.
func TestListedObjectIsFoundInItsCollection(t *testing.T) {
	api, meta := newAPI(t)

	gomock.InOrder(
		api.EXPECT().
			Do(gomock.Any(), request(http.MethodGet, "/things").withQuery("")).
			Return(page("c1", true, object("thing-1", "first")), nil),
		api.EXPECT().
			Do(gomock.Any(), request(http.MethodGet, "/things").withQuery("after=c1")).
			Return(page("", false, object("thing-2", "second")), nil),
	)

	source, schema := newDataSource(t, listedDefinition(), meta)

	config := configFor(t, schema, map[string]tftypes.Value{"id": tftypes.NewValue(tftypes.String, "thing-2")})

	resp := readDataSource(t, source, schema, config)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	var members map[string]tftypes.Value
	require.NoError(t, resp.State.Raw.As(&members))

	var name string
	require.NoError(t, members["name"].As(&name))
	require.Equal(t, "second", name)
}

// TestListedObjectReportsAMissingID keeps the collection walk from answering with
// nothing: a data source stands for an object, so an id that is not in the
// collection is an error rather than an empty state.
func TestListedObjectReportsAMissingID(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things")).
		Return(page("", false, object("thing-1", "first")), nil)

	source, schema := newDataSource(t, listedDefinition(), meta)

	config := configFor(t, schema, map[string]tftypes.Value{"id": tftypes.NewValue(tftypes.String, "missing")})

	resp := readDataSource(t, source, schema, config)

	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "No thing found", resp.Diagnostics.Errors()[0].Summary())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), `has the id "missing"`)
}

// TestListedObjectIsAddressableWithoutAFilterableCollection covers the id of an
// object the API only lists. The generated schema has it as a field of the object,
// because that is where the collection's elements carry it; nothing could ask by
// it unless the data source turns it into an argument, and a collection that
// cannot be filtered leaves no other way in.
func TestListedObjectIsAddressableWithoutAFilterableCollection(t *testing.T) {
	_, meta := newAPI(t)

	def := listedDefinition()
	def.Search = Operation{}
	def.Schema = func(_ context.Context) dschema.Schema {
		return dschema.Schema{
			Attributes: map[string]dschema.Attribute{
				"id":   dschema.StringAttribute{Computed: true},
				"name": dschema.StringAttribute{Computed: true},
			},
		}
	}

	_, schema := newDataSource(t, def, meta)

	id, ok := schema.Attributes[idAttribute]
	require.True(t, ok)
	require.True(t, id.IsRequired(), "there is nothing else to read one of these by")
	require.NotContains(t, schema.Attributes, filterAttribute, "the collection cannot be filtered")
}

// TestListedObjectCanStillBeFoundByFilter covers the other way to address one of
// these: the collection is filterable, so the filter goes to the API and the
// match has to be unique, exactly as it does for an object with a by-id endpoint.
func TestListedObjectCanStillBeFoundByFilter(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things").withQuery("filter=name+eq+%22second%22")).
		Return(page("", false, object("thing-2", "second")), nil)

	def := listedDefinition()
	def.Schema = func(_ context.Context) dschema.Schema {
		return dschema.Schema{
			Attributes: map[string]dschema.Attribute{
				"id":   dschema.StringAttribute{Optional: true, Computed: true},
				"name": dschema.StringAttribute{Computed: true},
			},
		}
	}

	source, schema := newDataSource(t, def, meta)

	config := configFor(t, schema, map[string]tftypes.Value{
		filterAttribute: tftypes.NewValue(tftypes.String, `name eq "second"`),
	})

	resp := readDataSource(t, source, schema, config)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	var members map[string]tftypes.Value
	require.NoError(t, resp.State.Raw.As(&members))

	var id string
	require.NoError(t, members["id"].As(&id))
	require.Equal(t, "thing-2", id)
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
