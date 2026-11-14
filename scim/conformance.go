package scim

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ConformanceOptions configures a live SCIM 2.0 conformance run against a
// target base URL (e.g. "http://127.0.0.1:8080/scim/v2" or
// "http://127.0.0.1:8080/tenants/acme/scim/v2").
type ConformanceOptions struct {
	BaseURL    string
	Token      string
	BasicUser  string
	BasicPass  string
	Profile    string // "all", "rfc", "entra", "okta", "keycloak"
	HTTPClient *http.Client
}

// ConformanceStep records the outcome of one conformance assertion.
type ConformanceStep struct {
	Profile    string  `json:"profile"`
	Name       string  `json:"name"`
	Passed     bool    `json:"passed"`
	Detail     string  `json:"detail,omitempty"`
	DurationMs float64 `json:"durationMs"`
}

// ConformanceReport summarises a live conformance run.
type ConformanceReport struct {
	TargetURL  string            `json:"targetUrl"`
	Profile    string            `json:"profile"`
	Passed     int               `json:"passed"`
	Failed     int               `json:"failed"`
	Total      int               `json:"total"`
	DurationMs float64           `json:"durationMs"`
	Timestamp  time.Time         `json:"timestamp"`
	Steps      []ConformanceStep `json:"steps"`
}

// RunConformance executes the selected conformance profile(s) against opts.BaseURL
// and cleans up every temporary resource created during the run.
func RunConformance(ctx context.Context, opts ConformanceOptions) ConformanceReport {
	start := time.Now()
	profile := strings.ToLower(strings.TrimSpace(opts.Profile))
	if profile == "" {
		profile = "all"
	}
	baseURL := strings.TrimRight(strings.TrimSpace(opts.BaseURL), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8080/scim/v2"
	}
	if !strings.Contains(baseURL, "/scim/v2") {
		baseURL += "/scim/v2"
	}
	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	runner := &conformanceRunner{
		ctx:       ctx,
		baseURL:   baseURL,
		token:     opts.Token,
		basicUser: opts.BasicUser,
		basicPass: opts.BasicPass,
		client:    client,
	}
	defer runner.cleanup()

	if profile == "all" || profile == "rfc" {
		runner.runRFCProfile()
	}
	if profile == "all" || profile == "entra" {
		runner.runEntraProfile()
	}
	if profile == "all" || profile == "okta" {
		runner.runOktaProfile()
	}
	if profile == "all" || profile == "keycloak" {
		runner.runKeycloakProfile()
	}

	passed, failed := 0, 0
	for _, st := range runner.steps {
		if st.Passed {
			passed++
		} else {
			failed++
		}
	}
	return ConformanceReport{
		TargetURL:  baseURL,
		Profile:    profile,
		Passed:     passed,
		Failed:     failed,
		Total:      len(runner.steps),
		DurationMs: float64(time.Since(start).Microseconds()) / 1000.0,
		Timestamp:  time.Now().UTC(),
		Steps:      runner.steps,
	}
}

type conformanceRunner struct {
	ctx         context.Context
	baseURL     string
	token       string
	basicUser   string
	basicPass   string
	client      *http.Client
	steps       []ConformanceStep
	cleanupURLs []string
}

func (r *conformanceRunner) step(profile, name string, fn func() error) {
	t0 := time.Now()
	err := fn()
	ms := float64(time.Since(t0).Microseconds()) / 1000.0
	st := ConformanceStep{
		Profile:    profile,
		Name:       name,
		Passed:     err == nil,
		DurationMs: ms,
	}
	if err != nil {
		st.Detail = err.Error()
	}
	r.steps = append(r.steps, st)
}

func (r *conformanceRunner) trackCleanup(subPath string) {
	r.cleanupURLs = append(r.cleanupURLs, subPath)
}

func (r *conformanceRunner) cleanup() {
	for i := len(r.cleanupURLs) - 1; i >= 0; i-- {
		_, _, _, _ = r.do(http.MethodDelete, r.cleanupURLs[i], nil, nil)
	}
}

func (r *conformanceRunner) do(method, subPath string, body interface{}, headers map[string]string) (int, http.Header, map[string]interface{}, error) {
	var rawBody io.Reader
	if body != nil {
		switch typed := body.(type) {
		case string:
			rawBody = strings.NewReader(typed)
		case []byte:
			rawBody = bytes.NewReader(typed)
		default:
			encoded, err := json.Marshal(body)
			if err != nil {
				return 0, nil, nil, err
			}
			rawBody = bytes.NewReader(encoded)
		}
	}
	req, err := http.NewRequestWithContext(r.ctx, method, r.baseURL+subPath, rawBody)
	if err != nil {
		return 0, nil, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", scimContentType)
	}
	req.Header.Set("Accept", scimContentType)
	if r.token != "" {
		req.Header.Set("Authorization", "Bearer "+r.token)
	} else if r.basicUser != "" || r.basicPass != "" {
		req.SetBasicAuth(r.basicUser, r.basicPass)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return resp.StatusCode, resp.Header, nil, err
	}
	var parsed map[string]interface{}
	if len(bytes.TrimSpace(respBytes)) > 0 {
		if err := json.Unmarshal(respBytes, &parsed); err != nil {
			return resp.StatusCode, resp.Header, nil, fmt.Errorf("HTTP %d non-JSON body: %s", resp.StatusCode, string(respBytes))
		}
	}
	return resp.StatusCode, resp.Header, parsed, nil
}

func (r *conformanceRunner) runRFCProfile() {
	tag := uuid.NewString()[:8]

	r.step("rfc", "Discovery endpoints (/ServiceProviderConfig, /Schemas, /ResourceTypes)", func() error {
		for _, ep := range []string{"/ServiceProviderConfig", "/Schemas", "/ResourceTypes"} {
			code, hdr, body, err := r.do(http.MethodGet, ep, nil, nil)
			if err != nil {
				return err
			}
			if code != http.StatusOK {
				return fmt.Errorf("GET %s returned %d", ep, code)
			}
			if !strings.Contains(strings.ToLower(hdr.Get("Content-Type")), "application/scim+json") {
				return fmt.Errorf("GET %s Content-Type = %q, want application/scim+json", ep, hdr.Get("Content-Type"))
			}
			if body == nil {
				return fmt.Errorf("GET %s returned empty JSON", ep)
			}
		}
		return nil
	})

	var userID, etag string
	uName := fmt.Sprintf("conf.rfc.%s@example.com", tag)

	r.step("rfc", "Create User (POST /Users) + ETag + Location + uniqueness conflict (409)", func() error {
		payload := map[string]interface{}{
			"schemas":  []string{coreUserSchemaURI},
			"userName": uName,
			"active":   true,
			"emails":   []interface{}{map[string]interface{}{"value": uName, "type": "work", "primary": true}},
		}
		code, hdr, created, err := r.do(http.MethodPost, "/Users", payload, nil)
		if err != nil {
			return err
		}
		if code != http.StatusCreated {
			return fmt.Errorf("POST /Users status = %d, want 201 (%v)", code, created)
		}
		userID, _ = created["id"].(string)
		if userID == "" {
			return fmt.Errorf("created User missing id")
		}
		r.trackCleanup("/Users/" + userID)
		etag = hdr.Get("ETag")
		if etag == "" || hdr.Get("Location") == "" {
			return fmt.Errorf("missing ETag (%q) or Location (%q)", etag, hdr.Get("Location"))
		}

		// Case-insensitive duplicate must return 409 scimType=uniqueness.
		dupPayload := map[string]interface{}{
			"schemas":  []string{coreUserSchemaURI},
			"userName": strings.ToUpper(uName),
		}
		dupCode, _, dupBody, err := r.do(http.MethodPost, "/Users", dupPayload, nil)
		if err != nil {
			return err
		}
		if dupCode != http.StatusConflict || dupBody["scimType"] != "uniqueness" {
			return fmt.Errorf("duplicate userName returned %d (%v), want 409 uniqueness", dupCode, dupBody)
		}
		return nil
	})

	r.step("rfc", "Conditional GET (If-None-Match -> 304) and stale If-Match (412)", func() error {
		if userID == "" {
			return fmt.Errorf("prerequisite user not created")
		}
		code, _, _, err := r.do(http.MethodGet, "/Users/"+userID, nil, map[string]string{"If-None-Match": etag})
		if err != nil {
			return err
		}
		if code != http.StatusNotModified {
			return fmt.Errorf("If-None-Match status = %d, want 304", code)
		}
		patchBody := map[string]interface{}{
			"schemas":    []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
			"Operations": []interface{}{map[string]interface{}{"op": "replace", "path": "displayName", "value": "Stale"}},
		}
		pCode, _, _, err := r.do(http.MethodPatch, "/Users/"+userID, patchBody, map[string]string{"If-Match": `W/"stale-version"`})
		if err != nil {
			return err
		}
		if pCode != http.StatusPreconditionFailed {
			return fmt.Errorf("stale If-Match status = %d, want 412", pCode)
		}
		return nil
	})

	r.step("rfc", "POST /.search with filter + attributes projection", func() error {
		if userID == "" {
			return fmt.Errorf("prerequisite user not created")
		}
		searchReq := map[string]interface{}{
			"schemas":    []string{searchRequestSchema},
			"filter":     fmt.Sprintf(`userName eq %q`, uName),
			"attributes": []string{"userName"},
		}
		code, _, body, err := r.do(http.MethodPost, "/Users/.search", searchReq, nil)
		if err != nil {
			return err
		}
		if code != http.StatusOK {
			return fmt.Errorf("POST /Users/.search status = %d (%v)", code, body)
		}
		total, _ := body["totalResults"].(float64)
		if int(total) != 1 {
			return fmt.Errorf("totalResults = %v, want 1", body["totalResults"])
		}
		return nil
	})

	r.step("rfc", "POST /Bulk with bulkId cross-references and Group.members -> User.groups sync", func() error {
		bulkUName := fmt.Sprintf("conf.bulk.%s@example.com", tag)
		bulkGName := fmt.Sprintf("ConfBulkGroup-%s", tag)
		bulkPayload := map[string]interface{}{
			"schemas": []string{bulkRequestSchema},
			"Operations": []interface{}{
				map[string]interface{}{
					"method": "POST",
					"path":   "/Users",
					"bulkId": "u1",
					"data": map[string]interface{}{
						"schemas":  []string{coreUserSchemaURI},
						"userName": bulkUName,
					},
				},
				map[string]interface{}{
					"method": "POST",
					"path":   "/Groups",
					"bulkId": "g1",
					"data": map[string]interface{}{
						"schemas":     []string{coreGroupSchemaURI},
						"displayName": bulkGName,
						"members":     []interface{}{map[string]interface{}{"value": "bulkId:u1"}},
					},
				},
			},
		}
		code, _, body, err := r.do(http.MethodPost, "/Bulk", bulkPayload, nil)
		if err != nil {
			return err
		}
		if code != http.StatusOK {
			return fmt.Errorf("POST /Bulk status = %d (%v)", code, body)
		}
		ops, _ := body["Operations"].([]interface{})
		if len(ops) != 2 {
			return fmt.Errorf("Bulk Operations len = %d, want 2", len(ops))
		}
		var createdUID, createdGID string
		for i, rawOp := range ops {
			opMap, _ := rawOp.(map[string]interface{})
			if opMap["status"] != "201" {
				return fmt.Errorf("Bulk op[%d] status = %v, want 201", i, opMap["status"])
			}
			loc, _ := opMap["location"].(string)
			parts := strings.Split(strings.TrimRight(loc, "/"), "/")
			id := parts[len(parts)-1]
			if i == 0 {
				createdUID = id
				r.trackCleanup("/Users/" + id)
			} else {
				createdGID = id
				r.trackCleanup("/Groups/" + id)
			}
		}
		if createdUID == "" || createdGID == "" {
			return fmt.Errorf("missing created IDs from Bulk Locations")
		}
		gCode, _, uDoc, err := r.do(http.MethodGet, "/Users/"+createdUID, nil, nil)
		if err != nil || gCode != http.StatusOK {
			return fmt.Errorf("GET bulk user failed: %d %v", gCode, err)
		}
		groups, _ := uDoc["groups"].([]interface{})
		if len(groups) != 1 {
			return fmt.Errorf("User.groups len = %d, want 1", len(groups))
		}
		return nil
	})
}

func (r *conformanceRunner) runEntraProfile() {
	tag := uuid.NewString()[:8]
	uName := fmt.Sprintf("conf.entra.%s@contoso.com", tag)
	var userID, groupID string

	r.step("entra", "Connectivity probe + mixed-case EnterpriseUser create", func() error {
		probeFilter := url.QueryEscape(fmt.Sprintf(`userName eq %q`, uName))
		code, _, list, err := r.do(http.MethodGet, "/Users?filter="+probeFilter, nil, nil)
		if err != nil || code != http.StatusOK {
			return fmt.Errorf("probe GET /Users failed: %d %v", code, err)
		}
		if total, _ := list["totalResults"].(float64); int(total) != 0 {
			return fmt.Errorf("probe totalResults = %v, want 0", total)
		}

		mixedExt := "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user"
		createBody := map[string]interface{}{
			"schemas":    []string{coreUserSchemaURI, mixedExt},
			"externalId": "entra-" + tag,
			"userName":   uName,
			"active":     true,
			mixedExt:     map[string]interface{}{"department": "Identity"},
		}
		cCode, _, created, err := r.do(http.MethodPost, "/Users", createBody, nil)
		if err != nil || cCode != http.StatusCreated {
			return fmt.Errorf("POST /Users status = %d (%v) %v", cCode, created, err)
		}
		userID, _ = created["id"].(string)
		r.trackCleanup("/Users/" + userID)
		return nil
	})

	r.step("entra", "Capitalized PATCH Op + value-selection filter emails[type eq \"work\"].value + Group Add/Remove", func() error {
		if userID == "" {
			return fmt.Errorf("prerequisite Entra user not created")
		}
		patchBody := map[string]interface{}{
			"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
			"Operations": []interface{}{
				map[string]interface{}{
					"Op":    "Add",
					"Path":  `emails[type eq "work"].value`,
					"Value": uName,
				},
				map[string]interface{}{
					"Op": "Replace",
					"Value": map[string]interface{}{
						"name.givenName": "Adele",
						"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department": "Cloud",
					},
				},
			},
		}
		pCode, _, patched, err := r.do(http.MethodPatch, "/Users/"+userID, patchBody, nil)
		if err != nil || pCode != http.StatusOK {
			return fmt.Errorf("Entra PATCH /Users status = %d (%v) %v", pCode, patched, err)
		}

		gCode, _, grp, err := r.do(http.MethodPost, "/Groups", map[string]interface{}{
			"schemas":     []string{coreGroupSchemaURI},
			"displayName": "EntraGroup-" + tag,
		}, nil)
		if err != nil || gCode != http.StatusCreated {
			return fmt.Errorf("POST /Groups status = %d (%v)", gCode, grp)
		}
		groupID, _ = grp["id"].(string)
		r.trackCleanup("/Groups/" + groupID)

		addCode, _, _, err := r.do(http.MethodPatch, "/Groups/"+groupID, map[string]interface{}{
			"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
			"Operations": []interface{}{
				map[string]interface{}{
					"op":    "Add",
					"path":  "members",
					"value": []interface{}{map[string]interface{}{"value": userID}},
				},
			},
		}, nil)
		if err != nil || addCode != http.StatusOK {
			return fmt.Errorf("Entra group member Add status = %d %v", addCode, err)
		}

		remCode, _, _, err := r.do(http.MethodPatch, "/Groups/"+groupID, map[string]interface{}{
			"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
			"Operations": []interface{}{
				map[string]interface{}{
					"op":    "Remove",
					"path":  "members",
					"value": []interface{}{map[string]interface{}{"value": userID}},
				},
			},
		}, nil)
		if err != nil || remCode != http.StatusOK {
			return fmt.Errorf("Entra group member Remove status = %d %v", remCode, err)
		}
		return nil
	})
}

func (r *conformanceRunner) runOktaProfile() {
	tag := uuid.NewString()[:8]
	uName := fmt.Sprintf("conf.okta.%s@example.com", tag)
	var userID string

	r.step("okta", "Create User with password (never echoed), PUT update, and pathless PATCH deactivation", func() error {
		cCode, _, created, err := r.do(http.MethodPost, "/Users", map[string]interface{}{
			"schemas":  []string{coreUserSchemaURI},
			"userName": uName,
			"password": "Okta-Conf-Password-99!",
			"active":   true,
			"name":     map[string]interface{}{"givenName": "Jane", "familyName": "Okta"},
			"emails":   []interface{}{map[string]interface{}{"value": uName, "type": "work", "primary": true}},
		}, nil)
		if err != nil || cCode != http.StatusCreated {
			return fmt.Errorf("Okta POST /Users status = %d (%v) %v", cCode, created, err)
		}
		userID, _ = created["id"].(string)
		r.trackCleanup("/Users/" + userID)
		if _, leaked := created["password"]; leaked {
			return fmt.Errorf("password was echoed in POST /Users response")
		}

		putCode, _, updated, err := r.do(http.MethodPut, "/Users/"+userID, map[string]interface{}{
			"schemas":     []string{coreUserSchemaURI},
			"id":          userID,
			"userName":    uName,
			"displayName": "Jane Okta Updated",
			"active":      true,
		}, nil)
		if err != nil || putCode != http.StatusOK {
			return fmt.Errorf("Okta PUT /Users status = %d (%v) %v", putCode, updated, err)
		}

		deactCode, _, deact, err := r.do(http.MethodPatch, "/Users/"+userID, map[string]interface{}{
			"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
			"Operations": []interface{}{
				map[string]interface{}{
					"op":    "replace",
					"value": map[string]interface{}{"active": false},
				},
			},
		}, nil)
		if err != nil || deactCode != http.StatusOK {
			return fmt.Errorf("Okta deactivation PATCH status = %d (%v) %v", deactCode, deact, err)
		}
		if active, _ := deact["active"].(bool); active {
			return fmt.Errorf("expected active=false after Okta deactivation PATCH")
		}
		return nil
	})
}

func (r *conformanceRunner) runKeycloakProfile() {
	tag := uuid.NewString()[:8]
	uName := fmt.Sprintf("conf.kc.%s@keycloak.local", tag)
	dept := "KCDept-" + tag

	r.step("keycloak", "Colon-delimited extension URN filter and PATCH reconciliation", func() error {
		cCode, _, created, err := r.do(http.MethodPost, "/Users", map[string]interface{}{
			"schemas":               []string{coreUserSchemaURI, enterpriseUserSchemaURI},
			"userName":              uName,
			"active":                true,
			enterpriseUserSchemaURI: map[string]interface{}{"department": dept},
		}, nil)
		if err != nil || cCode != http.StatusCreated {
			return fmt.Errorf("Keycloak POST /Users status = %d (%v) %v", cCode, created, err)
		}
		userID, _ := created["id"].(string)
		r.trackCleanup("/Users/" + userID)

		kcFilter := url.QueryEscape(fmt.Sprintf(`%s:department eq %q`, enterpriseUserSchemaURI, dept))
		sCode, _, searchRes, err := r.do(http.MethodGet, "/Users?filter="+kcFilter, nil, nil)
		if err != nil || sCode != http.StatusOK {
			return fmt.Errorf("Keycloak colon-URN filter status = %d (%v) %v", sCode, searchRes, err)
		}
		if total, _ := searchRes["totalResults"].(float64); int(total) != 1 {
			return fmt.Errorf("Keycloak colon-URN filter totalResults = %v, want 1", searchRes["totalResults"])
		}

		pCode, _, patched, err := r.do(http.MethodPatch, "/Users/"+userID, map[string]interface{}{
			"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
			"Operations": []interface{}{
				map[string]interface{}{
					"op":    "replace",
					"path":  enterpriseUserSchemaURI + ":department",
					"value": dept + "-v2",
				},
			},
		}, nil)
		if err != nil || pCode != http.StatusOK {
			return fmt.Errorf("Keycloak colon-URN PATCH status = %d (%v) %v", pCode, patched, err)
		}
		return nil
	})
}

// RunConformanceCLI parses `goscim conformance` flags, runs the live suite, and
// returns a process exit code (0 when every step passes, 1 on failure, 2 on bad
// CLI flags).
func RunConformanceCLI(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("conformance", flag.ContinueOnError)
	fs.SetOutput(stderr)
	targetURL := fs.String("url", "http://127.0.0.1:8080/scim/v2", "SCIM 2.0 base URL (e.g. http://127.0.0.1:8080/scim/v2)")
	token := fs.String("token", "", "Bearer token for Authorization header")
	basic := fs.String("basic", "", "HTTP Basic credentials in username:password form")
	profile := fs.String("profile", "all", "Conformance profile: all, rfc, entra, okta, or keycloak")
	asJSON := fs.Bool("json", false, "Emit machine-readable JSON report")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	var basicUser, basicPass string
	if *basic != "" {
		u, p, ok := strings.Cut(*basic, ":")
		if !ok {
			fmt.Fprintln(stderr, "error: -basic must be in username:password format")
			return 2
		}
		basicUser, basicPass = u, p
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	report := RunConformance(ctx, ConformanceOptions{
		BaseURL:   *targetURL,
		Token:     *token,
		BasicUser: basicUser,
		BasicPass: basicPass,
		Profile:   *profile,
	})

	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(report)
	} else {
		fmt.Fprintf(stdout, "GoSCIM Conformance Suite\n")
		fmt.Fprintf(stdout, "Target:  %s\n", report.TargetURL)
		fmt.Fprintf(stdout, "Profile: %s\n\n", report.Profile)
		for _, st := range report.Steps {
			mark := "PASS"
			if !st.Passed {
				mark = "FAIL"
			}
			fmt.Fprintf(stdout, "  [%s] (%-8s) %-68s %6.1f ms\n", mark, st.Profile, st.Name, st.DurationMs)
			if st.Detail != "" {
				fmt.Fprintf(stdout, "         -> %s\n", st.Detail)
			}
		}
		fmt.Fprintf(stdout, "\nSummary: %d/%d passed, %d failed (%.1f ms)\n",
			report.Passed, report.Total, report.Failed, report.DurationMs)
	}

	if report.Failed > 0 {
		return 1
	}
	return 0
}
