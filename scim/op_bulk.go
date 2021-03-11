package scim

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Bulk is 	POST https://example.com/{v}/Bulk
func Bulk(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
