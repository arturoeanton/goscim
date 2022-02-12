package main

import (
	"log"
	"os"

	"github.com/arturoeanton/goscim/commons"
	"github.com/arturoeanton/goscim/scim"
	"github.com/gin-gonic/gin"
)

func main() {

	commons.ExampleWalkMap()

	//panic("fin ok")

	scim.InitDB()

	log.Println("GoScim v0.1")
	folderConfig := "config"

	port := os.Getenv("SCIM_PORT")
	if port == "" {
		port = ":8080"
	}

	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1"})
	scim.ReadResourceType(folderConfig, r)
	//r.POST(scim.PREFIX+"/Bulk", operations.Bulk) // Bulk: 		POST https://example.com/{v}/Bulk

	r.GET("/ServiceProviderConfig", scim.DiscoveryServiceProviderConfig) // GET /ServiceProviderConfig -> Specification compliance, authentication schemes, data models.
	r.GET("/ResourceTypes", scim.DiscoveryResourceTypes)                 // GET /ResourceTypes 		-> An endpoint used to discover the types of resources available.
	r.GET("/Schemas", scim.DiscoverySchemas)                             // GET /Schemas 				-> Introspect resources and attribute extensions.
	r.Run(port)
}
