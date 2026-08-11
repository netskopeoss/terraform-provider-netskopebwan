package bwanclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := New(Config{
		Endpoint:   server.URL + "/api",
		Token:      "secret-token",
		UserAgent:  "terraform-provider-netskopebwan/test",
		HTTPClient: server.Client(),
	})
	require.NoError(t, err)

	return client
}

func TestDoAddsVersionPrefixAndCredentials(t *testing.T) {
	var (
		gotPath   string
		gotQuery  string
		gotAuth   string
		gotAgent  string
		gotType   string
		gotBody   string
		gotMethod string
	)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading the request body: %s", err)
		}

		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		gotAgent = r.Header.Get("User-Agent")
		gotType = r.Header.Get("Content-Type")
		gotBody = string(body)

		_, _ = w.Write([]byte(`{"id": "seg-1"}`))
	})

	document, err := client.Do(context.Background(), Request{
		Method: http.MethodPost,
		Path:   "/segments",
		Query:  url.Values{"filter": []string{"name eq x"}},
		Body:   map[string]any{"name": "segment"},
	})

	require.NoError(t, err)
	require.JSONEq(t, `{"id": "seg-1"}`, string(document))

	require.Equal(t, http.MethodPost, gotMethod)
	require.Equal(t, "/api/v2/segments", gotPath)
	require.Equal(t, "filter=name+eq+x", gotQuery)
	require.Equal(t, "Bearer secret-token", gotAuth)
	require.Equal(t, "terraform-provider-netskopebwan/test", gotAgent)
	require.Equal(t, "application/json", gotType)
	require.JSONEq(t, `{"name": "segment"}`, gotBody)
}

func TestDoReturnsNilForAnEmptyResponse(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	document, err := client.Do(context.Background(), Request{Method: http.MethodDelete, Path: "/segments/seg-1"})

	require.NoError(t, err)
	require.Nil(t, document)
}

func TestDoTranslatesStatusCodesToSentinels(t *testing.T) {
	for _, testCase := range []struct {
		status   int
		sentinel error
	}{
		{http.StatusNotFound, ErrNotFound},
		{http.StatusUnauthorized, ErrUnauthorized},
		{http.StatusForbidden, ErrUnauthorized},
		{http.StatusConflict, ErrConflict},
		{http.StatusBadRequest, ErrBadRequest},
	} {
		t.Run(http.StatusText(testCase.status), func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(testCase.status)
			})

			_, err := client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/segments/seg-1"})

			require.ErrorIs(t, err, testCase.sentinel)
		})
	}
}

func TestDoSurfacesTheAPIErrorDocument(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message": "name is already taken", "error_code": "DUPLICATE_NAME", "request_id": "abc123"}`))
	})

	_, err := client.Do(context.Background(), Request{Method: http.MethodPost, Path: "/segments"})

	require.ErrorIs(t, err, ErrBadRequest)
	require.ErrorContains(t, err, "POST /segments: 400 name is already taken")
	require.ErrorContains(t, err, "DUPLICATE_NAME")
	require.ErrorContains(t, err, "abc123")

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, "DUPLICATE_NAME", apiErr.Code)
}

func TestDoSurfacesANonJSONErrorBody(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("upstream connect error"))
	})

	_, err := client.Do(context.Background(), Request{Method: http.MethodGet, Path: "/segments"})

	require.ErrorContains(t, err, "upstream connect error")

	// 502 has no unambiguous meaning for a caller, so it stays a bare API error.
	require.NotErrorIs(t, err, ErrNotFound)
	require.NotErrorIs(t, err, ErrUnauthorized)
}

func TestNewRejectsAnIncompleteConfiguration(t *testing.T) {
	_, err := New(Config{Token: "t"})
	require.ErrorContains(t, err, "endpoint is required")

	_, err = New(Config{Endpoint: "https://example.com"})
	require.ErrorContains(t, err, "token is required")

	_, err = New(Config{Endpoint: "example.com", Token: "t"})
	require.ErrorContains(t, err, "must be an http or https URL")
}

func TestEscapePathSegment(t *testing.T) {
	require.Equal(t, "a%2Fb", EscapePathSegment("a/b"))
}
