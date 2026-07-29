package scim

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// MakeError ...
func MakeError(c *gin.Context, status int, message string) Error {
	return MakeTypedError(c, status, "", message)
}

// MakeTypedError writes a SCIM error carrying the scimType keyword defined in
// RFC 7644 3.12 ("invalidFilter", "invalidPath", "uniqueness", ...), which is
// how clients tell apart the several conditions that share a status code.
func MakeTypedError(c *gin.Context, status int, scimType string, message string) Error {
	scimError := Error{
		Schemas:  []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		ScimType: scimType,
		Status:   strconv.Itoa(status),
		Detail:   message,
	}
	c.JSON(status, scimError)
	return scimError
}
