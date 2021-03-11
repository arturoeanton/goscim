package scim

//ContainsString is
func ContainsString(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

//ContainsSchemaExtension is
func ContainsSchemaExtension(slice []SchemaExtension, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s.Schema] = struct{}{}
	}
	_, ok := set[item]
	return ok
}

// ContainsStringInArrayInterfase is
func ContainsStringInArrayInterfase(slice []interface{}, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s.(string)] = struct{}{}
	}
	_, ok := set[item]
	return ok
}
