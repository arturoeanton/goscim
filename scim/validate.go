package scim

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ValidateFieldSchemas is ..
func ValidateFieldSchemas(c *gin.Context, element map[string]interface{}, resourceType ResourceType) (bool, []string) {
	schemas, ok := element["schemas"]
	if !ok {
		MakeError(c, http.StatusBadRequest, "schemas no exist")
		return false, nil
	}
	if _, ok := schemas.([]interface{}); !ok {
		MakeError(c, http.StatusBadRequest, "schemas  is not array string")
		return false, nil
	}

	flagPrincipalSchema := false
	declared := make([]interface{}, 0)
	elementSchemas := make([]string, 0)
	for _, s := range schemas.([]interface{}) {
		name, ok := s.(string)
		if !ok {
			MakeError(c, http.StatusBadRequest, "schemas is not array string")
			return false, nil
		}
		// Store the URN as the server spells it, whatever case the client
		// used, so everything downstream can compare it directly.
		name = canonicalSchemaID(name)
		declared = append(declared, name)
		elementSchemas = append(elementSchemas, name)

		if strings.EqualFold(name, resourceType.Schema) {
			flagPrincipalSchema = true
			continue
		}
		if !containsSchemaFold(resourceType.SchemaExtensions, name) {
			MakeError(c, http.StatusBadRequest, "schema is not contained in schemaExtensions")
			return false, nil
		}
	}
	element["schemas"] = declared

	for _, se := range resourceType.SchemaExtensions {
		if !se.Required {
			continue
		}
		if !containsStringFold(elementSchemas, se.Schema) {
			MakeError(c, http.StatusBadRequest, se.Schema+" is required true")
			return false, nil
		}
	}

	if !flagPrincipalSchema {
		MakeError(c, http.StatusBadRequest, "there is not principal schema ")
		return false, nil
	}
	return true, elementSchemas
}

// ValidateSchemas is
func ValidateSchemas(c *gin.Context, element map[string]interface{}, schemaNameCore string, schemas []SchemaExtension) (bool, map[string]interface{}) {
	// Extension objects hang off the resource under their schema URN. Rename
	// those keys to the declared spelling first: the core schema's
	// unknown-attribute check runs next and recognises an extension key by
	// comparing it against the canonicalised "schemas" list.
	for _, schemaExtension := range schemas {
		canonicaliseKey(element, schemaExtension.Schema)
	}

	var flag bool
	schema, _ := LookupSchema(schemaNameCore)
	flag, element = validateSchema(c, element, schema, false)
	if !flag {
		return flag, nil
	}
	for _, schemaExtension := range schemas {
		raw, present := canonicaliseKey(element, schemaExtension.Schema)
		if !present {
			// An extension the resource type declares as optional may simply
			// be absent. Only a required one has to be there.
			if schemaExtension.Required {
				MakeTypedError(c, http.StatusBadRequest, "invalidValue",
					schemaExtension.Schema+" is required")
				return false, nil
			}
			continue
		}
		elementExtension, ok := raw.(map[string]interface{})
		if !ok {
			MakeTypedError(c, http.StatusBadRequest, "invalidValue",
				schemaExtension.Schema+" the extension should be object")
			return false, nil
		}
		var flag bool
		extensionSchema, _ := LookupSchema(schemaExtension.Schema)
		flag, elementExtension = validateSchema(c, elementExtension, extensionSchema, true)
		if !flag {
			return flag, nil
		}
		element[schemaExtension.Schema] = elementExtension
	}
	return true, element
}

func validateSchema(c *gin.Context, element map[string]interface{}, schema Schema, isExtension bool) (bool, map[string]interface{}) {
	fields := make([]string, 0)
	for _, attribute := range schema.Attributes {
		fields = append(fields, attribute.Name)
		var flag bool
		flag, element = validateAttribute(c, element, attribute)
		if !flag {
			return flag, nil
		}
	}
	for key := range element {
		if !isExtension {
			if strings.EqualFold(key, "schemas") {
				continue
			}
			// An extension object hangs off the resource under its own URN.
			if declaredSchemas, ok := element["schemas"].([]interface{}); ok &&
				ContainsStringInArrayInterfase(declaredSchemas, key) {
				continue
			}
		}
		// Keys were renamed to the declared spelling as each attribute was
		// validated, so anything left over really is undeclared.
		if ContainsString(fields, key) {
			continue
		}

		MakeError(c, http.StatusBadRequest, key+" no exist in schema "+schema.ID)
		return false, nil
	}

	return true, element
}

/*
   +-----------+-------------+-----------------------------------------+
   | SCIM Data | SCIM Schema | JSON Type                               |
   | Type      | "type"      |                                         |
   +-----------+-------------+-----------------------------------------+
   | String    | "string"    | String per Section 7 of [RFC7159]       |
   |           |             |                                         |
   | Boolean   | "boolean"   | Value per Section 3 of [RFC7159]        |
   |           |             |                                         |
   | Decimal   | "decimal"   | Number per Section 6 of [RFC7159]       |
   |           |             |                                         |
   | Integer   | "integer"   | Number per Section 6 of [RFC7159]       |
   |           |             |                                         |
   | DateTime  | "dateTime"  | String per Section 7 of [RFC7159]       |
   |           |             |                                         |
   | Binary    | "binary"    | Binary value base64 encoded per Section |
   |           |             | 4 of [RFC4648], or with URL and         |
   |           |             | filename safe alphabet URL per Section  |
   |           |             | 5 of [RFC4648] that is passed as a JSON |
   |           |             | string per Section 7 of [RFC7159]       |
   |           |             |                                         |
   | Reference | "reference" | String per Section 7 of [RFC7159]       |
   |           |             |                                         |
   | Complex   | "complex"   | Object per Section 4 of [RFC7159]       |
   +-----------+-------------+-----------------------------------------+

   Orthogonally to the type, RFC 7643 2.2 says an attribute declared
   "multiValued": true carries a JSON array whose elements each have the
   attribute's type.
*/

// validateAttribute checks one attribute of element in place, replacing its
// value with the normalised form (integers become int64, complex values are
// validated recursively).
func validateAttribute(c *gin.Context, element map[string]interface{}, attribute Attribute) (bool, map[string]interface{}) {
	raw, present := canonicaliseKey(element, attribute.Name)
	if !present {
		if attribute.Required {
			MakeTypedError(c, http.StatusBadRequest, "invalidValue", attribute.Name+" is required")
			return false, nil
		}
		return true, element
	}

	if attribute.MultiValued {
		items, ok := raw.([]interface{})
		if !ok {
			MakeTypedError(c, http.StatusBadRequest, "invalidValue",
				attribute.Name+" is multi-valued and should be an array")
			return false, nil
		}
		values := make([]interface{}, 0, len(items))
		for i, item := range items {
			value, ok := validateValue(c, item, attribute, fmt.Sprintf("%s[%d]", attribute.Name, i))
			if !ok {
				return false, nil
			}
			values = append(values, value)
		}
		if !validateSinglePrimary(c, values, attribute) {
			return false, nil
		}
		element[attribute.Name] = values
		return true, element
	}

	value, ok := validateValue(c, raw, attribute, attribute.Name)
	if !ok {
		return false, nil
	}
	element[attribute.Name] = value
	return true, element
}

// validateValue checks a single value against the attribute's declared type.
// path names the value being checked for the error message, so an element of a
// multi-valued attribute is reported as "emails[1]" rather than "emails".
func validateValue(c *gin.Context, value interface{}, attribute Attribute, path string) (interface{}, bool) {
	switch strings.ToLower(attribute.Type) {
	case "string":
		v, ok := value.(string)
		if !ok {
			return invalidValue(c, path+" should be string")
		}
		return v, true

	case "boolean":
		v, ok := value.(bool)
		if !ok {
			return invalidValue(c, path+" should be boolean")
		}
		return v, true

	case "decimal":
		v, ok := value.(float64)
		if !ok {
			return invalidValue(c, path+" should be decimal")
		}
		return v, true

	case "integer":
		v, ok := value.(float64)
		if !ok {
			return invalidValue(c, path+" should be integer")
		}
		if float64(int64(v))-v != 0 {
			return invalidValue(c, path+" should be integer")
		}
		return int64(v), true

	case "datetime":
		v, ok := value.(string)
		if !ok {
			return invalidValue(c, path+" should be datetime")
		}
		if _, err := time.Parse(time.RFC3339, v); err != nil {
			return invalidValue(c, path+" should be datetime. "+err.Error())
		}
		return v, true

	case "complex":
		sub, ok := value.(map[string]interface{})
		if !ok {
			return invalidValue(c, path+" should be complex")
		}
		fields := make([]string, 0, len(attribute.SubAttributes))
		for _, subAttribute := range attribute.SubAttributes {
			fields = append(fields, subAttribute.Name)
			var ok bool
			if ok, sub = validateAttribute(c, sub, subAttribute); !ok {
				return nil, false
			}
		}
		for key := range sub {
			if !ContainsString(fields, key) {
				return invalidValue(c, key+" no exist in attribute "+path)
			}
		}
		return sub, true

	default:
		// "binary" and "reference" carry no extra constraint beyond being
		// JSON strings, and an unknown type is left to the schema author.
		return value, true
	}
}

// validateSinglePrimary enforces RFC 7643 2.4: at most one value of a
// multi-valued attribute may be marked primary.
func validateSinglePrimary(c *gin.Context, values []interface{}, attribute Attribute) bool {
	if !hasSubAttribute(attribute, "primary") {
		return true
	}
	found := false
	for _, value := range values {
		item, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		if primary, _ := item["primary"].(bool); primary {
			if found {
				MakeTypedError(c, http.StatusBadRequest, "invalidValue",
					attribute.Name+" has more than one value marked primary")
				return false
			}
			found = true
		}
	}
	return true
}

func hasSubAttribute(attribute Attribute, name string) bool {
	for _, subAttribute := range attribute.SubAttributes {
		if subAttribute.Name == name {
			return true
		}
	}
	return false
}

func invalidValue(c *gin.Context, message string) (interface{}, bool) {
	MakeTypedError(c, http.StatusBadRequest, "invalidValue", message)
	return nil, false
}
