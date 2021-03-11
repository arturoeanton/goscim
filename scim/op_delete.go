package scim

import (
	"log"
	"net/http"

	"github.com/couchbase/gocb/v2"
	"github.com/couchbase/gocbcore/v9/memd"
	"github.com/gin-gonic/gin"
)

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
