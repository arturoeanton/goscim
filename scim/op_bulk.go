package scim

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// MaxBulkOperations bounds how many operations a single POST /Bulk request may
// carry. It is also advertised in ServiceProviderConfig as bulk.maxOperations.
const MaxBulkOperations = 1000

const (
	bulkRequestSchema  = "urn:ietf:params:scim:api:messages:2.0:BulkRequest"
	bulkResponseSchema = "urn:ietf:params:scim:api:messages:2.0:BulkResponse"
)

type bulkCtxKey int

const (
	bulkPrincipalKey bulkCtxKey = iota
	bulkRequestIDKey
)

func withBulkMeta(ctx context.Context, p *Principal, reqID string) context.Context {
	ctx = context.WithValue(ctx, bulkPrincipalKey, p)
	ctx = context.WithValue(ctx, bulkRequestIDKey, reqID)
	return ctx
}

func injectBulkContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		if p, ok := c.Request.Context().Value(bulkPrincipalKey).(*Principal); ok && p != nil {
			c.Set(principalContextKey, p)
		}
		if reqID, ok := c.Request.Context().Value(bulkRequestIDKey).(string); ok && reqID != "" {
			c.Set(requestIDContextKey, reqID)
		}
		c.Next()
	}
}

func newBulkEngine() *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())
	group := engine.Group(PREFIX, injectBulkContext(), AuditMiddleware())
	for endpoint := range Resources {
		group.POST(endpoint, Create(endpoint))
		group.PUT(endpoint+"/:id", Replace(endpoint))
		group.PATCH(endpoint+"/:id", Update(endpoint))
		group.DELETE(endpoint+"/:id", Delete(endpoint))
	}
	return engine
}

// BulkHandler serves POST /scim/v2/Bulk (RFC 7644 3.7).
func BulkHandler(subRouter http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		setAuditTarget(c, "bulk", "*", "")

		raw, ok := readBody(c)
		if !ok {
			return
		}
		var rawObj map[string]interface{}
		if err := json.Unmarshal(raw, &rawObj); err != nil || rawObj == nil {
			MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "request body must be a JSON object")
			return
		}
		if dup, hasDup := hasDuplicateCaseKeys(rawObj); hasDup {
			MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "duplicate attribute "+strconv.Quote(dup))
			return
		}

		var req BulkRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", err.Error())
			return
		}
		if !validateBulkSchemas(c, req.Schemas) {
			return
		}
		if req.FailOnErrors != nil && *req.FailOnErrors < 0 {
			MakeTypedError(c, http.StatusBadRequest, "invalidValue", "failOnErrors must be non-negative")
			return
		}
		if len(req.Operations) == 0 {
			MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "Operations is required and must not be empty")
			return
		}
		if len(req.Operations) > MaxBulkOperations {
			MakeError(c, http.StatusRequestEntityTooLarge,
				fmt.Sprintf("Operations count %d exceeds maxOperations (%d)", len(req.Operations), MaxBulkOperations))
			return
		}
		if !validateBulkEnvelopeOps(c, rawObj, req.Operations) {
			return
		}

		principal := PrincipalOf(c)
		reqID := RequestIDOf(c)
		resolvedBulkIDs := make(map[string]string)
		results := make([]BulkOperationResult, 0, len(req.Operations))
		errorCount := 0

		recordFailure := func(method, bulkID, location string, status int, scimType, detail string) {
			errorCount++
			results = append(results, BulkOperationResult{
				Method:   method,
				BulkID:   bulkID,
				Location: location,
				Status:   strconv.Itoa(status),
				Response: Error{
					Schemas:  []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
					ScimType: scimType,
					Status:   strconv.Itoa(status),
					Detail:   detail,
				},
			})
		}

		shouldStop := func() bool {
			return req.FailOnErrors != nil && *req.FailOnErrors > 0 && errorCount >= *req.FailOnErrors
		}

		for _, op := range req.Operations {
			method := strings.ToUpper(strings.TrimSpace(op.Method))
			bulkID := strings.TrimSpace(op.BulkID)

			rt, targetID, rawQuery, hasQuery, pathErr := resolveBulkTarget(method, op.Path, resolvedBulkIDs)
			if pathErr != nil {
				loc := ""
				if method != http.MethodPost && rt.Endpoint != "" && targetID != "" {
					loc = absoluteLocation(c, rt, targetID)
				}
				recordFailure(method, bulkID, loc, http.StatusBadRequest, pathErr.scimType, pathErr.message)
				if shouldStop() {
					break
				}
				continue
			}

			fallbackLoc := ""
			if method != http.MethodPost && targetID != "" {
				fallbackLoc = absoluteLocation(c, rt, targetID)
			}

			var subBody []byte
			if method != http.MethodDelete {
				dataMap, isObj := op.Data.(map[string]interface{})
				if !isObj || dataMap == nil {
					recordFailure(method, bulkID, fallbackLoc, http.StatusBadRequest, "invalidSyntax", "operation data must be a JSON object")
					if shouldStop() {
						break
					}
					continue
				}
				resolvedData, resolveErr := resolveBulkRefsInValue(dataMap, resolvedBulkIDs)
				if resolveErr != nil {
					recordFailure(method, bulkID, fallbackLoc, http.StatusBadRequest, "invalidValue", resolveErr.Error())
					if shouldStop() {
						break
					}
					continue
				}
				encoded, err := json.Marshal(resolvedData)
				if err != nil {
					recordFailure(method, bulkID, fallbackLoc, http.StatusBadRequest, "invalidSyntax", err.Error())
					if shouldStop() {
						break
					}
					continue
				}
				subBody = encoded
			}

			targetPath := PREFIX + rt.Endpoint
			if targetID != "" {
				targetPath += "/" + url.PathEscape(targetID)
			}
			if hasQuery && rawQuery != "" {
				targetPath += "?" + rawQuery
			}

			var bodyReader *bytes.Reader
			if subBody != nil {
				bodyReader = bytes.NewReader(subBody)
			} else {
				bodyReader = bytes.NewReader(nil)
			}

			subCtx := withBulkMeta(c.Request.Context(), principal, reqID)
			subReq, err := http.NewRequestWithContext(subCtx, method, targetPath, bodyReader)
			if err != nil {
				recordFailure(method, bulkID, fallbackLoc, http.StatusBadRequest, "invalidPath", err.Error())
				if shouldStop() {
					break
				}
				continue
			}
			subReq.Host = c.Request.Host
			subReq.TLS = c.Request.TLS
			subReq.RemoteAddr = c.Request.RemoteAddr
			if subBody != nil {
				subReq.Header.Set("Content-Type", scimContentType)
			}
			if ver := strings.TrimSpace(op.Version); ver != "" {
				subReq.Header.Set("If-Match", ver)
			}

			rec := newBulkResponseRecorder()
			subRouter.ServeHTTP(rec, subReq)

			loc := rec.Header().Get("Location")
			if loc == "" && method != http.MethodPost {
				loc = fallbackLoc
			}

			opRes := BulkOperationResult{
				Method:   method,
				BulkID:   bulkID,
				Location: loc,
				Status:   strconv.Itoa(rec.code),
			}

			if rec.code < http.StatusBadRequest {
				opRes.Version = rec.Header().Get("ETag")
				if method == http.MethodPost && rec.code == http.StatusCreated {
					var created map[string]interface{}
					if json.Unmarshal(rec.body.Bytes(), &created) == nil {
						if createdID, _ := created["id"].(string); createdID != "" {
							resolvedBulkIDs[bulkID] = createdID
						}
					}
				}
			} else {
				errorCount++
				var errPayload interface{}
				if json.Unmarshal(rec.body.Bytes(), &errPayload) == nil && errPayload != nil {
					opRes.Response = errPayload
				} else {
					opRes.Response = Error{
						Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
						Status:  strconv.Itoa(rec.code),
						Detail:  http.StatusText(rec.code),
					}
				}
			}

			results = append(results, opRes)
			if shouldStop() {
				break
			}
		}

		writeSCIM(c, http.StatusOK, BulkResponse{
			Schemas:    []string{bulkResponseSchema},
			Operations: results,
		})
	}
}

func validateBulkSchemas(c *gin.Context, schemas []string) bool {
	if len(schemas) == 0 {
		MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "schemas must include "+bulkRequestSchema)
		return false
	}
	found := false
	for _, s := range schemas {
		if strings.EqualFold(s, bulkRequestSchema) {
			found = true
			continue
		}
		MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "unexpected schema "+strconv.Quote(s))
		return false
	}
	if !found {
		MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "schemas must include "+bulkRequestSchema)
		return false
	}
	return true
}

func validateBulkEnvelopeOps(c *gin.Context, rawEnvelope map[string]interface{}, ops []BulkOperation) bool {
	if rawOpsVal, ok := lookupValueFold(rawEnvelope, "Operations"); ok {
		if rawOpsSlice, ok := rawOpsVal.([]interface{}); ok {
			for _, item := range rawOpsSlice {
				if opMap, ok := item.(map[string]interface{}); ok {
					if dup, hasDup := hasDuplicateCaseKeys(opMap); hasDup {
						MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "duplicate attribute "+strconv.Quote(dup))
						return false
					}
				}
			}
		}
	}

	seenPostBulkIDs := make(map[string]bool)
	for i, op := range ops {
		method := strings.ToUpper(strings.TrimSpace(op.Method))
		switch method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		default:
			MakeTypedError(c, http.StatusBadRequest, "invalidSyntax",
				fmt.Sprintf("Operations[%d].method must be POST, PUT, PATCH, or DELETE", i))
			return false
		}
		if strings.TrimSpace(op.Path) == "" {
			MakeTypedError(c, http.StatusBadRequest, "invalidSyntax",
				fmt.Sprintf("Operations[%d].path is required", i))
			return false
		}
		if method == http.MethodPost {
			bid := strings.TrimSpace(op.BulkID)
			if bid == "" {
				MakeTypedError(c, http.StatusBadRequest, "invalidSyntax",
					fmt.Sprintf("Operations[%d].bulkId is required when method is POST", i))
				return false
			}
			if seenPostBulkIDs[bid] {
				MakeTypedError(c, http.StatusBadRequest, "invalidSyntax",
					fmt.Sprintf("duplicate bulkId %q in BulkRequest", bid))
				return false
			}
			seenPostBulkIDs[bid] = true
		}
	}
	return true
}

type bulkOpError struct {
	scimType string
	message  string
}

func resolveBulkTarget(method, rawOpPath string, resolved map[string]string) (ResourceType, string, string, bool, *bulkOpError) {
	pathPart, rawQuery, hasQuery := strings.Cut(strings.TrimSpace(rawOpPath), "?")
	if pathPart == "" {
		return ResourceType{}, "", "", false, &bulkOpError{scimType: "invalidPath", message: "empty operation path"}
	}
	if !strings.HasPrefix(pathPart, "/") {
		pathPart = "/" + pathPart
	}
	if strings.EqualFold(pathPart, PREFIX) {
		return ResourceType{}, "", "", false, &bulkOpError{scimType: "invalidPath", message: "operation path must target a resource endpoint"}
	}
	if len(pathPart) > len(PREFIX) && strings.EqualFold(pathPart[:len(PREFIX)], PREFIX) && pathPart[len(PREFIX)] == '/' {
		pathPart = pathPart[len(PREFIX):]
	}

	if strings.EqualFold(pathPart, "/Bulk") || strings.HasSuffix(strings.ToLower(pathPart), "/.search") {
		return ResourceType{}, "", "", false, &bulkOpError{
			scimType: "invalidPath",
			message:  "path " + strconv.Quote(rawOpPath) + " is not permitted inside Bulk",
		}
	}

	trimmed := strings.TrimPrefix(pathPart, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 || parts[0] == "" || len(parts) > 2 {
		return ResourceType{}, "", "", false, &bulkOpError{
			scimType: "invalidPath",
			message:  "malformed resource path " + strconv.Quote(rawOpPath),
		}
	}

	rt, found := lookupResourceEndpoint("/" + parts[0])
	if !found {
		return ResourceType{}, "", "", false, &bulkOpError{
			scimType: "invalidPath",
			message:  "unknown resource endpoint " + strconv.Quote("/"+parts[0]),
		}
	}

	rawID := ""
	if len(parts) == 2 {
		if parts[1] == "" {
			return rt, "", "", false, &bulkOpError{
				scimType: "invalidPath",
				message:  "missing resource identifier in path " + strconv.Quote(rawOpPath),
			}
		}
		unescaped, err := url.PathUnescape(parts[1])
		if err != nil {
			return rt, "", "", false, &bulkOpError{scimType: "invalidPath", message: err.Error()}
		}
		rawID = unescaped
	}

	if method == http.MethodPost && rawID != "" {
		return rt, "", "", false, &bulkOpError{
			scimType: "invalidPath",
			message:  "POST path must target a resource collection, not " + strconv.Quote(rawOpPath),
		}
	}
	if method != http.MethodPost && rawID == "" {
		return rt, "", "", false, &bulkOpError{
			scimType: "invalidPath",
			message:  method + " path must include a resource identifier",
		}
	}

	if rawID == "" {
		return rt, "", rawQuery, hasQuery, nil
	}

	resolvedID, err := resolveBulkRefsInString(rawID, resolved)
	if err != nil {
		return rt, "", "", false, &bulkOpError{scimType: "invalidValue", message: err.Error()}
	}
	return rt, resolvedID, rawQuery, hasQuery, nil
}

func lookupResourceEndpoint(endpoint string) (ResourceType, bool) {
	if rt, ok := Resources[endpoint]; ok {
		return rt, true
	}
	for k, rt := range Resources {
		if strings.EqualFold(k, endpoint) {
			return rt, true
		}
	}
	return ResourceType{}, false
}

func resolveBulkRefsInValue(val interface{}, resolved map[string]string) (interface{}, error) {
	switch typed := val.(type) {
	case string:
		return resolveBulkRefsInString(typed, resolved)
	case []interface{}:
		out := make([]interface{}, len(typed))
		for i, elem := range typed {
			res, err := resolveBulkRefsInValue(elem, resolved)
			if err != nil {
				return nil, err
			}
			out[i] = res
		}
		return out, nil
	case map[string]interface{}:
		out := make(map[string]interface{}, len(typed))
		for k, elem := range typed {
			res, err := resolveBulkRefsInValue(elem, resolved)
			if err != nil {
				return nil, err
			}
			out[k] = res
		}
		return out, nil
	default:
		return val, nil
	}
}

// resolveBulkRefsInString resolves "bulkId:<token>" references when the whole
// string is a bulkId token, when it appears as a URI segment ("/.../bulkId:x"),
// or when it appears inside a quoted SCIM filter literal (`[value eq "bulkId:x"]`).
func resolveBulkRefsInString(s string, resolved map[string]string) (string, error) {
	const prefix = "bulkId:"
	if !strings.Contains(s, prefix) {
		return s, nil
	}
	if strings.HasPrefix(s, prefix) && !strings.ContainsAny(s[len(prefix):], `/?#"'] `) {
		token := s[len(prefix):]
		if token == "" {
			return "", fmt.Errorf("empty bulkId reference in %q", s)
		}
		id, ok := resolved[token]
		if !ok {
			return "", fmt.Errorf("unresolved bulkId reference %q", prefix+token)
		}
		return id, nil
	}

	var b strings.Builder
	i := 0
	for i < len(s) {
		rel := strings.Index(s[i:], prefix)
		if rel < 0 {
			b.WriteString(s[i:])
			break
		}
		pos := i + rel
		isBoundary := pos == 0 || s[pos-1] == '/' || s[pos-1] == '"'
		if !isBoundary {
			b.WriteString(s[i : pos+len(prefix)])
			i = pos + len(prefix)
			continue
		}
		b.WriteString(s[i:pos])
		tokenStart := pos + len(prefix)
		tokenEnd := tokenStart
		for tokenEnd < len(s) && !strings.ContainsRune(`/?#"'] `, rune(s[tokenEnd])) {
			tokenEnd++
		}
		token := s[tokenStart:tokenEnd]
		if token == "" {
			return "", fmt.Errorf("empty bulkId reference in %q", s)
		}
		id, ok := resolved[token]
		if !ok {
			return "", fmt.Errorf("unresolved bulkId reference %q", prefix+token)
		}
		b.WriteString(id)
		i = tokenEnd
	}
	return b.String(), nil
}

type bulkResponseRecorder struct {
	header      http.Header
	body        bytes.Buffer
	code        int
	wroteHeader bool
}

func newBulkResponseRecorder() *bulkResponseRecorder {
	return &bulkResponseRecorder{
		header: make(http.Header),
		code:   http.StatusOK,
	}
}

func (r *bulkResponseRecorder) Header() http.Header {
	return r.header
}

func (r *bulkResponseRecorder) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}
	r.code = statusCode
	r.wroteHeader = true
}

func (r *bulkResponseRecorder) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.body.Write(p)
}
