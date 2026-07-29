package scim

import "errors"

var (
	// ErrNotFound is returned by a Store when the requested document does not exist.
	ErrNotFound = errors.New("scim: resource not found")
	// ErrFilterUnsupported is returned by a Store that cannot evaluate SCIM filters.
	ErrFilterUnsupported = errors.New("scim: this store cannot evaluate filters")
)

// SearchQuery describes a paginated, sorted search over a single resource bucket.
// SortBy is a raw SCIM attribute path (optionally URN-prefixed), not a
// storage-specific expression: translating it is the Store's job.
type SearchQuery struct {
	Bucket         string
	Filter         string
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
	// Search returns the total number of matching resources (ignoring
	// pagination) along with the requested page.
	Search(q SearchQuery) (total int, resources []map[string]interface{}, err error)
	Close() error
}

// DB is the Store backing the SCIM handlers. Production wires a CouchbaseStore
// through InitDB; tests wire a MemoryStore.
var DB Store
