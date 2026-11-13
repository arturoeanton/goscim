package scim

import (
	"net/http"
	"testing"
)

func TestPatchValueFilterReplaceAddAndRemove(t *testing.T) {
	r, _ := newTestServer(t)

	userPayload := `{
		"schemas": ["` + schemaUser + `"],
		"userName": "entra.user",
		"emails": [
			{"value": "old.work@example.com", "type": "work", "display": "Work Mail"},
			{"value": "keep.home@example.com", "type": "home"}
		]
	}`
	w := do(t, r, http.MethodPost, usersPath, userPayload)
	requireStatus(t, w, http.StatusCreated)
	id := decode(t, w)["id"].(string)

	// 1. Replace a sub-attribute on the matching element only (case-insensitive type match).
	w = do(t, r, http.MethodPatch, usersPath+"/"+id, patchBody(map[string]interface{}{
		"op":    "replace",
		"path":  `emails[type eq "WORK"].value`,
		"value": "new.work@example.com",
	}))
	requireStatus(t, w, http.StatusOK)
	updated := decode(t, w)
	emails := updated["emails"].([]interface{})
	if len(emails) != 2 {
		t.Fatalf("len(emails) = %d, want 2", len(emails))
	}
	if emails[0].(map[string]interface{})["value"] != "new.work@example.com" {
		t.Errorf("work email was not updated: %v", emails[0])
	}
	if emails[1].(map[string]interface{})["value"] != "keep.home@example.com" {
		t.Errorf("home email was mutated: %v", emails[1])
	}

	// 2. Replace when no element matches returns 400 noTarget (RFC 7644 3.5.2.3).
	w = do(t, r, http.MethodPatch, usersPath+"/"+id, patchBody(map[string]interface{}{
		"op":    "replace",
		"path":  `emails[type eq "other"].value`,
		"value": "other@example.com",
	}))
	errBody := requireSCIMError(t, w, http.StatusBadRequest)
	if errBody["scimType"] != "noTarget" {
		t.Errorf("scimType = %v, want noTarget", errBody["scimType"])
	}

	// 3. Entra ID style "add" on an unmatched value-selection path synthesizes
	// the new entry from equality predicates in the filter.
	w = do(t, r, http.MethodPatch, usersPath+"/"+id, patchBody(
		map[string]interface{}{
			"op":    "add",
			"path":  `phoneNumbers[type eq "mobile" and primary eq true].value`,
			"value": "+1-555-0100",
		},
		map[string]interface{}{
			"op":    "add",
			"path":  "name.givenName",
			"value": "Entra",
		},
	))
	requireStatus(t, w, http.StatusOK)
	afterAdd := decode(t, w)
	phones, ok := afterAdd["phoneNumbers"].([]interface{})
	if !ok || len(phones) != 1 {
		t.Fatalf("phoneNumbers = %#v, want 1 synthesized entry", afterAdd["phoneNumbers"])
	}
	phone := phones[0].(map[string]interface{})
	if phone["type"] != "mobile" || phone["primary"] != true || phone["value"] != "+1-555-0100" {
		t.Errorf("synthesized phoneNumber = %v", phone)
	}
	nameObj, _ := afterAdd["name"].(map[string]interface{})
	if nameObj["givenName"] != "Entra" {
		t.Errorf("name.givenName = %v, want Entra", nameObj)
	}

	// 4. Remove a sub-attribute on a matching element keeps the element itself.
	w = do(t, r, http.MethodPatch, usersPath+"/"+id, patchBody(map[string]interface{}{
		"op":   "remove",
		"path": `emails[type eq "work"].display`,
	}))
	requireStatus(t, w, http.StatusOK)
	afterSubRemove := decode(t, w)
	workEmail := afterSubRemove["emails"].([]interface{})[0].(map[string]interface{})
	if _, hasDisplay := workEmail["display"]; hasDisplay {
		t.Errorf("display survived remove: %v", workEmail)
	}
	if workEmail["value"] != "new.work@example.com" {
		t.Errorf("value lost during sub-attribute remove: %v", workEmail)
	}
}

func TestPatchRemoveArrayElementsUnassignsEmptyAttribute(t *testing.T) {
	r, _ := newTestServer(t)

	groupPayload := `{
		"schemas": ["urn:ietf:params:scim:schemas:core:2.0:Group"],
		"displayName": "Engineering",
		"members": [
			{"value": "u-1", "type": "User"},
			{"value": "u-2", "type": "User"}
		]
	}`
	w := do(t, r, http.MethodPost, "/scim/v2/Groups", groupPayload)
	requireStatus(t, w, http.StatusCreated)
	groupID := decode(t, w)["id"].(string)

	// Remove one member by value filter.
	w = do(t, r, http.MethodPatch, "/scim/v2/Groups/"+groupID, patchBody(map[string]interface{}{
		"op":   "remove",
		"path": `members[value eq "u-1"]`,
	}))
	requireStatus(t, w, http.StatusOK)
	members := decode(t, w)["members"].([]interface{})
	if len(members) != 1 || members[0].(map[string]interface{})["value"] != "u-2" {
		t.Fatalf("members after first remove = %v", members)
	}

	// Removing a non-existent member yields 400 noTarget.
	w = do(t, r, http.MethodPatch, "/scim/v2/Groups/"+groupID, patchBody(map[string]interface{}{
		"op":   "remove",
		"path": `members[value eq "u-1"]`,
	}))
	errBody := requireSCIMError(t, w, http.StatusBadRequest)
	if errBody["scimType"] != "noTarget" {
		t.Errorf("scimType = %v, want noTarget", errBody["scimType"])
	}

	// Removing the final member unassigns members completely (RFC 7644 3.5.2.2).
	w = do(t, r, http.MethodPatch, "/scim/v2/Groups/"+groupID, patchBody(map[string]interface{}{
		"op":   "remove",
		"path": `members[value eq "u-2"]`,
	}))
	requireStatus(t, w, http.StatusOK)
	if _, present := decode(t, w)["members"]; present {
		t.Errorf("empty members array should have been unassigned")
	}
}

func TestPatchColonDelimitedURNAndExtensionAutoAttach(t *testing.T) {
	r, _ := newTestServer(t)

	w := do(t, r, http.MethodPost, usersPath, `{"schemas":["`+schemaUser+`"],"userName":"colon.user","emails":[{"type":"work","value":"a@b.com"}]}`)
	requireStatus(t, w, http.StatusCreated)
	id := decode(t, w)["id"].(string)

	w = do(t, r, http.MethodPatch, usersPath+"/"+id, patchBody(
		map[string]interface{}{
			"op":    "replace",
			"path":  enterpriseUserExt + ":department",
			"value": "Platform",
		},
		map[string]interface{}{
			"op":    "replace",
			"path":  schemaUser + `:emails[type eq "work"].value`,
			"value": "platform@b.com",
		},
	))
	requireStatus(t, w, http.StatusOK)
	got := decode(t, w)

	extObj, ok := got[enterpriseUserExt].(map[string]interface{})
	if !ok || extObj["department"] != "Platform" {
		t.Fatalf("enterprise extension = %#v, want department=Platform", got[enterpriseUserExt])
	}
	schemasList, _ := got["schemas"].([]interface{})
	foundExtSchema := false
	for _, s := range schemasList {
		if s == enterpriseUserExt {
			foundExtSchema = true
		}
	}
	if !foundExtSchema {
		t.Errorf("schemas = %v, expected %s to be attached", schemasList, enterpriseUserExt)
	}
	emails := got["emails"].([]interface{})
	if emails[0].(map[string]interface{})["value"] != "platform@b.com" {
		t.Errorf("emails = %v, want platform@b.com", emails)
	}
}

func TestPatchValueFilterEnforcesRBACAndFilterVisibility(t *testing.T) {
	adminRouter, store := newTestServerAs(t, &AnonymousAuthenticator{Roles: []string{"admin"}})
	userRouter := newRouterWithStore(t, store, &AnonymousAuthenticator{Roles: []string{"user"}})

	origSchema := Schemas[schemaUser]
	t.Cleanup(func() { Schemas[schemaUser] = origSchema })

	cloned := origSchema
	attrs := append([]Attribute(nil), origSchema.Attributes...)
	for i := range attrs {
		if attrs[i].Name == "emails" {
			subs := append([]Attribute(nil), attrs[i].SubAttributes...)
			for j := range subs {
				if subs[j].Name == "display" {
					subs[j].Writer = strs("admin")
					subs[j].Read = strs("admin")
				}
			}
			attrs[i].SubAttributes = subs
		}
	}
	cloned.Attributes = attrs
	Schemas[schemaUser] = cloned

	w := do(t, adminRouter, http.MethodPost, usersPath,
		`{"schemas":["`+schemaUser+`"],"userName":"rbac.user","emails":[{"type":"work","value":"r@b.com","display":"SecretTag"}]}`)
	requireStatus(t, w, http.StatusCreated)
	id := decode(t, w)["id"].(string)

	// Writing a $writer-restricted sub-attribute via a bracket path returns 403 Forbidden.
	w = do(t, userRouter, http.MethodPatch, usersPath+"/"+id, patchBody(map[string]interface{}{
		"op":    "replace",
		"path":  `emails[type eq "work"].display`,
		"value": "Tampered",
	}))
	requireSCIMError(t, w, http.StatusForbidden)

	// Probing a $reader-restricted sub-attribute inside the bracket filter returns 400 invalidFilter.
	w = do(t, userRouter, http.MethodPatch, usersPath+"/"+id, patchBody(map[string]interface{}{
		"op":    "replace",
		"path":  `emails[display eq "SecretTag"].value`,
		"value": "probe@b.com",
	}))
	errObj := requireSCIMError(t, w, http.StatusBadRequest)
	if errObj["scimType"] != "invalidFilter" {
		t.Errorf("scimType = %v, want invalidFilter", errObj["scimType"])
	}
}

func newRouterWithStore(t *testing.T, store *MemoryStore, auth Authenticator) http.Handler {
	t.Helper()
	DB = store
	r, _ := newTestServerAs(t, auth)
	// Restore DB to the shared store (newTestServerAs creates a fresh one).
	DB = store
	return r
}
