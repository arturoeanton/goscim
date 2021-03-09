package discovery

import "github.com/gin-gonic/gin"

//ServiceProviderConfig is  Specification compliance, authentication schemes, data models.
func ServiceProviderConfig(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

//ResourceTypes is  An endpoint used to discover the types of resources available.
func ResourceTypes(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

//Schemas is  Introspect resources and attribute extensions.
func Schemas(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
