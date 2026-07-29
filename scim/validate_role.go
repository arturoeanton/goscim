package scim

import "github.com/gin-gonic/gin"

// currentRoles returns the roles of the authenticated caller. A request that
// never went through the authentication middleware has none, so every
// restricted attribute stays hidden rather than being exposed by accident.
func currentRoles(c *gin.Context) []string {
	return PrincipalOf(c).rolesOrNone()
}

func (p *Principal) rolesOrNone() []string {
	if p == nil {
		return nil
	}
	return p.Roles
}

// ValidateReadRole returns a copy of element holding only what the caller may
// read, according to the $reader lists declared on the schema attributes.
//
// An attribute the caller cannot read is omitted rather than blanked: an empty
// string is indistinguishable from a legitimately empty value, and it is not
// even a valid value for a non-string attribute.
func ValidateReadRole(roles []string, resourceType ResoruceType, element map[string]interface{}) map[string]interface{} {
	if element == nil {
		return nil
	}
	filtered := filterAttributes(roles, Schemas[resourceType.Schema].Attributes, element)

	// Extension attributes live in an object keyed by the extension's URN, so
	// they have to be resolved against that schema rather than the core one.
	for _, extension := range resourceType.SchemaExtensions {
		raw, present := filtered[extension.Schema]
		if !present {
			continue
		}
		values, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		filtered[extension.Schema] = filterAttributes(roles, Schemas[extension.Schema].Attributes, values)
	}
	return filtered
}

// filterAttributes drops every entry whose declaring attribute restricts
// reading to roles the caller does not hold, and recurses into the values it
// keeps. Entries no schema declares -- id, meta, schemas -- are passed through.
func filterAttributes(roles []string, attributes []Attribute, element map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{}, len(element))
	for key, value := range element {
		attribute, declared := findAttribute(attributes, key)
		if !declared {
			filtered[key] = value
			continue
		}
		if !canRead(roles, attribute) {
			continue
		}
		filtered[key] = filterValue(roles, attribute, value)
	}
	return filtered
}

// filterValue recurses through complex values and, unlike the previous
// map-walking helper, through the elements of multi-valued ones.
func filterValue(roles []string, attribute Attribute, value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return filterAttributes(roles, attribute.SubAttributes, typed)
	case []interface{}:
		items := make([]interface{}, 0, len(typed))
		for _, item := range typed {
			items = append(items, filterValue(roles, attribute, item))
		}
		return items
	default:
		return value
	}
}

// canRead reports whether any of the caller's roles appears in the attribute's
// $reader list. An attribute without one is unrestricted, and "*" means any
// role, matching how $writer is spelled in the shipped schemas.
func canRead(roles []string, attribute Attribute) bool {
	if attribute.Read == nil {
		return true
	}
	for _, allowed := range *attribute.Read {
		if allowed == "*" {
			return true
		}
		for _, role := range roles {
			if allowed == role {
				return true
			}
		}
	}
	return false
}

func findAttribute(attributes []Attribute, name string) (Attribute, bool) {
	for _, attribute := range attributes {
		if attribute.Name == name {
			return attribute, true
		}
	}
	return Attribute{}, false
}
