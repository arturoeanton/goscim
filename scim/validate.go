package scim

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

//ValidateFieldSchemas is ..
func ValidateFieldSchemas(c *gin.Context, element map[string]interface{}, resourceType ResoruceType) (bool, []string) {
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
	elementSchemas := make([]string, 0)
	for _, s := range schemas.([]interface{}) {
		elementSchemas = append(elementSchemas, s.(string))
		if s.(string) == resourceType.Schema {
			flagPrincipalSchema = true
			continue
		}
		if !ContainsSchemaExtension(resourceType.SchemaExtensions, s.(string)) {
			MakeError(c, http.StatusBadRequest, "schema is not contained in schemaExtensions")
			return false, nil
		}
	}
	for _, se := range resourceType.SchemaExtensions {
		if !se.Required {
			continue
		}
		if !ContainsString(elementSchemas, se.Schema) {
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

//ValidateSchemas is
func ValidateSchemas(c *gin.Context, element map[string]interface{}, schemaNameCore string, schemas []SchemaExtension) (bool, map[string]interface{}) {
	var flag bool
	schema := Schemas[schemaNameCore]
	flag, element = validateSchema(c, element, schema, false)
	if !flag {
		return flag, nil
	}
	for _, schemaExtension := range schemas {
		var flag bool
		schema := Schemas[schemaExtension.Schema]
		elementExtension, ok := element[schemaExtension.Schema].(map[string]interface{})
		if !ok {
			MakeError(c, http.StatusBadRequest, schemaExtension.Schema+" the extension should be object")
			return false, nil
		}
		flag, elementExtension = validateSchema(c, elementExtension, schema, true)
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
			if key == "schemas" {
				continue
			}
			if ContainsStringInArrayInterfase(element["schemas"].([]interface{}), key) {
				continue
			}
		}
		if ContainsString(fields, key) {
			continue
		}

		MakeError(c, http.StatusBadRequest, key+" no exist in schema "+schema.ID)
		return false, nil
	}

	return true, element
}

func validateAttribute(c *gin.Context, element map[string]interface{}, attribute Attribute) (bool, map[string]interface{}) {

	_, ok := element[attribute.Name]
	if !ok && attribute.Required {
		MakeError(c, http.StatusBadRequest, attribute.Name+" is required")
		return false, nil
	}
	if !ok && !attribute.Required {
		return true, element
	}

	var flag bool
	flag, element = validateAttributeString(c, element, attribute)
	if !flag {
		return false, nil
	}
	flag, element = validateAttributeBoolean(c, element, attribute)
	if !flag {
		return false, nil
	}
	flag, element = validateAttributeDecimal(c, element, attribute)
	if !flag {
		return false, nil
	}
	flag, element = validateAttributeInteger(c, element, attribute)
	if !flag {
		return false, nil
	}
	flag, element = validateAttributeDateTime(c, element, attribute)
	if !flag {
		return false, nil
	}
	flag, element = validateAttributeBinary(c, element, attribute)
	if !flag {
		return false, nil
	}

	flag, element = validateAttributeReference(c, element, attribute)
	if !flag {
		return false, nil
	}
	flag, element = validateAttributeComplex(c, element, attribute)
	if !flag {
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
*/
func validateAttributeString(c *gin.Context, element map[string]interface{}, attribute Attribute) (bool, map[string]interface{}) {
	if attribute.Type == "string" {
		if _, ok := element[attribute.Name].(string); !ok {
			MakeError(c, http.StatusBadRequest, attribute.Name+" should be string")
			return false, nil
		}
	}
	return true, element
}

func validateAttributeBoolean(c *gin.Context, element map[string]interface{}, attribute Attribute) (bool, map[string]interface{}) {
	if attribute.Type == "boolean" {
		v, ok := element[attribute.Name].(bool)
		if !ok {
			MakeError(c, http.StatusBadRequest, attribute.Name+" should be boolean")
			return false, nil
		}
		element[attribute.Name] = bool(v)
	}
	return true, element
}

func validateAttributeDecimal(c *gin.Context, element map[string]interface{}, attribute Attribute) (bool, map[string]interface{}) {
	if attribute.Type == "decimal" {
		if _, ok := element[attribute.Name].(float64); !ok {
			MakeError(c, http.StatusBadRequest, attribute.Name+" should be decimal")
			return false, nil
		}
	}
	return true, element
}

func validateAttributeInteger(c *gin.Context, element map[string]interface{}, attribute Attribute) (bool, map[string]interface{}) {
	if attribute.Type == "integer" {
		v, ok := element[attribute.Name].(float64)
		if ok {
			intValue := int64(v)
			if (float64(intValue) - v) != 0 {
				MakeError(c, http.StatusBadRequest, attribute.Name+" should be integer")
				return false, nil
			}
		} else {
			MakeError(c, http.StatusBadRequest, attribute.Name+" should be integer")
			return false, nil
		}
		element[attribute.Name] = int64(v)
	}
	return true, element
}

func validateAttributeDateTime(c *gin.Context, element map[string]interface{}, attribute Attribute) (bool, map[string]interface{}) {
	if strings.ToLower(attribute.Type) == "datetime" {
		v, ok := element[attribute.Name].(string)
		if !ok {
			MakeError(c, http.StatusBadRequest, attribute.Name+" should be datetime")
			return false, nil
		}
		_, err := time.Parse(time.RFC3339, string(v))
		if err != nil {
			MakeError(c, http.StatusBadRequest, attribute.Name+" should be datetime. "+err.Error())
			return false, nil
		}
		//element[attribute.Name] = t.Unix()
	}
	return true, element
}

func validateAttributeBinary(c *gin.Context, element map[string]interface{}, attribute Attribute) (bool, map[string]interface{}) {
	/*if attribute.Type == "binary" {

	}*/
	return true, element
}

func validateAttributeReference(c *gin.Context, element map[string]interface{}, attribute Attribute) (bool, map[string]interface{}) {
	/*if attribute.Type == "reference" {

	}*/
	return true, element
}

func validateAttributeComplex(c *gin.Context, element map[string]interface{}, attribute Attribute) (bool, map[string]interface{}) {

	if attribute.Type == "complex" {
		subElement, ok := element[attribute.Name].(map[string]interface{})
		if !ok {
			MakeError(c, http.StatusBadRequest, attribute.Name+" should be complex")
			return false, nil
		}
		fields := make([]string, 0)
		for _, attribute := range attribute.SubAttributes {
			fields = append(fields, attribute.Name)
			var flag bool
			flag, subElement = validateAttribute(c, subElement, attribute)
			if !flag {
				return flag, nil
			}
		}

		for key := range subElement {
			if ContainsString(fields, key) {
				continue
			}
			MakeError(c, http.StatusBadRequest, key+" no exist in attribute "+attribute.Name)
			return false, nil
		}

		element[attribute.Name] = subElement
	}
	return true, element
}
