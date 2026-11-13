package scim

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Create is  POST https://example.com/{v}/{resource}
func Create(resource string) func(c *gin.Context) {
	return func(c *gin.Context) {
		if resource == "Bulk" {
			Bulk(c)
			return
		}
		resourceType := Resources[resource]
		setAuditTarget(c, "create", resourceType.Name, "")
		proj, ok := parseProjectionFromQuery(c)
		if !ok {
			return
		}
		raw, ok := readBody(c)
		if !ok {
			return
		}
		var element map[string]interface{}
		if err := json.Unmarshal(raw, &element); err != nil || element == nil {
			MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "request body must be a JSON object")
			return
		}
		ok, _ = ValidateFieldSchemas(c, element, resourceType)
		if !ok {
			return
		}
		deleteKeyFold(element, "id")
		deleteKeyFold(element, "meta")
		if !EnforceWriteAccess(c, resourceType, element, nil) {
			return
		}
		ok, element = ValidateSchemas(c, element, resourceType.Schema, resourceType.SchemaExtensions)
		if !ok {
			return
		}
		if !ApplyPasswordPolicy(c, resourceType, element, nil, nil) {
			return
		}

		id := uuid.New().String()
		element["id"] = id
		meta := generateMeta(element, resourceType)
		element["meta"] = meta

		claims := ExtractUniqueClaims(resourceType, element)
		if err := DB.CreateUnique(resourceType.Name, id, element, claims); err != nil {
			writeStoreUniqueError(c, resourceType, err)
			return
		}
		setAuditTarget(c, "create", resourceType.Name, id)

		// RFC 7644 3.3: a successful create answers 201 with the URI of the
		// new resource in Location.
		c.Header("Location", absoluteLocation(c, resourceType, id))
		setVersionHeader(c, meta)
		writeSCIM(c, http.StatusCreated, FormatResource(currentRoles(c), resourceType, element, proj))
	}
}
