package scim

import (
	"fmt"
	"strings"

	"github.com/arturoeanton/goscim/scim/parser"
)

// commonAttributes are the RFC 7643 3.1 attributes every resource carries and
// that no schema declares, so they must be accepted without being found in one.
var commonAttributes = map[string]bool{
	"id":                true,
	"externalId":        true,
	"meta":              true,
	"meta.resourceType": true,
	"meta.created":      true,
	"meta.lastModified": true,
	"meta.location":     true,
	"meta.version":      true,
}

// NormalizeSortBy validates a comma-separated sortBy against the attributes the
// resource type actually declares and returns the cleaned value.
//
// sortBy reaches the query as an identifier rather than a bound parameter, so
// an unchecked value is an injection vector: the quoting used to wrap the path
// in backticks without escaping the ones the client sent, letting it close the
// identifier and append arbitrary N1QL. Restricting the value to declared
// attribute names removes the vector at the source instead of trying to
// sanitise it.
func NormalizeSortBy(resourceType ResourceType, sortBy string) (string, error) {
	if strings.TrimSpace(sortBy) == "" {
		return "", nil
	}
	paths := strings.Split(sortBy, ",")
	cleaned := make([]string, 0, len(paths))
	for _, raw := range paths {
		path := strings.TrimSpace(raw)
		if path == "" {
			return "", fmt.Errorf("sortBy contains an empty attribute path")
		}
		if !attributeIsDeclared(resourceType, path) {
			return "", fmt.Errorf("sortBy: %q is not an attribute of %s", path, resourceType.Name)
		}
		cleaned = append(cleaned, path)
	}
	return strings.Join(cleaned, ","), nil
}

// attributeIsDeclared reports whether path names an attribute of the resource
// type's core schema, of one of its declared extensions, or a common attribute.
func attributeIsDeclared(resourceType ResourceType, path string) bool {
	urn, attrPath := parser.SplitURNPath(path)
	if attrPath == "" {
		return false
	}
	if urn == "" {
		if commonAttributes[attrPath] {
			return true
		}
		return declaresAttribute(resourceType.Schema, attrPath)
	}
	// A URN prefix is only acceptable if the resource type actually uses that
	// schema; otherwise a client could name any schema loaded on the server.
	if urn != resourceType.Schema && !ContainsSchemaExtension(resourceType.SchemaExtensions, urn) {
		return false
	}
	return declaresAttribute(urn, attrPath)
}

func declaresAttribute(schemaID, attrPath string) bool {
	schema, ok := Schemas[schemaID]
	if !ok {
		return false
	}
	return GetAttribute(schema.Attributes, attrPath).Name != ""
}
