package scim

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/couchbase/gocb/v2"
	"github.com/couchbase/gocbcore/v10/memd"
	"github.com/gin-gonic/gin"
)

// Replace is PUT https://example.com/{v}/{resource}/{id}
func Replace(c *gin.Context) {
	resource := c.Param("resource")
	id := c.Param("id")
	if resource == "Bulk" {
		Bulk(c)
		return
	}
	resourceType := Resources["/"+resource]
	var element map[string]interface{}
	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	json.Unmarshal(buf.Bytes(), &element)
	replace(c, resourceType, id, element)
}

func replace(c *gin.Context, resourceType ResoruceType, id string, element map[string]interface{}) {
	ok, _ := ValidateFieldSchemas(c, element, resourceType)
	if !ok {
		return
	}
	meta := element["meta"].(map[string]interface{})
	delete(element, "id")
	delete(element, "meta")
	ok, element = ValidateSchemas(c, element, resourceType.Schema, resourceType.SchemaExtensions)
	if !ok {
		return
	}
	element["id"] = id
	element["meta"] = updateMeta(meta, element, resourceType)

	bucket := Cluster.Bucket(resourceType.Name)
	collection := bucket.DefaultCollection()
	_, err := collection.Replace(element["id"].(string), element, &gocb.ReplaceOptions{})
	if err != nil {
		if se, ok := err.(*gocb.KeyValueError); ok {
			if se.StatusCode == memd.StatusKeyNotFound {
				MakeError(c, http.StatusNotFound, se.ErrorDescription)
				return
			}
		}
		MakeError(c, http.StatusInternalServerError, err.Error())
		log.Println(err.Error())
		return
	}
	c.JSON(http.StatusOK, element)
}
