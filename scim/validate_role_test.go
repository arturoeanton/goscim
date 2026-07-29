package scim

import (
	"testing"
)

func strs(values ...string) *[]string { return &values }

// The read filter used to walk only nested objects, so anything inside a
// multi-valued attribute was returned untouched no matter what $reader said.
func TestReadFilterRecursesIntoArrays(t *testing.T) {
	attributes := []Attribute{
		{
			Name:        "contacts",
			Type:        "complex",
			MultiValued: true,
			SubAttributes: []Attribute{
				{Name: "label", Type: "string"},
				{Name: "secret", Type: "string", Read: strs("privileged")},
			},
		},
	}
	element := map[string]interface{}{
		"contacts": []interface{}{
			map[string]interface{}{"label": "work", "secret": "classified-1"},
			map[string]interface{}{"label": "home", "secret": "classified-2"},
		},
	}

	filtered := filterAttributes([]string{"user"}, attributes, element)

	contacts, ok := filtered["contacts"].([]interface{})
	if !ok || len(contacts) != 2 {
		t.Fatalf("contacts = %#v", filtered["contacts"])
	}
	for i, raw := range contacts {
		item := raw.(map[string]interface{})
		if _, present := item["secret"]; present {
			t.Errorf("contacts[%d] leaked the restricted sub-attribute: %v", i, item)
		}
		if item["label"] == nil {
			t.Errorf("contacts[%d] lost the readable sub-attribute: %v", i, item)
		}
	}
}

// A caller holding a listed role keeps the attribute.
func TestReadFilterKeepsPermittedAttributes(t *testing.T) {
	attributes := []Attribute{
		{Name: "open", Type: "string"},
		{Name: "restricted", Type: "string", Read: strs("role2", "role1")},
		{Name: "wildcard", Type: "string", Read: strs("*")},
		{Name: "forbidden", Type: "string", Read: strs("nobody")},
	}
	element := map[string]interface{}{
		"open":       "a",
		"restricted": "b",
		"wildcard":   "c",
		"forbidden":  "d",
		"undeclared": "e",
	}

	filtered := filterAttributes([]string{"role1"}, attributes, element)

	for _, key := range []string{"open", "restricted", "wildcard", "undeclared"} {
		if _, present := filtered[key]; !present {
			t.Errorf("%s should have been kept", key)
		}
	}
	if _, present := filtered["forbidden"]; present {
		t.Errorf("forbidden should have been dropped: %v", filtered)
	}
}

// Nested complex attributes are filtered at every level.
func TestReadFilterRecursesIntoComplexValues(t *testing.T) {
	attributes := []Attribute{
		{
			Name: "manager",
			Type: "complex",
			SubAttributes: []Attribute{
				{Name: "displayName", Type: "string"},
				{Name: "salary", Type: "integer", Read: strs("hr")},
			},
		},
	}
	element := map[string]interface{}{
		"manager": map[string]interface{}{"displayName": "Ada", "salary": float64(1)},
	}

	filtered := filterAttributes([]string{"user"}, attributes, element)
	manager := filtered["manager"].(map[string]interface{})
	if _, present := manager["salary"]; present {
		t.Errorf("salary leaked: %v", manager)
	}
	if manager["displayName"] != "Ada" {
		t.Errorf("displayName = %v", manager["displayName"])
	}
}

// Extension attributes hang off the resource under their schema URN and have to
// be resolved against that schema, not the core one. They used to be skipped
// entirely, so a $reader declared on an extension attribute did nothing.
func TestReadFilterAppliesToExtensionAttributes(t *testing.T) {
	newTestServer(t) // loads the shipped schemas into the package globals

	const extension = "urn:test:extension"
	Schemas[extension] = Schema{
		ID: extension,
		Attributes: []Attribute{
			{Name: "open", Type: "string"},
			{Name: "restricted", Type: "string", Read: strs("privileged")},
		},
	}
	resourceType := ResourceType{
		Name:             "Thing",
		Schema:           "urn:test:core",
		SchemaExtensions: []SchemaExtension{{Schema: extension}},
	}
	Schemas["urn:test:core"] = Schema{
		ID:         "urn:test:core",
		Attributes: []Attribute{{Name: "name", Type: "string"}},
	}

	element := map[string]interface{}{
		"id":   "1",
		"name": "thing",
		extension: map[string]interface{}{
			"open":       "visible",
			"restricted": "classified",
		},
	}

	filtered := ValidateReadRole([]string{"user"}, resourceType, element)

	ext, ok := filtered[extension].(map[string]interface{})
	if !ok {
		t.Fatalf("the extension object is gone: %v", filtered)
	}
	if _, present := ext["restricted"]; present {
		t.Errorf("the extension leaked a restricted attribute: %v", ext)
	}
	if ext["open"] != "visible" {
		t.Errorf("the extension lost a readable attribute: %v", ext)
	}
	if filtered["id"] != "1" {
		t.Errorf("common attributes must survive: %v", filtered)
	}
}

// Filtering must not mutate the document it was handed, which in the search
// path comes straight from the store.
func TestReadFilterDoesNotMutateItsInput(t *testing.T) {
	attributes := []Attribute{{Name: "secret", Type: "string", Read: strs("nobody")}}
	element := map[string]interface{}{"secret": "value"}

	filterAttributes([]string{"user"}, attributes, element)

	if element["secret"] != "value" {
		t.Errorf("the input was modified: %v", element)
	}
}

func TestReadFilterHandlesNil(t *testing.T) {
	if got := ValidateReadRole(nil, ResourceType{}, nil); got != nil {
		t.Errorf("ValidateReadRole(nil) = %v", got)
	}
}
