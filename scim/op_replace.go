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
	ok, _ := ValidateFieldSchemas(c, element, resourceType)
	if !ok {
		return
	}
	//TODO: Validate _write of all fields of element

	meta := element["meta"].(map[string]interface{})
	delete(element, "id")
	delete(element, "meta")
	ok, element = ValidateSchemas(c, element, resourceType.Schema, resourceType.SchemaExtensions)
	if !ok {
		return
	}
	element["id"] = id
	element["meta"] = updateMeta(meta, element, resourceType)

	if err := DB.Replace(resourceType.Name, id, element); err != nil {
		if errors.Is(err, ErrNotFound) {
			MakeError(c, http.StatusNotFound, "Resource "+id+" not found")
			return
		}
		MakeError(c, http.StatusInternalServerError, err.Error())
		log.Println(err.Error())
		return
	}
	c.JSON(http.StatusOK, element)
}
