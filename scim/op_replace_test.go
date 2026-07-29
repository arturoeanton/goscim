package scim

import (
	"encoding/json"
	"net/http"
	"testing"
)

// replace() read the previous meta straight out of the request body with an
// unchecked type assertion, so a PUT without meta panicked. RFC 7644 does not
// require a client to echo meta back, so the ordinary case was the one that
// crashed.
func TestReplaceWithoutMetaInTheBody(t *testing.T) {
	r, store := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)
	originalMeta := created["meta"].(map[string]interface{})

	// Exactly what a client sends: the resource, no meta.
	w := do(t, r, http.MethodPut, elementsPath+"/"+id, validElement("Element1-replaced", 9))
	requireStatus(t, w, http.StatusOK)
	body := decode(t, w)

	if body["name"] != "Element1-replaced" {
		t.Errorf("name = %v", body["name"])
	}
	meta, ok := body["meta"].(map[string]interface{})
	if !ok {
		t.Fatalf("meta missing from the response: %v", body)
	}
	if meta["created"] != originalMeta["created"] {
		t.Errorf("created = %v, want the original %v", meta["created"], originalMeta["created"])
	}
	if meta["version"] == originalMeta["version"] {
		t.Error("version did not change")
	}

	stored, err := store.Get("Element", id)
	if err != nil {
		t.Fatal(err)
	}
	if stored["name"] != "Element1-replaced" {
		t.Errorf("stored name = %v", stored["name"])
	}
}

// meta belongs to the server. A client that sends one must not be able to
// rewrite the resource's history.
func TestReplaceIgnoresClientSuppliedMeta(t *testing.T) {
	r, _ := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)
	originalMeta := created["meta"].(map[string]interface{})

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(validElement("Element1", 1)), &payload); err != nil {
		t.Fatal(err)
	}
	payload["meta"] = map[string]interface{}{
		"created":      "1999-01-01T00:00:00Z",
		"version":      "forged",
		"resourceType": "NotAnElement",
		"location":     "/somewhere/else",
	}
	payload["id"] = "forged-id"
	raw, _ := json.Marshal(payload)

	w := do(t, r, http.MethodPut, elementsPath+"/"+id, string(raw))
	requireStatus(t, w, http.StatusOK)
	body := decode(t, w)

	if body["id"] != id {
		t.Errorf("the client changed the id: %v", body["id"])
	}
	meta := body["meta"].(map[string]interface{})
	if meta["created"] != originalMeta["created"] {
		t.Errorf("the client rewrote created: %v", meta["created"])
	}
	if meta["version"] == "forged" {
		t.Error("the client set the version")
	}
	if meta["resourceType"] != "Element" {
		t.Errorf("resourceType = %v", meta["resourceType"])
	}
}

// A PUT to an id that does not exist is a 404 whatever the body looks like.
func TestReplaceMissingResource(t *testing.T) {
	r, _ := newTestServer(t)

	w := do(t, r, http.MethodPut, elementsPath+"/does-not-exist", validElement("Element1", 1))
	requireStatus(t, w, http.StatusNotFound)

	w = do(t, r, http.MethodPut, elementsPath+"/does-not-exist", `{"nonsense":true}`)
	requireStatus(t, w, http.StatusNotFound)
}

// The body is still validated against the schema.
func TestReplaceValidatesTheBody(t *testing.T) {
	r, store := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)

	// $ref is required by the core Element schema.
	body := `{"schemas":["` + schemaCore + `","` + schemaExt + `"],"name":"x","` + schemaExt + `":{"required":1}}`
	w := do(t, r, http.MethodPut, elementsPath+"/"+id, body)
	requireSCIMError(t, w, http.StatusBadRequest)

	stored, _ := store.Get("Element", id)
	if stored["name"] != "Element1" {
		t.Errorf("an invalid replace was persisted: %v", stored["name"])
	}
}

// updateMeta used to assert metaOld["created"].(string), so a resource stored
// without meta -- written by an earlier version, say -- panicked on update.
func TestUpdateMetaToleratesAMissingCreated(t *testing.T) {
	resourceType := ResoruceType{Name: "Element", Endpoint: "/Elements"}
	element := map[string]interface{}{"id": "abc"}

	cases := []struct {
		name    string
		metaOld map[string]interface{}
	}{
		{"nil meta", nil},
		{"empty meta", map[string]interface{}{}},
		{"created of the wrong type", map[string]interface{}{"created": 12345}},
		{"created present", map[string]interface{}{"created": "2020-01-01T00:00:00Z"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			meta := updateMeta(tc.metaOld, element, resourceType)
			if meta.Created == "" {
				t.Error("created left empty")
			}
			if meta.Version == "" || meta.LastModified == "" {
				t.Errorf("meta = %+v", meta)
			}
			if meta.Location != PREFIX+"/Elements/abc" {
				t.Errorf("location = %s", meta.Location)
			}
		})
	}

	// When it is there, it is carried forward untouched.
	meta := updateMeta(map[string]interface{}{"created": "2020-01-01T00:00:00Z"}, element, resourceType)
	if meta.Created != "2020-01-01T00:00:00Z" {
		t.Errorf("created = %s", meta.Created)
	}
}
