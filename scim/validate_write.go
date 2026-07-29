package scim

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// EnforceWriteAccess applies the write-side rules to an incoming resource:
// SCIM's own mutability keyword (RFC 7643 7) and this project's $writer role
// lists, the mirror of $reader.
//
// existing is the stored resource on an update and nil on a create. It reports
// whether the caller may proceed; when it returns false the response has
// already been written.
func EnforceWriteAccess(c *gin.Context, resourceType ResoruceType, incoming, existing map[string]interface{}) bool {
	roles := currentRoles(c)

	if !enforceAttributeWrites(c, roles, Schemas[resourceType.Schema].Attributes, incoming, existing) {
		return false
	}
	for _, extension := range resourceType.SchemaExtensions {
		values, ok := incoming[extension.Schema].(map[string]interface{})
		if !ok {
			continue
		}
		storedValues, _ := existing[extension.Schema].(map[string]interface{})
		if !enforceAttributeWrites(c, roles, Schemas[extension.Schema].Attributes, values, storedValues) {
			return false
		}
	}
	return true
}

func enforceAttributeWrites(c *gin.Context, roles []string, attributes []Attribute, incoming, existing map[string]interface{}) bool {
	for _, attribute := range attributes {
		value, present := incoming[attribute.Name]

		if isReadOnly(attribute) {
			// A read-only attribute is server-owned. The client's value is
			// dropped rather than refused, because read-modify-write clients
			// echo the whole resource back on a PUT and rejecting that would
			// make an ordinary update impossible.
			delete(incoming, attribute.Name)
			if stored, ok := existing[attribute.Name]; ok {
				incoming[attribute.Name] = stored
			}
			continue
		}
		if !present {
			continue
		}
		if !canWrite(roles, attribute) {
			MakeError(c, http.StatusForbidden,
				"the caller is not allowed to write "+attribute.Name)
			return false
		}
		if existing != nil && strings.EqualFold(attribute.Mutability, "immutable") {
			if stored, ok := existing[attribute.Name]; ok && !reflect.DeepEqual(stored, value) {
				MakeTypedError(c, http.StatusBadRequest, "mutability",
					attribute.Name+" is immutable and cannot be changed")
				return false
			}
		}
		if !enforceNestedWrites(c, roles, attribute, value, existing[attribute.Name]) {
			return false
		}
	}
	return true
}

// enforceNestedWrites applies the same rules to the sub-attributes of a complex
// value, and to every element of a multi-valued one.
func enforceNestedWrites(c *gin.Context, roles []string, attribute Attribute, value, existing interface{}) bool {
	if len(attribute.SubAttributes) == 0 {
		return true
	}
	switch typed := value.(type) {
	case map[string]interface{}:
		storedValues, _ := existing.(map[string]interface{})
		return enforceAttributeWrites(c, roles, attribute.SubAttributes, typed, storedValues)
	case []interface{}:
		for _, item := range typed {
			values, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			// Elements of a multi-valued attribute have no stable identity
			// here, so there is nothing to compare an immutable sub-attribute
			// against; the role and read-only rules still apply.
			if !enforceAttributeWrites(c, roles, attribute.SubAttributes, values, nil) {
				return false
			}
		}
	}
	return true
}

// canWrite reports whether any of the caller's roles appears in the attribute's
// $writer list, with "*" meaning any role.
func canWrite(roles []string, attribute Attribute) bool {
	if attribute.Writer == nil {
		return true
	}
	for _, allowed := range *attribute.Writer {
		if allowed == "*" {
			return true
		}
		for _, role := range roles {
			if allowed == role {
				return true
			}
		}
	}
	return false
}

func isReadOnly(attribute Attribute) bool {
	return strings.EqualFold(attribute.Mutability, "readOnly")
}

// isReturnable reports whether an attribute may appear in a response at all,
// before any role check. RFC 7643 7 defines "returned": "never" for values the
// service provider must not disclose, and a writeOnly attribute -- a password
// being the archetype -- is never returned either.
func isReturnable(attribute Attribute) bool {
	if strings.EqualFold(attribute.Returned, "never") {
		return false
	}
	return !strings.EqualFold(attribute.Mutability, "writeOnly")
}
