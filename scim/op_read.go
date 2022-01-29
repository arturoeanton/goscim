package scim

import (
	"log"
	"net/http"

	"github.com/couchbase/gocb/v2"
	"github.com/couchbase/gocbcore/v10/memd"
	"github.com/gin-gonic/gin"
)

// Read is GET https://example.com/{v}/{resource}/{id}
func Read(c *gin.Context) {
	resource := c.Param("resource")
	resourceType := Resources["/"+resource]
	id := c.Param("id")
	element, err := getElementByID(c, id, resourceType)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, element)
}
func getElementByID(c *gin.Context, id string, resourceType ResoruceType) (map[string]interface{}, error) {
	bucket := Cluster.Bucket(resourceType.Name)
	data, err := bucket.DefaultCollection().Get(id, &gocb.GetOptions{})
	if err != nil {
		if se, ok := err.(*gocb.KeyValueError); ok {
			if se.StatusCode == memd.StatusKeyNotFound {
				MakeError(c, http.StatusNotFound, se.ErrorDescription)
				return nil, err
			}
		}
		MakeError(c, http.StatusInternalServerError, err.Error())
		log.Println(err.Error())
		return nil, err
	}
	var result interface{}
	data.Content(&result)
	return result.(map[string]interface{}), nil
}
