package bwanclient

import (
	"context"
	"encoding/json"
)

// API is the whole of the API as the rest of the provider sees it.
//
// The provider is generated, so nothing above this line knows about individual
// objects: a request is a method, a path and a document, whichever resource it
// belongs to. That makes one method enough to stand for the entire API, and it
// makes the seam a test can replace — a mock here substitutes for every endpoint
// at once while the engine under test runs unchanged.
type API interface {
	Do(ctx context.Context, req Request) (json.RawMessage, error)
}

var _ API = (*Client)(nil)
