package scim

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MakeError ...
func MakeError(c *gin.Context, status int, message string) Error {
	return MakeTypedError(c, status, "", message)
}

// internalError logs the underlying failure and sends a generic 500 to the
// client, keeping driver errors, bucket names and cluster addresses out of
// responses.
func internalError(c *gin.Context, err error) {
	log.Println(err.Error())
	MakeError(c, http.StatusInternalServerError, "internal server error")
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
	writeSCIM(c, status, scimError)
	return scimError
}
