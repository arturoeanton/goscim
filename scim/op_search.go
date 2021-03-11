package scim

import (
	"log"
	"net/http"

	"github.com/arturoeanton/goscim/scim/parser"
	"github.com/gin-gonic/gin"
)

// Search is 	GET https://example.com/{v}/{resource}?ﬁlter={attribute}{op}{value}&sortBy={attributeName}&sortOrder={ascending|descending}
func Search(c *gin.Context) {

	resource := c.Param("resource")
	filter := c.Query("filter")
	resourceType := Resources["/"+resource]
	query := parser.FilterToN1QL(resourceType.Name, filter)
	//log.Println(filter + " - " + query)
	rows, err := Cluster.Query(query, nil)
	defer rows.Close()
	if err != nil {
		MakeError(c, http.StatusInternalServerError, err.Error())
		log.Println(err.Error())
		return
	}

	var result ListResponse
	result.Schemas = append(result.Schemas, "urn:ietf:params:scim:api:messages:2.0:ListResponse")
	result.TotalResults = 0
	result.Resources = make([]interface{}, 0)
	for rows.Next() {
		var item Resource
		err := rows.Row(&item)
		if err != nil {
			MakeError(c, http.StatusInternalServerError, err.Error())
			log.Println(err.Error())
			return
		}
		result.Resources = append(result.Resources, item[resourceType.Name])
		result.TotalResults++
	}
	c.JSON(http.StatusOK, result)
}
