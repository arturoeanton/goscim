package scim

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// NewRouter builds the SCIM router from the JSON config in folderConfig, with
// every resource endpoint behind the given Authenticator. The caller is
// responsible for having installed a Store in DB beforehand.
//
// The discovery endpoints are mounted under the same /scim/v2 prefix as the
// resources, where RFC 7644 4 puts them, but outside the authenticated group:
// section 2 allows a service provider to serve them anonymously, and a client
// needs them precisely to work out how to authenticate.
func NewRouter(folderConfig string, r *gin.Engine, authenticator Authenticator) (*gin.Engine, error) {
	if authenticator == nil {
		return nil, errors.New("scim: an Authenticator is required; use AnonymousAuthenticator to serve requests unauthenticated")
	}
	if r == nil {
		r = gin.Default()
	}
	discoveryAuthenticator = authenticator

	discovery := r.Group(PREFIX)
	discovery.GET("/ServiceProviderConfig", DiscoveryServiceProviderConfig(folderConfig))
	discovery.GET("/ResourceTypes", DiscoveryResourceTypes)
	discovery.GET("/ResourceTypes/:id", DiscoveryResourceType)
	discovery.GET("/Schemas", DiscoverySchemas)
	discovery.GET("/Schemas/:id", DiscoverySchema)

	scim := r.Group(PREFIX, Authenticate(authenticator))
	if err := ReadResourceType(folderConfig, scim); err != nil {
		return nil, err
	}
	//scim.POST("/Bulk", Bulk) // Bulk: 		POST https://example.com/{v}/Bulk

	return r, nil
}
