package scim

import (
	"fmt"
	"strings"

	"github.com/arturoeanton/goscim/scim/parser"
)

// commonAttributes are the RFC 7643 3.1 attributes every resource carries and
// that no schema declares, so they must be accepted without being found in one.
// Keys are lower-cased: RFC 7643 2.1 makes attribute names case-insensitive.
var commonAttributes = map[string]bool{
	"id":                true,
	"externalid":        true,
	"meta":              true,
	"meta.resourcetype": true,
	"meta.created":      true,
	"meta.lastmodified": true,
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
		if !attributeIsSortable(resourceType, path) {
			return "", fmt.Errorf("sortBy: %q cannot be used for sorting", path)
		}
		cleaned = append(cleaned, CanonicalAttributePath(resourceType, path))
	}
	return strings.Join(cleaned, ","), nil
}

// CanonicalAttributePath rewrites an attribute path with the spelling the
// schema declares, leaving it untouched when nothing matches.
//
// Attribute names are case-insensitive to a SCIM client but not to N1QL, so a
// filter or sort naming "USERNAME" has to become `userName` before it reaches
// the query, or it names a key no stored document has.
func CanonicalAttributePath(resourceType ResourceType, path string) string {
	urn, attrPath := parser.SplitURNPath(path)

	schemaID := resourceType.Schema
	prefix := ""
	if urn != "" {
		schemaID = urn
		prefix = canonicalSchemaID(urn) + "."
	}
	schema, ok := LookupSchema(schemaID)
	if !ok {
		return path
	}

	segments := strings.Split(attrPath, ".")
	attributes := schema.Attributes
	for i, segment := range segments {
		attribute, found := FindAttribute(attributes, segment)
		if !found {
			return path
		}
		segments[i] = attribute.Name
		attributes = attribute.SubAttributes
	}
	return prefix + strings.Join(segments, ".")
}

// attributeIsDeclared reports whether path names an attribute of the resource
// type's core schema, of one of its declared extensions, or a common attribute.
func attributeIsDeclared(resourceType ResourceType, path string) bool {
	urn, attrPath := parser.SplitURNPath(path)
	if attrPath == "" {
		return false
	}
	if urn == "" {
		if commonAttributes[strings.ToLower(attrPath)] {
			return true
		}
		return declaresAttribute(resourceType.Schema, attrPath)
	}
	// A URN prefix is only acceptable if the resource type actually uses that
	// schema; otherwise a client could name any schema loaded on the server.
	if !strings.EqualFold(urn, resourceType.Schema) && !containsSchemaFold(resourceType.SchemaExtensions, urn) {
		return false
	}
	return declaresAttribute(urn, attrPath)
}

func declaresAttribute(schemaID, attrPath string) bool {
	schema, ok := LookupSchema(schemaID)
	if !ok {
		return false
	}
	return GetAttribute(schema.Attributes, attrPath).Name != ""
}

// attributeIsSortable refuses writeOnly and returned:"never" attributes (such
// as User.password), whose ordering would otherwise disclose secret values.
func attributeIsSortable(resourceType ResourceType, path string) bool {
	for _, attr := range resolveAttributeChain(resourceType, path) {
		if !isReturnable(attr) {
			return false
		}
	}
	return true
}

// ValidateFilterAttributes ensures a SCIM filter cannot be used as a blind
// oracle against writeOnly / returned:"never" secrets (like password) or
// attributes hidden from the caller by $reader.
func ValidateFilterAttributes(resourceType ResourceType, roles []string, filter string) error {
	if filter == "" {
		return nil
	}
	paths, err := parser.ReferencedAttributes(filter)
	if err != nil {
		return err
	}
	for _, path := range paths {
		urn, attrPath := parser.SplitURNPath(path)
		if attrPath == "" {
			return fmt.Errorf("invalid attribute path %q in filter", path)
		}
		if urn != "" && !strings.EqualFold(urn, resourceType.Schema) &&
			!containsSchemaFold(resourceType.SchemaExtensions, urn) {
			return fmt.Errorf("schema %q is not part of %s", urn, resourceType.Name)
		}
		for _, attr := range resolveAttributeChain(resourceType, path) {
			if !isReturnable(attr) {
				return fmt.Errorf("attribute %q cannot be filtered", attr.Name)
			}
			if !canRead(roles, attr) {
				return fmt.Errorf("the caller is not allowed to filter by %q", attr.Name)
			}
		}
	}
	return nil
}

func resolveAttributeChain(resourceType ResourceType, path string) []Attribute {
	urn, attrPath := parser.SplitURNPath(path)
	if urn == "" && commonAttributes[strings.ToLower(attrPath)] {
		return nil
	}
	schemaID := resourceType.Schema
	if urn != "" {
		schemaID = urn
	}
	schema, ok := LookupSchema(schemaID)
	if !ok {
		return nil
	}
	segments := strings.Split(attrPath, ".")
	attributes := schema.Attributes
	chain := make([]Attribute, 0, len(segments))
	for _, segment := range segments {
		attr, found := FindAttribute(attributes, segment)
		if !found {
			break
		}
		chain = append(chain, attr)
		attributes = attr.SubAttributes
	}
	return chain
}

// ValidateRootSortBy validates sortBy on a server-root POST /.search: every
// path must be declared by at least one resource type and must be sortable on
// every resource type that declares it.
func ValidateRootSortBy(resourceTypes []ResourceType, sortBy string) error {
	if strings.TrimSpace(sortBy) == "" {
		return nil
	}
	for _, raw := range strings.Split(sortBy, ",") {
		path := strings.TrimSpace(raw)
		if path == "" {
			return fmt.Errorf("sortBy contains an empty attribute path")
		}
		declaredAnywhere := false
		for _, rt := range resourceTypes {
			if attributeIsDeclared(rt, path) {
				declaredAnywhere = true
				if !attributeIsSortable(rt, path) {
					return fmt.Errorf("sortBy: %q cannot be used for sorting", path)
				}
			}
		}
		if !declaredAnywhere {
			return fmt.Errorf("sortBy: %q is not a declared attribute", path)
		}
	}
	return nil
}

// ValidateRootFilterAttributes validates a filter on a server-root POST
// /.search: every referenced attribute must be declared by at least one
// resource type, and on every resource type that declares it the attribute must
// be returnable and readable by the caller's roles.
func ValidateRootFilterAttributes(resourceTypes []ResourceType, roles []string, filter string) error {
	if filter == "" {
		return nil
	}
	paths, err := parser.ReferencedAttributes(filter)
	if err != nil {
		return err
	}
	for _, path := range paths {
		_, attrPath := parser.SplitURNPath(path)
		if attrPath == "" {
			return fmt.Errorf("invalid attribute path %q in filter", path)
		}
		declaredAnywhere := false
		for _, rt := range resourceTypes {
			if !attributeIsDeclared(rt, path) {
				continue
			}
			declaredAnywhere = true
			for _, attr := range resolveAttributeChain(rt, path) {
				if !isReturnable(attr) {
					return fmt.Errorf("attribute %q cannot be filtered", attr.Name)
				}
				if !canRead(roles, attr) {
					return fmt.Errorf("the caller is not allowed to filter by %q", attr.Name)
				}
			}
		}
		if !declaredAnywhere {
			return fmt.Errorf("attribute %q is not declared", path)
		}
	}
	return nil
}
