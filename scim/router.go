package scim

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// NewRouter builds the SCIM router from the JSON config in folderConfig, with
// every resource endpoint behind the given Authenticator. The caller is
// responsible for having installed a Store in DB beforehand.
//
// The discovery endpoints stay outside the authenticated group: RFC 7644 2
// allows a service provider to serve them anonymously, and they carry no
// tenant data.
func NewRouter(folderConfig string, r *gin.Engine, authenticator Authenticator) (*gin.Engine, error) {
	if authenticator == nil {
		return nil, errors.New("scim: an Authenticator is required; use AnonymousAuthenticator to serve requests unauthenticated")
	}
	if r == nil {
		r = gin.Default()
	}

	scim := r.Group(PREFIX, Authenticate(authenticator))
	if err := ReadResourceType(folderConfig, scim); err != nil {
		return nil, err
	}
	//scim.POST("/Bulk", Bulk) // Bulk: 		POST https://example.com/{v}/Bulk

	r.GET("/ServiceProviderConfig", DiscoveryServiceProviderConfig) // GET /ServiceProviderConfig -> Specification compliance, authentication schemes, data models.
	r.GET("/ResourceTypes", DiscoveryResourceTypes)                 // GET /ResourceTypes 		-> An endpoint used to discover the types of resources available.
	r.GET("/Schemas", DiscoverySchemas)                             // GET /Schemas 				-> Introspect resources and attribute extensions.
	return r, nil
}
