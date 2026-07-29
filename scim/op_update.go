package scim

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// patchError is a client error in a PATCH request, carrying the RFC 7644 3.12
// scimType keyword that tells the client which condition it hit.
type patchError struct {
	scimType string
	message  string
}

func (e *patchError) Error() string { return e.message }

func patchFailure(scimType, message string) error {
	return &patchError{scimType: scimType, message: message}
}

// Update is	PATCH https://example.com/{v}/{resource}/{id}
func Update(resource string) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		resourceType := Resources[resource]
		var patchRequest Patch
		buf := new(bytes.Buffer)
		buf.ReadFrom(c.Request.Body)
		if err := json.Unmarshal(buf.Bytes(), &patchRequest); err != nil {
			MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", err.Error())
			return
		}
		if len(patchRequest.Operations) == 0 {
			MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "Operations is required and must not be empty")
			return
		}
		element, err := getElementByID(c, id, resourceType)
		if err != nil {
			return
		}
		for _, op := range patchRequest.Operations {
			element, err = applyPatchOperation(op, element)
			if err != nil {
				var failure *patchError
				if errors.As(err, &failure) {
					MakeTypedError(c, http.StatusBadRequest, failure.scimType, failure.message)
					return
				}
				MakeError(c, http.StatusBadRequest, err.Error())
				return
			}
		}

		replace(c, resourceType, id, element)
	}
}

// applyPatchOperation applies one RFC 7644 3.5.2 operation to element. Every
// failure is reported as an error: the resource is walked with checked lookups
// rather than bare type assertions, so a path that does not resolve is a 400
// instead of a panic.
func applyPatchOperation(op Operation, element map[string]interface{}) (map[string]interface{}, error) {
	operation := strings.ToLower(op.Op)
	switch operation {
	case "add", "replace":
		if op.Path == "" {
			// RFC 7644 3.5.2.1/3.5.2.3: with no path the value is an object
			// whose members are applied to the resource itself.
			values, ok := op.Value.(map[string]interface{})
			if !ok {
				return nil, patchFailure("invalidSyntax",
					`an "`+operation+`" operation without a path must carry an object value`)
			}
			for key, value := range values {
				element[key] = value
			}
			return element, nil
		}
	case "remove":
		if op.Path == "" {
			return nil, patchFailure("noTarget", `a "remove" operation requires a path`)
		}
	default:
		return nil, patchFailure("invalidSyntax", "unknown patch operation "+strconv.Quote(op.Op))
	}

	field, parent, err := resolvePatchPath(op.Path, element)
	if err != nil {
		return nil, err
	}

	switch operation {
	case "add":
		parent[field] = addPatchValue(parent[field], op.Value)
	case "replace":
		parent[field] = op.Value
	case "remove":
		delete(parent, field)
	}
	return element, nil
}

// resolvePatchPath walks path through element and returns the final attribute
// name together with the object holding it.
func resolvePatchPath(path string, element map[string]interface{}) (string, map[string]interface{}, error) {
	if strings.ContainsAny(path, "[]") {
		// e.g. emails[type eq "work"].value. Rejecting these is wrong per the
		// RFC but honest: silently treating them as a literal attribute name
		// would write to an attribute the client never named.
		return "", nil, patchFailure("invalidPath",
			"value filters in patch paths are not supported: "+strconv.Quote(path))
	}
	segments := opPathTopathArray(path)
	for i, field := range segments {
		if field == "" {
			return "", nil, patchFailure("invalidPath", "empty segment in path "+strconv.Quote(path))
		}
		if i == len(segments)-1 {
			return field, element, nil
		}
		next, ok := element[field].(map[string]interface{})
		if !ok {
			return "", nil, patchFailure("noTarget",
				"path "+strconv.Quote(path)+" does not resolve: "+strconv.Quote(field)+" is not an object")
		}
		element = next
	}
	return "", nil, patchFailure("invalidPath", "empty path")
}

// addPatchValue implements the "add" semantics: appending to an existing array,
// and otherwise setting the value.
func addPatchValue(current interface{}, value interface{}) interface{} {
	existing, ok := current.([]interface{})
	if !ok {
		return value
	}
	if values, ok := value.([]interface{}); ok {
		return append(existing, values...)
	}
	return append(existing, value)
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
