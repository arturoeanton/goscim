package scim

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestHTMXConsoleEndToEndAndAuthGuard(t *testing.T) {
	t.Setenv("SCIM_UI", "true")
	DefaultInspector.Clear()

	r, _ := newTestServer(t)

	// 1. Offline vendored assets (htmx.min.js + style.css) must load without CDN.
	wJS := do(t, r, http.MethodGet, "/ui/assets/htmx.min.js", "")
	requireStatus(t, wJS, http.StatusOK)
	if len(wJS.Body.Bytes()) < 10000 {
		t.Fatalf("htmx.min.js unexpectedly small: %d bytes", len(wJS.Body.Bytes()))
	}
	wCSS := do(t, r, http.MethodGet, "/ui/assets/style.css", "")
	requireStatus(t, wCSS, http.StatusOK)

	// 2. Overview & live stats partial.
	wOverview := do(t, r, http.MethodGet, "/ui", "")
	requireStatus(t, wOverview, http.StatusOK)
	if !strings.Contains(wOverview.Body.String(), "Operational Overview") {
		t.Fatalf("overview HTML missing heading")
	}

	// 3. Provision a User and a Group via HTMX form endpoints in tenant "acme".
	formUser := url.Values{
		"tenant":      {"acme"},
		"userName":    {"htmx.admin@acme.io"},
		"displayName": {"HTMX Admin"},
		"givenName":   {"HTMX"},
		"familyName":  {"Admin"},
		"email":       {"htmx.admin@acme.io"},
		"department":  {"Identity"},
		"password":    {"ConsoleSecret123!"},
	}
	reqCreateU := httptest.NewRequest(http.MethodPost, "/ui/resources/users", strings.NewReader(formUser.Encode()))
	reqCreateU.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	wCreateU := serve(r, reqCreateU)
	requireStatus(t, wCreateU, http.StatusOK)
	if !strings.Contains(wCreateU.Body.String(), "htmx.admin@acme.io") {
		t.Fatalf("HTMX user table partial did not render created user: %s", wCreateU.Body.String())
	}

	// Lookup the created user ID in tenant acme.
	_, docs, err := DB.Search(SearchQuery{Bucket: "t:acme:User", Limit: 10})
	if err != nil || len(docs) != 1 {
		t.Fatalf("expected 1 user in t:acme:User, got %d (%v)", len(docs), err)
	}
	uID := docs[0]["id"].(string)

	// Create Group referencing uID via HTMX form.
	formGroup := url.Values{
		"tenant":      {"acme"},
		"displayName": {"AcmeAdmins"},
		"memberIDs":   {uID},
	}
	reqCreateG := httptest.NewRequest(http.MethodPost, "/ui/resources/groups", strings.NewReader(formGroup.Encode()))
	reqCreateG.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	wCreateG := serve(r, reqCreateG)
	requireStatus(t, wCreateG, http.StatusOK)
	if !strings.Contains(wCreateG.Body.String(), "AcmeAdmins") {
		t.Fatalf("HTMX group table partial did not render created group")
	}

	// 4. Inspect JSON detail drawer -> must display computed direct group badge.
	wDetail := do(t, r, http.MethodGet, "/ui/resources/detail?type=Users&id="+uID+"&tenant=acme", "")
	requireStatus(t, wDetail, http.StatusOK)
	if !strings.Contains(wDetail.Body.String(), "AcmeAdmins (direct)") {
		t.Fatalf("detail drawer missing computed group membership badge: %s", wDetail.Body.String())
	}

	// 5. Filter & N1QL Lab -> compiles parameterized N1QL and returns matched user.
	pgForm := url.Values{
		"tenant":       {"acme"},
		"resourceType": {"Users"},
		"filter":       {`emails[type eq "work" and value co "acme.io"]`},
	}
	reqPG := httptest.NewRequest(http.MethodPost, "/ui/playground/eval", strings.NewReader(pgForm.Encode()))
	reqPG.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	wPG := serve(r, reqPG)
	requireStatus(t, wPG, http.StatusOK)
	if !strings.Contains(wPG.Body.String(), "ANY") || !strings.Contains(wPG.Body.String(), "1 matched") {
		t.Fatalf("unexpected playground output: %s", wPG.Body.String())
	}

	// 6. Live Inspector must capture exchanges and redact the password field.
	exchanges := DefaultInspector.Recent(10)
	if len(exchanges) == 0 {
		t.Fatalf("expected inspector buffer to record exchanges")
	}
	var createSeq uint64
	for _, ex := range exchanges {
		if ex.Action == "create" && ex.ResourceType == "User" {
			createSeq = ex.Seq
			if strings.Contains(ex.RequestPreview, "ConsoleSecret123!") || !strings.Contains(ex.RequestPreview, "[REDACTED]") {
				t.Fatalf("inspector failed to redact password: %s", ex.RequestPreview)
			}
		}
	}
	if createSeq == 0 {
		t.Fatalf("did not find user create exchange in inspector")
	}
	wInspDetail := do(t, r, http.MethodGet, fmt.Sprintf("/ui/inspector/detail?seq=%d", createSeq), "")
	requireStatus(t, wInspDetail, http.StatusOK)

	// 7. Conformance Runner HTMX endpoint executes all profiles with 0 failures.
	confForm := url.Values{"tenant": {"acme"}, "profile": {"all"}}
	reqConf := httptest.NewRequest(http.MethodPost, "/ui/conformance/run", strings.NewReader(confForm.Encode()))
	reqConf.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	wConf := serve(r, reqConf)
	requireStatus(t, wConf, http.StatusOK)
	if !strings.Contains(wConf.Body.String(), "0 failed") {
		t.Fatalf("UI conformance run reported failures: %s", wConf.Body.String())
	}

	// 8. Optional SCIM_UI_TOKEN guard redirects unauthenticated browser requests to /ui/login.
	t.Setenv("SCIM_UI_TOKEN", "console-admin-secret")
	wGuarded := do(t, r, http.MethodGet, "/ui", "")
	requireStatus(t, wGuarded, http.StatusSeeOther)
	if loc := wGuarded.Header().Get("Location"); loc != "/ui/login" {
		t.Fatalf("expected redirect to /ui/login, got %q", loc)
	}

	loginForm := url.Values{"token": {"console-admin-secret"}}
	reqLogin := httptest.NewRequest(http.MethodPost, "/ui/login", strings.NewReader(loginForm.Encode()))
	reqLogin.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	wLogin := serve(r, reqLogin)
	requireStatus(t, wLogin, http.StatusSeeOther)
	cookies := wLogin.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie after /ui/login")
	}

	reqAuthed := httptest.NewRequest(http.MethodGet, "/ui", nil)
	reqAuthed.AddCookie(cookies[0])
	requireStatus(t, serve(r, reqAuthed), http.StatusOK)
}
