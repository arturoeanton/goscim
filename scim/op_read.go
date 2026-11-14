package scim

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Read is GET https://example.com/{v}/{resource}/{id}
func Read(resource string) func(c *gin.Context) {
	return func(c *gin.Context) {
		resourceType := Resources[resource]
		id := c.Param("id")
		setAuditTarget(c, "read", resourceType.Name, id)
		proj, ok := parseProjectionFromQuery(c)
		if !ok {
			return
		}
		element, err := getElementByID(c, id, resourceType)
		if err != nil {
			return
		}
		version := versionOf(element)
		if ifNoneMatch := c.GetHeader("If-None-Match"); ifNoneMatch != "" && etagMatches(ifNoneMatch, version) {
			c.Status(http.StatusNotModified)
			return
		}
		if tag := entityTag(version); tag != "" {
			c.Header("ETag", tag)
		}
		writeSCIM(c, http.StatusOK, FormatResource(currentRoles(c), resourceType, element, proj))
	}
}

func getElementByID(c *gin.Context, id string, resourceType ResourceType) (map[string]interface{}, error) {
	element, err := DB.Get(storeBucketForContext(c, resourceType.Name), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			MakeError(c, http.StatusNotFound, "Resource "+id+" not found")
			return nil, err
		}
		internalError(c, err)
		return nil, err
	}
	return element, nil
}
