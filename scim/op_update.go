package scim

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// Update is	PATCH https://example.com/{v}/{resource}/{id}
func Update(resource string) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		resourceType := Resources[resource]
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

		//TODO: Validate _write of all fields of element

		replace(c, resourceType, id, element)
	}
}
func opPathTopathArray(value string) []string {
	re := regexp.MustCompile(`^(urn[:\w\.\_]*)(:-*)?(:[\w]*)(\.)(.*)$`)
	urn := ""
	if re.MatchString(value) {
		urn = re.ReplaceAllString(value, `${1}${2}${3}`)
	}
	path := re.ReplaceAllString(value, `${5}`)
	pathArray := make([]string, 0)
	if urn != "" {
		pathArray = append(pathArray, urn)
	}
	pathArray = append(pathArray, strings.Split(path, ".")...)
	return pathArray
}
func pointValue(opPath string, elemOld interface{}) (string, interface{}) {
	pathArray := opPathTopathArray(opPath)
	var elemPoint interface{}
	var elemPointPrev interface{}
	elemPoint = elemOld
	lastField := ""
	for _, field := range pathArray {
		lastField = field
		elemPointPrev = elemPoint
		elemPoint = (elemPoint.(map[string]interface{})[field])
	}
	return lastField, elemPointPrev
}

func patchAdd(opPath string, opValue interface{}, elemOld map[string]interface{}) map[string]interface{} {
	lastField, elemPointPrev := pointValue(opPath, elemOld)
	if lastField != "" {
		if elemPointPrev.(map[string]interface{})[lastField] != nil {
			arrayElemPoint, ok := elemPointPrev.(map[string]interface{})[lastField].([]interface{})
			if ok {
				arrayValue, ok := opValue.([]interface{})
				if ok {
					arrayElemPoint = append(arrayElemPoint, arrayValue...)
				} else {
					arrayElemPoint = append(arrayElemPoint, opValue)
				}
				//TODO: check if arrayElemPoint is []interface{}
				elemPointPrev.(map[string]interface{})[lastField] = arrayElemPoint
			} else {
				elemPointPrev.(map[string]interface{})[lastField] = opValue
			}
		} else {
			elemPointPrev.(map[string]interface{})[lastField] = opValue
		}
	}
	return elemOld
}
func patchReplace(opPath string, opValue interface{}, elemOld map[string]interface{}) map[string]interface{} {
	lastField, elemPointPrev := pointValue(opPath, elemOld)
	if lastField != "" {
		elemPointPrev.(map[string]interface{})[lastField] = opValue
	}
	return elemOld
}
func patchRemove(opPath string, opValue interface{}, elemOld map[string]interface{}) map[string]interface{} {
	lastField, elemPointPrev := pointValue(opPath, elemOld)
	if lastField != "" {
		delete(elemPointPrev.(map[string]interface{}), lastField)
	}
	return elemOld
}
