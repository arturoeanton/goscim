package scim

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// newTestServer builds the real SCIM router from the repository config, backed
// by an in-memory store. Everything below the HTTP layer is the production code
// path: routing, validation, meta generation and role filtering all run.
func newTestServer(t *testing.T) (*gin.Engine, *MemoryStore) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	store := NewMemoryStore()
	DB = store
	r := gin.New()
	if _, err := NewRouter("../config", r); err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	return r, store
}

// newRequest builds a SCIM request, for the cases that need extra headers.
func newRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/scim+json")
	return req
}

// serve runs a prepared request through the router.
func serve(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// do issues a request against the router and returns the recorder.
func do(t *testing.T, r http.Handler, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	return serve(r, newRequest(method, target, body))
}

// decode unmarshals a JSON response body, failing the test if it is not JSON.
func decode(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var out map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("response is not JSON (status %d): %v\nbody: %s", w.Code, err, w.Body.String())
	}
	return out
}

// requireStatus asserts the response status, printing the body on mismatch.
func requireStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("status = %d, want %d\nbody: %s", w.Code, want, w.Body.String())
	}
}

// requireSCIMError asserts the response is a SCIM error object with the given status.
func requireSCIMError(t *testing.T, w *httptest.ResponseRecorder, want int) map[string]interface{} {
	t.Helper()
	requireStatus(t, w, want)
	out := decode(t, w)
	schemas, _ := out["schemas"].([]interface{})
	if len(schemas) != 1 || schemas[0] != "urn:ietf:params:scim:api:messages:2.0:Error" {
		t.Errorf("not a SCIM error object: %v", out)
	}
	return out
}

const (
	elementsPath = "/scim/v2/Elements"
	schemaCore   = "urn:ietf:params:scim:schemas:core:2.0:Element"
	schemaExt    = "urn:ietf:params:scim:schemas:extension:one:2.0:Element"
)

// validElement is a payload that satisfies the shipped Element resource type:
// name and $ref are required by the core schema, and the "one" extension is
// declared required by config/resourceType/Element.json.
func validElement(name string, required int) string {
	body := map[string]interface{}{
		"schemas":     []string{schemaCore, schemaExt},
		"name":        name,
		"description": "description of " + name,
		"$ref":        "/" + name,
		schemaExt:     map[string]interface{}{"required": required},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

// createElement posts a valid element and returns the decoded response.
func createElement(t *testing.T, r http.Handler, name string, required int) map[string]interface{} {
	t.Helper()
	w := do(t, r, http.MethodPost, elementsPath, validElement(name, required))
	requireStatus(t, w, http.StatusCreated)
	return decode(t, w)
}
