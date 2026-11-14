package scim

import (
	"fmt"

	"github.com/google/uuid"
)

const enterpriseUserSchemaURI = "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"

// SeedDemoResources populates the target tenant ("" for default) with sample
// SCIM 2.0 Users and nested Groups when the User store in that tenant is empty.
// If force is true, sample records with fresh identifiers are inserted only when
// their userName / displayName does not already exist. Returns the number of
// resources created.
func SeedDemoResources(tenant string, force bool) (int, error) {
	if DB == nil {
		return 0, fmt.Errorf("store is not initialised")
	}
	userRT, groupRT, ok := findUserAndGroupResourceTypes()
	if !ok {
		return 0, nil
	}

	userBucket := ScopedBucket(tenant, userRT.Name)
	groupBucket := ScopedBucket(tenant, groupRT.Name)
	prefix := PREFIX
	if tenant != "" {
		prefix = "/tenants/" + tenant + PREFIX
	}

	if !force {
		total, _, err := DB.Search(SearchQuery{Bucket: userBucket, Limit: 1})
		if err != nil {
			return 0, err
		}
		if total > 0 {
			return 0, nil
		}
	}

	createdCount := 0
	createdUserIDs := make(map[string]string)

	demoUsers := []map[string]interface{}{
		{
			"schemas":    []string{coreUserSchemaURI, enterpriseUserSchemaURI},
			"userName":   "adele.vance@contoso.com",
			"externalId": "entra-obj-adele-001",
			"name": map[string]interface{}{
				"formatted":  "Adele Vance",
				"givenName":  "Adele",
				"familyName": "Vance",
			},
			"displayName":       "Adele Vance",
			"title":             "Principal Identity Architect",
			"preferredLanguage": "en-US",
			"active":            true,
			"emails": []interface{}{
				map[string]interface{}{"value": "adele.vance@contoso.com", "type": "work", "primary": true},
			},
			"roles": []interface{}{
				map[string]interface{}{"value": "admin", "display": "SCIM Administrator", "primary": true},
			},
			enterpriseUserSchemaURI: map[string]interface{}{
				"employeeNumber": "E-1001",
				"department":     "Platform Engineering",
				"organization":   "Contoso",
			},
		},
		{
			"schemas":    []string{coreUserSchemaURI, enterpriseUserSchemaURI},
			"userName":   "megan.bowen@contoso.com",
			"externalId": "entra-obj-megan-002",
			"name": map[string]interface{}{
				"formatted":  "Megan Bowen",
				"givenName":  "Megan",
				"familyName": "Bowen",
			},
			"displayName":       "Megan Bowen",
			"title":             "Staff Distributed Systems Engineer",
			"preferredLanguage": "en-US",
			"active":            true,
			"emails": []interface{}{
				map[string]interface{}{"value": "megan.bowen@contoso.com", "type": "work", "primary": true},
			},
			"roles": []interface{}{
				map[string]interface{}{"value": "user", "display": "Standard User", "primary": true},
			},
			enterpriseUserSchemaURI: map[string]interface{}{
				"employeeNumber": "E-1002",
				"department":     "Backend Services",
				"organization":   "Contoso",
			},
		},
		{
			"schemas":    []string{coreUserSchemaURI, enterpriseUserSchemaURI},
			"userName":   "diego.siciliani@example.org",
			"externalId": "okta-00u9diego",
			"name": map[string]interface{}{
				"formatted":  "Diego Siciliani",
				"givenName":  "Diego",
				"familyName": "Siciliani",
			},
			"displayName":       "Diego Siciliani",
			"title":             "Security Operations Lead",
			"preferredLanguage": "es-AR",
			"active":            false,
			"emails": []interface{}{
				map[string]interface{}{"value": "diego.siciliani@example.org", "type": "work", "primary": true},
			},
			enterpriseUserSchemaURI: map[string]interface{}{
				"employeeNumber": "E-1042",
				"department":     "Security",
				"organization":   "Example Corp",
			},
		},
	}

	for _, u := range demoUsers {
		uName, _ := u["userName"].(string)
		id := uuid.NewString()
		u["id"] = id
		meta := generateMetaForPrefix(prefix, u, userRT)
		if metaMap, err := structToMap(meta); err == nil {
			u["meta"] = metaMap
		}
		claims := ExtractUniqueClaims(userRT, u)
		if err := DB.CreateUnique(userBucket, id, u, claims); err != nil {
			continue
		}
		createdUserIDs[uName] = id
		createdCount++
	}

	// Create child group "Backend" containing Megan, then parent group
	// "Engineering" containing Adele + child group "Backend" to demonstrate
	// direct and indirect User.groups memberships out of the box.
	var backendGroupID string
	if meganID := createdUserIDs["megan.bowen@contoso.com"]; meganID != "" {
		gID := uuid.NewString()
		backendGroup := map[string]interface{}{
			"schemas":     []string{coreGroupSchemaURI},
			"id":          gID,
			"externalId":  "grp-backend",
			"displayName": "Backend",
			"members": []interface{}{
				map[string]interface{}{
					"value":   meganID,
					"display": "Megan Bowen",
					"type":    "User",
					"$ref":    resourcePathWithPrefix(prefix, userRT, meganID),
				},
			},
		}
		meta := generateMetaForPrefix(prefix, backendGroup, groupRT)
		if metaMap, err := structToMap(meta); err == nil {
			backendGroup["meta"] = metaMap
		}
		if err := DB.CreateUnique(groupBucket, gID, backendGroup, ExtractUniqueClaims(groupRT, backendGroup)); err == nil {
			backendGroupID = gID
			syncAfterCreateInTenant(tenant, prefix, groupRT, gID, backendGroup)
			createdCount++
		}
	}

	var engMembers []interface{}
	if adeleID := createdUserIDs["adele.vance@contoso.com"]; adeleID != "" {
		engMembers = append(engMembers, map[string]interface{}{
			"value":   adeleID,
			"display": "Adele Vance",
			"type":    "User",
			"$ref":    resourcePathWithPrefix(prefix, userRT, adeleID),
		})
	}
	if backendGroupID != "" {
		engMembers = append(engMembers, map[string]interface{}{
			"value":   backendGroupID,
			"display": "Backend",
			"type":    "Group",
			"$ref":    resourcePathWithPrefix(prefix, groupRT, backendGroupID),
		})
	}
	if len(engMembers) > 0 {
		gID := uuid.NewString()
		engGroup := map[string]interface{}{
			"schemas":     []string{coreGroupSchemaURI},
			"id":          gID,
			"externalId":  "grp-engineering",
			"displayName": "Engineering",
			"members":     engMembers,
		}
		meta := generateMetaForPrefix(prefix, engGroup, groupRT)
		if metaMap, err := structToMap(meta); err == nil {
			engGroup["meta"] = metaMap
		}
		if err := DB.CreateUnique(groupBucket, gID, engGroup, ExtractUniqueClaims(groupRT, engGroup)); err == nil {
			syncAfterCreateInTenant(tenant, prefix, groupRT, gID, engGroup)
			createdCount++
		}
	}

	return createdCount, nil
}
