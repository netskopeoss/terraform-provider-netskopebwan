package genresource

import (
	"context"
	"fmt"
	"net/url"

	"infiot.com/infiot/mgmt/tf-provider/internal/bwanclient"
	"infiot.com/infiot/mgmt/tf-provider/internal/tfjson"
)

// Field names of the paginated collection envelope every list endpoint returns.
const (
	dataField      = "data"
	pageInfoField  = "page_info"
	endCursorField = "end_cursor"
	hasNextField   = "has_next"
)

// Query parameters shared by every list endpoint.
const (
	afterParam = "after"
	firstParam = "first"
	// filterAttribute is both the query parameter the API filters a collection
	// with and the attribute a data source exposes it as.
	filterAttribute = "filter"
)

// maxPages bounds a collection walk so a server that keeps reporting another
// page cannot spin forever.
const maxPages = 1000

// fetch performs a request and decodes the response document. The second result
// reports whether the response carried one at all: a delete, and some updates,
// answer with an empty body.
func fetch(ctx context.Context, client bwanclient.API, req bwanclient.Request) (any, bool, error) {
	raw, err := client.Do(ctx, req)
	if err != nil {
		return nil, false, err
	}

	if raw == nil {
		return nil, false, nil
	}

	document, err := tfjson.Unmarshal(raw)
	if err != nil {
		return nil, false, err
	}

	return document, true, nil
}

// collection is one page of a list endpoint.
type collection struct {
	Data     []any
	PageInfo map[string]any
}

// Document rebuilds the envelope so it can be decoded against a data source
// schema.
func (c collection) Document() map[string]any {
	document := map[string]any{dataField: c.Data}

	if c.PageInfo != nil {
		document[pageInfoField] = c.PageInfo
	}

	return document
}

// fetchPage reads a single page of a collection.
func fetchPage(ctx context.Context, client bwanclient.API, path string, query url.Values) (collection, error) {
	document, found, err := fetch(ctx, client, bwanclient.Request{Method: "GET", Path: path, Query: query})
	if err != nil {
		return collection{}, err
	}

	if !found {
		return collection{}, fmt.Errorf("GET %s: expected a collection object, got no document", path)
	}

	envelope, ok := document.(map[string]any)
	if !ok {
		return collection{}, fmt.Errorf("GET %s: expected a collection object, got %T", path, document)
	}

	page := collection{}

	if items, ok := envelope[dataField].([]any); ok {
		page.Data = items
	}

	if info, ok := envelope[pageInfoField].(map[string]any); ok {
		page.PageInfo = info
	}

	return page, nil
}

// fetchAll walks every page of a collection, so a caller that did not ask for a
// specific page gets the whole list rather than whatever the server's default
// page size happens to be.
func fetchAll(ctx context.Context, client bwanclient.API, path string, query url.Values) (collection, error) {
	all := collection{}
	cursor := ""

	for range maxPages {
		pageQuery := cloneQuery(query)
		if cursor != "" {
			pageQuery.Set(afterParam, cursor)
		}

		page, err := fetchPage(ctx, client, path, pageQuery)
		if err != nil {
			return collection{}, err
		}

		all.Data = append(all.Data, page.Data...)
		all.PageInfo = page.PageInfo

		hasNext, _ := page.PageInfo[hasNextField].(bool)
		cursor, _ = page.PageInfo[endCursorField].(string)

		if !hasNext || cursor == "" || len(page.Data) == 0 {
			return all, nil
		}
	}

	return collection{}, fmt.Errorf("GET %s: gave up after %d pages", path, maxPages)
}

// findByID walks a collection looking for the element carrying id, which is how
// an object the API exposes no single-object GET for is refreshed. An element of
// another kind is skipped: one endpoint can serve several.
func findByID(ctx context.Context, client bwanclient.API, path, id string, variant *Variant) (map[string]any, bool, error) {
	all, err := fetchAll(ctx, client, path, nil)
	if err != nil {
		return nil, false, err
	}

	for _, item := range all.Data {
		element, ok := item.(map[string]any)
		if !ok {
			continue
		}

		if elementID, ok := element[idAttribute].(string); ok && elementID == id && variant.Matches(element) {
			return element, true, nil
		}
	}

	return nil, false, nil
}

func cloneQuery(query url.Values) url.Values {
	out := make(url.Values, len(query))

	for key, values := range query {
		out[key] = append([]string(nil), values...)
	}

	return out
}
