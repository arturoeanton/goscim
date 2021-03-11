package scim

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DiscoveryServiceProviderConfig is  Specification compliance, authentication schemes, data models.
func DiscoveryServiceProviderConfig(c *gin.Context) {
	MakeError(c, http.StatusNotImplemented, "Not Implemented")
}

// DiscoveryResourceTypes is  An endpoint used to discover the types of resources available.
func DiscoveryResourceTypes(c *gin.Context) {
	MakeError(c, http.StatusNotImplemented, "Not Implemented")
}

// DiscoverySchemas is  Introspect resources and attribute extensions.
func DiscoverySchemas(c *gin.Context) {
	MakeError(c, http.StatusNotImplemented, "Not Implemented")
}
