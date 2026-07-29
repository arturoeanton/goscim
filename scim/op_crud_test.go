package scim

import (
	"encoding/json"
	"net/http"
	"testing"
)

// These tests drive the real SCIM router end to end against the in-memory
// store. They pin current behaviour so the rest of the release plan changes it
// deliberately and visibly: assertions covering known RFC violations are marked
// TODO(Bn) after the corresponding entry in RELEASE-1.0.md.

func TestCreateElement(t *testing.T) {
	r, store := newTestServer(t)

	body := createElement(t, r, "Element1", 1)

	id, ok := body["id"].(string)
	if !ok || id == "" {
		t.Fatalf("response carries no id: %v", body)
	}
	if body["name"] != "Element1" {
		t.Errorf("name = %v", body["name"])
	}
	// Every response goes through the same read filter, so an attribute the
	// caller may not read is absent from the create response too. description
	// declares $reader ["role2","role3"] and no caller role matches.
	if _, present := body["description"]; present {
		t.Errorf("description should not be readable: %v", body["description"])
	}

	meta, ok := body["meta"].(map[string]interface{})
	if !ok {
		t.Fatalf("meta missing or malformed: %v", body["meta"])
	}
	if meta["resourceType"] != "Element" {
		t.Errorf("meta.resourceType = %v", meta["resourceType"])
	}
	if meta["created"] == "" || meta["created"] != meta["lastModified"] {
		t.Errorf("created/lastModified inconsistent: %v", meta)
	}
	if meta["version"] == "" {
		t.Error("meta.version is empty")
	}
	if meta["location"] != PREFIX+"/Elements/"+id {
		t.Errorf("meta.location = %v", meta["location"])
	}

	// The document is stored in the bucket named after the resource type
	// ("Element"), not after the endpoint ("/Elements").
	stored, err := store.Get("Element", id)
	if err != nil {
		t.Fatalf("not persisted in the Element bucket: %v", err)
	}
	if stored["name"] != "Element1" {
		t.Errorf("stored document = %v", stored)
	}
}

func TestCreateElementRejectsInvalidPayloads(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{
			"no schemas",
			`{"name":"x","$ref":"/x"}`,
		},
		{
			"missing a required core attribute",
			`{"schemas":["` + schemaCore + `","` + schemaExt + `"],"name":"x","` + schemaExt + `":{"required":1}}`,
		},
		{
			"attribute not declared in the schema",
			`{"schemas":["` + schemaCore + `","` + schemaExt + `"],"name":"x","$ref":"/x","doesNotExist":1,"` + schemaExt + `":{"required":1}}`,
		},
		{
			"wrong type in the core schema",
			`{"schemas":["` + schemaCore + `","` + schemaExt + `"],"name":123,"$ref":"/x","` + schemaExt + `":{"required":1}}`,
		},
		{
			"wrong type in the extension",
			`{"schemas":["` + schemaCore + `","` + schemaExt + `"],"name":"x","$ref":"/x","` + schemaExt + `":{"required":"not-an-integer"}}`,
		},
		{
			"schema not declared by the resource type",
			`{"schemas":["` + schemaCore + `","urn:ietf:params:scim:schemas:extension:made-up:2.0:X"],"name":"x","$ref":"/x"}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := newTestServer(t)
			w := do(t, r, http.MethodPost, elementsPath, tc.body)
			requireSCIMError(t, w, http.StatusBadRequest)
		})
	}
}

func TestReadElement(t *testing.T) {
	r, _ := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)

	w := do(t, r, http.MethodGet, elementsPath+"/"+id, "")
	requireStatus(t, w, http.StatusOK)
	body := decode(t, w)

	if body["id"] != id {
		t.Errorf("id = %v", body["id"])
	}
	// name declares $reader ["role2","role1"] and the hardcoded roles include
	// role1, so the value is returned.
	if body["name"] != "Element1" {
		t.Errorf("name = %v", body["name"])
	}
	// description declares $reader ["role2","role3"]: no role matches, so the
	// key is omitted entirely rather than returned empty.
	if _, present := body["description"]; present {
		t.Errorf("description should have been omitted, got %v", body["description"])
	}
	// $ref declares no $reader, so it is not filtered.
	if body["$ref"] != "/Element1" {
		t.Errorf("$ref = %v", body["$ref"])
	}
}

func TestReadElementNotFound(t *testing.T) {
	r, _ := newTestServer(t)
	w := do(t, r, http.MethodGet, elementsPath+"/does-not-exist", "")
	requireStatus(t, w, http.StatusNotFound)
}

func TestReplaceElement(t *testing.T) {
	r, _ := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)
	originalMeta := created["meta"].(map[string]interface{})

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(validElement("Element1-modified", 7)), &payload); err != nil {
		t.Fatal(err)
	}
	payload["meta"] = originalMeta
	raw, _ := json.Marshal(payload)

	w := do(t, r, http.MethodPut, elementsPath+"/"+id, string(raw))
	requireStatus(t, w, http.StatusOK)
	body := decode(t, w)

	if body["id"] != id {
		t.Errorf("PUT changed the id: %v", body["id"])
	}
	if body["name"] != "Element1-modified" {
		t.Errorf("name = %v", body["name"])
	}
	meta := body["meta"].(map[string]interface{})
	if meta["created"] != originalMeta["created"] {
		t.Errorf("created not preserved: %v vs %v", meta["created"], originalMeta["created"])
	}
	if meta["version"] == originalMeta["version"] {
		t.Error("version did not change after the PUT")
	}
}

func TestReplaceElementNotFound(t *testing.T) {
	r, _ := newTestServer(t)
	created := createElement(t, r, "Element1", 1)

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(validElement("Element1", 1)), &payload); err != nil {
		t.Fatal(err)
	}
	payload["meta"] = created["meta"]
	raw, _ := json.Marshal(payload)

	w := do(t, r, http.MethodPut, elementsPath+"/does-not-exist", string(raw))
	requireStatus(t, w, http.StatusNotFound)
}

func TestPatchElement(t *testing.T) {
	r, store := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)

	patch := `{"schemas":["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
	           "Operations":[{"op":"replace","path":"description","value":"new description"}]}`
	w := do(t, r, http.MethodPatch, elementsPath+"/"+id, patch)
	requireStatus(t, w, http.StatusOK)

	stored, err := store.Get("Element", id)
	if err != nil {
		t.Fatal(err)
	}
	if stored["description"] != "new description" {
		t.Errorf("stored description = %v", stored["description"])
	}
	if stored["name"] != "Element1" {
		t.Errorf("the patch touched an attribute it should not have: %v", stored["name"])
	}
}

func TestPatchElementOnExtension(t *testing.T) {
	r, store := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)

	patch := `{"schemas":["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
	           "Operations":[{"op":"replace","path":"` + schemaExt + `.required","value":42}]}`
	w := do(t, r, http.MethodPatch, elementsPath+"/"+id, patch)
	requireStatus(t, w, http.StatusOK)

	stored, err := store.Get("Element", id)
	if err != nil {
		t.Fatal(err)
	}
	ext := stored[schemaExt].(map[string]interface{})
	if ext["required"] != float64(42) {
		t.Errorf("extension after the patch = %v", ext)
	}
}

func TestDeleteElement(t *testing.T) {
	r, _ := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)

	w := do(t, r, http.MethodDelete, elementsPath+"/"+id, "")
	requireStatus(t, w, http.StatusNoContent)
	if w.Body.Len() != 0 {
		t.Errorf("204 with a body: %s", w.Body.String())
	}

	w = do(t, r, http.MethodGet, elementsPath+"/"+id, "")
	requireStatus(t, w, http.StatusNotFound)
}

func TestDeleteElementNotFound(t *testing.T) {
	r, _ := newTestServer(t)
	w := do(t, r, http.MethodDelete, elementsPath+"/does-not-exist", "")
	requireStatus(t, w, http.StatusNotFound)
}

func TestSearchElements(t *testing.T) {
	r, _ := newTestServer(t)
	createElement(t, r, "AAA", 1)
	createElement(t, r, "BBB", 2)
	createElement(t, r, "CCC", 3)

	w := do(t, r, http.MethodGet, elementsPath+"?sortBy=name", "")
	requireStatus(t, w, http.StatusOK)
	body := decode(t, w)

	schemas := body["schemas"].([]interface{})
	if len(schemas) != 1 || schemas[0] != "urn:ietf:params:scim:api:messages:2.0:ListResponse" {
		t.Errorf("schemas = %v", schemas)
	}
	if body["totalResults"] != float64(3) {
		t.Errorf("totalResults = %v", body["totalResults"])
	}
	if body["startIndex"] != float64(1) {
		t.Errorf("startIndex = %v", body["startIndex"])
	}

	resources := body["Resources"].([]interface{})
	if len(resources) != 3 {
		t.Fatalf("returned %d resources", len(resources))
	}
	names := make([]string, 0, 3)
	for _, res := range resources {
		names = append(names, res.(map[string]interface{})["name"].(string))
	}
	if names[0] != "AAA" || names[1] != "BBB" || names[2] != "CCC" {
		t.Errorf("wrong ascending order: %v", names)
	}
}

func TestSearchElementsDescending(t *testing.T) {
	r, _ := newTestServer(t)
	createElement(t, r, "AAA", 1)
	createElement(t, r, "BBB", 2)

	w := do(t, r, http.MethodGet, elementsPath+"?sortBy=name&sortOrder=descending", "")
	requireStatus(t, w, http.StatusOK)
	resources := decode(t, w)["Resources"].([]interface{})
	if resources[0].(map[string]interface{})["name"] != "BBB" {
		t.Errorf("wrong descending order: %v", resources)
	}
}

func TestSearchElementsPagination(t *testing.T) {
	r, _ := newTestServer(t)
	createElement(t, r, "AAA", 1)
	createElement(t, r, "BBB", 2)
	createElement(t, r, "CCC", 3)

	w := do(t, r, http.MethodGet, elementsPath+"?sortBy=name&startIndex=2&count=2", "")
	requireStatus(t, w, http.StatusOK)
	body := decode(t, w)

	if body["totalResults"] != float64(3) {
		t.Errorf("totalResults must count the whole set, not the page: %v", body["totalResults"])
	}
	if body["startIndex"] != float64(2) {
		t.Errorf("startIndex = %v", body["startIndex"])
	}
	resources := body["Resources"].([]interface{})
	if len(resources) != 2 {
		t.Fatalf("want 2 resources, got %d", len(resources))
	}
	if resources[0].(map[string]interface{})["name"] != "BBB" {
		t.Errorf("the page starts at %v", resources[0])
	}
}

func TestSearchElementsMasksByRole(t *testing.T) {
	r, _ := newTestServer(t)
	createElement(t, r, "AAA", 1)

	w := do(t, r, http.MethodGet, elementsPath, "")
	requireStatus(t, w, http.StatusOK)
	resources := decode(t, w)["Resources"].([]interface{})
	first := resources[0].(map[string]interface{})
	if _, present := first["description"]; present {
		t.Errorf("description should have been omitted in search: %v", first["description"])
	}
	if first["name"] != "AAA" {
		t.Errorf("a readable attribute was dropped: %v", first)
	}
}

func TestSearchElementsInvalidParameters(t *testing.T) {
	cases := []string{"?startIndex=abc", "?count=abc"}
	for _, query := range cases {
		t.Run(query, func(t *testing.T) {
			r, _ := newTestServer(t)
			w := do(t, r, http.MethodGet, elementsPath+query, "")
			requireStatus(t, w, http.StatusBadRequest)
		})
	}
}

func TestDiscoveryNotImplemented(t *testing.T) {
	r, _ := newTestServer(t)
	// TODO(B9/improvement 8): all three must serve the real configuration and
	// hang off /scim/v2.
	for _, path := range []string{"/ServiceProviderConfig", "/ResourceTypes", "/Schemas"} {
		w := do(t, r, http.MethodGet, path, "")
		requireStatus(t, w, http.StatusNotImplemented)
	}
}

func TestRoutesRegisteredForEveryResourceType(t *testing.T) {
	r, _ := newTestServer(t)
	routes := make(map[string]bool)
	for _, info := range r.Routes() {
		routes[info.Method+" "+info.Path] = true
	}
	for _, endpoint := range []string{"/Users", "/Groups", "/Elements"} {
		for _, want := range []string{
			"POST " + PREFIX + endpoint,
			"GET " + PREFIX + endpoint,
			"GET " + PREFIX + endpoint + "/:id",
			"PUT " + PREFIX + endpoint + "/:id",
			"PATCH " + PREFIX + endpoint + "/:id",
			"DELETE " + PREFIX + endpoint + "/:id",
		} {
			if !routes[want] {
				t.Errorf("route not registered: %s", want)
			}
		}
	}
}
