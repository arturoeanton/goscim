package scim

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Delete is  	DELETE https://example.com/{v}/{resource}/{id}
func Delete(resource string) func(c *gin.Context) {
	return func(c *gin.Context) {

		//TODO: Validate _remove of all fields of element

		id := c.Param("id")
		if resource == "Bulk" {
			Bulk(c)
			return
		}
		resourceType := Resources[resource]
		if c.GetHeader("If-Match") != "" {
			stored, err := getElementByID(c, id, resourceType)
			if err != nil {
				return
			}
			if !checkPrecondition(c, stored) {
				return
			}
		}
		if err := DB.Remove(resourceType.Name, id); err != nil {
			if errors.Is(err, ErrNotFound) {
				MakeError(c, http.StatusNotFound, "Resource "+id+" not found")
				return
			}
			MakeError(c, http.StatusInternalServerError, err.Error())
			log.Println(err.Error())
			return
		}
		c.Writer.WriteHeader(http.StatusNoContent)
	}
}
