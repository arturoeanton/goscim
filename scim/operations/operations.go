package operations

import (
	"bytes"
	"encoding/json"

	"github.com/arturoeanton/goscim/scim/config"
	"github.com/arturoeanton/goscim/scim/meta"
	"github.com/arturoeanton/goscim/scim/validates"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//Create is  POST https://example.com/{v}/{resource}
func Create(c *gin.Context) {
	resource := c.Param("resource")
	if resource == "Bulk" {
		Bulk(c)
		return
	}
	resourceType := config.Resources["/"+resource]
	var element map[string]interface{}
	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	json.Unmarshal(buf.Bytes(), &element)
	ok, _ := validates.ValidateFieldSchemas(c, element, resourceType)
	if !ok {
		return
	}
	delete(element, "id")
	ok, element = validates.ValidateSchemas(c, element, resourceType.Schema, resourceType.SchemaExtensions)
	if !ok {
		return
	}
	element["id"] = uuid.New().String()
	element["meta"] = meta.GenerateMeta(element, resourceType)
	c.JSON(200, element)
}

// Read is GET https://example.com/{v}/{resource}/{id}
func Read(c *gin.Context) {
	resource := c.Param("resource")
	id := c.Param("id")
	c.JSON(200, gin.H{
		"message": resource,
		"id":      id,
	})
}

// Replace is PUT https://example.com/{v}/{resource}/{id}
func Replace(c *gin.Context) {
	resource := c.Param("resource")
	id := c.Param("id")
	c.JSON(200, gin.H{
		"message": resource,
		"id":      id,
	})
}

// Delete is  	DELETE https://example.com/{v}/{resource}/{id}
func Delete(c *gin.Context) {
	resource := c.Param("resource")
	id := c.Param("id")
	c.JSON(200, gin.H{
		"message": resource,
		"id":      id,
	})
}

// Update is	PATCH https://example.com/{v}/{resource}/{id}
func Update(c *gin.Context) {
	resource := c.Param("resource")
	id := c.Param("id")
	c.JSON(200, gin.H{
		"message": resource,
		"id":      id,
	})
}

// Search is 	GET https://example.com/{v}/{resource}?ﬁlter={attribute}{op}{value}&sortBy={attributeName}&sortOrder={ascending|descending}
func Search(c *gin.Context) {
	resource := c.Param("resource")
	c.JSON(200, gin.H{
		"message": resource,
	})
}

// Bulk is 	POST https://example.com/{v}/Bulk
func Bulk(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
