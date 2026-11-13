package scim

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Replace is PUT https://example.com/{v}/{resource}/{id}
func Replace(resource string) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		resourceType := Resources[resource]
		setAuditTarget(c, "replace", resourceType.Name, id)
		raw, ok := readBody(c)
		if !ok {
			return
		}
		var element map[string]interface{}
		if err := json.Unmarshal(raw, &element); err != nil || element == nil {
			// Still answer 404 when the target resource does not exist.
			if _, getErr := getElementByID(c, id, resourceType); getErr != nil {
				return
			}
			MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "request body must be a JSON object")
			return
		}
		replaceWithMode(c, resourceType, id, element, false, nil)
	}
}

func replace(c *gin.Context, resourceType ResourceType, id string, element map[string]interface{}) {
	replaceWithMode(c, resourceType, id, element, false, nil)
}

func replaceWithMode(c *gin.Context, resourceType ResourceType, id string, element map[string]interface{}, isPatch bool, patchedPaths map[string]bool) {
	proj, ok := parseProjectionFromQuery(c)
	if !ok {
		return
	}
	// The previous meta comes from the stored resource, never from the
	// request: RFC 7644 does not require a client to echo meta back on a PUT,
	// and meta.created is not the client's to set.
	stored, err := getElementByID(c, id, resourceType)
	if err != nil {
		return
	}
	if !checkPrecondition(c, stored) {
		return
	}
	previousMeta, _ := stored["meta"].(map[string]interface{})

	ok, _ = ValidateFieldSchemas(c, element, resourceType)
	if !ok {
		return
	}

	deleteKeyFold(element, "id")
	deleteKeyFold(element, "meta")
	if !enforceWriteAccess(c, resourceType, element, stored, isPatch) {
		return
	}
	ok, element = ValidateSchemas(c, element, resourceType.Schema, resourceType.SchemaExtensions)
	if !ok {
		return
	}
	if !ApplyPasswordPolicy(c, resourceType, element, stored, patchedPaths) {
		return
	}
	element["id"] = id
	meta := updateMeta(previousMeta, element, resourceType)
	element["meta"] = meta

	newClaims := ExtractUniqueClaims(resourceType, element)
	err = DB.ReplaceUnique(resourceType.Name, id, element, newClaims, func(current map[string]interface{}) []UniqueClaim {
		return ExtractUniqueClaims(resourceType, current)
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			MakeError(c, http.StatusNotFound, "Resource "+id+" not found")
			return
		}
		writeStoreUniqueError(c, resourceType, err)
		return
	}
	syncAfterReplace(resourceType, id, stored, element)
	c.Header("Location", absoluteLocation(c, resourceType, id))
	setVersionHeader(c, meta)
	writeSCIM(c, http.StatusOK, FormatResource(currentRoles(c), resourceType, element, proj))
}
