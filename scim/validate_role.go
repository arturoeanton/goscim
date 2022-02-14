package scim

import "github.com/arturoeanton/goscim/commons"

func ValidateReadRole(roles []string, resourceType ResoruceType, element map[string]interface{}) map[string]interface{} {
	schema := Schemas[resourceType.Schema]
	element = commons.WalkMap("", element, func(path, currentKey string, v interface{}) interface{} {
		attr := GetAttribute(schema.Attributes, path)
		if attr.Read != nil {
			for _, read := range *attr.Read {
				for _, role := range roles {
					if read == role {
						return v
					}
				}
			}
			return ""
		}
		return v
	})
	return element
}
