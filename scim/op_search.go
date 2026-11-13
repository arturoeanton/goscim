package scim

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// maxSearchCount bounds how many resources one page may carry, whatever the
// client asks for. It is also what the ServiceProviderConfig advertises as
// filter.maxResults.
const maxSearchCount = 200

const searchRequestSchema = "urn:ietf:params:scim:api:messages:2.0:SearchRequest"

type searchParams struct {
	Filter         string
	SortBy         string
	SortDescending bool
	StartIndex     int
	Count          int
	Projection     Projection
}

// Search is GET https://example.com/{v}/{resource}?filter=...
func Search(resource string) func(c *gin.Context) {
	return func(c *gin.Context) {
		rt := Resources[resource]
		setAuditTarget(c, "search", rt.Name, "")
		params, ok := parseQuerySearchParams(c)
		if !ok {
			return
		}
		executeResourceSearch(c, rt, params)
	}
}

// PostSearch is POST https://example.com/{v}/{resource}/.search (RFC 7644 3.4.3).
func PostSearch(resource string) func(c *gin.Context) {
	return func(c *gin.Context) {
		rt := Resources[resource]
		setAuditTarget(c, "search", rt.Name, "")
		params, ok := parseBodySearchParams(c)
		if !ok {
			return
		}
		executeResourceSearch(c, rt, params)
	}
}

// RootPostSearch is POST https://example.com/{v}/.search across all configured
// resource types (RFC 7644 3.4.3).
func RootPostSearch(c *gin.Context) {
	setAuditTarget(c, "search", "*", "")
	params, ok := parseBodySearchParams(c)
	if !ok {
		return
	}

	endpoints := make([]string, 0, len(Resources))
	for ep := range Resources {
		endpoints = append(endpoints, ep)
	}
	sort.Strings(endpoints)

	resourceTypes := make([]ResourceType, 0, len(endpoints))
	for _, ep := range endpoints {
		resourceTypes = append(resourceTypes, Resources[ep])
	}

	if err := ValidateRootSortBy(resourceTypes, params.SortBy); err != nil {
		MakeTypedError(c, http.StatusBadRequest, "invalidValue", err.Error())
		return
	}
	roles := currentRoles(c)
	if err := ValidateRootFilterAttributes(resourceTypes, roles, params.Filter); err != nil {
		MakeTypedError(c, http.StatusBadRequest, "invalidFilter", err.Error())
		return
	}

	type taggedItem struct {
		rt  ResourceType
		doc map[string]interface{}
	}

	offset := params.StartIndex - 1
	fetchWindow := offset + params.Count
	totalAll := 0
	var combined []taggedItem

	for _, rt := range resourceTypes {
		bucketSortBy := ""
		if params.SortBy != "" {
			if normalized, err := NormalizeSortBy(rt, params.SortBy); err == nil {
				bucketSortBy = normalized
			}
		}
		total, docs, err := DB.Search(SearchQuery{
			Bucket: rt.Name,
			Filter: params.Filter,
			ResolveName: func(path string) string {
				return CanonicalAttributePath(rt, path)
			},
			SortBy:         bucketSortBy,
			SortDescending: params.SortDescending,
			Offset:         0,
			Limit:          fetchWindow,
		})
		if err != nil {
			if errors.Is(err, ErrInvalidFilter) {
				MakeTypedError(c, http.StatusBadRequest, "invalidFilter", err.Error())
				return
			}
			internalError(c, err)
			return
		}
		totalAll += total
		for _, doc := range docs {
			combined = append(combined, taggedItem{rt: rt, doc: doc})
		}
	}

	sortKey := params.SortBy
	if sortKey == "" {
		sortKey = "id"
	}
	firstSortField := strings.TrimSpace(strings.Split(sortKey, ",")[0])
	sort.SliceStable(combined, func(i, j int) bool {
		pathI := opPathTopathArray(CanonicalAttributePath(combined[i].rt, firstSortField))
		pathJ := opPathTopathArray(CanonicalAttributePath(combined[j].rt, firstSortField))
		cmp := compareValues(resolvePath(combined[i].doc, pathI), resolvePath(combined[j].doc, pathJ))
		if cmp == 0 {
			return false
		}
		less := cmp < 0
		if params.SortDescending {
			return !less
		}
		return less
	})

	page := []taggedItem(nil)
	if offset < len(combined) {
		page = combined[offset:]
		if params.Count < len(page) {
			page = page[:params.Count]
		}
	}

	outResources := make([]interface{}, 0, len(page))
	for _, item := range page {
		outResources = append(outResources, FormatResource(roles, item.rt, item.doc, params.Projection))
	}

	writeSCIM(c, http.StatusOK, ListResponse{
		Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		TotalResults: totalAll,
		ItemsPerPage: len(outResources),
		StartIndex:   params.StartIndex,
		Resources:    outResources,
	})
}

func executeResourceSearch(c *gin.Context, resourceType ResourceType, params searchParams) {
	sortBy, err := NormalizeSortBy(resourceType, params.SortBy)
	if err != nil {
		MakeTypedError(c, http.StatusBadRequest, "invalidValue", err.Error())
		return
	}
	roles := currentRoles(c)
	if err := ValidateFilterAttributes(resourceType, roles, params.Filter); err != nil {
		MakeTypedError(c, http.StatusBadRequest, "invalidFilter", err.Error())
		return
	}

	total, resources, err := DB.Search(SearchQuery{
		Bucket: resourceType.Name,
		Filter: params.Filter,
		ResolveName: func(path string) string {
			return CanonicalAttributePath(resourceType, path)
		},
		SortBy:         sortBy,
		SortDescending: params.SortDescending,
		Offset:         params.StartIndex - 1,
		Limit:          params.Count,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidFilter) {
			MakeTypedError(c, http.StatusBadRequest, "invalidFilter", err.Error())
			return
		}
		internalError(c, err)
		return
	}

	outResources := make([]interface{}, 0, len(resources))
	for _, item := range resources {
		outResources = append(outResources, FormatResource(roles, resourceType, item, params.Projection))
	}
	// RFC 7644 3.4.2.4: itemsPerPage is how many resources this page
	// carries, not how many were asked for.
	writeSCIM(c, http.StatusOK, ListResponse{
		Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		TotalResults: total,
		ItemsPerPage: len(outResources),
		StartIndex:   params.StartIndex,
		Resources:    outResources,
	})
}

func parseQuerySearchParams(c *gin.Context) (searchParams, bool) {
	proj, ok := parseProjectionFromQuery(c)
	if !ok {
		return searchParams{}, false
	}

	startIndexStr := c.Query("startIndex")
	if startIndexStr == "" {
		startIndexStr = "1"
	}
	startIndex, err := strconv.Atoi(startIndexStr)
	if err != nil {
		MakeError(c, http.StatusBadRequest, err.Error())
		log.Println(err.Error())
		return searchParams{}, false
	}
	if startIndex < 1 {
		startIndex = 1
	}

	countStr := c.Query("count")
	if countStr == "" {
		countStr = "100"
	}
	requestedCount, err := strconv.Atoi(countStr)
	if err != nil {
		MakeError(c, http.StatusBadRequest, err.Error())
		log.Println(err.Error())
		return searchParams{}, false
	}
	requestedCount = clampSearchCount(requestedCount)

	descending, ok := parseSortOrder(c, c.Query("sortOrder"))
	if !ok {
		return searchParams{}, false
	}

	return searchParams{
		Filter:         c.Query("filter"),
		SortBy:         c.Query("sortBy"),
		SortDescending: descending,
		StartIndex:     startIndex,
		Count:          requestedCount,
		Projection:     proj,
	}, true
}

func parseBodySearchParams(c *gin.Context) (searchParams, bool) {
	raw, ok := readBody(c)
	if !ok {
		return searchParams{}, false
	}
	var rawObj map[string]interface{}
	if err := json.Unmarshal(raw, &rawObj); err != nil || rawObj == nil {
		MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "request body must be a JSON object")
		return searchParams{}, false
	}
	if dup, hasDup := hasDuplicateCaseKeys(rawObj); hasDup {
		MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "duplicate attribute "+strconv.Quote(dup))
		return searchParams{}, false
	}

	var req SearchRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", err.Error())
		return searchParams{}, false
	}
	if len(req.Schemas) == 0 {
		MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "schemas must include "+searchRequestSchema)
		return searchParams{}, false
	}
	hasSearchSchema := false
	for _, s := range req.Schemas {
		if strings.EqualFold(s, searchRequestSchema) {
			hasSearchSchema = true
			continue
		}
		MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "unexpected schema "+strconv.Quote(s))
		return searchParams{}, false
	}
	if !hasSearchSchema {
		MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "schemas must include "+searchRequestSchema)
		return searchParams{}, false
	}

	proj, err := ParseProjectionParams(req.Attributes, req.ExcludedAttributes)
	if err != nil {
		MakeTypedError(c, http.StatusBadRequest, "invalidValue", err.Error())
		return searchParams{}, false
	}

	startIndex := 1
	if req.StartIndex != nil {
		startIndex = *req.StartIndex
		if startIndex < 1 {
			startIndex = 1
		}
	}
	requestedCount := 100
	if req.Count != nil {
		requestedCount = clampSearchCount(*req.Count)
	}

	descending, ok := parseSortOrder(c, req.SortOrder)
	if !ok {
		return searchParams{}, false
	}

	return searchParams{
		Filter:         req.Filter,
		SortBy:         req.SortBy,
		SortDescending: descending,
		StartIndex:     startIndex,
		Count:          requestedCount,
		Projection:     proj,
	}, true
}

func clampSearchCount(requested int) int {
	if requested < 0 {
		return 0
	}
	if requested > maxSearchCount {
		return maxSearchCount
	}
	return requested
}

func parseSortOrder(c *gin.Context, raw string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "ascending":
		return false, true
	case "descending":
		return true, true
	default:
		MakeTypedError(c, http.StatusBadRequest, "invalidValue", "sortOrder must be ascending or descending")
		return false, false
	}
}
