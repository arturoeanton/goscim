package scim

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Update is	PATCH https://example.com/{v}/{resource}/{id}
func Update(c *gin.Context) {
	resource := c.Param("resource")
	id := c.Param("id")
	resourceType := Resources["/"+resource]
	var patchRequest Patch
	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	err := json.Unmarshal(buf.Bytes(), &patchRequest)
	if err != nil {
		MakeError(c, http.StatusBadRequest, err.Error())
	}
	element, err := getElementByID(c, id, resourceType)
	if err != nil {
		return
	}
	for _, op := range patchRequest.Operations {
		switch op.Op {
		case "add":
			{
				element = patchAdd(op.Path, op.Value, element)
			}
		case "replace":
			{
				element = patchReplace(op.Path, op.Value, element)
			}
		case "remove":
			{
				element = patchRemove(op.Path, op.Value, element)
			}
		}

	}
	replace(c, resourceType, id, element)
}

func patchAdd(opPath string, opValue interface{}, element map[string]interface{}) map[string]interface{} {
	return element
}
func patchReplace(opPath string, opValue interface{}, element map[string]interface{}) map[string]interface{} {
	return element
}
func patchRemove(opPath string, opValue interface{}, element map[string]interface{}) map[string]interface{} {
	return element
}
