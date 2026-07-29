package scim

import (
	"errors"
	"fmt"
	"testing"

	"github.com/couchbase/gocb/v2"
	"github.com/couchbase/gocbcore/v10/memd"
)

// Translating the driver's "document missing" error is what separates a 404
// from a 500, and it is the one piece of the Couchbase store that can be
// exercised without a cluster. It also pins the behaviour across driver
// upgrades: the shape of this error is the part most likely to move.
func TestIsKeyNotFound(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"the driver's sentinel", gocb.ErrDocumentNotFound, true},
		{"a wrapped sentinel", fmt.Errorf("get: %w", gocb.ErrDocumentNotFound), true},
		{"a key-value error carrying the status", &gocb.KeyValueError{StatusCode: memd.StatusKeyNotFound}, true},
		{"a key-value error with another status", &gocb.KeyValueError{StatusCode: memd.StatusBusy}, false},
		{"an unrelated driver error", gocb.ErrTimeout, false},
		{"an unrelated error", errors.New("connection refused"), false},
		{"no error at all", nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isKeyNotFound(tc.err); got != tc.want {
				t.Errorf("isKeyNotFound(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// The store must satisfy the interface the handlers depend on, and its search
// must reject a malformed filter before it needs a cluster at all.
func TestCouchbaseStoreConstruction(t *testing.T) {
	store := NewCouchbaseStore(nil)
	if _, _, err := store.Search(SearchQuery{Bucket: "Element", Filter: "@@@"}); !errors.Is(err, ErrInvalidFilter) {
		t.Errorf("Search with a malformed filter = %v, want ErrInvalidFilter", err)
	}
}
