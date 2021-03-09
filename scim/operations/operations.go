package operations

import "github.com/gin-gonic/gin"

//Create is  POST https://example.com/{v}/{resource}
func Create(c *gin.Context) {
	resource := c.Param("resource")
	if resource == "Bulk" {
		Bulk(c)
		return
	}

	c.JSON(200, gin.H{
		"message": resource,
	})
}

// Read is GET https://example.com/{v}/{resource}/{id}
func Read(c *gin.Context) {
	resource := c.Param("resource")
	id := c.Param("id")
	c.JSON(200, gin.H{
		"message": resource,
		"id":      id,
	})
}

// Replace is PUT https://example.com/{v}/{resource}/{id}
func Replace(c *gin.Context) {
	resource := c.Param("resource")
	id := c.Param("id")
	c.JSON(200, gin.H{
		"message": resource,
		"id":      id,
	})
}

// Delete is  	DELETE https://example.com/{v}/{resource}/{id}
func Delete(c *gin.Context) {
	resource := c.Param("resource")
	id := c.Param("id")
	c.JSON(200, gin.H{
		"message": resource,
		"id":      id,
	})
}

// Update is	PATCH https://example.com/{v}/{resource}/{id}
func Update(c *gin.Context) {
	resource := c.Param("resource")
	id := c.Param("id")
	c.JSON(200, gin.H{
		"message": resource,
		"id":      id,
	})
}

// Search is 	GET https://example.com/{v}/{resource}?ﬁlter={attribute}{op}{value}&sortBy={attributeName}&sortOrder={ascending|descending}
func Search(c *gin.Context) {
	resource := c.Param("resource")
	c.JSON(200, gin.H{
		"message": resource,
	})
}

// Bulk is 	POST https://example.com/{v}/Bulk
func Bulk(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
