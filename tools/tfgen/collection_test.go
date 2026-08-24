package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// runPrepFor runs prep over a document, claiming an element of each named
// collection the way a data source entry does.
func runPrepFor(t *testing.T, document string, elements ...string) (map[string]any, []string) {
	t.Helper()

	doc := parse(t, document)
	prep := NewPrep(doc)

	for _, collection := range elements {
		prep.ClaimElement(collection)
	}

	prep.Run()

	return doc, prep.Warnings
}

func operation(t *testing.T, doc map[string]any, path, method string) map[string]any {
	t.Helper()

	out, ok := nestedMap(doc, "paths", path, method)
	require.True(t, ok, "no %s on %s", method, path)

	return out
}

// parameterNames lists the parameters of an operation, in order, as "in:name".
func parameterNames(operation map[string]any) []string {
	var out []string

	for _, parameter := range parametersOf(operation) {
		out = append(out, stringOr(parameter["in"])+":"+stringOr(parameter["name"]))
	}

	return out
}

const listDocument = `
components:
  schemas:
    _PageInfo:
      type: object
      required: [end_cursor, has_next, total_count]
      properties:
        end_cursor: {type: string, readOnly: true}
        has_next: {type: boolean, readOnly: true}
        total_count: {type: integer, readOnly: true}
    Segment:
      type: object
      properties:
        id: {type: string}
        name: {type: string}
paths:
  /segments:
    get:
      parameters:
        - {name: after, in: query, schema: {type: string}}
        - {name: first, in: query, schema: {type: integer}}
        - {name: sort, in: query, schema: {type: array, items: {type: string}}}
        - {name: filter, in: query, schema: {type: string}}
      responses:
        "200":
          content:
            application/json:
              schema:
                type: object
                required: [page_info, data]
                properties:
                  data: {type: array, items: {$ref: '#/components/schemas/Segment'}}
                  page_info: {$ref: '#/components/schemas/_PageInfo'}
`

// TestPrepReducesACollectionToFiltersAndACount covers what a list data source is
// generated from: the two parameters that narrow a collection, and an envelope
// holding the elements beside the count.
func TestPrepReducesACollectionToFiltersAndACount(t *testing.T) {
	doc, warnings := runPrepFor(t, listDocument)

	require.Empty(t, warnings)

	list := operation(t, doc, "/segments", "get")

	// The cursor pair is gone: the runtime walks every page, so there is no page
	// for a configuration to ask for.
	require.Equal(t, []string{"query:sort", "query:filter"}, parameterNames(list))

	envelope, ok := nestedMap(list, "responses", "200", "content", "application/json", "schema")
	require.True(t, ok)

	properties, _ := envelope["properties"].(map[string]any)
	require.Contains(t, properties, "data")
	require.NotContains(t, properties, "page_info")

	total, _ := properties["total_count"].(map[string]any)
	require.Equal(t, "integer", total["type"], "the count keeps the type the spec gave it")
	require.Equal(t, totalCountDescription, total["description"])

	require.Equal(t, []string{"data", "total_count"}, stringList(envelope["required"]))
}

// TestPrepLiftsOneElementOfACollectionOntoItsOwnPath covers the object that can
// only be listed: the data source claims an element of the collection, and the
// schema the generator maps for it is the collection's element schema.
func TestPrepLiftsOneElementOfACollectionOntoItsOwnPath(t *testing.T) {
	doc, warnings := runPrepFor(t, listDocument, "/segments")

	require.Empty(t, warnings)

	paths, _ := doc["paths"].(map[string]any)

	// The path is one only the generator sees, so it carries the separator no API
	// path can contain — the same one the paths emitted per variant carry.
	require.Contains(t, paths, "/segments"+ElementSuffix)
	require.NotContains(t, paths, "/segments/{id}",
		"nothing may invent the endpoint the API does not serve")

	read := operation(t, doc, "/segments"+ElementSuffix, "get")

	// Nothing addresses the element in the request, because no request is made
	// with this path: the runtime walks the collection.
	require.Empty(t, parameterNames(read))

	schema, ok := nestedMap(read, "responses", "200", "content", "application/json", "schema")
	require.True(t, ok)
	require.Equal(t, "#/components/schemas/Segment", schema["$ref"],
		"one object of the collection is what the data source stands for")

	// The collection itself is untouched by any of this, and still lists.
	list := operation(t, doc, "/segments", "get")
	require.Equal(t, []string{"query:sort", "query:filter"}, parameterNames(list))
}

// TestPrepKeepsWhatItTakesToReachAnElementUnderAParent covers an address object:
// its collection hangs off a group, so the group is still needed to reach one.
func TestPrepKeepsWhatItTakesToReachAnElementUnderAParent(t *testing.T) {
	doc, warnings := runPrepFor(t, `
components:
  schemas:
    AddressObject:
      type: object
      properties:
        id: {type: string}
paths:
  /address-groups/{group-id}/address-objects:
    get:
      parameters:
        - {name: filter, in: query, schema: {type: string}}
        - {name: group-id, in: path, required: true, schema: {type: string}}
      responses:
        "200":
          content:
            application/json:
              schema:
                type: object
                properties:
                  data: {type: array, items: {$ref: '#/components/schemas/AddressObject'}}
                  page_info: {type: object, properties: {total_count: {type: integer}}}
`, "/address-groups/{group_id}/address-objects")

	require.Empty(t, warnings)

	read := operation(t, doc, "/address-groups/{group_id}/address-objects"+ElementSuffix, "get")

	// The group the collection lives under survives; the filter that narrows the
	// list does not, because one object is not a list.
	require.Equal(t, []string{"path:group_id"}, parameterNames(read))
}

// TestPrepReportsAnElementItCannotTake covers the three ways the claim can be
// wrong about the API. Each is a warning rather than a rewrite, so the build
// fails on the data source the generator then produces nothing for, saying which
// entry to look at.
func TestPrepReportsAnElementItCannotTake(t *testing.T) {
	_, missing := runPrepFor(t, listDocument, "/segmnets")
	require.Len(t, missing, 1)
	require.Contains(t, missing[0], "claims an element of /segmnets, which the spec does not have")

	_, unlisted := runPrepFor(t, `
paths:
  /segments:
    post:
      responses:
        "200": {}
`, "/segments")
	require.Len(t, unlisted, 1)
	require.Contains(t, unlisted[0], "which the spec does not list")

	// A collection whose elements carry no id cannot have one of them addressed.
	_, anonymous := runPrepFor(t, `
paths:
  /segments:
    get:
      responses:
        "200":
          content:
            application/json:
              schema:
                type: object
                properties:
                  data: {type: array, items: {type: object, properties: {name: {type: string}}}}
                  page_info: {type: object, properties: {total_count: {type: integer}}}
`, "/segments")
	require.Len(t, anonymous, 1)
	require.Contains(t, anonymous[0], `have no "id" to address one by`)
}
