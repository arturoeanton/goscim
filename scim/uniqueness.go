package scim

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// EnforceUniqueness refuses a write that would duplicate a value the schema
// declares unique (RFC 7643 7, "uniqueness": "server" or "global").
//
// currentID is the resource being replaced, or "" on a create: a resource is
// allowed to keep its own value.
//
// Without this, an identity provider that retries a provisioning request --
// because the first attempt timed out, say -- silently creates a second user
// with the same userName, and every later lookup by that name is ambiguous.
//
// It reports whether the caller may proceed; when it returns false the
// response has already been written.
func EnforceUniqueness(c *gin.Context, resourceType ResourceType, element map[string]interface{}, currentID string) bool {
	if !checkUniqueAttributes(c, resourceType, Schemas[resourceType.Schema].Attributes, element, "", currentID) {
		return false
	}
	for _, extension := range resourceType.SchemaExtensions {
		values, ok := element[extension.Schema].(map[string]interface{})
		if !ok {
			continue
		}
		if !checkUniqueAttributes(c, resourceType, Schemas[extension.Schema].Attributes, values, extension.Schema+".", currentID) {
			return false
		}
	}
	return true
}

// checkUniqueAttributes looks up every unique attribute present in element.
//
// prefix carries the extension URN so the lookup path matches how the value is
// stored, nested under that key.
func checkUniqueAttributes(c *gin.Context, resourceType ResourceType, attributes []Attribute,
	element map[string]interface{}, prefix, currentID string) bool {

	for _, attribute := range attributes {
		if !isUniqueAttribute(attribute) {
			continue
		}
		value, present := element[attribute.Name]
		if !present || value == nil {
			continue
		}

		path := prefix + attribute.Name
		existingID, err := DB.FindIDByAttribute(resourceType.Name, path, value)
		if err != nil {
			MakeError(c, http.StatusInternalServerError, err.Error())
			log.Println(err.Error())
			return false
		}
		if existingID != "" && existingID != currentID {
			MakeTypedError(c, http.StatusConflict, "uniqueness",
				"a "+resourceType.Name+" with this "+path+" already exists")
			return false
		}
	}
	return true
}

// isUniqueAttribute reports whether the schema asks for this attribute to be
// unique. Multi-valued attributes are skipped: uniqueness across the elements
// of an array is a different question, and nothing in the shipped schemas asks
// for it.
func isUniqueAttribute(attribute Attribute) bool {
	if attribute.MultiValued {
		return false
	}
	switch strings.ToLower(attribute.Uniqueness) {
	case "server", "global":
		return true
	default:
		return false
	}
}
