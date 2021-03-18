package main

import (
	"log"

	"github.com/arturoeanton/goscim/scim"
	"github.com/gin-gonic/gin"
)

func main() {
	scim.InitDB()
	PREFIX := "/scim/v2"
	log.Println("GoScim v0.1")
	folderConfig := "config"

	scim.ReadResourceType(folderConfig)

	r := gin.Default()
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
	r.Run(":8080")
}

// wget http://www.antlr.org/download/antlr-4.7-complete.jar
// alias antlr='java -jar $PWD/antlr-4.7-complete.jar'
// antlr -Dlanguage=Go -o scim/parser ScimFilter.g4
// docker run -d --name db -p 8091-8094:8091-8094 -p 11210:11210 couchbase
