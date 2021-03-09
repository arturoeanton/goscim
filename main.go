package main

import (
	"log"

	"github.com/arturoeanton/goscim/scim/discovery"
	"github.com/arturoeanton/goscim/scim/operations"
	"github.com/gin-gonic/gin"
)

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func main() {
	PREFIX := "/scim/v2"
	log.Println("GoScim v0.1")
	r := gin.Default()
	r.POST(PREFIX+"/:resource", operations.Create)       // Create:  POST https://example.com/{v}/{resource}
	r.GET(PREFIX+"/:resource/:id", operations.Read)      // Read: 	GET https://example.com/{v}/{resource}/{id}
	r.PUT(PREFIX+"/:resource/:id", operations.Replace)   // Replace: PUT https://example.com/{v}/{resource}/{id}
	r.DELETE(PREFIX+"/:resource/:id", operations.Delete) // Delete: 	DELETE https://example.com/{v}/{resource}/{id}
	r.PATCH(PREFIX+"/:resource/:id", operations.Update)  // Update: 	PATCH https://example.com/{v}/{resource}/{id}
	r.GET(PREFIX+"/:resource", operations.Search)        // Search: 	GET https://example.com/{v}/{resource}?ﬁlter={attribute}{op}{value}&sortBy={attributeName}&sortOrder={ascending|descending}
	//r.POST(PREFIX+"/Bulk", operations.Bulk)              // Bulk: 	POST https://example.com/{v}/Bulk

	r.GET("/ServiceProviderConfig", discovery.ServiceProviderConfig) // GET /ServiceProviderConfig -> Specification compliance, authentication schemes, data models.
	r.GET("/ResourceTypes", discovery.ResourceTypes)                 // GET /ResourceTypes 		-> An endpoint used to discover the types of resources available.
	r.GET("/Schemas", discovery.Schemas)                             // GET /Schemas 				-> Introspect resources and attribute extensions.

	r.GET("/ping", ping)
	r.Run(":8080")
}

// TODO: Hacer POST
// TODO: Hacer POST
// TODO: Hacer POST
// TODO: Hacer POST
// TODO: Hacer POST
// TODO: Hacer POST
// TODO: Hacer POST
