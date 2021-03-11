package scim

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// MakeError ...
func MakeError(c *gin.Context, status int, message string) Error {
	scimError := Error{
		Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		Status:  strconv.Itoa(status),
		Detail:  message,
	}
	c.JSON(status, scimError)
	return scimError
}
