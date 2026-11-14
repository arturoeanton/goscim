package scim

import (
	"context"
	"errors"
)

var (
	// ErrNotFound is returned by a Store when the requested document does not exist.
	ErrNotFound = errors.New("scim: resource not found")
	// ErrFilterUnsupported is returned by a Store that cannot evaluate SCIM filters.
	ErrFilterUnsupported = errors.New("scim: this store cannot evaluate filters")
	// ErrInvalidFilter is returned when the client's filter does not parse. It
	// is a client error, so handlers must map it to 400 and not 500.
	ErrInvalidFilter = errors.New("scim: invalid filter")
)

// UniqueClaim identifies one unique attribute value claimed by a resource.
type UniqueClaim struct {
	Attribute string
	LockID    string
	Key       string
	RawValue  interface{}
	CaseExact bool
}

// UniqueConflictError reports that a unique attribute value is already owned by
// another resource in the same bucket.
type UniqueConflictError struct {
	Attribute string
}

func (e *UniqueConflictError) Error() string {
	return "scim: duplicate unique value for " + e.Attribute
}

// SearchQuery describes a paginated, sorted search over a single resource bucket.
// SortBy is a raw SCIM attribute path (optionally URN-prefixed), not a
// storage-specific expression: translating it is the Store's job.
type SearchQuery struct {
	Bucket string
	Filter string
	// ResolveName maps an attribute path to the spelling its schema declares.
	// Optional: without it the filter is translated verbatim.
	ResolveName    func(path string) string
	IsCaseExact    func(path string) bool
	IsMultiValued  func(path string) bool
	SortBy         string
	SortDescending bool
	Offset         int
	Limit          int
}

// Store is the persistence layer used by the SCIM operations. Resources are
// plain decoded JSON documents keyed by their SCIM id inside a per-resource-type
// bucket. Implementations must return ErrNotFound rather than a driver-specific
// error when a document is missing, so the handlers can map it to a 404.
type Store interface {
	// EnsureBucket creates the bucket for a resource type if it does not exist.
	EnsureBucket(name string) error
	Get(bucket, id string) (map[string]interface{}, error)
	Upsert(bucket, id string, doc map[string]interface{}) error
	Replace(bucket, id string, doc map[string]interface{}) error
	Remove(bucket, id string) error
	// CreateUnique atomically acquires claims and inserts doc, returning
	// *UniqueConflictError if any claim is already held by another resource.
	CreateUnique(bucket, id string, doc map[string]interface{}, claims []UniqueClaim) error
	// ReplaceUnique atomically updates unique claims and replaces doc,
	// releasing superseded claims once the document write succeeds.
	ReplaceUnique(bucket, id string, doc map[string]interface{}, newClaims []UniqueClaim, extractClaims func(map[string]interface{}) []UniqueClaim) error
	// RemoveUnique removes doc and releases any unique claims owned by id.
	RemoveUnique(bucket, id string, claims []UniqueClaim) error
	// Search returns the total number of matching resources (ignoring
	// pagination) along with the requested page.
	Search(q SearchQuery) (total int, resources []map[string]interface{}, err error)
	// FindIDByAttribute returns the id of a resource whose attribute holds
	// value, or "" when there is none.
	FindIDByAttribute(bucket, attributePath string, value interface{}) (string, error)
	// Ping checks that the underlying store is reachable within ctx.
	Ping(ctx context.Context) error
	Close() error
}

// DB is the Store backing the SCIM handlers. Production wires a CouchbaseStore
// through InitDB; tests wire a MemoryStore.
var DB Store
