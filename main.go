package main

import (
	"log"
	"os"

	"github.com/arturoeanton/goscim/scim"
	"github.com/gin-gonic/gin"
)

func main() {
	scim.InitDB()
	PREFIX := "/scim/v2"
	log.Println("GoScim v0.1")
	folderConfig := "config"

	port := os.Getenv("SCIM_PORT")
	if port == "" {
		port = ":8080"
	}

	trustedProxies := os.Getenv("SCIM_TRUSTED_PROXIES")
	if trustedProxies == "" {
		trustedProxies = "127.0.0.1"
	}
	log.Println("Trusted Proxies:", trustedProxies)
	gin.SetMode(gin.ReleaseMode)

	scim.ReadResourceType(folderConfig)

	r := gin.Default()

	r.SetTrustedProxies([]string{trustedProxies})
	r.POST(PREFIX+"/:resource", scim.Create)       // Create:  	POST https://example.com/{v}/{resource}
	r.GET(PREFIX+"/:resource/:id", scim.Read)      // Read: 		GET https://example.com/{v}/{resource}/{id}
	r.PUT(PREFIX+"/:resource/:id", scim.Replace)   // Replace: 	PUT https://example.com/{v}/{resource}/{id}
	r.DELETE(PREFIX+"/:resource/:id", scim.Delete) // Delete: 	DELETE https://example.com/{v}/{resource}/{id}
	r.PATCH(PREFIX+"/:resource/:id", scim.Update)  // Update: 	PATCH https://example.com/{v}/{resource}/{id}
	r.GET(PREFIX+"/:resource", scim.Search)        // Search: 	GET https://example.com/{v}/{resource}?ﬁlter={attribute}{op}{value}&sortBy={attributeName}&sortOrder={ascending|descending}
	//r.POST(PREFIX+"/Bulk", operations.Bulk)            // Bulk: 		POST https://example.com/{v}/Bulk

	r.GET("/ServiceProviderConfig", scim.DiscoveryServiceProviderConfig) // GET /ServiceProviderConfig -> Specification compliance, authentication schemes, data models.
	r.GET("/ResourceTypes", scim.DiscoveryResourceTypes)                 // GET /ResourceTypes 		-> An endpoint used to discover the types of resources available.
	r.GET("/Schemas", scim.DiscoverySchemas)                             // GET /Schemas 				-> Introspect resources and attribute extensions.
	r.Run(port)
}
