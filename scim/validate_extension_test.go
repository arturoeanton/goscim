package scim

import (
	"encoding/json"
	"net/http"
	"testing"
)

// ValidateSchemas iterated every declared extension and type-asserted its
// object without looking at Required, so an extension the resource type marks
// optional was in practice mandatory. The shipped User declares the enterprise
// extension with "required": false.
func TestOptionalExtensionMayBeOmitted(t *testing.T) {
	r, store := newTestServer(t)

	body := `{"schemas":["` + schemaUser + `"],"userName":"jane.doe"}`
	w := do(t, r, http.MethodPost, usersPath, body)
	requireStatus(t, w, http.StatusOK) // TODO(B9): must become 201

	out := decode(t, w)
	if out["userName"] != "jane.doe" {
		t.Errorf("userName = %v", out["userName"])
	}
	if _, err := store.Get("User", out["id"].(string)); err != nil {
		t.Errorf("user not persisted: %v", err)
	}
}

// Optional does not mean unchecked: when the extension is sent it is validated
// like any other schema.
func TestOptionalExtensionIsValidatedWhenPresent(t *testing.T) {
	cases := []struct {
		name string
		ext  interface{}
		want int
	}{
		{"valid extension", map[string]interface{}{"department": "Engineering"}, http.StatusOK},
		{"empty extension", map[string]interface{}{}, http.StatusOK},
		{"attribute of the wrong type", map[string]interface{}{"department": 42}, http.StatusBadRequest},
		{"undeclared attribute", map[string]interface{}{"nope": "x"}, http.StatusBadRequest},
		{"not an object", "not-an-object", http.StatusBadRequest},
		{"an array", []interface{}{1, 2}, http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := newTestServer(t)
			body := map[string]interface{}{
				"schemas":     []string{schemaUser, schemaUserEnt},
				"userName":    "jane.doe",
				schemaUserEnt: tc.ext,
			}
			raw, _ := json.Marshal(body)
			w := do(t, r, http.MethodPost, usersPath, string(raw))
			requireStatus(t, w, tc.want)
		})
	}
}

// A required extension is still required. The Element resource type declares
// the "one" extension with "required": true.
func TestRequiredExtensionIsStillEnforced(t *testing.T) {
	r, _ := newTestServer(t)

	// Neither listed in "schemas" nor present as an object.
	body := `{"schemas":["` + schemaCore + `"],"name":"Element1","$ref":"/Element1"}`
	w := do(t, r, http.MethodPost, elementsPath, body)
	requireSCIMError(t, w, http.StatusBadRequest)

	// Listed in "schemas" but the object is missing.
	body = `{"schemas":["` + schemaCore + `","` + schemaExt + `"],"name":"Element1","$ref":"/Element1"}`
	w = do(t, r, http.MethodPost, elementsPath, body)
	requireSCIMError(t, w, http.StatusBadRequest)
}

// PUT goes through the same validation, so the same rule has to hold there.
func TestOptionalExtensionMayBeOmittedOnReplace(t *testing.T) {
	r, _ := newTestServer(t)

	w := do(t, r, http.MethodPost, usersPath, `{"schemas":["`+schemaUser+`"],"userName":"jane.doe"}`)
	requireStatus(t, w, http.StatusOK)
	created := decode(t, w)

	payload := map[string]interface{}{
		"schemas":  []string{schemaUser},
		"userName": "jane.roe",
		"meta":     created["meta"],
	}
	raw, _ := json.Marshal(payload)
	w = do(t, r, http.MethodPut, usersPath+"/"+created["id"].(string), string(raw))
	requireStatus(t, w, http.StatusOK)
	if decode(t, w)["userName"] != "jane.roe" {
		t.Error("the replace did not apply")
	}
}
