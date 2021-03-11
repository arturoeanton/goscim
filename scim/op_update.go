package scim

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Update is	PATCH https://example.com/{v}/{resource}/{id}
func Update(c *gin.Context) {
	resource := c.Param("resource")
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": resource,
		"id":      id,
	})
}
