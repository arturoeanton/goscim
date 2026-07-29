package scim

import "strings"

// RFC 7643 2.1: attribute names are case-insensitive, and so are the schema
// URNs that identify them. Comparing them exactly meant a client sending
// "username" or an extension URN in a different case was told the attribute
// does not exist. Identity providers do both.
//
// Resolution is case-insensitive, but the declared spelling wins: a payload
// key is renamed to the name the schema uses before the resource is stored.
// Otherwise the document would carry whatever casing the client happened to
// send, and a later filter -- which asks the schema for the attribute path --
// would not find it.

// FindAttribute resolves an attribute by name, ignoring case.
func FindAttribute(attributes []Attribute, name string) (Attribute, bool) {
	for _, attribute := range attributes {
		if strings.EqualFold(attribute.Name, name) {
			return attribute, true
		}
	}
	return Attribute{}, false
}

// LookupSchema resolves a schema by URN, ignoring case.
func LookupSchema(id string) (Schema, bool) {
	if schema, ok := Schemas[id]; ok {
		return schema, true
	}
	for declared, schema := range Schemas {
		if strings.EqualFold(declared, id) {
			return schema, true
		}
	}
	return Schema{}, false
}

// canonicalSchemaID returns the URN as the server declares it, so a resource
// is stored with the canonical spelling whatever the client sent.
func canonicalSchemaID(id string) string {
	if schema, ok := LookupSchema(id); ok {
		return schema.ID
	}
	return id
}

// lookupKey finds the key of element that matches name ignoring case, and
// reports whether there was one.
func lookupKey(element map[string]interface{}, name string) (string, bool) {
	if _, ok := element[name]; ok {
		return name, true
	}
	for key := range element {
		if strings.EqualFold(key, name) {
			return key, true
		}
	}
	return "", false
}

// canonicaliseKey renames the entry matching name to exactly name. It returns
// the value and whether the attribute was present at all.
func canonicaliseKey(element map[string]interface{}, name string) (interface{}, bool) {
	key, present := lookupKey(element, name)
	if !present {
		return nil, false
	}
	value := element[key]
	if key != name {
		delete(element, key)
		element[name] = value
	}
	return value, true
}

// containsSchemaFold reports whether the resource type declares this extension,
// ignoring case.
func containsSchemaFold(extensions []SchemaExtension, id string) bool {
	for _, extension := range extensions {
		if strings.EqualFold(extension.Schema, id) {
			return true
		}
	}
	return false
}

// containsStringFold is ContainsString, ignoring case.
func containsStringFold(values []string, item string) bool {
	for _, value := range values {
		if strings.EqualFold(value, item) {
			return true
		}
	}
	return false
}
