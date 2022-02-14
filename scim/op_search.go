package scim

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arturoeanton/goscim/scim/parser"
	"github.com/gin-gonic/gin"
)

// Search is 	GET https://example.com/{v}/{resource}?ﬁlter={attribute}{op}{value}&sortBy={attributeName}&sortOrder={ascending|descending}
func Search(resource string) func(c *gin.Context) {
	return func(c *gin.Context) {
		var result ListResponse
		filter := c.Query("filter")
		startIndex := c.Query("startIndex")
		count := c.Query("count")
		sortBy := c.Query("sortBy")
		sortOrder := c.Query("sortOrder")
		resourceType := Resources[resource]
		queryPage, queryCount := parser.FilterToN1QL(resourceType.Name, filter)

		if sortBy == "" {
			sortBy = "id"
		} else {
			sortByArray := strings.Split(sortBy, ",")
			cache := make([]string, 0)
			for _, s := range sortByArray {
				cache = append(cache, parser.AddQuote(s))
			}
			sortBy = strings.Join(cache, ",")
		}

		sortBy = strings.Trim(sortBy, " ")
		sortBy = strings.ReplaceAll(sortBy, ";", "")

		if sortOrder == "descending" {
			sortOrder = "DESC"
		} else {
			sortOrder = "ASC"
		}

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
		queryPage += "\nORDER BY " + sortBy + " " + sortOrder
		queryPage += "\nOFFSET " + strconv.Itoa(result.StartIndex-1)
		queryPage += "\nLIMIT " + count

		//log.Println(queryCount)
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
		//log.Println(queryPage)
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

			//TODO: Validate _read of all fields of element
			roles := []string{"user", "admin", "superadmin", "role1"} // TODO: get the user roles from the token
			element := ValidateReadRole(roles, resourceType, item[resourceType.Name].(map[string]interface{}))
			result.Resources = append(result.Resources, element)
		}
		c.JSON(http.StatusOK, result)
	}
}
