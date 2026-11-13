package scim

import (
	"net/http"
	"testing"
)

const enterpriseUserExt = "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"

func TestAttributeProjection(t *testing.T) {
	r, store := newTestServerAs(t, &AnonymousAuthenticator{Roles: []string{"role1"}})

	userBody := `{
		"schemas": ["` + schemaUser + `", "` + enterpriseUserExt + `"],
		"userName": "jane.doe",
		"displayName": "Jane Doe",
		"active": true,
		"password": "secret-password",
		"name": {"givenName": "Jane", "familyName": "Doe"},
		"emails": [
			{"value": "jane@work.example", "type": "work", "primary": true},
			{"value": "jane@home.example", "type": "home"}
		],
		"` + enterpriseUserExt + `": {
			"department": "Engineering",
			"costCenter": "CC-42"
		}
	}`
	w := do(t, r, http.MethodPost, usersPath, userBody)
	requireStatus(t, w, http.StatusCreated)
	id := decode(t, w)["id"].(string)

	t.Run("attributes selects top-level, nested, multi-valued sub-attributes and colon-separated extension paths", func(t *testing.T) {
		query := "attributes=USERNAME,name.givenName,emails.value,meta.resourceType," + enterpriseUserExt + ":department"
		w := do(t, r, http.MethodGet, usersPath+"/"+id+"?"+query, "")
		requireStatus(t, w, http.StatusOK)
		got := decode(t, w)

		if got["id"] != id {
			t.Errorf("id = %v, want %s", got["id"], id)
		}
		if got["userName"] != "jane.doe" {
			t.Errorf("userName = %v", got["userName"])
		}
		if _, present := got["displayName"]; present {
			t.Errorf("unrequested displayName returned: %v", got["displayName"])
		}
		if _, present := got["active"]; present {
			t.Errorf("unrequested active returned: %v", got["active"])
		}

		nameMap := got["name"].(map[string]interface{})
		if nameMap["givenName"] != "Jane" || nameMap["familyName"] != nil {
			t.Errorf("name = %v, want givenName only", nameMap)
		}

		emails := got["emails"].([]interface{})
		if len(emails) != 2 {
			t.Fatalf("len(emails) = %d, want 2", len(emails))
		}
		firstEmail := emails[0].(map[string]interface{})
		if firstEmail["value"] != "jane@work.example" || firstEmail["type"] != nil {
			t.Errorf("emails[0] = %v, want value only", firstEmail)
		}

		meta := got["meta"].(map[string]interface{})
		if meta["resourceType"] != "User" || meta["location"] != nil {
			t.Errorf("meta = %v, want resourceType only", meta)
		}

		ext := got[enterpriseUserExt].(map[string]interface{})
		if ext["department"] != "Engineering" || ext["costCenter"] != nil {
			t.Errorf("enterprise extension = %v, want department only", ext)
		}
	})

	t.Run("omitting extension attributes drops the extension object and trims schemas", func(t *testing.T) {
		w := do(t, r, http.MethodGet, usersPath+"/"+id+"?attributes=urn:ietf:params:scim:schemas:core:2.0:User:userName", "")
		requireStatus(t, w, http.StatusOK)
		got := decode(t, w)

		if _, present := got[enterpriseUserExt]; present {
			t.Errorf("unrequested extension returned: %v", got[enterpriseUserExt])
		}
		if _, present := got["meta"]; present {
			t.Errorf("unrequested meta returned: %v", got["meta"])
		}
		schemas := got["schemas"].([]interface{})
		if len(schemas) != 1 || schemas[0] != schemaUser {
			t.Errorf("schemas = %v, want [%s]", schemas, schemaUser)
		}
	})

	t.Run("excludedAttributes drops fields, sub-attributes and whole extensions while keeping id and schemas", func(t *testing.T) {
		query := "excludedAttributes=id,schemas,displayName,emails.type,meta.version," + enterpriseUserExt
		w := do(t, r, http.MethodGet, usersPath+"/"+id+"?"+query, "")
		requireStatus(t, w, http.StatusOK)
		got := decode(t, w)

		// RFC 7643 3.1: id and schemas are returned:always and cannot be excluded.
		if got["id"] != id {
			t.Errorf("id was excluded: %v", got)
		}
		if _, present := got["displayName"]; present {
			t.Errorf("excluded displayName survived: %v", got["displayName"])
		}
		if _, present := got[enterpriseUserExt]; present {
			t.Errorf("excluded extension survived: %v", got[enterpriseUserExt])
		}
		firstEmail := got["emails"].([]interface{})[0].(map[string]interface{})
		if _, present := firstEmail["type"]; present {
			t.Errorf("excluded emails.type survived: %v", firstEmail)
		}
		if firstEmail["value"] != "jane@work.example" {
			t.Errorf("emails.value missing: %v", firstEmail)
		}
		meta := got["meta"].(map[string]interface{})
		if _, present := meta["version"]; present {
			t.Errorf("excluded meta.version survived: %v", meta)
		}
		if meta["resourceType"] != "User" {
			t.Errorf("meta.resourceType missing: %v", meta)
		}
	})

	t.Run("attributes cannot expose writeOnly passwords or $reader-hidden attributes", func(t *testing.T) {
		w := do(t, r, http.MethodGet, usersPath+"/"+id+"?attributes=userName,password", "")
		requireStatus(t, w, http.StatusOK)
		if _, present := decode(t, w)["password"]; present {
			t.Fatal("attributes=password disclosed password")
		}

		// Element.description requires role2 or role3; caller only holds role1.
		elem := createElement(t, r, "ElemSec", 1)
		elemID := elem["id"].(string)
		// Store a description directly to prove $reader hides it even when requested.
		doc, _ := store.Get("Element", elemID)
		doc["description"] = "classified"
		_ = store.Replace("Element", elemID, doc)

		w = do(t, r, http.MethodGet, elementsPath+"/"+elemID+"?attributes=name,description", "")
		requireStatus(t, w, http.StatusOK)
		if _, present := decode(t, w)["description"]; present {
			t.Fatal("attributes=description bypassed $reader")
		}
	})

	t.Run("returned=always and returned=request are honoured", func(t *testing.T) {
		schema := Schemas[schemaUser]
		orig := append([]Attribute(nil), schema.Attributes...)
		t.Cleanup(func() {
			s := Schemas[schemaUser]
			s.Attributes = orig
			Schemas[schemaUser] = s
		})
		modified := append([]Attribute(nil), schema.Attributes...)
		for i := range modified {
			switch modified[i].Name {
			case "userName":
				modified[i].Returned = "always"
			case "displayName":
				modified[i].Returned = "request"
			}
		}
		schema.Attributes = modified
		Schemas[schemaUser] = schema

		// Default read omits returned:request (displayName) and keeps returned:always (userName).
		got := decode(t, do(t, r, http.MethodGet, usersPath+"/"+id, ""))
		if _, present := got["displayName"]; present {
			t.Errorf("returned:request attribute returned by default: %v", got["displayName"])
		}
		if got["userName"] != "jane.doe" {
			t.Errorf("userName missing: %v", got)
		}

		// Explicitly requesting displayName returns it, AND userName (always) stays too!
		got = decode(t, do(t, r, http.MethodGet, usersPath+"/"+id+"?attributes=displayName", ""))
		if got["displayName"] != "Jane Doe" {
			t.Errorf("explicitly requested displayName missing: %v", got)
		}
		if got["userName"] != "jane.doe" {
			t.Errorf("returned:always userName omitted under attributes=displayName: %v", got)
		}

		// Attempting to exclude a returned:always attribute keeps it anyway.
		got = decode(t, do(t, r, http.MethodGet, usersPath+"/"+id+"?excludedAttributes=userName", ""))
		if got["userName"] != "jane.doe" {
			t.Errorf("returned:always userName excluded: %v", got)
		}
	})

	t.Run("attributes and excludedAttributes together return 400 before mutating state", func(t *testing.T) {
		beforeTotal, _, _ := store.Search(SearchQuery{Bucket: "User", Limit: 50})

		w := do(t, r, http.MethodPost, usersPath+"?attributes=userName&excludedAttributes=active",
			`{"schemas":["`+schemaUser+`"],"userName":"never-created"}`)
		out := requireSCIMError(t, w, http.StatusBadRequest)
		if out["scimType"] != "invalidValue" {
			t.Errorf("scimType = %v, want invalidValue", out["scimType"])
		}

		afterTotal, _, _ := store.Search(SearchQuery{Bucket: "User", Limit: 50})
		if afterTotal != beforeTotal {
			t.Fatalf("failed query parameter check still created a resource (%d -> %d)", beforeTotal, afterTotal)
		}
	})

	t.Run("projection works on POST, PUT, PATCH and GET list", func(t *testing.T) {
		w := do(t, r, http.MethodPost, usersPath+"?attributes=userName",
			`{"schemas":["`+schemaUser+`"],"userName":"bob","displayName":"Bob"}`)
		requireStatus(t, w, http.StatusCreated)
		created := decode(t, w)
		if created["userName"] != "bob" || created["displayName"] != nil {
			t.Errorf("POST projection = %v", created)
		}
		bobID := created["id"].(string)

		w = do(t, r, http.MethodPut, usersPath+"/"+bobID+"?excludedAttributes=displayName",
			`{"schemas":["`+schemaUser+`"],"userName":"bob","displayName":"Bobby"}`)
		requireStatus(t, w, http.StatusOK)
		if _, present := decode(t, w)["displayName"]; present {
			t.Errorf("PUT excludedAttributes failed")
		}

		w = do(t, r, http.MethodPatch, usersPath+"/"+bobID+"?attributes=displayName",
			patchBody(map[string]interface{}{"op": "replace", "path": "displayName", "value": "Robert"}))
		requireStatus(t, w, http.StatusOK)
		patched := decode(t, w)
		if patched["displayName"] != "Robert" || patched["userName"] != nil {
			t.Errorf("PATCH projection = %v", patched)
		}

		w = do(t, r, http.MethodGet, usersPath+"?attributes=userName", "")
		requireStatus(t, w, http.StatusOK)
		for _, item := range decode(t, w)["Resources"].([]interface{}) {
			m := item.(map[string]interface{})
			if _, present := m["displayName"]; present {
				t.Errorf("GET list projection leaked displayName: %v", m)
			}
		}
	})
}

func TestPostSearchEndpoints(t *testing.T) {
	t.Run("authentication is required on resource and root /.search", func(t *testing.T) {
		r, _ := newTestServerAs(t, &BasicAuthenticator{})
		body := `{"schemas":["` + searchRequestSchema + `"]}`
		requireSCIMError(t, do(t, r, http.MethodPost, usersPath+"/.search", body), http.StatusUnauthorized)
		requireSCIMError(t, do(t, r, http.MethodPost, PREFIX+"/.search", body), http.StatusUnauthorized)
	})

	r, _ := newTestServerAs(t, &AnonymousAuthenticator{Roles: []string{"role1"}})
	for _, uname := range []string{"charlie", "alice", "bob"} {
		requireStatus(t, do(t, r, http.MethodPost, usersPath,
			`{"schemas":["`+schemaUser+`"],"userName":"`+uname+`","displayName":"Display `+uname+`","password":"pw-`+uname+`"}`),
			http.StatusCreated)
	}
	requireStatus(t, do(t, r, http.MethodPost, PREFIX+"/Groups",
		`{"schemas":["urn:ietf:params:scim:schemas:core:2.0:Group"],"displayName":"Admins"}`),
		http.StatusCreated)
	createElement(t, r, "Elem1", 1)

	t.Run("resource POST /.search sorts, paginates and projects attributes", func(t *testing.T) {
		body := `{
			"schemas": ["` + searchRequestSchema + `"],
			"attributes": ["userName"],
			"sortBy": "userName",
			"sortOrder": "descending",
			"startIndex": 2,
			"count": 1
		}`
		w := do(t, r, http.MethodPost, usersPath+"/.search", body)
		requireStatus(t, w, http.StatusOK)
		list := decode(t, w)

		if list["totalResults"] != float64(3) || list["itemsPerPage"] != float64(1) || list["startIndex"] != float64(2) {
			t.Fatalf("unexpected ListResponse envelope: %v", list)
		}
		res := list["Resources"].([]interface{})[0].(map[string]interface{})
		// Descending of [alice, bob, charlie] -> index 2 is bob.
		if res["userName"] != "bob" {
			t.Errorf("userName = %v, want bob", res["userName"])
		}
		if _, present := res["displayName"]; present {
			t.Errorf("displayName should have been omitted by attributes: %v", res)
		}
	})

	t.Run("root POST /scim/v2/.search aggregates across all resource types", func(t *testing.T) {
		body := `{
			"schemas": ["` + searchRequestSchema + `"],
			"attributes": ["userName", "displayName", "name"],
			"sortBy": "meta.resourceType",
			"sortOrder": "ascending",
			"startIndex": 1,
			"count": 10
		}`
		w := do(t, r, http.MethodPost, PREFIX+"/.search", body)
		requireStatus(t, w, http.StatusOK)
		list := decode(t, w)

		// 3 Users + 1 Group + 1 Element = 5 total resources.
		if list["totalResults"] != float64(5) || list["itemsPerPage"] != float64(5) {
			t.Fatalf("root search counts = %v", list)
		}
		items := list["Resources"].([]interface{})
		// First item sorted ascending by meta.resourceType is Element ("Element" < "Group" < "User").
		if items[0].(map[string]interface{})["name"] != "Elem1" {
			t.Errorf("first item = %v, want Elem1", items[0])
		}
	})

	t.Run("validation rejects bad schemas, bad sortOrder, secret filters and mutually exclusive projection", func(t *testing.T) {
		badCases := []struct {
			name     string
			endpoint string
			body     string
			scimType string
		}{
			{
				"missing SearchRequest schema",
				usersPath + "/.search",
				`{"schemas":["urn:wrong"],"count":10}`,
				"invalidSyntax",
			},
			{
				"invalid sortOrder",
				usersPath + "/.search",
				`{"schemas":["` + searchRequestSchema + `"],"sortOrder":"sideways"}`,
				"invalidValue",
			},
			{
				"attributes and excludedAttributes together",
				usersPath + "/.search",
				`{"schemas":["` + searchRequestSchema + `"],"attributes":["userName"],"excludedAttributes":["active"]}`,
				"invalidValue",
			},
			{
				"filtering on password on root /.search",
				PREFIX + "/.search",
				`{"schemas":["` + searchRequestSchema + `"],"filter":"password pr"}`,
				"invalidFilter",
			},
			{
				"sorting on password on root /.search",
				PREFIX + "/.search",
				`{"schemas":["` + searchRequestSchema + `"],"sortBy":"password"}`,
				"invalidValue",
			},
			{
				"filtering on $reader-restricted attribute on root /.search",
				PREFIX + "/.search",
				`{"schemas":["` + searchRequestSchema + `"],"filter":"description pr"}`,
				"invalidFilter",
			},
		}

		for _, tc := range badCases {
			t.Run(tc.name, func(t *testing.T) {
				out := requireSCIMError(t, do(t, r, http.MethodPost, tc.endpoint, tc.body), http.StatusBadRequest)
				if out["scimType"] != tc.scimType {
					t.Errorf("scimType = %v, want %s", out["scimType"], tc.scimType)
				}
			})
		}
	})
}
