package scim

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Replace is PUT https://example.com/{v}/{resource}/{id}
func Replace(resource string) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		if resource == "Bulk" {
			Bulk(c)
			return
		}
		resourceType := Resources[resource]
		var element map[string]interface{}
		buf := new(bytes.Buffer)
		buf.ReadFrom(c.Request.Body)
		json.Unmarshal(buf.Bytes(), &element)
		replace(c, resourceType, id, element)
	}
}
func replace(c *gin.Context, resourceType ResoruceType, id string, element map[string]interface{}) {
	// The previous meta comes from the stored resource, never from the
	// request: RFC 7644 does not require a client to echo meta back on a PUT,
	// and meta.created is not the client's to set.
	stored, err := getElementByID(c, id, resourceType)
	if err != nil {
		return
	}
	previousMeta, _ := stored["meta"].(map[string]interface{})

	ok, _ := ValidateFieldSchemas(c, element, resourceType)
	if !ok {
		return
	}
	//TODO: Validate _write of all fields of element

	delete(element, "id")
	delete(element, "meta")
	ok, element = ValidateSchemas(c, element, resourceType.Schema, resourceType.SchemaExtensions)
	if !ok {
		return
	}
	element["id"] = id
	element["meta"] = updateMeta(previousMeta, element, resourceType)

	if err := DB.Replace(resourceType.Name, id, element); err != nil {
		if errors.Is(err, ErrNotFound) {
			MakeError(c, http.StatusNotFound, "Resource "+id+" not found")
			return
		}
		MakeError(c, http.StatusInternalServerError, err.Error())
		log.Println(err.Error())
		return
	}
	c.JSON(http.StatusOK, ValidateReadRole(currentRoles(c), resourceType, element))
}
