package scim

import (
	"log"
	"net/http"
	"strconv"

	"github.com/arturoeanton/goscim/scim/parser"
	"github.com/gin-gonic/gin"
)

// Search is 	GET https://example.com/{v}/{resource}?ﬁlter={attribute}{op}{value}&sortBy={attributeName}&sortOrder={ascending|descending}
func Search(c *gin.Context) {
	var result ListResponse
	resource := c.Param("resource")
	filter := c.Query("filter")
	startIndex := c.Query("startIndex")
	count := c.Query("count")
	resourceType := Resources["/"+resource]
	queryPage, queryCount := parser.FilterToN1QL(resourceType.Name, filter)

	log.Println(queryPage)
	log.Println(queryCount)

	//pagination
	if startIndex == "" {
		startIndex = "1"
	}
	if count == "" {
		count = "100"
	}
	var err error
	result.StartIndex, err = strconv.Atoi(startIndex)
	if err != nil {
		MakeError(c, http.StatusBadRequest, err.Error())
		log.Println(err.Error())
		return
	}
	if result.StartIndex < 1 {
		result.StartIndex = 1
	}
	result.ItemsPerPage, err = strconv.Atoi(count)
	if err != nil {
		MakeError(c, http.StatusBadRequest, err.Error())
		log.Println(err.Error())
		return
	}
	queryPage += "\nOFFSET " + strconv.Itoa(result.StartIndex-1)
	queryPage += "\nLIMIT " + count

	rowsCount, err := Cluster.Query(queryCount, nil)
	if err != nil {
		MakeError(c, http.StatusInternalServerError, err.Error())
		log.Println(err.Error())
		return
	}
	defer rowsCount.Close()

	var countResult struct {
		Count int
	}
	rowsCount.One(&countResult)
	rows, err := Cluster.Query(queryPage, nil)
	if err != nil {
		MakeError(c, http.StatusInternalServerError, err.Error())
		log.Println(err.Error())
		return
	}
	defer rows.Close()

	result.Schemas = append(result.Schemas, "urn:ietf:params:scim:api:messages:2.0:ListResponse")
	result.TotalResults = countResult.Count
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
	}
	c.JSON(http.StatusOK, result)
}
