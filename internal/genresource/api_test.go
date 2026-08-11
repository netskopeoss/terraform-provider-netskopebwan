package genresource

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"testing"

	"go.uber.org/mock/gomock"

	"infiot.com/infiot/mgmt/tf-provider/internal/bwanclient"
	mock_bwanclient "infiot.com/infiot/mgmt/tf-provider/internal/bwanclient/mock"
)

// newAPI returns a mock of the whole API, plus the provider data a resource or
// data source is configured with.
//
// The engine works out what to call entirely from a schema and a pair of
// operations, so the calls it makes — and the ones it does not — are most of what
// there is to test. Mocking the one interface it reaches the API through means a
// test states the requests it expects up front rather than sifting through the
// ones that arrived, and any request it did not ask for fails it. Serving HTTP
// instead would answer a spurious call quietly.
func newAPI(t *testing.T) (*mock_bwanclient.MockAPI, *Meta) {
	t.Helper()

	api := mock_bwanclient.NewMockAPI(gomock.NewController(t))

	return api, &Meta{Client: api, TypeName: "bwan", RawEnabled: map[string]bool{}}
}

// requestMatcher matches a request by the parts a test means to pin down. A part
// left unset matches anything, so a test that is about a path says nothing about
// the body.
type requestMatcher struct {
	method string
	path   string
	query  *string
	body   *string
}

var (
	_ gomock.Matcher      = (*requestMatcher)(nil)
	_ gomock.GotFormatter = (*requestMatcher)(nil)
)

func request(method, path string) *requestMatcher {
	return &requestMatcher{method: method, path: path}
}

// withQuery pins the encoded query string; "" means there must not be one.
func (m *requestMatcher) withQuery(query string) *requestMatcher {
	m.query = &query

	return m
}

// withBody pins the request document, compared as JSON so that a test can write
// what it expects to be sent rather than the Go value the engine builds it out
// of. "" means the request must carry no body at all.
func (m *requestMatcher) withBody(document string) *requestMatcher {
	m.body = &document

	return m
}

func (m *requestMatcher) Matches(x any) bool {
	req, ok := x.(bwanclient.Request)
	if !ok {
		return false
	}

	if req.Method != m.method || req.Path != m.path {
		return false
	}

	if m.query != nil && req.Query.Encode() != *m.query {
		return false
	}

	return m.body == nil || sameDocument(*m.body, req.Body)
}

func (m *requestMatcher) String() string {
	out := m.method + " " + m.path

	if m.query != nil && *m.query != "" {
		out += "?" + *m.query
	}

	if m.body != nil {
		out += " with body " + *m.body
	}

	return out
}

// Got renders the request the way String renders the expectation, so a mismatch
// reads as two comparable lines instead of a Go value beside a JSON document.
func (m *requestMatcher) Got(x any) string {
	req, ok := x.(bwanclient.Request)
	if !ok {
		return fmt.Sprintf("%v", x)
	}

	out := req.Method + " " + req.Path

	if encoded := req.Query.Encode(); encoded != "" {
		out += "?" + encoded
	}

	if req.Body != nil {
		encoded, err := json.Marshal(req.Body)
		if err != nil {
			return fmt.Sprintf("%s with body %v", out, req.Body)
		}

		out += " with body " + string(encoded)
	}

	return out
}

func sameDocument(want string, body any) bool {
	if want == "" {
		return body == nil
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return false
	}

	var wanted, got any

	if json.Unmarshal([]byte(want), &wanted) != nil || json.Unmarshal(encoded, &got) != nil {
		return false
	}

	return reflect.DeepEqual(wanted, got)
}

// apiError is the error the real client reports for a failed response. Building
// it directly lets a test provoke the status codes the engine treats specially —
// a 404 on a read means the object is gone, a 404 on a delete means it already
// was — without standing up a server to return them.
func apiError(status int, message string) error {
	return &bwanclient.APIError{StatusCode: status, Message: message}
}

// page renders one page of a collection. The envelope is the same for every list
// endpoint, so spelling it out in each test would bury what the test is about.
func page(cursor string, hasNext bool, elements ...map[string]any) json.RawMessage {
	items := make([]any, 0, len(elements))

	for _, element := range elements {
		items = append(items, element)
	}

	encoded, err := json.Marshal(map[string]any{
		dataField: items,
		pageInfoField: map[string]any{
			endCursorField: cursor,
			hasNextField:   hasNext,
			"total_count":  len(elements),
		},
	})
	if err != nil {
		panic(err)
	}

	return encoded
}

// object is a shorthand for one element of a collection.
func object(id, name string, rest ...map[string]any) map[string]any {
	out := map[string]any{idAttribute: id, "name": name}

	for _, extra := range rest {
		maps.Copy(out, extra)
	}

	return out
}
