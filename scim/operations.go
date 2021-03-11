package scim

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/couchbase/gocb/v2"
	"github.com/couchbase/gocbcore/v9/memd"
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

// Read is GET https://example.com/{v}/{resource}/{id}
func Read(c *gin.Context) {
	resource := c.Param("resource")
	resourceType := Resources["/"+resource]
	id := c.Param("id")
	bucket := Cluster.Bucket(resourceType.Name)
	data, err := bucket.DefaultCollection().Get(id, &gocb.GetOptions{})
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
	var result interface{}
	data.Content(&result)
	c.JSON(http.StatusOK, result)
}

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
	ok, _ := ValidateFieldSchemas(c, element, resourceType)
	if !ok {
		return
	}
	delete(element, "id")
	ok, element = ValidateSchemas(c, element, resourceType.Schema, resourceType.SchemaExtensions)
	if !ok {
		return
	}
	element["id"] = id
	element["meta"] = GenerateMeta(element, resourceType)

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

// Delete is  	DELETE https://example.com/{v}/{resource}/{id}
func Delete(c *gin.Context) {
	resource := c.Param("resource")
	id := c.Param("id")
	if resource == "Bulk" {
		Bulk(c)
		return
	}
	resourceType := Resources["/"+resource]
	bucket := Cluster.Bucket(resourceType.Name)
	collection := bucket.DefaultCollection()
	_, err := collection.Remove(id, &gocb.RemoveOptions{})
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
	c.Writer.WriteHeader(http.StatusNoContent)
}

// Update is	PATCH https://example.com/{v}/{resource}/{id}
func Update(c *gin.Context) {
	resource := c.Param("resource")
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": resource,
		"id":      id,
	})
}

// Search is 	GET https://example.com/{v}/{resource}?ﬁlter={attribute}{op}{value}&sortBy={attributeName}&sortOrder={ascending|descending}
func Search(c *gin.Context) {
	resource := c.Param("resource")
	resourceType := Resources["/"+resource]
	rows, err := Cluster.Query("SELECT * FROM `"+resourceType.Name+"`;", nil)
	defer rows.Close()
	if err != nil {
		MakeError(c, http.StatusInternalServerError, err.Error())
		log.Println(err.Error())
		return
	}

	var result ListResponse
	result.Schemas = append(result.Schemas, "urn:ietf:params:scim:api:messages:2.0:ListResponse")
	result.TotalResults = 0
	for rows.Next() {
		var item Resource
		err := rows.Row(&item)
		if err != nil {
			MakeError(c, http.StatusInternalServerError, err.Error())
			log.Println(err.Error())
			return
		}
		result.Resources = append(result.Resources, item)
		result.TotalResults++
	}
	c.JSON(http.StatusOK, result)
}

// Bulk is 	POST https://example.com/{v}/Bulk
func Bulk(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
