package scim

import (
	"encoding/json"
	"net/http"
	"testing"
)

// patchBody builds a PatchOp request from raw operations.
func patchBody(ops ...map[string]interface{}) string {
	body := map[string]interface{}{
		"schemas":    []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		"Operations": ops,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

// pointValue walked the resource with bare type assertions, so as soon as a
// path segment was missing the interface conversion panicked and the request
// died as a 500. Every one of these used to take the handler down.
func TestPatchRejectsUnresolvablePathsWithoutPanicking(t *testing.T) {
	cases := []struct {
		name string
		op   map[string]interface{}
	}{
		{"missing intermediate segment", map[string]interface{}{"op": "replace", "path": "doesNotExist.sub", "value": "x"}},
		{"segment through a scalar", map[string]interface{}{"op": "replace", "path": "name.deeper", "value": "x"}},
		{"deeply missing path", map[string]interface{}{"op": "add", "path": "a.b.c.d", "value": "x"}},
		{"remove through a missing path", map[string]interface{}{"op": "remove", "path": "nope.sub"}},
		{"empty segment", map[string]interface{}{"op": "replace", "path": "name..sub", "value": "x"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, store := newTestServer(t)
			created := createElement(t, r, "Element1", 1)
			id := created["id"].(string)

			w := do(t, r, http.MethodPatch, elementsPath+"/"+id, patchBody(tc.op))
			requireSCIMError(t, w, http.StatusBadRequest)

			// A rejected patch must leave the resource untouched.
			stored, err := store.Get("Element", id)
			if err != nil {
				t.Fatal(err)
			}
			if stored["name"] != "Element1" {
				t.Errorf("the resource was modified by a rejected patch: %v", stored)
			}
		})
	}
}

// Malformed requests are client errors too, and must not be silently ignored.
func TestPatchRejectsMalformedRequests(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"not JSON", `{`},
		{"no operations", `{"schemas":["urn:ietf:params:scim:api:messages:2.0:PatchOp"]}`},
		{"empty operations", `{"schemas":["urn:ietf:params:scim:api:messages:2.0:PatchOp"],"Operations":[]}`},
		{"unknown op", patchBody(map[string]interface{}{"op": "frobnicate", "path": "name", "value": "x"})},
		{"remove without a path", patchBody(map[string]interface{}{"op": "remove"})},
		{"pathless replace with a scalar", patchBody(map[string]interface{}{"op": "replace", "value": "x"})},
		{"value filter in the path", patchBody(map[string]interface{}{"op": "replace", "path": `emails[type eq "work"].value`, "value": "x"})},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := newTestServer(t)
			created := createElement(t, r, "Element1", 1)
			w := do(t, r, http.MethodPatch, elementsPath+"/"+created["id"].(string), tc.body)
			requireSCIMError(t, w, http.StatusBadRequest)
		})
	}
}

// The operations that used to work must keep working.
func TestPatchOperations(t *testing.T) {
	t.Run("replace a top-level attribute", func(t *testing.T) {
		r, store := newTestServer(t)
		created := createElement(t, r, "Element1", 1)
		id := created["id"].(string)

		w := do(t, r, http.MethodPatch, elementsPath+"/"+id,
			patchBody(map[string]interface{}{"op": "replace", "path": "description", "value": "changed"}))
		requireStatus(t, w, http.StatusOK)

		stored, _ := store.Get("Element", id)
		if stored["description"] != "changed" {
			t.Errorf("description = %v", stored["description"])
		}
	})

	t.Run("replace inside an extension", func(t *testing.T) {
		r, store := newTestServer(t)
		created := createElement(t, r, "Element1", 1)
		id := created["id"].(string)

		w := do(t, r, http.MethodPatch, elementsPath+"/"+id,
			patchBody(map[string]interface{}{"op": "replace", "path": schemaExt + ".required", "value": 42}))
		requireStatus(t, w, http.StatusOK)

		stored, _ := store.Get("Element", id)
		if stored[schemaExt].(map[string]interface{})["required"] != float64(42) {
			t.Errorf("extension = %v", stored[schemaExt])
		}
	})

	t.Run("remove an attribute", func(t *testing.T) {
		r, store := newTestServer(t)
		created := createElement(t, r, "Element1", 1)
		id := created["id"].(string)

		w := do(t, r, http.MethodPatch, elementsPath+"/"+id,
			patchBody(map[string]interface{}{"op": "remove", "path": "description"}))
		requireStatus(t, w, http.StatusOK)

		stored, _ := store.Get("Element", id)
		if _, present := stored["description"]; present {
			t.Errorf("description survived the remove: %v", stored)
		}
	})

	t.Run("add appends to an array", func(t *testing.T) {
		r, store := newTestServer(t)
		body := map[string]interface{}{
			"schemas":  []string{schemaUser, schemaUserEnt},
			"userName": "jane",
			"emails": []interface{}{
				map[string]interface{}{"value": "a@example.com"},
			},
		}
		raw, _ := json.Marshal(body)
		w := do(t, r, http.MethodPost, usersPath, string(raw))
		requireStatus(t, w, http.StatusCreated)
		id := decode(t, w)["id"].(string)

		w = do(t, r, http.MethodPatch, usersPath+"/"+id, patchBody(map[string]interface{}{
			"op":    "add",
			"path":  "emails",
			"value": []interface{}{map[string]interface{}{"value": "b@example.com"}},
		}))
		requireStatus(t, w, http.StatusOK)

		stored, _ := store.Get("User", id)
		emails := stored["emails"].([]interface{})
		if len(emails) != 2 {
			t.Fatalf("emails = %v", emails)
		}
		if emails[1].(map[string]interface{})["value"] != "b@example.com" {
			t.Errorf("appended value = %v", emails[1])
		}
	})

	// RFC 7644 3.5.2.1: an operation with no path applies the members of its
	// object value to the resource. This used to be silently ignored, so the
	// server answered 200 without having changed anything.
	t.Run("pathless replace applies every member", func(t *testing.T) {
		r, store := newTestServer(t)
		created := createElement(t, r, "Element1", 1)
		id := created["id"].(string)

		w := do(t, r, http.MethodPatch, elementsPath+"/"+id, patchBody(map[string]interface{}{
			"op":    "replace",
			"value": map[string]interface{}{"description": "from a pathless op"},
		}))
		requireStatus(t, w, http.StatusOK)

		stored, _ := store.Get("Element", id)
		if stored["description"] != "from a pathless op" {
			t.Errorf("the pathless operation was ignored: %v", stored["description"])
		}
	})

	t.Run("several operations apply in order", func(t *testing.T) {
		r, store := newTestServer(t)
		created := createElement(t, r, "Element1", 1)
		id := created["id"].(string)

		w := do(t, r, http.MethodPatch, elementsPath+"/"+id, patchBody(
			map[string]interface{}{"op": "replace", "path": "description", "value": "first"},
			map[string]interface{}{"op": "replace", "path": "description", "value": "second"},
		))
		requireStatus(t, w, http.StatusOK)

		stored, _ := store.Get("Element", id)
		if stored["description"] != "second" {
			t.Errorf("description = %v", stored["description"])
		}
	})
}

// A patch that produces a resource violating the schema must still be refused.
func TestPatchStillValidatesTheResult(t *testing.T) {
	r, store := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)

	w := do(t, r, http.MethodPatch, elementsPath+"/"+id,
		patchBody(map[string]interface{}{"op": "replace", "path": "name", "value": 12345}))
	requireSCIMError(t, w, http.StatusBadRequest)

	stored, _ := store.Get("Element", id)
	if stored["name"] != "Element1" {
		t.Errorf("an invalid patch was persisted: %v", stored["name"])
	}
}
