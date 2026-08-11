package genresource

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// searchableDefinition mirrors a singular data source whose collection the API
// lets you filter, so the object can be addressed by id or found by name.
func searchableDefinition() DataSourceDefinition {
	def := singularDefinition()
	def.Search = Operation{Method: http.MethodGet, Path: "/things"}

	return def
}

func TestSearchableDataSourceAcceptsEitherIDOrFilter(t *testing.T) {
	_, meta := newAPI(t)

	_, schema := newDataSource(t, searchableDefinition(), meta)

	id, ok := schema.Attributes["id"].(dschema.StringAttribute)
	require.True(t, ok)
	require.True(t, id.IsOptional(), "an object can be found by filter instead")
	require.True(t, id.IsComputed(), "the id is filled in when the object was found by filter")
	require.Len(t, id.Validators, 1, "exactly one of id and filter has to be set")

	filter, ok := schema.Attributes["filter"].(dschema.StringAttribute)
	require.True(t, ok)
	require.True(t, filter.IsOptional())
	require.Contains(t, filter.GetDescription(), `name eq "corporate"`)
}

func TestDataSourceWithoutASearchableCollectionStillRequiresTheID(t *testing.T) {
	_, meta := newAPI(t)

	_, schema := newDataSource(t, singularDefinition(), meta)

	require.NotContains(t, schema.Attributes, "filter")
	require.True(t, schema.Attributes["id"].IsRequired())
}

func TestSearchableDataSourceFindsTheObjectByFilter(t *testing.T) {
	api, meta := newAPI(t)

	// The collection is searched, not the single-object path, and the filter travels
	// as the query parameter the API declared.
	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things").withQuery("filter=name+eq+%22corporate%22")).
		Return(page("", false, object("thing-7", "corporate")), nil)

	source, schema := newDataSource(t, searchableDefinition(), meta)

	config := configFor(t, schema, map[string]tftypes.Value{
		"filter": tftypes.NewValue(tftypes.String, `name eq "corporate"`),
	})

	resp := readDataSource(t, source, schema, config)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	var members map[string]tftypes.Value
	require.NoError(t, resp.State.Raw.As(&members))

	var id, name string
	require.NoError(t, members["id"].As(&id))
	require.NoError(t, members["name"].As(&name))
	require.Equal(t, "thing-7", id, "the id of the object that was found has to reach state")
	require.Equal(t, "corporate", name)
}

func TestSearchableDataSourceStillReadsByIDWhenGivenOne(t *testing.T) {
	api, meta := newAPI(t)

	// An id travels in the path, never as a filter, so the collection is not touched.
	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things/thing-7").withQuery("")).
		Return(json.RawMessage(`{"id": "thing-7", "name": "corporate"}`), nil).
		Times(1)

	source, schema := newDataSource(t, searchableDefinition(), meta)

	config := configFor(t, schema, map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, "thing-7"),
	})

	resp := readDataSource(t, source, schema, config)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
}

func TestSearchableDataSourceRefusesAnAmbiguousFilter(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things")).
		Return(page("", false, object("thing-1", "corporate"), object("thing-2", "corporate")), nil)

	source, schema := newDataSource(t, searchableDefinition(), meta)

	config := configFor(t, schema, map[string]tftypes.Value{
		"filter": tftypes.NewValue(tftypes.String, `name eq "corporate"`),
	})

	resp := readDataSource(t, source, schema, config)

	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "Multiple thing found", resp.Diagnostics.Errors()[0].Summary())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "matches 2 objects")
}

func TestSearchableDataSourceReportsAFilterThatMatchesNothing(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/things")).
		Return(page("", false), nil)

	source, schema := newDataSource(t, searchableDefinition(), meta)

	config := configFor(t, schema, map[string]tftypes.Value{
		"filter": tftypes.NewValue(tftypes.String, `name eq "missing"`),
	})

	resp := readDataSource(t, source, schema, config)

	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "No thing found", resp.Diagnostics.Errors()[0].Summary())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), `name eq \"missing\"`)
}

// TestSearchableDataSourceWalksEveryPageOfTheCollection matters because a filter
// has to be judged against the whole collection: stopping at the first page could
// call a name unique when it is not.
func TestSearchableDataSourceWalksEveryPageOfTheCollection(t *testing.T) {
	api, meta := newAPI(t)

	gomock.InOrder(
		api.EXPECT().
			Do(gomock.Any(), request(http.MethodGet, "/things").withQuery("filter=name+eq+%22corporate%22")).
			Return(page("c1", true, object("thing-1", "corporate")), nil),
		api.EXPECT().
			Do(gomock.Any(), request(http.MethodGet, "/things").withQuery("after=c1&filter=name+eq+%22corporate%22")).
			Return(page("c2", false, object("thing-2", "corporate")), nil),
	)

	source, schema := newDataSource(t, searchableDefinition(), meta)

	config := configFor(t, schema, map[string]tftypes.Value{
		"filter": tftypes.NewValue(tftypes.String, `name eq "corporate"`),
	})

	resp := readDataSource(t, source, schema, config)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "matches 2 objects")
}

// TestSearchableDataSourceSearchesUnderItsParent covers a collection that itself
// sits under a parent, so the search path needs filling in too.
func TestSearchableDataSourceSearchesUnderItsParent(t *testing.T) {
	api, meta := newAPI(t)

	api.EXPECT().
		Do(gomock.Any(), request(http.MethodGet, "/groups/group-1/things").withQuery("filter=name+eq+%22corporate%22")).
		Return(page("", false, object("thing-7", "corporate")), nil)

	def := DataSourceDefinition{
		Name: "thing",
		Schema: func(_ context.Context) dschema.Schema {
			return dschema.Schema{
				Attributes: map[string]dschema.Attribute{
					"group_id": dschema.StringAttribute{Required: true},
					"id":       dschema.StringAttribute{Required: true},
					"name":     dschema.StringAttribute{Computed: true},
				},
			}
		},
		Read:   Operation{Method: http.MethodGet, Path: "/groups/{group_id}/things/{id}"},
		Search: Operation{Method: http.MethodGet, Path: "/groups/{group_id}/things"},
	}

	source, schema := newDataSource(t, def, meta)

	config := configFor(t, schema, map[string]tftypes.Value{
		"group_id": tftypes.NewValue(tftypes.String, "group-1"),
		"filter":   tftypes.NewValue(tftypes.String, `name eq "corporate"`),
	})

	resp := readDataSource(t, source, schema, config)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
}
