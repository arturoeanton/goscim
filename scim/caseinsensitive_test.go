package scim

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// RFC 7643 2.1: attribute names and schema URNs are case-insensitive.
// Comparison was exact, so a client sending "username" was told the attribute
// does not exist. Identity providers vary the casing in practice.

func TestAttributeNamesAreCaseInsensitive(t *testing.T) {
	r, store := newTestServer(t)

	body := `{"schemas":["` + schemaUser + `"],"USERNAME":"jane.doe","Active":true}`
	w := do(t, r, http.MethodPost, usersPath, body)
	requireStatus(t, w, http.StatusCreated)

	// The resource is stored with the spelling the schema declares, not the
	// one the client happened to send: a filter asks the schema for the path,
	// and would not find "USERNAME".
	stored, err := store.Get("User", decode(t, w)["id"].(string))
	if err != nil {
		t.Fatal(err)
	}
	if stored["userName"] != "jane.doe" {
		t.Errorf("stored keys were not canonicalised: %v", stored)
	}
	if _, present := stored["USERNAME"]; present {
		t.Errorf("the client's spelling survived: %v", stored)
	}
	if stored["active"] != true {
		t.Errorf("active = %v", stored["active"])
	}
}

func TestSchemaURNsAreCaseInsensitive(t *testing.T) {
	r, store := newTestServer(t)

	// The URN in a different case, and the extension object keyed the same way.
	body := `{"schemas":["` + strings.ToUpper(schemaUser) + `","` + strings.ToUpper(schemaUserEnt) + `"],
	          "userName":"jane.doe",
	          "` + strings.ToUpper(schemaUserEnt) + `":{"department":"Engineering"}}`
	w := do(t, r, http.MethodPost, usersPath, body)
	requireStatus(t, w, http.StatusCreated)

	stored, err := store.Get("User", decode(t, w)["id"].(string))
	if err != nil {
		t.Fatal(err)
	}
	// Both the schemas list and the extension key are canonicalised.
	schemas := stored["schemas"].([]interface{})
	for _, s := range schemas {
		if s == strings.ToUpper(schemaUser) || s == strings.ToUpper(schemaUserEnt) {
			t.Errorf("a schema URN kept the client's casing: %v", schemas)
		}
	}
	extension, ok := stored[schemaUserEnt].(map[string]interface{})
	if !ok {
		t.Fatalf("the extension is not under its canonical URN: %v", stored)
	}
	if extension["department"] != "Engineering" {
		t.Errorf("extension = %v", extension)
	}
}

// Sub-attributes of a complex value follow the same rule.
func TestSubAttributeNamesAreCaseInsensitive(t *testing.T) {
	r, store := newTestServer(t)

	body := `{"schemas":["` + schemaUser + `"],"userName":"jane",
	          "emails":[{"VALUE":"jane@example.com","Type":"work"}]}`
	w := do(t, r, http.MethodPost, usersPath, body)
	requireStatus(t, w, http.StatusCreated)

	stored, _ := store.Get("User", decode(t, w)["id"].(string))
	email := stored["emails"].([]interface{})[0].(map[string]interface{})
	if email["value"] != "jane@example.com" || email["type"] != "work" {
		t.Errorf("sub-attribute keys were not canonicalised: %v", email)
	}
}

// An attribute that really is undeclared is still refused, whatever its case.
func TestUndeclaredAttributesAreStillRefused(t *testing.T) {
	r, _ := newTestServer(t)

	body := `{"schemas":["` + schemaUser + `"],"userName":"jane","NoSuchAttribute":1}`
	requireSCIMError(t, do(t, r, http.MethodPost, usersPath, body), http.StatusBadRequest)
}

// A 200 is not enough here: sortBy has to be canonicalised or the ORDER BY
// names a key no document has, and the results come back in an arbitrary order
// while the request still looks successful.
func TestSortByIsCaseInsensitive(t *testing.T) {
	r, _ := newTestServer(t)
	createElement(t, r, "CCC", 3)
	createElement(t, r, "AAA", 1)
	createElement(t, r, "BBB", 2)

	for _, sortBy := range []string{"name", "NAME", "Name"} {
		t.Run(sortBy, func(t *testing.T) {
			w := do(t, r, http.MethodGet, elementsPath+"?sortBy="+url.QueryEscape(sortBy), "")
			requireStatus(t, w, http.StatusOK)

			names := make([]string, 0, 3)
			for _, raw := range decode(t, w)["Resources"].([]interface{}) {
				names = append(names, raw.(map[string]interface{})["name"].(string))
			}
			if strings.Join(names, ",") != "AAA,BBB,CCC" {
				t.Errorf("sortBy=%s ordered %v", sortBy, names)
			}
		})
	}

	// Common attributes and an undeclared one behave as before.
	requireStatus(t, do(t, r, http.MethodGet, elementsPath+"?sortBy=META.lastModified", ""), http.StatusOK)
	requireStatus(t, do(t, r, http.MethodGet, elementsPath+"?sortBy=NOPE", ""), http.StatusBadRequest)
}

func TestCanonicalAttributePath(t *testing.T) {
	newTestServer(t)
	users := Resources["/Users"]

	cases := []struct{ in, want string }{
		{"userName", "userName"},
		{"USERNAME", "userName"},
		{"UserName", "userName"},
		{"EMAILS.VALUE", "emails.value"},
		{"name.FAMILYNAME", "name.familyName"},
		{strings.ToUpper(schemaUserEnt) + ".DEPARTMENT", schemaUserEnt + ".department"},
		{"notAnAttribute", "notAnAttribute"}, // left alone rather than guessed at
	}
	for _, tc := range cases {
		if got := CanonicalAttributePath(users, tc.in); got != tc.want {
			t.Errorf("CanonicalAttributePath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestLookupSchemaIgnoresCase(t *testing.T) {
	newTestServer(t)

	for _, id := range []string{schemaUser, strings.ToUpper(schemaUser), strings.ToLower(schemaUser)} {
		schema, ok := LookupSchema(id)
		if !ok {
			t.Errorf("LookupSchema(%q) found nothing", id)
			continue
		}
		if schema.ID != schemaUser {
			t.Errorf("LookupSchema(%q).ID = %q, want the declared spelling", id, schema.ID)
		}
	}
	if _, ok := LookupSchema("urn:made:up"); ok {
		t.Error("LookupSchema invented a schema")
	}
}

func TestFindAttributeIgnoresCase(t *testing.T) {
	attributes := []Attribute{{Name: "userName"}, {Name: "displayName"}}
	for _, name := range []string{"userName", "USERNAME", "username", "UserName"} {
		attribute, ok := FindAttribute(attributes, name)
		if !ok || attribute.Name != "userName" {
			t.Errorf("FindAttribute(%q) = %v, %v", name, attribute, ok)
		}
	}
	if _, ok := FindAttribute(attributes, "nope"); ok {
		t.Error("FindAttribute invented an attribute")
	}
}
