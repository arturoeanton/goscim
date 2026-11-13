package scim

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

const bulkEndpoint = "/scim/v2/Bulk"

func TestBulkEndToEndWithCrossReferencesAndGroupSync(t *testing.T) {
	r, _ := newTestServer(t)

	payload, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{bulkRequestSchema},
		"Operations": []map[string]interface{}{
			{
				"method": "POST",
				"path":   "/Users",
				"bulkId": "u-ada",
				"data": map[string]interface{}{
					"schemas":  []string{coreUserSchemaURI},
					"userName": "ada.bulk@example.com",
				},
			},
			{
				"method": "POST",
				"path":   "/Users",
				"bulkId": "u-grace",
				"data": map[string]interface{}{
					"schemas":  []string{coreUserSchemaURI},
					"userName": "grace.bulk@example.com",
				},
			},
			{
				"method": "POST",
				"path":   "/Groups",
				"bulkId": "g-compilers",
				"data": map[string]interface{}{
					"schemas":     []string{groupSchemaURI},
					"displayName": "Compilers",
					"members": []map[string]interface{}{
						{
							"value":   "bulkId:u-ada",
							"$ref":    "https://example.com/scim/v2/Users/bulkId:u-ada",
							"type":    "User",
							"display": "Ada",
						},
					},
				},
			},
			{
				"method": "PATCH",
				"path":   "/Groups/bulkId:g-compilers",
				"data": map[string]interface{}{
					"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
					"Operations": []map[string]interface{}{
						{
							"op":    "add",
							"path":  "members",
							"value": []map[string]interface{}{{"value": "bulkId:u-grace", "type": "User"}},
						},
						{
							"op":   "remove",
							"path": `members[value eq "bulkId:u-ada"]`,
						},
					},
				},
			},
			{
				"method": "PUT",
				"path":   "/scim/v2/Users/bulkId:u-grace",
				"data": map[string]interface{}{
					"schemas":     []string{coreUserSchemaURI},
					"userName":    "grace.bulk@example.com",
					"displayName": "Grace Hopper",
				},
			},
			{
				"method": "DELETE",
				"path":   "/Users/bulkId:u-ada",
			},
		},
	})

	w := do(t, r, http.MethodPost, bulkEndpoint, string(payload))
	requireStatus(t, w, http.StatusOK)
	resp := decode(t, w)

	schemas, _ := resp["schemas"].([]interface{})
	if len(schemas) != 1 || schemas[0] != bulkResponseSchema {
		t.Fatalf("schemas = %v, want [%s]", schemas, bulkResponseSchema)
	}

	ops, _ := resp["Operations"].([]interface{})
	if len(ops) != 6 {
		t.Fatalf("len(Operations) = %d, want 6: %s", len(ops), w.Body.String())
	}

	wantStatuses := []string{"201", "201", "201", "200", "200", "204"}
	for i, want := range wantStatuses {
		op := ops[i].(map[string]interface{})
		if op["status"] != want {
			t.Fatalf("op[%d] status = %v, want %s (body: %v)", i, op["status"], want, op)
		}
		loc, _ := op["location"].(string)
		if loc == "" {
			t.Errorf("op[%d] (%s) missing location: %v", i, op["method"], op)
		}
		if i < 5 && op["version"] == "" {
			t.Errorf("op[%d] (%s) missing ETag version: %v", i, op["method"], op)
		}
	}

	graceLoc := ops[1].(map[string]interface{})["location"].(string)
	graceID := graceLoc[strings.LastIndex(graceLoc, "/")+1:]
	compilersLoc := ops[2].(map[string]interface{})["location"].(string)
	compilersID := compilersLoc[strings.LastIndex(compilersLoc, "/")+1:]

	// Ada was deleted (404), Grace has displayName AND synced Compilers membership.
	adaLoc := ops[0].(map[string]interface{})["location"].(string)
	adaID := adaLoc[strings.LastIndex(adaLoc, "/")+1:]
	requireStatus(t, do(t, r, http.MethodGet, usersEndpoint+"/"+adaID, ""), http.StatusNotFound)

	wGrace := do(t, r, http.MethodGet, usersEndpoint+"/"+graceID, "")
	requireStatus(t, wGrace, http.StatusOK)
	graceDoc := decode(t, wGrace)
	if graceDoc["displayName"] != "Grace Hopper" {
		t.Errorf("grace displayName = %v", graceDoc["displayName"])
	}
	graceGroups := readUserGroupsMap(t, r, graceID)
	if len(graceGroups) != 1 || graceGroups[compilersID]["display"] != "Compilers" {
		t.Fatalf("grace synced groups = %#v, want Compilers (%s)", graceGroups, compilersID)
	}
}

func TestBulkFailOnErrorsAndUnresolvedBulkId(t *testing.T) {
	r, _ := newTestServer(t)

	// Operation 0 creates duplicate@example.com (201).
	// Operation 1 tries to create duplicate@example.com again with bulkId "u-dup" (409).
	// Operation 2 references bulkId:u-dup with failOnErrors=2 -> fails with 400 invalidValue, hitting failOnErrors=2.
	// Operation 3 would create survivor@example.com, MUST NOT execute because errorCount == 2.
	failLimit := 2
	payload, _ := json.Marshal(map[string]interface{}{
		"schemas":      []string{bulkRequestSchema},
		"failOnErrors": failLimit,
		"Operations": []map[string]interface{}{
			{
				"method": "POST",
				"path":   "/Users",
				"bulkId": "u-ok",
				"data": map[string]interface{}{
					"schemas":  []string{coreUserSchemaURI},
					"userName": "duplicate@example.com",
				},
			},
			{
				"method": "POST",
				"path":   "/Users",
				"bulkId": "u-dup",
				"data": map[string]interface{}{
					"schemas":  []string{coreUserSchemaURI},
					"userName": "DUPLICATE@example.com",
				},
			},
			{
				"method": "PATCH",
				"path":   "/Users/bulkId:u-dup",
				"data": map[string]interface{}{
					"schemas":    []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
					"Operations": []map[string]interface{}{{"op": "replace", "path": "active", "value": false}},
				},
			},
			{
				"method": "POST",
				"path":   "/Users",
				"bulkId": "u-skipped",
				"data": map[string]interface{}{
					"schemas":  []string{coreUserSchemaURI},
					"userName": "skipped@example.com",
				},
			},
		},
	})

	w := do(t, r, http.MethodPost, bulkEndpoint, string(payload))
	requireStatus(t, w, http.StatusOK)
	ops := decode(t, w)["Operations"].([]interface{})
	if len(ops) != 3 {
		t.Fatalf("expected processing to stop at 3 operations (2 errors), got %d: %s", len(ops), w.Body.String())
	}

	op1 := ops[1].(map[string]interface{})
	if op1["status"] != "409" {
		t.Errorf("op[1] status = %v, want 409", op1["status"])
	}
	if _, hasLoc := op1["location"]; hasLoc {
		t.Errorf("failed POST must omit location per RFC 7644 3.7.3: %v", op1)
	}
	op1Err := op1["response"].(map[string]interface{})
	if op1Err["scimType"] != "uniqueness" {
		t.Errorf("op[1] scimType = %v, want uniqueness", op1Err["scimType"])
	}

	op2 := ops[2].(map[string]interface{})
	if op2["status"] != "400" {
		t.Errorf("op[2] status = %v, want 400", op2["status"])
	}
	op2Err := op2["response"].(map[string]interface{})
	if op2Err["scimType"] != "invalidValue" {
		t.Errorf("op[2] scimType = %v, want invalidValue", op2Err["scimType"])
	}

	// Verify u-skipped was never created.
	wList := do(t, r, http.MethodGet, usersEndpoint, "")
	requireStatus(t, wList, http.StatusOK)
	if decode(t, wList)["totalResults"] != float64(1) {
		t.Errorf("expected only 1 user in store, got %s", wList.Body.String())
	}
}

func TestBulkVersionPreconditionAndAuditCorrelation(t *testing.T) {
	r, _ := newTestServer(t)

	var auditBuf bytes.Buffer
	SetAuditLogger(slog.New(slog.NewJSONHandler(&auditBuf, nil)))
	t.Cleanup(func() { SetAuditLogger(nil) })

	// Create a user first to obtain its initial ETag version.
	uID := postUserForSync(t, r, "etag.bulk@example.com")
	wGet := do(t, r, http.MethodGet, usersEndpoint+"/"+uID, "")
	requireStatus(t, wGet, http.StatusOK)
	etag := wGet.Header().Get("ETag")
	auditBuf.Reset()

	payload, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{bulkRequestSchema},
		"Operations": []map[string]interface{}{
			{
				"method":  "PATCH",
				"path":    "/Users/" + uID,
				"version": etag,
				"data": map[string]interface{}{
					"schemas":    []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
					"Operations": []map[string]interface{}{{"op": "replace", "path": "active", "value": true}},
				},
			},
			{
				"method":  "PATCH",
				"path":    "/Users/" + uID,
				"version": `W/"stale-version"`,
				"data": map[string]interface{}{
					"schemas":    []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
					"Operations": []map[string]interface{}{{"op": "replace", "path": "active", "value": false}},
				},
			},
		},
	})

	req := newRequest(http.MethodPost, bulkEndpoint, string(payload))
	req.Header.Set(RequestIDHeader, "bulk-corr-42")
	w := serve(r, req)
	requireStatus(t, w, http.StatusOK)
	if w.Header().Get(RequestIDHeader) != "bulk-corr-42" {
		t.Errorf("X-Request-Id = %q, want bulk-corr-42", w.Header().Get(RequestIDHeader))
	}

	ops := decode(t, w)["Operations"].([]interface{})
	if ops[0].(map[string]interface{})["status"] != "200" {
		t.Fatalf("matching version should succeed: %v", ops[0])
	}
	if ops[1].(map[string]interface{})["status"] != "412" {
		t.Fatalf("stale version should return 412: %v", ops[1])
	}

	lines := strings.Split(strings.TrimSpace(auditBuf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 audit events (2 sub-ops + 1 outer bulk), got %d:\n%s", len(lines), auditBuf.String())
	}
	for i, line := range lines {
		var rec map[string]interface{}
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("audit line %d invalid JSON: %v", i, err)
		}
		if rec["request_id"] != "bulk-corr-42" {
			t.Errorf("audit line %d request_id = %v, want bulk-corr-42", i, rec["request_id"])
		}
	}
}

func TestBulkValidationAndMaxOperationsLimit(t *testing.T) {
	r, _ := newTestServer(t)

	// Exceeding MaxBulkOperations must answer 413 Payload Too Large (RFC 7644 3.7.2).
	manyOps := make([]map[string]interface{}, MaxBulkOperations+1)
	for i := range manyOps {
		manyOps[i] = map[string]interface{}{
			"method": "DELETE",
			"path":   "/Users/nonexistent",
		}
	}
	tooManyBody, _ := json.Marshal(map[string]interface{}{
		"schemas":    []string{bulkRequestSchema},
		"Operations": manyOps,
	})
	w := do(t, r, http.MethodPost, bulkEndpoint, string(tooManyBody))
	requireSCIMError(t, w, http.StatusRequestEntityTooLarge)

	invalidEnvelopes := []struct {
		name string
		body string
	}{
		{"wrong schema", `{"schemas":["urn:wrong"],"Operations":[{"method":"DELETE","path":"/Users/1"}]}`},
		{"empty operations", `{"schemas":["` + bulkRequestSchema + `"],"Operations":[]}`},
		{"negative failOnErrors", `{"schemas":["` + bulkRequestSchema + `"],"failOnErrors":-1,"Operations":[{"method":"DELETE","path":"/Users/1"}]}`},
		{"POST missing bulkId", `{"schemas":["` + bulkRequestSchema + `"],"Operations":[{"method":"POST","path":"/Users","data":{}}]}`},
		{"duplicate POST bulkId", `{"schemas":["` + bulkRequestSchema + `"],"Operations":[{"method":"POST","bulkId":"x","path":"/Users","data":{}},{"method":"POST","bulkId":"x","path":"/Users","data":{}}]}`},
		{"unsupported GET method", `{"schemas":["` + bulkRequestSchema + `"],"Operations":[{"method":"GET","path":"/Users/1"}]}`},
	}
	for _, tc := range invalidEnvelopes {
		t.Run(tc.name, func(t *testing.T) {
			resp := do(t, r, http.MethodPost, bulkEndpoint, tc.body)
			requireSCIMError(t, resp, http.StatusBadRequest)
		})
	}

	// Disallowed sub-operation paths (/Bulk recursion, /.search, missing ID on DELETE)
	// report per-operation 400 invalidPath rather than crashing.
	badPathsBody, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{bulkRequestSchema},
		"Operations": []map[string]interface{}{
			{"method": "POST", "bulkId": "b1", "path": "/Bulk", "data": map[string]interface{}{}},
			{"method": "POST", "bulkId": "b2", "path": "/Users/.search", "data": map[string]interface{}{}},
			{"method": "DELETE", "path": "/Users"},
		},
	})
	wBad := do(t, r, http.MethodPost, bulkEndpoint, string(badPathsBody))
	requireStatus(t, wBad, http.StatusOK)
	for i, raw := range decode(t, wBad)["Operations"].([]interface{}) {
		op := raw.(map[string]interface{})
		if op["status"] != "400" {
			t.Errorf("bad path op[%d] status = %v, want 400", i, op["status"])
		}
	}
}

func TestBulkEnforcesRBACPerOperation(t *testing.T) {
	// Caller holds only "user" role; restrict Element.description to "superadmin".
	r, _ := newTestServerAs(t, &AnonymousAuthenticator{Roles: []string{"user"}})
	core := Schemas[schemaCore]
	for i := range core.Attributes {
		if core.Attributes[i].Name == "description" {
			restricted := []string{"superadmin"}
			core.Attributes[i].Writer = &restricted
		}
	}
	Schemas[schemaCore] = core

	payload, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{bulkRequestSchema},
		"Operations": []map[string]interface{}{
			{
				"method": "POST",
				"bulkId": "el-allowed",
				"path":   "/Elements",
				"data": map[string]interface{}{
					"schemas": []string{schemaCore, schemaExt},
					"name":    "public-el",
					"$ref":    "/public-el",
					schemaExt: map[string]interface{}{"required": 1},
				},
			},
			{
				"method": "POST",
				"bulkId": "el-forbidden",
				"path":   "/Elements",
				"data": map[string]interface{}{
					"schemas":     []string{schemaCore, schemaExt},
					"name":        "secret-el",
					"description": "restricted",
					"$ref":        "/secret-el",
					schemaExt:     map[string]interface{}{"required": 1},
				},
			},
		},
	})

	w := do(t, r, http.MethodPost, bulkEndpoint, string(payload))
	requireStatus(t, w, http.StatusOK)
	ops := decode(t, w)["Operations"].([]interface{})
	if ops[0].(map[string]interface{})["status"] != "201" {
		t.Errorf("op[0] = %v, want 201", ops[0])
	}
	if ops[1].(map[string]interface{})["status"] != "403" {
		t.Errorf("op[1] = %v, want 403", ops[1])
	}
}
