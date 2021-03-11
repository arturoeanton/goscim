package scim

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Search is 	GET https://example.com/{v}/{resource}?ﬁlter={attribute}{op}{value}&sortBy={attributeName}&sortOrder={ascending|descending}
func Search(c *gin.Context) {
	resource := c.Param("resource")
	resourceType := Resources["/"+resource]
	rows, err := Cluster.Query("SELECT * FROM `"+resourceType.Name+"`;", nil)
	defer rows.Close()
	if err != nil {
		MakeError(c, http.StatusInternalServerError, err.Error())
		log.Println(err.Error())
		return
	}

	var result ListResponse
	result.Schemas = append(result.Schemas, "urn:ietf:params:scim:api:messages:2.0:ListResponse")
	result.TotalResults = 0
	for rows.Next() {
		var item Resource
		err := rows.Row(&item)
		if err != nil {
			MakeError(c, http.StatusInternalServerError, err.Error())
			log.Println(err.Error())
			return
		}
		result.Resources = append(result.Resources, item)
		result.TotalResults++
	}
	c.JSON(http.StatusOK, result)
}
