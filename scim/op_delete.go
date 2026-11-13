package scim

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Delete is  	DELETE https://example.com/{v}/{resource}/{id}
func Delete(resource string) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		if resource == "Bulk" {
			Bulk(c)
			return
		}
		resourceType := Resources[resource]
		setAuditTarget(c, "delete", resourceType.Name, id)
		stored, err := getElementByID(c, id, resourceType)
		if err != nil {
			return
		}
		if !checkPrecondition(c, stored) {
			return
		}
		if !EnforceDeleteAccess(c, resourceType, stored) {
			return
		}
		claims := ExtractUniqueClaims(resourceType, stored)
		if err := DB.RemoveUnique(resourceType.Name, id, claims); err != nil {
			if errors.Is(err, ErrNotFound) {
				MakeError(c, http.StatusNotFound, "Resource "+id+" not found")
				return
			}
			internalError(c, err)
			return
		}
		c.Writer.WriteHeader(http.StatusNoContent)
	}
}
