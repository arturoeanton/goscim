package scim

import "strings"

// standardRoleClaims are the claims RFC 9068 2.2.3 / 2.2.3.1 and major OAuth2
// identity providers use for authorization attributes:
//   - "roles", "groups", "entitlements" (RFC 9068 / RFC 7643)
//   - "scope" (RFC 6749 space-delimited scopes)
//   - "scp" (Microsoft Entra ID v2.0 space-delimited scopes and Okta scope arrays)
//   - "realm_access.roles" (Keycloak default realm roles claim)
var standardRoleClaims = []string{"roles", "groups", "entitlements", "scope", "scp", "realm_access.roles"}

// rolesFromClaims collects the caller's roles out of a validated token.
//
// extraClaim, when set, names one or more comma-separated deployment-specific
// claims (top-level keys such as "https://example.com/roles" or dotted nested
// paths such as "resource_access.goscim.roles") to read as well.
func rolesFromClaims(claims map[string]interface{}, extraClaim string) []string {
	names := standardRoleClaims
	if strings.TrimSpace(extraClaim) != "" {
		names = append([]string{}, standardRoleClaims...)
		for _, part := range strings.Split(extraClaim, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				names = append(names, trimmed)
			}
		}
	}

	roles := make([]string, 0, 8)
	seen := make(map[string]struct{})
	for _, name := range names {
		for _, role := range claimValues(lookupClaim(claims, name)) {
			if _, duplicate := seen[role]; duplicate {
				continue
			}
			seen[role] = struct{}{}
			roles = append(roles, role)
		}
	}
	return roles
}

// lookupClaim checks for an exact top-level claim key first (so URI-shaped
// claim names containing dots such as "https://example.com/roles" keep working)
// and otherwise walks dotted object segments (e.g. "realm_access.roles" or
// "resource_access.goscim.roles").
func lookupClaim(claims map[string]interface{}, path string) interface{} {
	if claims == nil || path == "" {
		return nil
	}
	if direct, ok := claims[path]; ok {
		return direct
	}
	if !strings.Contains(path, ".") {
		return nil
	}
	var current interface{} = claims
	for _, seg := range strings.Split(path, ".") {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		val, exists := m[seg]
		if !exists {
			return nil
		}
		current = val
	}
	return current
}

// claimValues normalises one claim into a list of role names.
func claimValues(claim interface{}) []string {
	switch value := claim.(type) {
	case string:
		return strings.Fields(value)
	case []interface{}:
		out := make([]string, 0, len(value))
		for _, item := range value {
			switch typed := item.(type) {
			case string:
				if trimmed := strings.TrimSpace(typed); trimmed != "" {
					out = append(out, trimmed)
				}
			case map[string]interface{}:
				// A SCIM multi-valued attribute: {"value": "...", "type": ...}
				if inner, ok := typed["value"].(string); ok {
					if trimmed := strings.TrimSpace(inner); trimmed != "" {
						out = append(out, trimmed)
					}
				}
			}
		}
		return out
	case []string:
		return value
	default:
		return nil
	}
}
