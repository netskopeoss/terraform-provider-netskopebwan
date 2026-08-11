// Package bwanclient is a thin JSON-over-HTTP client for the BWAN v2 API.
//
// It deliberately knows nothing about individual resources: the Terraform
// provider drives it with paths and documents derived from the OpenAPI spec.
package bwanclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// apiPrefix is the path the v2 API is mounted under. It matches the `servers`
// entry of the OpenAPI document the provider is generated from.
const apiPrefix = "/v2"

const (
	defaultTimeout = 60 * time.Second

	// maxErrorBody bounds how much of a failed response is read before giving up
	// on decoding an error document.
	maxErrorBody = 1 << 20
)

// Sentinels callers translate into Terraform diagnostics. Every APIError
// unwraps to one of these where the status code has an unambiguous meaning.
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("conflict")
	ErrBadRequest   = errors.New("bad request")
)

// APIError is a non-2xx response, decoded from the API's BaseError document
// where the body carries one.
type APIError struct {
	Method     string
	Path       string
	StatusCode int
	Code       string
	Message    string
	RequestID  string
}

func (e *APIError) Error() string {
	message := e.Message
	if message == "" {
		message = http.StatusText(e.StatusCode)
	}

	out := fmt.Sprintf("%s %s: %d %s", e.Method, e.Path, e.StatusCode, message)

	if e.Code != "" {
		out += fmt.Sprintf(" (%s)", e.Code)
	}

	if e.RequestID != "" {
		out += fmt.Sprintf(" [request %s]", e.RequestID)
	}

	return out
}

func (e *APIError) Unwrap() error {
	switch e.StatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrUnauthorized
	case http.StatusConflict:
		return ErrConflict
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return ErrBadRequest
	default:
		return nil
	}
}

// Config describes how to reach the API.
type Config struct {
	// Endpoint is the base URL the v2 API is served from, without the "/v2"
	// prefix, e.g. https://foo.api.infiot.net.
	Endpoint string
	// Token is the bearer token presented on every request.
	Token string
	// Timeout bounds a single request. Zero selects defaultTimeout.
	Timeout time.Duration
	// InsecureSkipVerify disables TLS certificate verification, for tenants
	// serving a self-signed certificate.
	InsecureSkipVerify bool
	UserAgent          string
	// HTTPClient replaces the client built from the fields above. Tests use it;
	// when set, Timeout and InsecureSkipVerify are ignored.
	HTTPClient *http.Client
}

// Client talks to one BWAN tenant.
type Client struct {
	baseURL   *url.URL
	token     string
	userAgent string
	http      *http.Client
}

func New(cfg Config) (*Client, error) {
	if cfg.Endpoint == "" {
		return nil, errors.New("endpoint is required")
	}

	if cfg.Token == "" {
		return nil, errors.New("token is required")
	}

	baseURL, err := url.Parse(strings.TrimRight(cfg.Endpoint, "/") + apiPrefix)
	if err != nil {
		return nil, fmt.Errorf("parsing endpoint %q: %w", cfg.Endpoint, err)
	}

	if baseURL.Scheme != "https" && baseURL.Scheme != "http" {
		return nil, fmt.Errorf("endpoint %q must be an http or https URL", cfg.Endpoint)
	}

	return &Client{
		baseURL:   baseURL,
		token:     cfg.Token,
		userAgent: cfg.UserAgent,
		http:      httpClient(cfg),
	}, nil
}

func httpClient(cfg Config) *http.Client {
	if cfg.HTTPClient != nil {
		return cfg.HTTPClient
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	client := &http.Client{Timeout: timeout}

	if cfg.InsecureSkipVerify {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // opt-in, for tenants with a self-signed certificate
		client.Transport = transport
	}

	return client
}

// Request is one API call. Path is relative to the API root and must already be
// escaped; use EscapePathSegment for values interpolated into it.
type Request struct {
	Method string
	Path   string
	Query  url.Values
	// Body is marshalled as JSON when non-nil.
	Body any
}

// Do performs req and returns the raw response document, or nil when the
// response carries no body.
func (c *Client) Do(ctx context.Context, req Request) (json.RawMessage, error) {
	httpReq, err := c.newRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", req.Method, req.Path, err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, apiError(req, resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s %s: reading response: %w", req.Method, req.Path, err)
	}

	if len(bytes.TrimSpace(body)) == 0 {
		return nil, nil
	}

	return body, nil
}

func (c *Client) newRequest(ctx context.Context, req Request) (*http.Request, error) {
	target := *c.baseURL
	target.Path += req.Path

	if len(req.Query) > 0 {
		target.RawQuery = req.Query.Encode()
	}

	var body io.Reader

	if req.Body != nil {
		encoded, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("%s %s: encoding request body: %w", req.Method, req.Path, err)
		}

		body = bytes.NewReader(encoded)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, target.String(), body)
	if err != nil {
		return nil, fmt.Errorf("%s %s: building request: %w", req.Method, req.Path, err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Accept", "application/json")

	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	if c.userAgent != "" {
		httpReq.Header.Set("User-Agent", c.userAgent)
	}

	return httpReq, nil
}

func apiError(req Request, resp *http.Response) error {
	apiErr := &APIError{
		Method:     req.Method,
		Path:       req.Path,
		StatusCode: resp.StatusCode,
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
	if err != nil {
		return apiErr
	}

	var decoded struct {
		Message   string `json:"message"`
		ErrorCode string `json:"error_code"`
		RequestID string `json:"request_id"`
	}

	if err := json.Unmarshal(body, &decoded); err != nil {
		// A non-JSON error body is still worth surfacing, trimmed to something
		// that fits in a diagnostic.
		apiErr.Message = truncate(strings.TrimSpace(string(body)), 512)

		return apiErr
	}

	apiErr.Message = decoded.Message
	apiErr.Code = decoded.ErrorCode
	apiErr.RequestID = decoded.RequestID

	return apiErr
}

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}

	return s[:limit] + "..."
}

// EscapePathSegment escapes a value being interpolated into a request path.
func EscapePathSegment(value string) string {
	return url.PathEscape(value)
}
