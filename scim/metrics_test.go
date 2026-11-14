package scim

import (
	"net/http"
	"strings"
	"testing"
)

func TestPrometheusMetricsEndpointAndTokenGuard(t *testing.T) {
	DefaultMetrics.Reset()
	DefaultMetrics.SetBuildVersion("1.4.0-test")

	auth, err := NewBearerAuthenticator([]BearerTokenEntry{{
		Subject: "metrics-tester",
		Token:   "scim-tok",
		Roles:   testRoles,
	}})
	if err != nil {
		t.Fatalf("NewBearerAuthenticator: %v", err)
	}
	r, _ := newTestServerAs(t, auth)

	// Trigger a 401 auth failure, a 201 create, and a 409 uniqueness conflict.
	wUnauth := do(t, r, http.MethodGet, "/scim/v2/Users", "")
	requireStatus(t, wUnauth, http.StatusUnauthorized)

	createReq := newRequest(http.MethodPost, "/scim/v2/Users",
		`{"schemas":["urn:ietf:params:scim:schemas:core:2.0:User"],"userName":"metrics@example.com"}`)
	createReq.Header.Set("Authorization", "Bearer scim-tok")
	requireStatus(t, serve(r, createReq), http.StatusCreated)

	dupReq := newRequest(http.MethodPost, "/scim/v2/Users",
		`{"schemas":["urn:ietf:params:scim:schemas:core:2.0:User"],"userName":"METRICS@example.com"}`)
	dupReq.Header.Set("Authorization", "Bearer scim-tok")
	requireStatus(t, serve(r, dupReq), http.StatusConflict)

	wMetrics := do(t, r, http.MethodGet, "/metrics", "")
	requireStatus(t, wMetrics, http.StatusOK)
	body := wMetrics.Body.String()

	expectedSnippets := []string{
		`goscim_build_info{version="1.4.0-test"} 1`,
		`goscim_store_up 1`,
		`goscim_auth_failures_total 1`,
		`goscim_http_requests_total{action="create",method="POST",resource_type="User",status="201"} 1`,
		`goscim_uniqueness_conflicts_total{resource_type="User"} 1`,
		`goscim_http_request_duration_seconds_count{action="create"} 2`,
	}
	for _, want := range expectedSnippets {
		if !strings.Contains(body, want) {
			t.Errorf("metrics output missing %q\nfull output:\n%s", want, body)
		}
	}

	// Enable SCIM_METRICS_TOKEN and verify 401 without token and 200 with token.
	t.Setenv("SCIM_METRICS_TOKEN", "prom-scrape-secret")
	wBlocked := do(t, r, http.MethodGet, "/metrics", "")
	requireStatus(t, wBlocked, http.StatusUnauthorized)

	promReq := newRequest(http.MethodGet, "/metrics", "")
	promReq.Header.Set("Authorization", "Bearer prom-scrape-secret")
	requireStatus(t, serve(r, promReq), http.StatusOK)
}
