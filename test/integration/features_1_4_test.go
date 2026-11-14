//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/arturoeanton/goscim/scim"
)

func TestCouchbaseAutoSecondaryIndexes(t *testing.T) {
	cbStore, ok := scim.DB.(*scim.CouchbaseStore)
	if !ok {
		t.Fatalf("expected *scim.CouchbaseStore in integration suite, got %T", scim.DB)
	}

	userIndexes, err := cbStore.ListBucketIndexNames("User")
	if err != nil {
		t.Fatalf("ListBucketIndexNames(User): %v", err)
	}
	haveUser := make(map[string]bool, len(userIndexes))
	for _, name := range userIndexes {
		haveUser[name] = true
	}
	for _, want := range []string{
		"idx_scim_User_tenant",
		"idx_scim_User_externalId",
		"idx_scim_User_uniq_userName",
		"idx_scim_User_arr_emails_lower",
	} {
		if !haveUser[want] {
			t.Errorf("Couchbase User bucket missing auto-provisioned GSI index %q (have: %v)", want, userIndexes)
		}
	}

	groupIndexes, err := cbStore.ListBucketIndexNames("Group")
	if err != nil {
		t.Fatalf("ListBucketIndexNames(Group): %v", err)
	}
	haveGroup := make(map[string]bool, len(groupIndexes))
	for _, name := range groupIndexes {
		haveGroup[name] = true
	}
	for _, want := range []string{
		"idx_scim_Group_tenant",
		"idx_scim_Group_externalId",
		"idx_scim_Group_arr_members_lower",
	} {
		if !haveGroup[want] {
			t.Errorf("Couchbase Group bucket missing auto-provisioned GSI index %q (have: %v)", want, groupIndexes)
		}
	}
}

func TestCouchbaseMultiTenantIsolationAndConformance(t *testing.T) {
	auth, err := scim.NewBearerAuthenticator([]scim.BearerTokenEntry{
		{Subject: "acme-client", Token: "cb-tok-acme", Roles: []string{"admin", "superadmin", "user", "role1", "role2"}, Tenant: "acme"},
		{Subject: "globex-client", Token: "cb-tok-globex", Roles: []string{"admin", "superadmin", "user", "role1", "role2"}, Tenant: "globex"},
		{Subject: "root-client", Token: "cb-tok-root", Roles: []string{"admin", "superadmin", "user", "role1", "role2"}},
	})
	if err != nil {
		t.Fatalf("NewBearerAuthenticator: %v", err)
	}
	router := newAuthenticatedIntegrationRouter(t, auth)

	sharedUserPayload := `{"schemas":["urn:ietf:params:scim:schemas:core:2.0:User"],"userName":"tenant.shared@contoso.com","active":true}`

	// Create identical userName in acme, globex, and default root namespace on real Couchbase.
	wAcme := doAuth(t, router, "cb-tok-acme", http.MethodPost, "/tenants/acme/scim/v2/Users", sharedUserPayload)
	if wAcme.Code != http.StatusCreated {
		t.Fatalf("acme create status = %d: %s", wAcme.Code, wAcme.Body.String())
	}
	var acmeDoc map[string]interface{}
	_ = json.Unmarshal(wAcme.Body.Bytes(), &acmeDoc)
	acmeID := acmeDoc["id"].(string)
	t.Cleanup(func() {
		doAuth(t, router, "cb-tok-acme", http.MethodDelete, "/tenants/acme/scim/v2/Users/"+acmeID, "")
	})

	wGlobex := doAuth(t, router, "cb-tok-globex", http.MethodPost, "/tenants/globex/scim/v2/Users", sharedUserPayload)
	if wGlobex.Code != http.StatusCreated {
		t.Fatalf("globex create status = %d: %s", wGlobex.Code, wGlobex.Body.String())
	}
	var globexDoc map[string]interface{}
	_ = json.Unmarshal(wGlobex.Body.Bytes(), &globexDoc)
	globexID := globexDoc["id"].(string)
	t.Cleanup(func() {
		doAuth(t, router, "cb-tok-globex", http.MethodDelete, "/tenants/globex/scim/v2/Users/"+globexID, "")
	})

	wRoot := doAuth(t, router, "cb-tok-root", http.MethodPost, "/scim/v2/Users", sharedUserPayload)
	if wRoot.Code != http.StatusCreated {
		t.Fatalf("root create status = %d: %s", wRoot.Code, wRoot.Body.String())
	}
	var rootDoc map[string]interface{}
	_ = json.Unmarshal(wRoot.Body.Bytes(), &rootDoc)
	rootID := rootDoc["id"].(string)
	t.Cleanup(func() {
		doAuth(t, router, "cb-tok-root", http.MethodDelete, "/scim/v2/Users/"+rootID, "")
	})

	// Verify N1QL OR filter cannot leak documents across tenants.
	orFilter := url.QueryEscape(`userName eq "tenant.shared@contoso.com" or active eq true`)
	wSearchAcme := doAuth(t, router, "cb-tok-acme", http.MethodGet, "/tenants/acme/scim/v2/Users?filter="+orFilter, "")
	if wSearchAcme.Code != http.StatusOK {
		t.Fatalf("acme N1QL search status = %d: %s", wSearchAcme.Code, wSearchAcme.Body.String())
	}
	var acmeList map[string]interface{}
	_ = json.Unmarshal(wSearchAcme.Body.Bytes(), &acmeList)
	if int(acmeList["totalResults"].(float64)) != 1 {
		t.Fatalf("acme N1QL OR search leaked records: totalResults=%v, want 1", acmeList["totalResults"])
	}
	if _, leaked := acmeList["Resources"].([]interface{})[0].(map[string]interface{})["_scim_tenant"]; leaked {
		t.Fatalf("internal _scim_tenant marker leaked in response")
	}

	// Run the full IdP Conformance Suite (RFC + Entra + Okta + Keycloak) against
	// tenant "acme" backed by real Couchbase KV + N1QL.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	report := scim.RunConformance(ctx, scim.ConformanceOptions{
		BaseURL: "http://cb.internal/tenants/acme/scim/v2",
		Token:   "cb-tok-acme",
		Profile: "all",
		HTTPClient: &http.Client{
			Transport: cbRouterTransport{router: router},
			Timeout:   15 * time.Second,
		},
	})
	if report.Failed != 0 {
		t.Fatalf("Couchbase multi-tenant conformance failed: %+v", report.Steps)
	}
}

type cbRouterTransport struct {
	router http.Handler
}

func (tr cbRouterTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	tr.router.ServeHTTP(rec, req)
	res := rec.Result()
	if res.Body == nil {
		res.Body = io.NopCloser(bytes.NewReader(nil))
	}
	return res, nil
}
