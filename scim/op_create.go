package scim

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/couchbase/gocb/v2"
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
	resourceType := Resources["/"+resource]
	var element map[string]interface{}
	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	json.Unmarshal(buf.Bytes(), &element)
	ok, _ := ValidateFieldSchemas(c, element, resourceType)
	if !ok {
		return
	}
	delete(element, "id")
	ok, element = ValidateSchemas(c, element, resourceType.Schema, resourceType.SchemaExtensions)
	if !ok {
		return
	}
	element["id"] = uuid.New().String()
	element["meta"] = GenerateMeta(element, resourceType)

	bucket := Cluster.Bucket(resourceType.Name)
	collection := bucket.DefaultCollection()
	_, err := collection.Upsert(element["id"].(string), element, &gocb.UpsertOptions{})
	if err != nil {
		MakeError(c, http.StatusInternalServerError, err.Error())
		log.Println(err.Error())
		return
	}
	c.JSON(http.StatusOK, element)
}
