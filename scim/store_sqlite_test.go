package scim

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestSQLiteStorePersistenceAndUniqueness(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "goscim-test.db")
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	if err := store.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	if err := store.EnsureBucket("User"); err != nil {
		t.Fatalf("EnsureBucket: %v", err)
	}

	u1 := map[string]interface{}{
		"id":       "u-1",
		"userName": "Alice@Example.com",
		"active":   true,
	}
	claims1 := []UniqueClaim{{Attribute: "userName", Key: "alice@example.com"}}
	if err := store.CreateUnique("User", "u-1", u1, claims1); err != nil {
		t.Fatalf("CreateUnique u-1: %v", err)
	}

	// Duplicate in same tenant fails with UniqueConflictError.
	if err := store.CreateUnique("User", "u-2", map[string]interface{}{"id": "u-2"}, claims1); err == nil {
		t.Fatalf("expected uniqueness conflict in same tenant")
	}

	// Same userName in another tenant ("t:acme:User") succeeds for server uniqueness.
	if err := store.CreateUnique("t:acme:User", "u-acme-1", map[string]interface{}{
		"id":       "u-acme-1",
		"userName": "Alice@Example.com",
	}, claims1); err != nil {
		t.Fatalf("CreateUnique in tenant acme failed: %v", err)
	}

	// Global uniqueness claim blocks across tenants.
	globalClaim := []UniqueClaim{{Attribute: "employeeGlobalId", Key: "g-900", Global: true}}
	if err := store.CreateUnique("t:acme:User", "u-g1", map[string]interface{}{"id": "u-g1"}, globalClaim); err != nil {
		t.Fatalf("CreateUnique global claim: %v", err)
	}
	if err := store.CreateUnique("t:globex:User", "u-g2", map[string]interface{}{"id": "u-g2"}, globalClaim); err == nil {
		t.Fatalf("expected global uniqueness conflict across tenants")
	}

	// Close and reopen the database file to verify durable persistence.
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	reopened, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Reopen NewSQLiteStore: %v", err)
	}
	defer reopened.Close()

	got, err := reopened.Get("User", "u-1")
	if err != nil {
		t.Fatalf("Get after reopen: %v", err)
	}
	if got["userName"] != "Alice@Example.com" {
		t.Fatalf("userName after reopen = %v, want Alice@Example.com", got["userName"])
	}

	// Default tenant search must not return tenant acme records.
	totalDefault, _, err := reopened.Search(SearchQuery{Bucket: "User", Limit: 50})
	if err != nil {
		t.Fatalf("Search default: %v", err)
	}
	if totalDefault != 1 {
		t.Fatalf("totalDefault = %d, want 1", totalDefault)
	}
	totalAcme, _, err := reopened.Search(SearchQuery{Bucket: "t:acme:User", Limit: 50})
	if err != nil {
		t.Fatalf("Search acme: %v", err)
	}
	if totalAcme != 2 {
		t.Fatalf("totalAcme = %d, want 2", totalAcme)
	}
}

func TestSQLiteFullRouterConformanceSuite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prevAudit := AuditLogger()
	SetAuditLogger(nil)
	t.Cleanup(func() { SetAuditLogger(prevAudit) })

	dbPath := filepath.Join(t.TempDir(), "conformance.db")
	t.Setenv("SCIM_STORE", "sqlite")
	t.Setenv("SCIM_SQLITE_PATH", dbPath)

	backend, err := InitStoreFromEnv()
	if err != nil {
		t.Fatalf("InitStoreFromEnv: %v", err)
	}
	if backend != StoreBackendSQLite {
		t.Fatalf("backend = %q, want sqlite", backend)
	}
	t.Cleanup(func() { _ = DB.Close() })

	auth, err := NewBearerAuthenticator([]BearerTokenEntry{{
		Subject: "sqlite-tester",
		Token:   "sqlite-token",
		Roles:   testRoles,
	}})
	if err != nil {
		t.Fatalf("NewBearerAuthenticator: %v", err)
	}
	engine := gin.New()
	if _, err := NewRouter("../config", engine, auth); err != nil {
		t.Fatalf("NewRouter: %v", err)
	}

	report := RunConformance(context.Background(), ConformanceOptions{
		BaseURL: "http://sqlite.local/scim/v2",
		Token:   "sqlite-token",
		Profile: "all",
		HTTPClient: &http.Client{
			Transport: handlerTransport{h: engine},
			Timeout:   10 * time.Second,
		},
	})
	if report.Failed != 0 || report.Passed == 0 {
		t.Fatalf("SQLite conformance failed: passed=%d failed=%d steps=%+v", report.Passed, report.Failed, report.Steps)
	}
}

type handlerTransport struct {
	h http.Handler
}

func (tr handlerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	tr.h.ServeHTTP(rec, req)
	res := rec.Result()
	if res.Body == nil {
		res.Body = io.NopCloser(bytes.NewReader(nil))
	}
	return res, nil
}
