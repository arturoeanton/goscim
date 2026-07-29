package scim

import (
	"encoding/json"
	"net/http"
	"testing"
)

const (
	usersPath      = "/scim/v2/Users"
	schemaUser     = "urn:ietf:params:scim:schemas:core:2.0:User"
	schemaUserEnt  = "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
	multiComplex   = "multiValued"  // complex, multi-valued, in the Element extension
	multiScalarInt = "multiValued2" // integer, multi-valued, in the Element extension
)

// The validator ignored Attribute.MultiValued entirely, so a declared
// multi-valued attribute was checked as if it were scalar and every array was
// rejected. That made a standard SCIM User impossible to create: emails,
// phoneNumbers, addresses and groups are all multi-valued in the shipped
// core schema.
func TestCreateUserWithMultiValuedAttributes(t *testing.T) {
	r, store := newTestServer(t)

	body := map[string]interface{}{
		"schemas":  []string{schemaUser, schemaUserEnt},
		"userName": "jane.doe@example.com",
		"active":   true,
		"emails": []interface{}{
			map[string]interface{}{"value": "jane@example.com", "type": "work", "primary": true},
			map[string]interface{}{"value": "jane@home.example", "type": "home", "primary": false},
		},
		"phoneNumbers": []interface{}{
			map[string]interface{}{"value": "+34 600 000 000", "type": "mobile"},
		},
		schemaUserEnt: map[string]interface{}{"department": "Engineering"},
	}
	raw, _ := json.Marshal(body)

	w := do(t, r, http.MethodPost, usersPath, string(raw))
	requireStatus(t, w, http.StatusCreated)
	out := decode(t, w)

	emails, ok := out["emails"].([]interface{})
	if !ok || len(emails) != 2 {
		t.Fatalf("emails came back as %#v", out["emails"])
	}
	first := emails[0].(map[string]interface{})
	if first["value"] != "jane@example.com" || first["type"] != "work" || first["primary"] != true {
		t.Errorf("first email = %v", first)
	}

	stored, err := store.Get("User", out["id"].(string))
	if err != nil {
		t.Fatalf("user not persisted: %v", err)
	}
	if len(stored["emails"].([]interface{})) != 2 {
		t.Errorf("stored emails = %v", stored["emails"])
	}
}

// Each element of a multi-valued attribute is validated with the attribute's
// own type rules.
func TestMultiValuedElementsAreValidated(t *testing.T) {
	cases := []struct {
		name  string
		value interface{}
		want  int
	}{
		{
			"complex elements",
			[]interface{}{
				map[string]interface{}{"data": 1, "$ref": "/a"},
				map[string]interface{}{"data": 2, "$ref": "/b"},
			},
			http.StatusCreated,
		},
		{
			"an element missing a required sub-attribute",
			[]interface{}{map[string]interface{}{"data": 1}}, // $ref is required
			http.StatusBadRequest,
		},
		{
			"an element with a sub-attribute of the wrong type",
			[]interface{}{map[string]interface{}{"data": "not-an-integer", "$ref": "/a"}},
			http.StatusBadRequest,
		},
		{
			"an element with an undeclared sub-attribute",
			[]interface{}{map[string]interface{}{"$ref": "/a", "nope": 1}},
			http.StatusBadRequest,
		},
		{
			"a scalar where the array element must be complex",
			[]interface{}{"just a string"},
			http.StatusBadRequest,
		},
		{
			"an empty array",
			[]interface{}{},
			http.StatusCreated,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := newTestServer(t)
			w := do(t, r, http.MethodPost, elementsPath, elementWithExtra(t, multiComplex, tc.value))
			requireStatus(t, w, tc.want)
		})
	}
}

// A multi-valued attribute must be an array, and a single-valued one must not be.
func TestMultiValuedRequiresAnArray(t *testing.T) {
	r, _ := newTestServer(t)

	// Scalar where an array is declared.
	w := do(t, r, http.MethodPost, elementsPath,
		elementWithExtra(t, multiScalarInt, 42))
	requireStatus(t, w, http.StatusBadRequest)

	// Array where a scalar is declared.
	w = do(t, r, http.MethodPost, elementsPath,
		elementWithExtra(t, "numberInteger", []interface{}{1, 2}))
	requireStatus(t, w, http.StatusBadRequest)

	// The declared array of integers works, and the elements keep their type.
	w = do(t, r, http.MethodPost, elementsPath,
		elementWithExtra(t, multiScalarInt, []interface{}{1, 2, 3}))
	requireStatus(t, w, http.StatusCreated)

	// An array of integers must reject a non-integer element.
	w = do(t, r, http.MethodPost, elementsPath,
		elementWithExtra(t, multiScalarInt, []interface{}{1, 2.5}))
	requireStatus(t, w, http.StatusBadRequest)
}

// RFC 7643 2.4: at most one value of a multi-valued attribute may be primary.
func TestAtMostOnePrimaryValue(t *testing.T) {
	r, _ := newTestServer(t)

	twoPrimaries := map[string]interface{}{
		"schemas":  []string{schemaUser, schemaUserEnt},
		"userName": "jane",
		"emails": []interface{}{
			map[string]interface{}{"value": "a@example.com", "primary": true},
			map[string]interface{}{"value": "b@example.com", "primary": true},
		},
		schemaUserEnt: map[string]interface{}{},
	}
	raw, _ := json.Marshal(twoPrimaries)
	w := do(t, r, http.MethodPost, usersPath, string(raw))
	requireSCIMError(t, w, http.StatusBadRequest)

	// One primary is fine.
	onePrimary := map[string]interface{}{
		"schemas":  []string{schemaUser, schemaUserEnt},
		"userName": "jane",
		"emails": []interface{}{
			map[string]interface{}{"value": "a@example.com", "primary": true},
			map[string]interface{}{"value": "b@example.com", "primary": false},
		},
		schemaUserEnt: map[string]interface{}{},
	}
	raw, _ = json.Marshal(onePrimary)
	w = do(t, r, http.MethodPost, usersPath, string(raw))
	requireStatus(t, w, http.StatusCreated)
}

// elementWithExtra builds a valid Element payload with one extra attribute set
// inside the "one" extension.
func elementWithExtra(t *testing.T, attribute string, value interface{}) string {
	t.Helper()
	body := map[string]interface{}{
		"schemas": []string{schemaCore, schemaExt},
		"name":    "Element1",
		"$ref":    "/Element1",
		schemaExt: map[string]interface{}{
			"required": 1,
			attribute:  value,
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
