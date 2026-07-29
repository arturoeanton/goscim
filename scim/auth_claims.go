package scim

import "strings"

// standardRoleClaims are the claims RFC 9068 2.2.3 and 2.2.3.1 define for an
// OAuth 2.0 access token's authorization attributes.
//
// "roles", "groups" and "entitlements" are the JWT profile's reuse of the SCIM
// attributes of the same names from RFC 7643 4.1.2, which is why they are the
// natural source here: they are already SCIM's own vocabulary. "scope" comes
// from RFC 6749 and carries a space-delimited list.
var standardRoleClaims = []string{"roles", "groups", "entitlements", "scope"}

// rolesFromClaims collects the caller's roles out of a validated token.
//
// extraClaim, when set, names a deployment-specific claim to read as well, for
// issuers that put authorization attributes somewhere of their own choosing.
//
// Each claim may be a list of strings, a list of SCIM multi-valued objects
// ({"value": "..."} per RFC 7643 2.4), or a space-delimited string, because all
// three are in the wild.
func rolesFromClaims(claims map[string]interface{}, extraClaim string) []string {
	names := standardRoleClaims
	if extraClaim != "" {
		names = append(append([]string{}, standardRoleClaims...), extraClaim)
	}

	roles := make([]string, 0, 8)
	seen := make(map[string]struct{})
	for _, name := range names {
		for _, role := range claimValues(claims[name]) {
			if _, duplicate := seen[role]; duplicate {
				continue
			}
			seen[role] = struct{}{}
			roles = append(roles, role)
		}
	}
	return roles
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
