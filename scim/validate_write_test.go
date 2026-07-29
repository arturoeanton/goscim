package scim

import (
	"encoding/json"
	"net/http"
	"testing"
)

// User.password is declared writeOnly with "returned": "never", and was being
// echoed back in every response anyway.
func TestWriteOnlyAttributesAreNeverReturned(t *testing.T) {
	r, store := newTestServer(t)

	body := `{"schemas":["` + schemaUser + `"],"userName":"jane","password":"correct-horse"}`
	w := do(t, r, http.MethodPost, usersPath, body)
	requireStatus(t, w, http.StatusCreated)

	created := decode(t, w)
	if _, present := created["password"]; present {
		t.Errorf("the create response returned the password: %v", created["password"])
	}
	id := created["id"].(string)

	// It was stored, it is just never disclosed.
	stored, err := store.Get("User", id)
	if err != nil {
		t.Fatal(err)
	}
	if stored["password"] != "correct-horse" {
		t.Errorf("the password was not stored: %v", stored["password"])
	}

	w = do(t, r, http.MethodGet, usersPath+"/"+id, "")
	requireStatus(t, w, http.StatusOK)
	if _, present := decode(t, w)["password"]; present {
		t.Error("the read response returned the password")
	}

	w = do(t, r, http.MethodGet, usersPath, "")
	requireStatus(t, w, http.StatusOK)
	resources := decode(t, w)["Resources"].([]interface{})
	if _, present := resources[0].(map[string]interface{})["password"]; present {
		t.Error("the search response returned the password")
	}
}

// User.groups is readOnly: it is derived from Group membership, not set by the
// client. A value sent for it is ignored rather than refused, because a
// read-modify-write client echoes the whole resource back on a PUT.
func TestReadOnlyAttributesAreIgnoredNotRefused(t *testing.T) {
	r, store := newTestServer(t)

	body := map[string]interface{}{
		"schemas":  []string{schemaUser},
		"userName": "jane",
		"groups":   []interface{}{map[string]interface{}{"value": "smuggled-in"}},
	}
	raw, _ := json.Marshal(body)

	w := do(t, r, http.MethodPost, usersPath, string(raw))
	requireStatus(t, w, http.StatusCreated)

	created := decode(t, w)
	if _, present := created["groups"]; present {
		t.Errorf("the client set a read-only attribute: %v", created["groups"])
	}
	stored, _ := store.Get("User", created["id"].(string))
	if _, present := stored["groups"]; present {
		t.Errorf("a read-only attribute was persisted from the client: %v", stored["groups"])
	}
}

// The same on update: the stored value wins and the request still succeeds.
func TestReadOnlyAttributesSurviveAnUpdate(t *testing.T) {
	r, store := newTestServer(t)

	w := do(t, r, http.MethodPost, usersPath, `{"schemas":["`+schemaUser+`"],"userName":"jane"}`)
	requireStatus(t, w, http.StatusCreated)
	id := decode(t, w)["id"].(string)

	// Give the stored resource a server-owned value.
	stored, _ := store.Get("User", id)
	stored["groups"] = []interface{}{map[string]interface{}{"value": "engineering"}}
	if err := store.Upsert("User", id, stored); err != nil {
		t.Fatal(err)
	}

	// The client echoes the resource back with groups changed.
	body := map[string]interface{}{
		"schemas":  []string{schemaUser},
		"userName": "jane.updated",
		"groups":   []interface{}{map[string]interface{}{"value": "administrators"}},
	}
	raw, _ := json.Marshal(body)
	w = do(t, r, http.MethodPut, usersPath+"/"+id, string(raw))
	requireStatus(t, w, http.StatusOK)

	after, _ := store.Get("User", id)
	if after["userName"] != "jane.updated" {
		t.Errorf("the writable attribute was not updated: %v", after["userName"])
	}
	groups := after["groups"].([]interface{})
	if groups[0].(map[string]interface{})["value"] != "engineering" {
		t.Errorf("the client changed a read-only attribute: %v", groups)
	}
}

// --- $writer, the mirror of $reader ------------------------------------

func writerSchemas(t *testing.T) ResoruceType {
	t.Helper()
	const core = "urn:test:write:core"
	Schemas[core] = Schema{
		ID: core,
		Attributes: []Attribute{
			{Name: "open", Type: "string"},
			{Name: "restricted", Type: "string", Writer: strs("privileged")},
			{Name: "anyone", Type: "string", Writer: strs("*")},
			{Name: "identifier", Type: "string", Mutability: "immutable"},
			{Name: "profile", Type: "complex", SubAttributes: []Attribute{
				{Name: "nickname", Type: "string"},
				{Name: "clearance", Type: "string", Writer: strs("privileged")},
			}},
		},
	}
	return ResoruceType{Name: "Thing", Schema: core}
}

func TestWriterRolesGateWrites(t *testing.T) {
	r, _ := newTestServer(t) // loads the shipped schemas
	_ = r
	resourceType := writerSchemas(t)

	cases := []struct {
		name     string
		roles    []string
		incoming map[string]interface{}
		allowed  bool
	}{
		{"an unrestricted attribute", nil, map[string]interface{}{"open": "x"}, true},
		{"a restricted one without the role", []string{"user"}, map[string]interface{}{"restricted": "x"}, false},
		{"a restricted one with the role", []string{"privileged"}, map[string]interface{}{"restricted": "x"}, true},
		{"a wildcard writer", []string{"anyone-at-all"}, map[string]interface{}{"anyone": "x"}, true},
		{"a restricted sub-attribute without the role", []string{"user"},
			map[string]interface{}{"profile": map[string]interface{}{"clearance": "top"}}, false},
		{"an unrestricted sub-attribute", []string{"user"},
			map[string]interface{}{"profile": map[string]interface{}{"nickname": "jj"}}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := testContext(tc.roles)
			if got := EnforceWriteAccess(c, resourceType, tc.incoming, nil); got != tc.allowed {
				t.Fatalf("EnforceWriteAccess = %v, want %v", got, tc.allowed)
			}
			if !tc.allowed && w.Code != http.StatusForbidden {
				t.Errorf("status = %d, want 403", w.Code)
			}
		})
	}
}

// RFC 7643 7: an immutable attribute may be set once and never changed.
func TestImmutableAttributes(t *testing.T) {
	r, _ := newTestServer(t)
	_ = r
	resourceType := writerSchemas(t)

	t.Run("may be set on create", func(t *testing.T) {
		c, _ := testContext(nil)
		if !EnforceWriteAccess(c, resourceType, map[string]interface{}{"identifier": "abc"}, nil) {
			t.Error("setting an immutable attribute at creation was refused")
		}
	})

	t.Run("may be repeated unchanged on update", func(t *testing.T) {
		c, _ := testContext(nil)
		existing := map[string]interface{}{"identifier": "abc"}
		if !EnforceWriteAccess(c, resourceType, map[string]interface{}{"identifier": "abc"}, existing) {
			t.Error("echoing an unchanged immutable attribute was refused")
		}
	})

	t.Run("may not be changed", func(t *testing.T) {
		c, w := testContext(nil)
		existing := map[string]interface{}{"identifier": "abc"}
		if EnforceWriteAccess(c, resourceType, map[string]interface{}{"identifier": "different"}, existing) {
			t.Fatal("an immutable attribute was changed")
		}
		requireStatus(t, w, http.StatusBadRequest)
		if decode(t, w)["scimType"] != "mutability" {
			t.Errorf("scimType = %v, want mutability", decode(t, w)["scimType"])
		}
	})
}

// End to end: a caller without the writing role is refused with 403 and
// nothing is stored.
func TestWriteDenialIsRefusedOverHTTP(t *testing.T) {
	r, store := newTestServer(t)

	// Tighten $writer on description to a role the caller does not hold. The
	// handlers consult Schemas per request, so this takes effect immediately.
	schema := Schemas[schemaCore]
	attributes := append([]Attribute(nil), schema.Attributes...)
	for i := range attributes {
		if attributes[i].Name == "description" {
			attributes[i].Writer = strs("privileged")
		}
	}
	schema.Attributes = attributes
	Schemas[schemaCore] = schema

	w := do(t, r, http.MethodPost, elementsPath, validElement("Element1", 1))
	requireSCIMError(t, w, http.StatusForbidden)

	total, _, err := store.Search(SearchQuery{Bucket: "Element", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 {
		t.Errorf("a refused create was persisted (%d resources)", total)
	}
}
