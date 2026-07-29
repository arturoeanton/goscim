package scim

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Read is GET https://example.com/{v}/{resource}/{id}
func Read(resource string) func(c *gin.Context) {
	return func(c *gin.Context) {
		resourceType := Resources[resource]
		id := c.Param("id")
		element, err := getElementByID(c, id, resourceType)
		if err != nil {
			return
		}
		element = ValidateReadRole(currentRoles(c), resourceType, element)
		c.JSON(http.StatusOK, element)
	}
}

func getElementByID(c *gin.Context, id string, resourceType ResoruceType) (map[string]interface{}, error) {
	element, err := DB.Get(resourceType.Name, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			MakeError(c, http.StatusNotFound, "Resource "+id+" not found")
			return nil, err
		}
		MakeError(c, http.StatusInternalServerError, err.Error())
		log.Println(err.Error())
		return nil, err
	}
	return element, nil
}
