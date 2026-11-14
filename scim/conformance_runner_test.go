package scim

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	configembed "github.com/arturoeanton/goscim/config"
	"github.com/gin-gonic/gin"
)

func TestDemoSeedAndEmbeddedConfigFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prevAudit := AuditLogger()
	SetAuditLogger(nil)
	t.Cleanup(func() { SetAuditLogger(prevAudit) })

	// Move into an empty temp directory where ./config does not exist on disk
	// to prove single-binary standalone execution works via embedded config.
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
		SetFallbackConfigFS(nil)
	})

	SetFallbackConfigFS(configembed.FS)
	DB = NewMemoryStore()

	r := gin.New()
	if _, err := NewRouter("config", r, &AnonymousAuthenticator{Roles: testRoles}); err != nil {
		t.Fatalf("NewRouter with embedded fallback failed: %v", err)
	}

	seeded, err := SeedDemoResources("", false)
	if err != nil {
		t.Fatalf("SeedDemoResources: %v", err)
	}
	if seeded != 5 {
		t.Fatalf("seeded = %d, want 5 (3 Users + 2 nested Groups)", seeded)
	}

	// Verify Megan Bowen has both direct ("Backend") and indirect ("Engineering") groups.
	wUsers := do(t, r, http.MethodGet, `/scim/v2/Users?filter=userName+eq+"megan.bowen@contoso.com"`, "")
	requireStatus(t, wUsers, http.StatusOK)
	res := decode(t, wUsers)["Resources"].([]interface{})
	if len(res) != 1 {
		t.Fatalf("expected Megan Bowen in demo dataset")
	}
	groups := res[0].(map[string]interface{})["groups"].([]interface{})
	if len(groups) != 2 {
		t.Fatalf("Megan groups len = %d, want 2 (direct Backend + indirect Engineering)", len(groups))
	}
}

func TestConformanceRunnerAllProfilesAndIndexSpecs(t *testing.T) {
	r, store := newTestServer(t)

	report := RunConformance(context.Background(), ConformanceOptions{
		BaseURL: "http://conformance.internal/scim/v2",
		Profile: "all",
		HTTPClient: &http.Client{
			Transport: handlerTransport{h: r},
			Timeout:   10 * time.Second,
		},
	})
	if report.Failed != 0 || report.Passed < 9 {
		t.Fatalf("RunConformance failed: passed=%d failed=%d steps=%+v", report.Passed, report.Failed, report.Steps)
	}

	// Verify cleanup removed all temporary conformance Users and Groups.
	uTotal, _, _ := store.Search(SearchQuery{Bucket: "User", Limit: 10})
	gTotal, _, _ := store.Search(SearchQuery{Bucket: "Group", Limit: 10})
	if uTotal != 0 || gTotal != 0 {
		t.Fatalf("conformance runner left temporary records: users=%d groups=%d", uTotal, gTotal)
	}

	// Verify schema-driven secondary GSI index generator produces expected specs.
	userSpecs := BuildSchemaIndexSpecs(Resources["/Users"])
	foundNames := make(map[string]bool, len(userSpecs))
	for _, sp := range userSpecs {
		foundNames[sp.Name] = true
	}
	for _, want := range []string{
		"idx_scim_User_tenant",
		"idx_scim_User_externalId",
		"idx_scim_User_uniq_userName",
		"idx_scim_User_arr_emails_lower",
	} {
		if !foundNames[want] {
			t.Errorf("missing generated Couchbase index spec %q in %v", want, foundNames)
		}
	}
}
