package scim

import "github.com/gin-gonic/gin"

// NewRouter builds the SCIM router from the JSON config in folderConfig. The
// caller is responsible for having installed a Store in DB beforehand.
func NewRouter(folderConfig string, r *gin.Engine) (*gin.Engine, error) {
	if r == nil {
		r = gin.Default()
	}
	if err := ReadResourceType(folderConfig, r); err != nil {
		return nil, err
	}
	//r.POST(PREFIX+"/Bulk", Bulk) // Bulk: 		POST https://example.com/{v}/Bulk

	r.GET("/ServiceProviderConfig", DiscoveryServiceProviderConfig) // GET /ServiceProviderConfig -> Specification compliance, authentication schemes, data models.
	r.GET("/ResourceTypes", DiscoveryResourceTypes)                 // GET /ResourceTypes 		-> An endpoint used to discover the types of resources available.
	r.GET("/Schemas", DiscoverySchemas)                             // GET /Schemas 				-> Introspect resources and attribute extensions.
	return r, nil
}
