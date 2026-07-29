package scim

import (
	"bytes"
	"encoding/json"
	"log"
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
		var element map[string]interface{}
		buf := new(bytes.Buffer)
		buf.ReadFrom(c.Request.Body)
		json.Unmarshal(buf.Bytes(), &element)
		ok, _ := ValidateFieldSchemas(c, element, resourceType)
		if !ok {
			return
		}
		delete(element, "id")
		if !EnforceWriteAccess(c, resourceType, element, nil) {
			return
		}
		ok, element = ValidateSchemas(c, element, resourceType.Schema, resourceType.SchemaExtensions)
		if !ok {
			return
		}
		if !EnforceUniqueness(c, resourceType, element, "") {
			return
		}

		id := uuid.New().String()
		element["id"] = id
		meta := generateMeta(element, resourceType)
		element["meta"] = meta

		if err := DB.Upsert(resourceType.Name, id, element); err != nil {
			MakeError(c, http.StatusInternalServerError, err.Error())
			log.Println(err.Error())
			return
		}

		// RFC 7644 3.3: a successful create answers 201 with the URI of the
		// new resource in Location.
		c.Header("Location", absoluteLocation(c, resourceType, id))
		setVersionHeader(c, meta)
		writeSCIM(c, http.StatusCreated, ValidateReadRole(currentRoles(c), resourceType, element))
	}
}
