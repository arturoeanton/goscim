package scim

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/arturoeanton/goscim/scim/parser"
)

// MemoryStore is an in-memory Store used by the test suite. Documents are
// round-tripped through JSON on the way in and out, exactly as the Couchbase
// driver does: callers cannot mutate stored state by holding on to a returned
// map, and Go structs handed to Upsert come back as map[string]interface{}
// with float64 numbers, so the fake cannot mask type-assertion bugs that the
// real store would expose.
type MemoryStore struct {
	mu      sync.RWMutex
	buckets map[string]map[string]map[string]interface{}
	unique  map[string]map[string]string
}

// NewMemoryStore returns an empty in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		buckets: make(map[string]map[string]map[string]interface{}),
		unique:  make(map[string]map[string]string),
	}
}

// EnsureBucket implements Store.
func (s *MemoryStore) EnsureBucket(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.buckets[name]; !ok {
		s.buckets[name] = make(map[string]map[string]interface{})
	}
	if _, ok := s.unique[name]; !ok {
		s.unique[name] = make(map[string]string)
	}
	return nil
}

// Ping implements Store.
func (s *MemoryStore) Ping(context.Context) error { return nil }

// Close implements Store.
func (s *MemoryStore) Close() error { return nil }

// Get implements Store.
func (s *MemoryStore) Get(bucket, id string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	doc, ok := s.buckets[bucket][id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneDoc(doc)
}

// Upsert implements Store.
func (s *MemoryStore) Upsert(bucket, id string, doc map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureLocked(bucket)
	stored, err := cloneDoc(doc)
	if err != nil {
		return err
	}
	s.buckets[bucket][id] = stored
	return nil
}

// Replace implements Store.
func (s *MemoryStore) Replace(bucket, id string, doc map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.buckets[bucket][id]; !ok {
		return ErrNotFound
	}
	stored, err := cloneDoc(doc)
	if err != nil {
		return err
	}
	s.buckets[bucket][id] = stored
	return nil
}

// Remove implements Store.
func (s *MemoryStore) Remove(bucket, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.buckets[bucket][id]; !ok {
		return ErrNotFound
	}
	delete(s.buckets[bucket], id)
	s.releaseOwnerLocksLocked(bucket, id)
	return nil
}

func memoryOwnerToken(bucket, id string) string {
	return bucket + "\x00" + id
}

func splitMemoryOwnerToken(token string) (bucket, id string) {
	idx := strings.IndexByte(token, 0)
	if idx < 0 {
		return "", token
	}
	return token[:idx], token[idx+1:]
}

func memoryUniqueScope(bucket string, claim UniqueClaim) string {
	if claim.Global {
		_, phys := SplitScopedBucket(bucket)
		return "__global__:" + phys
	}
	return bucket
}

// CreateUnique implements Store.
func (s *MemoryStore) CreateUnique(bucket, id string, doc map[string]interface{}, claims []UniqueClaim) error {
	stored, err := cloneDoc(doc)
	if err != nil {
		return err
	}
	owner := memoryOwnerToken(bucket, id)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureLocked(bucket)
	if _, exists := s.buckets[bucket][id]; exists {
		return &UniqueConflictError{Attribute: "id"}
	}
	for _, claim := range claims {
		if err := s.checkClaimAvailableLocked(bucket, id, claim); err != nil {
			return err
		}
	}
	s.buckets[bucket][id] = stored
	for _, claim := range claims {
		scope := memoryUniqueScope(bucket, claim)
		s.ensureLocked(scope)
		s.unique[scope][claim.LockID] = owner
	}
	return nil
}

// ReplaceUnique implements Store.
func (s *MemoryStore) ReplaceUnique(bucket, id string, doc map[string]interface{}, newClaims []UniqueClaim, extractClaims func(map[string]interface{}) []UniqueClaim) error {
	stored, err := cloneDoc(doc)
	if err != nil {
		return err
	}
	owner := memoryOwnerToken(bucket, id)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureLocked(bucket)
	current, exists := s.buckets[bucket][id]
	if !exists {
		return ErrNotFound
	}
	for _, claim := range newClaims {
		if err := s.checkClaimAvailableLocked(bucket, id, claim); err != nil {
			return err
		}
	}
	var oldClaims []UniqueClaim
	if extractClaims != nil {
		oldClaims = extractClaims(current)
	}
	s.buckets[bucket][id] = stored
	for _, old := range oldClaims {
		scope := memoryUniqueScope(bucket, old)
		if s.unique[scope][old.LockID] == owner {
			delete(s.unique[scope], old.LockID)
		}
	}
	for _, claim := range newClaims {
		scope := memoryUniqueScope(bucket, claim)
		s.ensureLocked(scope)
		s.unique[scope][claim.LockID] = owner
	}
	return nil
}

// RemoveUnique implements Store.
func (s *MemoryStore) RemoveUnique(bucket, id string, claims []UniqueClaim) error {
	owner := memoryOwnerToken(bucket, id)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.buckets[bucket][id]; !ok {
		return ErrNotFound
	}
	delete(s.buckets[bucket], id)
	for _, claim := range claims {
		scope := memoryUniqueScope(bucket, claim)
		if s.unique[scope][claim.LockID] == owner {
			delete(s.unique[scope], claim.LockID)
		}
	}
	s.releaseOwnerLocksLocked(bucket, id)
	return nil
}

func (s *MemoryStore) ensureLocked(bucket string) {
	if _, ok := s.buckets[bucket]; !ok {
		s.buckets[bucket] = make(map[string]map[string]interface{})
	}
	if _, ok := s.unique[bucket]; !ok {
		s.unique[bucket] = make(map[string]string)
	}
}

func (s *MemoryStore) releaseOwnerLocksLocked(bucket, ownerID string) {
	owner := memoryOwnerToken(bucket, ownerID)
	_, phys := SplitScopedBucket(bucket)
	for _, scope := range []string{bucket, "__global__:" + phys} {
		locks := s.unique[scope]
		for lockID, heldBy := range locks {
			if heldBy == owner {
				delete(locks, lockID)
			}
		}
	}
}

func (s *MemoryStore) checkClaimAvailableLocked(bucket, selfID string, claim UniqueClaim) error {
	selfOwner := memoryOwnerToken(bucket, selfID)
	scope := memoryUniqueScope(bucket, claim)
	s.ensureLocked(scope)
	if owner, taken := s.unique[scope][claim.LockID]; taken && owner != selfOwner {
		ownerBucket, ownerID := splitMemoryOwnerToken(owner)
		if ownerBucket == "" {
			ownerBucket = bucket
		}
		if _, alive := s.buckets[ownerBucket][ownerID]; alive {
			return &UniqueConflictError{Attribute: claim.Attribute}
		}
		delete(s.unique[scope], claim.LockID)
	}
	wanted, err := normalizeValue(claim.RawValue)
	if err != nil {
		return err
	}
	path := opPathTopathArray(claim.Attribute)
	_, targetPhys := SplitScopedBucket(bucket)
	for bName, docs := range s.buckets {
		if strings.HasPrefix(bName, "__global__:") {
			continue
		}
		if claim.Global {
			_, candidatePhys := SplitScopedBucket(bName)
			if candidatePhys != targetPhys {
				continue
			}
		} else if bName != bucket {
			continue
		}
		for existingID, doc := range docs {
			if bName == bucket && existingID == selfID {
				continue
			}
			existingVal := resolvePath(doc, path)
			if existingVal == nil {
				continue
			}
			if !claim.CaseExact {
				sA, okA := existingVal.(string)
				sB, okB := wanted.(string)
				if okA && okB && strings.EqualFold(sA, sB) {
					return &UniqueConflictError{Attribute: claim.Attribute}
				}
			}
			if existingVal == wanted {
				return &UniqueConflictError{Attribute: claim.Attribute}
			}
		}
	}
	return nil
}

// Search implements Store, evaluating SCIM filters in memory via parser.CompileMatcherWithOptions.
func (s *MemoryStore) Search(q SearchQuery) (int, []map[string]interface{}, error) {
	var matcher func(map[string]interface{}) bool
	if q.Filter != "" {
		var err error
		matcher, err = parser.CompileMatcherWithOptions(q.Filter, parser.EvalOptions{
			Resolve:     q.ResolveName,
			IsCaseExact: q.IsCaseExact,
		})
		if err != nil {
			return 0, nil, fmt.Errorf("%w: %s", ErrInvalidFilter, err)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := make([]map[string]interface{}, 0, len(s.buckets[q.Bucket]))
	for _, doc := range s.buckets[q.Bucket] {
		if matcher != nil && !matcher(doc) {
			continue
		}
		copied, err := cloneDoc(doc)
		if err != nil {
			return 0, nil, err
		}
		all = append(all, copied)
	}

	sortBy := q.SortBy
	if sortBy == "" {
		sortBy = "id"
	}
	path := opPathTopathArray(sortBy)
	sort.SliceStable(all, func(i, j int) bool {
		less := compareValues(resolvePath(all[i], path), resolvePath(all[j], path)) < 0
		if q.SortDescending {
			return !less
		}
		return less
	})

	total := len(all)
	if q.Offset >= total {
		return total, []map[string]interface{}{}, nil
	}
	page := all[q.Offset:]
	if q.Limit >= 0 && q.Limit < len(page) {
		page = page[:q.Limit]
	}
	return total, page, nil
}

// FindIDByAttribute implements Store.
//
// The value is normalised through JSON before comparing, because that is what
// storing it did: an int64 handed in by the validator is a float64 once it has
// been through the store, and comparing the two as interfaces would never
// match even though N1QL would compare them numerically.
func (s *MemoryStore) FindIDByAttribute(bucket, attributePath string, value interface{}) (string, error) {
	wanted, err := normalizeValue(value)
	if err != nil {
		return "", err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	path := opPathTopathArray(attributePath)
	for id, doc := range s.buckets[bucket] {
		if resolvePath(doc, path) == wanted {
			return id, nil
		}
	}
	return "", nil
}

// normalizeValue round-trips a scalar through JSON, matching how the store
// holds it.
func normalizeValue(value interface{}) (interface{}, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var out interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// resolvePath walks a dotted attribute path through nested objects.
func resolvePath(doc map[string]interface{}, path []string) interface{} {
	var current interface{} = doc
	for _, field := range path {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current = m[field]
	}
	return current
}

// compareValues orders two decoded JSON values, grouping unlike types by a
// stable type rank so that sorting never panics on heterogeneous documents.
func compareValues(a, b interface{}) int {
	ra, rb := typeRank(a), typeRank(b)
	if ra != rb {
		return ra - rb
	}
	switch va := a.(type) {
	case string:
		vb := b.(string)
		switch {
		case va < vb:
			return -1
		case va > vb:
			return 1
		}
	case float64:
		vb := b.(float64)
		switch {
		case va < vb:
			return -1
		case va > vb:
			return 1
		}
	case bool:
		vb := b.(bool)
		switch {
		case !va && vb:
			return -1
		case va && !vb:
			return 1
		}
	}
	return 0
}

func typeRank(v interface{}) int {
	switch v.(type) {
	case nil:
		return 0
	case bool:
		return 1
	case float64:
		return 2
	case string:
		return 3
	default:
		return 4
	}
}

// cloneDoc round-trips a document through JSON, mirroring what the Couchbase
// driver does on write and read.
func cloneDoc(in map[string]interface{}) (map[string]interface{}, error) {
	if in == nil {
		return nil, nil
	}
	raw, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}
