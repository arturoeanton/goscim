package scim

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arturoeanton/goscim/scim/parser"
	_ "modernc.org/sqlite"
)

const globalTenantSentinel = "__global__"

// SQLiteStore is a single-file persistent Store backed by pure-Go SQLite
// (modernc.org/sqlite, CGO_ENABLED=0). It supports ACID unique locks,
// multi-tenancy, and full SCIM filter evaluation without requiring an external
// database server.
type SQLiteStore struct {
	db   *sql.DB
	path string
}

// NewSQLiteStore opens (or creates) a SQLite database at dsnPath (use ":memory:"
// for an ephemeral database) and initialises the SCIM schema.
func NewSQLiteStore(dsnPath string) (*SQLiteStore, error) {
	dsnPath = strings.TrimSpace(dsnPath)
	if dsnPath == "" {
		dsnPath = "goscim.db"
	}
	if dsnPath != ":memory:" && !strings.HasPrefix(dsnPath, "file:") {
		if dir := filepath.Dir(dsnPath); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("creating sqlite directory %q: %w", dir, err)
			}
		}
	}

	db, err := sql.Open("sqlite", dsnPath)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite %q: %w", dsnPath, err)
	}
	// Serialise writers cleanly inside SQLite rather than surfacing SQLITE_BUSY
	// under concurrent provisioning bursts.
	db.SetMaxOpenConns(1)

	pragmas := []string{
		`PRAGMA busy_timeout = 5000`,
		`PRAGMA journal_mode = WAL`,
		`PRAGMA synchronous = NORMAL`,
		`PRAGMA foreign_keys = ON`,
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("applying sqlite pragma %q: %w", p, err)
		}
	}

	ddl := []string{
		`CREATE TABLE IF NOT EXISTS scim_buckets (
			name TEXT PRIMARY KEY
		)`,
		`CREATE TABLE IF NOT EXISTS scim_resources (
			bucket TEXT NOT NULL,
			tenant TEXT NOT NULL,
			id     TEXT NOT NULL,
			doc    TEXT NOT NULL,
			PRIMARY KEY (bucket, tenant, id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_scim_resources_bucket_tenant
			ON scim_resources (bucket, tenant)`,
		`CREATE TABLE IF NOT EXISTS scim_unique (
			bucket    TEXT NOT NULL,
			tenant    TEXT NOT NULL,
			lock_id   TEXT NOT NULL,
			owner_key TEXT NOT NULL,
			attribute TEXT NOT NULL,
			PRIMARY KEY (bucket, tenant, lock_id)
		)`,
	}
	for _, stmt := range ddl {
		if _, err := db.Exec(stmt); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("initialising sqlite schema: %w", err)
		}
	}

	return &SQLiteStore{db: db, path: dsnPath}, nil
}

// Path returns the configured SQLite database path.
func (s *SQLiteStore) Path() string { return s.path }

// EnsureBucket implements Store.
func (s *SQLiteStore) EnsureBucket(name string) error {
	_, phys := SplitScopedBucket(name)
	_, err := s.db.Exec(`INSERT OR IGNORE INTO scim_buckets (name) VALUES (?)`, phys)
	return err
}

// Ping implements Store.
func (s *SQLiteStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close implements Store.
func (s *SQLiteStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func sqliteOwnerKey(tenant, id string) string {
	return tenant + "\x00" + id
}

func sqliteClaimTenant(tenant string, claim UniqueClaim) string {
	if claim.Global {
		return globalTenantSentinel
	}
	return tenant
}

// Get implements Store.
func (s *SQLiteStore) Get(bucket, id string) (map[string]interface{}, error) {
	tenant, phys := SplitScopedBucket(bucket)
	var raw string
	err := s.db.QueryRow(
		`SELECT doc FROM scim_resources WHERE bucket = ? AND tenant = ? AND id = ?`,
		phys, tenant, id,
	).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// Upsert implements Store.
func (s *SQLiteStore) Upsert(bucket, id string, doc map[string]interface{}) error {
	tenant, phys := SplitScopedBucket(bucket)
	raw, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO scim_resources (bucket, tenant, id, doc) VALUES (?, ?, ?, ?)
		 ON CONFLICT(bucket, tenant, id) DO UPDATE SET doc = excluded.doc`,
		phys, tenant, id, string(raw),
	)
	return err
}

// Replace implements Store.
func (s *SQLiteStore) Replace(bucket, id string, doc map[string]interface{}) error {
	tenant, phys := SplitScopedBucket(bucket)
	raw, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(
		`UPDATE scim_resources SET doc = ? WHERE bucket = ? AND tenant = ? AND id = ?`,
		string(raw), phys, tenant, id,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Remove implements Store.
func (s *SQLiteStore) Remove(bucket, id string) error {
	tenant, phys := SplitScopedBucket(bucket)
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(
		`DELETE FROM scim_resources WHERE bucket = ? AND tenant = ? AND id = ?`,
		phys, tenant, id,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	_, err = tx.Exec(
		`DELETE FROM scim_unique WHERE bucket = ? AND owner_key = ?`,
		phys, sqliteOwnerKey(tenant, id),
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// CreateUnique implements Store.
func (s *SQLiteStore) CreateUnique(bucket, id string, doc map[string]interface{}, claims []UniqueClaim) error {
	tenant, phys := SplitScopedBucket(bucket)
	raw, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	owner := sqliteOwnerKey(tenant, id)

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var existingID string
	err = tx.QueryRow(
		`SELECT id FROM scim_resources WHERE bucket = ? AND tenant = ? AND id = ?`,
		phys, tenant, id,
	).Scan(&existingID)
	if err == nil {
		return &UniqueConflictError{Attribute: "id"}
	}
	if err != sql.ErrNoRows {
		return err
	}

	for _, claim := range claims {
		claimTenant := sqliteClaimTenant(tenant, claim)
		var lockOwner string
		err := tx.QueryRow(
			`SELECT owner_key FROM scim_unique WHERE bucket = ? AND tenant = ? AND lock_id = ?`,
			phys, claimTenant, claim.LockID,
		).Scan(&lockOwner)
		if err == nil && lockOwner != owner {
			return &UniqueConflictError{Attribute: claim.Attribute}
		}
		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}

	if _, err := tx.Exec(
		`INSERT INTO scim_resources (bucket, tenant, id, doc) VALUES (?, ?, ?, ?)`,
		phys, tenant, id, string(raw),
	); err != nil {
		return err
	}

	for _, claim := range claims {
		claimTenant := sqliteClaimTenant(tenant, claim)
		if _, err := tx.Exec(
			`INSERT OR REPLACE INTO scim_unique (bucket, tenant, lock_id, owner_key, attribute) VALUES (?, ?, ?, ?, ?)`,
			phys, claimTenant, claim.LockID, owner, claim.Attribute,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ReplaceUnique implements Store.
func (s *SQLiteStore) ReplaceUnique(bucket, id string, doc map[string]interface{}, newClaims []UniqueClaim, extractClaims func(map[string]interface{}) []UniqueClaim) error {
	tenant, phys := SplitScopedBucket(bucket)
	raw, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	owner := sqliteOwnerKey(tenant, id)

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var currentRaw string
	err = tx.QueryRow(
		`SELECT doc FROM scim_resources WHERE bucket = ? AND tenant = ? AND id = ?`,
		phys, tenant, id,
	).Scan(&currentRaw)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	for _, claim := range newClaims {
		claimTenant := sqliteClaimTenant(tenant, claim)
		var lockOwner string
		err := tx.QueryRow(
			`SELECT owner_key FROM scim_unique WHERE bucket = ? AND tenant = ? AND lock_id = ?`,
			phys, claimTenant, claim.LockID,
		).Scan(&lockOwner)
		if err == nil && lockOwner != owner {
			return &UniqueConflictError{Attribute: claim.Attribute}
		}
		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}

	var currentDoc map[string]interface{}
	if err := json.Unmarshal([]byte(currentRaw), &currentDoc); err != nil {
		return err
	}
	var oldClaims []UniqueClaim
	if extractClaims != nil {
		oldClaims = extractClaims(currentDoc)
	}

	if _, err := tx.Exec(
		`UPDATE scim_resources SET doc = ? WHERE bucket = ? AND tenant = ? AND id = ?`,
		string(raw), phys, tenant, id,
	); err != nil {
		return err
	}

	for _, old := range oldClaims {
		claimTenant := sqliteClaimTenant(tenant, old)
		if _, err := tx.Exec(
			`DELETE FROM scim_unique WHERE bucket = ? AND tenant = ? AND lock_id = ? AND owner_key = ?`,
			phys, claimTenant, old.LockID, owner,
		); err != nil {
			return err
		}
	}
	for _, claim := range newClaims {
		claimTenant := sqliteClaimTenant(tenant, claim)
		if _, err := tx.Exec(
			`INSERT OR REPLACE INTO scim_unique (bucket, tenant, lock_id, owner_key, attribute) VALUES (?, ?, ?, ?, ?)`,
			phys, claimTenant, claim.LockID, owner, claim.Attribute,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RemoveUnique implements Store.
func (s *SQLiteStore) RemoveUnique(bucket, id string, claims []UniqueClaim) error {
	return s.Remove(bucket, id)
}

// Search implements Store.
func (s *SQLiteStore) Search(q SearchQuery) (int, []map[string]interface{}, error) {
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

	tenant, phys := SplitScopedBucket(q.Bucket)
	rows, err := s.db.Query(
		`SELECT doc FROM scim_resources WHERE bucket = ? AND tenant = ?`,
		phys, tenant,
	)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	all := make([]map[string]interface{}, 0)
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return 0, nil, err
		}
		var doc map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &doc); err != nil {
			return 0, nil, err
		}
		if matcher != nil && !matcher(doc) {
			continue
		}
		all = append(all, doc)
	}
	if err := rows.Err(); err != nil {
		return 0, nil, err
	}

	total := len(all)
	sortKey := q.SortBy
	if sortKey == "" {
		sortKey = "id"
	}
	keyPath := opPathTopathArray(sortKey)
	sort.Slice(all, func(i, j int) bool {
		cmp := compareValues(resolvePath(all[i], keyPath), resolvePath(all[j], keyPath))
		if q.SortDescending {
			return cmp > 0
		}
		return cmp < 0
	})

	if q.Offset >= len(all) {
		return total, []map[string]interface{}{}, nil
	}
	end := len(all)
	if q.Limit > 0 && q.Offset+q.Limit < end {
		end = q.Offset + q.Limit
	}
	return total, all[q.Offset:end], nil
}

// FindIDByAttribute implements Store.
func (s *SQLiteStore) FindIDByAttribute(bucket, attributePath string, value interface{}) (string, error) {
	wanted, err := normalizeValue(value)
	if err != nil {
		return "", err
	}
	tenant, phys := SplitScopedBucket(bucket)
	rows, err := s.db.Query(
		`SELECT id, doc FROM scim_resources WHERE bucket = ? AND tenant = ?`,
		phys, tenant,
	)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	path := opPathTopathArray(attributePath)
	for rows.Next() {
		var id, raw string
		if err := rows.Scan(&id, &raw); err != nil {
			return "", err
		}
		var doc map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &doc); err != nil {
			continue
		}
		if resolvePath(doc, path) == wanted {
			return id, nil
		}
	}
	return "", rows.Err()
}
