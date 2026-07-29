package scim

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// maxSearchCount bounds how many resources one page may carry, whatever the
// client asks for. It is also what the ServiceProviderConfig advertises as
// filter.maxResults.
const maxSearchCount = 200

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
		requestedCount, err := strconv.Atoi(count)
		if err != nil {
			MakeError(c, http.StatusBadRequest, err.Error())
			log.Println(err.Error())
			return
		}
		// A client asking for everything must not be able to make the server
		// materialise an unbounded page.
		if requestedCount < 0 {
			requestedCount = 0
		}
		if requestedCount > maxSearchCount {
			requestedCount = maxSearchCount
		}

		sortBy, err = NormalizeSortBy(resourceType, sortBy)
		if err != nil {
			MakeTypedError(c, http.StatusBadRequest, "invalidValue", err.Error())
			return
		}

		total, resources, err := DB.Search(SearchQuery{
			Bucket:         resourceType.Name,
			Filter:         filter,
			SortBy:         sortBy,
			SortDescending: sortOrder == "descending",
			Offset:         result.StartIndex - 1,
			Limit:          requestedCount,
		})
		if err != nil {
			if errors.Is(err, ErrInvalidFilter) {
				MakeTypedError(c, http.StatusBadRequest, "invalidFilter", err.Error())
				return
			}
			MakeError(c, http.StatusInternalServerError, err.Error())
			log.Println(err.Error())
			return
		}

		result.Schemas = append(result.Schemas, "urn:ietf:params:scim:api:messages:2.0:ListResponse")
		result.TotalResults = total
		result.Resources = make([]interface{}, 0)
		roles := currentRoles(c)
		for _, item := range resources {
			result.Resources = append(result.Resources, ValidateReadRole(roles, resourceType, item))
		}
		// RFC 7644 3.4.2.4: itemsPerPage is how many resources this page
		// carries, not how many were asked for.
		result.ItemsPerPage = len(result.Resources)
		writeSCIM(c, http.StatusOK, result)
	}
}
