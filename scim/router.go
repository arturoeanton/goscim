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
	r.Use(RequestID())
	discoveryAuthenticator = authenticator

	r.GET("/healthz", Healthz)
	r.GET("/readyz", Readyz)
	if MetricsEnabledFromEnv() {
		r.GET("/metrics", MetricsHandler(DefaultMetrics))
	}

	mountDiscoveryRoutes := func(group *gin.RouterGroup) {
		group.GET("/ServiceProviderConfig", DiscoveryServiceProviderConfig(folderConfig))
		group.GET("/ResourceTypes", DiscoveryResourceTypes)
		group.GET("/ResourceTypes/:id", DiscoveryResourceType)
		group.GET("/Schemas", DiscoverySchemas)
		group.GET("/Schemas/:id", DiscoverySchema)
	}

	mountDiscoveryRoutes(r.Group(PREFIX))

	scimGroup := r.Group(PREFIX, AuditMiddleware(), LimitBody(DefaultMaxBodyBytes), Authenticate(authenticator))
	if err := ReadResourceType(folderConfig, scimGroup); err != nil {
		return nil, err
	}
	bulkEngine := newBulkEngine()
	scimGroup.POST("/.search", RootPostSearch)
	scimGroup.POST("/Bulk", BulkHandler(bulkEngine))

	// Multi-tenant B2B routes under /tenants/:tenant/scim/v2/...
	tenantDiscovery := r.Group("/tenants/:tenant"+PREFIX, TenantMiddleware())
	mountDiscoveryRoutes(tenantDiscovery)

	tenantScim := r.Group("/tenants/:tenant"+PREFIX, TenantMiddleware(), AuditMiddleware(), LimitBody(DefaultMaxBodyBytes), Authenticate(authenticator))
	mountResourceRoutes(tenantScim)
	tenantScim.POST("/.search", RootPostSearch)
	tenantScim.POST("/Bulk", BulkHandler(bulkEngine))

	if UIEnabledFromEnv() {
		if err := MountUI(r, UIOptions{}); err != nil {
			return nil, err
		}
	}

	return r, nil
}
