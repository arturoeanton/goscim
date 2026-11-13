package scim

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"testing"
)

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestRequestIDGenerationAndSanitisation(t *testing.T) {
	r, _ := newTestServer(t)

	t.Run("generates a UUID when X-Request-Id is omitted", func(t *testing.T) {
		w := do(t, r, http.MethodGet, "/healthz", "")
		requireStatus(t, w, http.StatusOK)
		got := w.Header().Get(RequestIDHeader)
		if len(got) != 36 {
			t.Fatalf("X-Request-Id = %q, want 36-char UUID", got)
		}
	})

	t.Run("echoes a safe caller-supplied X-Request-Id", func(t *testing.T) {
		req := newRequest(http.MethodGet, elementsPath, "")
		req.Header.Set(RequestIDHeader, "trace-2026_abc:01.99")
		w := serve(r, req)
		requireStatus(t, w, http.StatusOK)
		if got := w.Header().Get(RequestIDHeader); got != "trace-2026_abc:01.99" {
			t.Errorf("X-Request-Id = %q, want trace-2026_abc:01.99", got)
		}
	})

	t.Run("replaces unsafe or oversized X-Request-Id values", func(t *testing.T) {
		unsafe := []string{
			"bad id with spaces",
			"inject\r\nheader: x",
			"semi;colon",
			`"quoted"`,
			strings.Repeat("a", maxRequestIDLen+1),
		}
		for _, bad := range unsafe {
			req := newRequest(http.MethodGet, elementsPath, "")
			req.Header.Set(RequestIDHeader, bad)
			w := serve(r, req)
			requireStatus(t, w, http.StatusOK)
			got := w.Header().Get(RequestIDHeader)
			if got == "" || got == bad {
				t.Errorf("unsafe X-Request-Id %q was not replaced: got %q", bad, got)
			}
		}
	})
}

func TestAuditLogCapturesPrincipalOperationsAndNeverLeaksSecrets(t *testing.T) {
	auth := &BasicAuthenticator{users: map[string]basicUser{
		"provisioner": {password: "top-secret-basic-pw", roles: []string{"admin", "role1"}},
	}}
	r, _ := newTestServerAs(t, auth)

	out := &lockedBuffer{}
	SetAuditLogger(slog.New(slog.NewJSONHandler(out, nil)))

	// Health probes must not emit audit events.
	requireStatus(t, do(t, r, http.MethodGet, "/healthz", ""), http.StatusOK)
	if out.String() != "" {
		t.Fatalf("/healthz emitted audit log: %s", out.String())
	}

	// 1. Unauthenticated request -> 401 denied
	req401 := newRequest(http.MethodGet, usersPath, "")
	req401.Header.Set(RequestIDHeader, "req-unauth-1")
	requireStatus(t, serve(r, req401), http.StatusUnauthorized)

	// 2. Create User with sensitive password -> 201 success
	createBody := `{"schemas":["` + schemaUser + `"],"userName":"audit.user","password":"SuperSecretUserPass!99"}`
	reqCreate := newRequest(http.MethodPost, usersPath, createBody)
	reqCreate.SetBasicAuth("provisioner", "top-secret-basic-pw")
	reqCreate.Header.Set(RequestIDHeader, "req-create-1")
	wCreate := serve(r, reqCreate)
	requireStatus(t, wCreate, http.StatusCreated)
	userID := decode(t, wCreate)["id"].(string)

	// 3. Duplicate create -> 409 conflict with scim_type=uniqueness
	reqDup := newRequest(http.MethodPost, usersPath, createBody)
	reqDup.SetBasicAuth("provisioner", "top-secret-basic-pw")
	reqDup.Header.Set(RequestIDHeader, "req-dup-1")
	requireStatus(t, serve(r, reqDup), http.StatusConflict)

	// 4. Patch user -> 200 success
	reqPatch := newRequest(http.MethodPatch, usersPath+"/"+userID,
		patchBody(map[string]interface{}{"op": "replace", "path": "displayName", "value": "Audited"}))
	reqPatch.SetBasicAuth("provisioner", "top-secret-basic-pw")
	reqPatch.Header.Set(RequestIDHeader, "req-patch-1")
	requireStatus(t, serve(r, reqPatch), http.StatusOK)

	// 5. Search with sensitive query parameter -> query value must NOT appear in logs
	reqSearch := newRequest(http.MethodGet, usersPath+"?filter=userName+eq+%22SensitiveFilterPII%22", "")
	reqSearch.SetBasicAuth("provisioner", "top-secret-basic-pw")
	reqSearch.Header.Set(RequestIDHeader, "req-search-1")
	serve(r, reqSearch)

	// 6. Delete user -> 204 success
	reqDel := newRequest(http.MethodDelete, usersPath+"/"+userID, "")
	reqDel.SetBasicAuth("provisioner", "top-secret-basic-pw")
	reqDel.Header.Set(RequestIDHeader, "req-del-1")
	requireStatus(t, serve(r, reqDel), http.StatusNoContent)

	rawLogs := out.String()
	for _, secret := range []string{"SuperSecretUserPass!99", "top-secret-basic-pw", "SensitiveFilterPII"} {
		if strings.Contains(rawLogs, secret) {
			t.Fatalf("audit log leaked secret %q:\n%s", secret, rawLogs)
		}
	}

	lines := strings.Split(strings.TrimSpace(rawLogs), "\n")
	byReqID := make(map[string]map[string]interface{}, len(lines))
	for _, line := range lines {
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("invalid JSON audit line %q: %v", line, err)
		}
		if entry["msg"] == "scim.audit" {
			if id, ok := entry["request_id"].(string); ok {
				byReqID[id] = entry
			}
		}
	}

	unauthEntry := byReqID["req-unauth-1"]
	if unauthEntry["outcome"] != "denied" || unauthEntry["status"] != float64(401) || unauthEntry["level"] != "WARN" {
		t.Errorf("401 audit entry = %v", unauthEntry)
	}

	createEntry := byReqID["req-create-1"]
	if createEntry["subject"] != "provisioner" || createEntry["action"] != "create" ||
		createEntry["resource_type"] != "User" || createEntry["resource_id"] != userID ||
		createEntry["outcome"] != "success" || createEntry["status"] != float64(201) {
		t.Errorf("create audit entry = %v", createEntry)
	}

	dupEntry := byReqID["req-dup-1"]
	if dupEntry["outcome"] != "conflict" || dupEntry["scim_type"] != "uniqueness" || dupEntry["status"] != float64(409) {
		t.Errorf("conflict audit entry = %v", dupEntry)
	}

	patchEntry := byReqID["req-patch-1"]
	if patchEntry["action"] != "patch" || patchEntry["resource_id"] != userID || patchEntry["outcome"] != "success" {
		t.Errorf("patch audit entry = %v", patchEntry)
	}

	delEntry := byReqID["req-del-1"]
	if delEntry["action"] != "delete" || delEntry["resource_id"] != userID || delEntry["status"] != float64(204) {
		t.Errorf("delete audit entry = %v", delEntry)
	}
}

func TestAuditModeFromEnv(t *testing.T) {
	prev := AuditLogger()
	t.Cleanup(func() { SetAuditLogger(prev) })

	for _, val := range []string{"", "json", "text", "off", "none"} {
		t.Setenv("SCIM_AUDIT_LOG", val)
		var buf bytes.Buffer
		if _, err := ConfigureAuditLoggerFromEnv(&buf); err != nil {
			t.Errorf("SCIM_AUDIT_LOG=%q rejected: %v", val, err)
		}
	}

	t.Setenv("SCIM_AUDIT_LOG", "invalid-format")
	if _, err := ConfigureAuditLoggerFromEnv(nil); err == nil {
		t.Error("invalid SCIM_AUDIT_LOG was accepted")
	}
}
