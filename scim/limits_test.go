package scim

import (
	"net/http"
	"strings"
	"testing"
)

// The handlers read the whole body into memory before parsing it, so an
// unbounded POST was enough to exhaust the server.
func TestOversizedBodyIsRefused(t *testing.T) {
	r, store := newTestServer(t)

	// A syntactically valid payload with one enormous attribute value.
	huge := `{"schemas":["` + schemaUser + `"],"userName":"` + strings.Repeat("a", int(DefaultMaxBodyBytes)+1024) + `"}`
	w := do(t, r, http.MethodPost, usersPath, huge)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413\nbody: %s", w.Code, w.Body.String())
	}

	total, _, err := store.Search(SearchQuery{Bucket: "User", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 {
		t.Errorf("an oversized request was persisted (%d resources)", total)
	}
}

// A payload under the limit is unaffected.
func TestBodyUnderTheLimitIsAccepted(t *testing.T) {
	r, _ := newTestServer(t)

	body := `{"schemas":["` + schemaUser + `"],"userName":"` + strings.Repeat("a", 1024) + `"}`
	requireStatus(t, do(t, r, http.MethodPost, usersPath, body), http.StatusCreated)
}

// The limit applies to every verb that carries a payload.
func TestBodyLimitAppliesToEveryWrite(t *testing.T) {
	r, _ := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)
	huge := strings.Repeat("a", int(DefaultMaxBodyBytes)+1024)

	cases := []struct {
		method string
		target string
		body   string
	}{
		{http.MethodPut, elementsPath + "/" + id,
			`{"schemas":["` + schemaCore + `"],"name":"` + huge + `"}`},
		{http.MethodPatch, elementsPath + "/" + id,
			`{"schemas":["urn:ietf:params:scim:api:messages:2.0:PatchOp"],"Operations":[{"op":"replace","path":"name","value":"` + huge + `"}]}`},
	}

	for _, tc := range cases {
		t.Run(tc.method, func(t *testing.T) {
			w := do(t, r, tc.method, tc.target, tc.body)
			if w.Code != http.StatusRequestEntityTooLarge {
				t.Errorf("status = %d, want 413", w.Code)
			}
		})
	}
}
