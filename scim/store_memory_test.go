package scim

import (
	"errors"
	"testing"
)

// MemoryStore is test infrastructure: if it lies, every test built on it lies
// too. These cases pin its contract against Store.

func TestMemoryStoreCRUD(t *testing.T) {
	s := NewMemoryStore()
	if err := s.EnsureBucket("Element"); err != nil {
		t.Fatal(err)
	}

	if _, err := s.Get("Element", "1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get on an empty bucket = %v, want ErrNotFound", err)
	}
	if err := s.Replace("Element", "1", map[string]interface{}{}); !errors.Is(err, ErrNotFound) {
		t.Errorf("Replace on a missing document = %v, want ErrNotFound", err)
	}
	if err := s.Remove("Element", "1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Remove on a missing document = %v, want ErrNotFound", err)
	}

	if err := s.Upsert("Element", "1", map[string]interface{}{"name": "one"}); err != nil {
		t.Fatal(err)
	}
	doc, err := s.Get("Element", "1")
	if err != nil || doc["name"] != "one" {
		t.Fatalf("Get after Upsert = %v, %v", doc, err)
	}

	if err := s.Replace("Element", "1", map[string]interface{}{"name": "two"}); err != nil {
		t.Fatal(err)
	}
	doc, _ = s.Get("Element", "1")
	if doc["name"] != "two" {
		t.Errorf("Replace did not apply: %v", doc)
	}

	if err := s.Remove("Element", "1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get("Element", "1"); !errors.Is(err, ErrNotFound) {
		t.Error("the document survived Remove")
	}
}

// Upserting into a bucket nobody declared must work, matching Couchbase where
// every bucket already exists from startup.
func TestMemoryStoreUpsertCreatesBucket(t *testing.T) {
	s := NewMemoryStore()
	if err := s.Upsert("New", "1", map[string]interface{}{"a": 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get("New", "1"); err != nil {
		t.Fatal(err)
	}
}

// The store must isolate callers from its internal state, as a real driver
// that serializes would.
func TestMemoryStoreIsolatesState(t *testing.T) {
	s := NewMemoryStore()
	original := map[string]interface{}{"nested": map[string]interface{}{"v": "initial"}}
	if err := s.Upsert("Element", "1", original); err != nil {
		t.Fatal(err)
	}

	original["nested"].(map[string]interface{})["v"] = "mutated-outside"
	doc, _ := s.Get("Element", "1")
	if doc["nested"].(map[string]interface{})["v"] != "initial" {
		t.Error("mutating the original map altered stored state")
	}

	doc["nested"].(map[string]interface{})["v"] = "mutated-in-the-copy"
	doc2, _ := s.Get("Element", "1")
	if doc2["nested"].(map[string]interface{})["v"] != "initial" {
		t.Error("mutating the returned map altered stored state")
	}
}

// A Go struct written to the store must come back as a map with float64
// numbers, exactly as it would through Couchbase. Without this the fake would
// hide the type assertions that do break in production.
func TestMemoryStoreNormalizesLikeJSON(t *testing.T) {
	s := NewMemoryStore()
	doc := map[string]interface{}{
		"meta":    Meta{ResourceType: "Element", Created: "2020-01-01T00:00:00Z"},
		"integer": int64(42),
	}
	if err := s.Upsert("Element", "1", doc); err != nil {
		t.Fatal(err)
	}
	out, err := s.Get("Element", "1")
	if err != nil {
		t.Fatal(err)
	}
	meta, ok := out["meta"].(map[string]interface{})
	if !ok {
		t.Fatalf("meta came back as %T, want map[string]interface{}", out["meta"])
	}
	if meta["resourceType"] != "Element" {
		t.Errorf("meta = %v", meta)
	}
	if _, ok := out["integer"].(float64); !ok {
		t.Errorf("integer came back as %T, want float64", out["integer"])
	}
}

func TestMemoryStoreSearchOrderAndPagination(t *testing.T) {
	s := NewMemoryStore()
	for _, name := range []string{"c", "a", "b"} {
		if err := s.Upsert("Element", name, map[string]interface{}{"id": name, "name": name}); err != nil {
			t.Fatal(err)
		}
	}

	total, page, err := s.Search(SearchQuery{Bucket: "Element", SortBy: "name", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 {
		t.Errorf("total = %d", total)
	}
	if page[0]["name"] != "a" || page[2]["name"] != "c" {
		t.Errorf("ascending order = %v", page)
	}

	_, page, _ = s.Search(SearchQuery{Bucket: "Element", SortBy: "name", SortDescending: true, Limit: 10})
	if page[0]["name"] != "c" {
		t.Errorf("descending order = %v", page)
	}

	total, page, _ = s.Search(SearchQuery{Bucket: "Element", SortBy: "name", Offset: 1, Limit: 1})
	if total != 3 {
		t.Errorf("total must ignore pagination, got %d", total)
	}
	if len(page) != 1 || page[0]["name"] != "b" {
		t.Errorf("page = %v", page)
	}

	// An offset past the end yields an empty page, not an error.
	total, page, err = s.Search(SearchQuery{Bucket: "Element", Offset: 99, Limit: 10})
	if err != nil || total != 3 || len(page) != 0 {
		t.Errorf("offset out of range = %d, %v, %v", total, page, err)
	}

	// An unknown bucket is an empty set, not an error.
	total, page, err = s.Search(SearchQuery{Bucket: "DoesNotExist", Limit: 10})
	if err != nil || total != 0 || len(page) != 0 {
		t.Errorf("unknown bucket = %d, %v, %v", total, page, err)
	}
}

// The fake does not evaluate filters: it must say so out loud rather than
// silently return wrong results.
func TestMemoryStoreRejectsFilters(t *testing.T) {
	s := NewMemoryStore()
	if _, _, err := s.Search(SearchQuery{Bucket: "Element", Filter: `name eq "x"`}); !errors.Is(err, ErrFilterUnsupported) {
		t.Errorf("Search with a filter = %v, want ErrFilterUnsupported", err)
	}
}

// Compile-time check that both implementations satisfy Store.
var (
	_ Store = (*MemoryStore)(nil)
	_ Store = (*CouchbaseStore)(nil)
)
