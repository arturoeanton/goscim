//go:build integration

// These tests run the server against a real Couchbase, started in a container.
// They are behind a build tag because they take minutes rather than
// milliseconds: `make integration`, or `go test -tags integration ./scim`.
//
// They live in their own package on purpose. Sharing a test binary with the
// unit suite would let its newTestServer swap the global scim.DB for the
// in-memory fake, and these tests would then quietly exercise the fake instead
// of Couchbase -- a false green, which is worse than no test at all.
//
// They exist because that in-memory store cannot answer the questions that
// matter most here. It does not evaluate SCIM
// filters, so the whole N1QL translation -- the part this release changed the
// most and the part carrying the injection fixes -- had no end-to-end coverage
// at all. Neither did startup: connecting, creating buckets and indexes.
package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arturoeanton/goscim/scim"
	"github.com/gin-gonic/gin"
	tccouchbase "github.com/testcontainers/testcontainers-go/modules/couchbase"
)

const couchbaseImage = "couchbase:community-7.1.1"

var (
	testRouter *gin.Engine
	usersPath  = scim.PREFIX + "/Users"
)

// TestMain starts one Couchbase for the whole package: the container takes a
// couple of minutes to become healthy, so paying that per test is not viable.
func TestMain(m *testing.M) {
	// os.Exit skips deferred calls, so the setup lives in a function that
	// returns a code and cleans up after itself. Without this a failed setup
	// exited zero and the package reported success having run nothing.
	os.Exit(runIntegrationSuite(m))
}

func runIntegrationSuite(m *testing.M) int {
	ctx := context.Background()

	container, err := tccouchbase.Run(ctx, couchbaseImage,
		tccouchbase.WithAdminCredentials("Administrator", "password"))
	if err != nil {
		fmt.Printf("could not start Couchbase: %v\n", err)
		return 1
	}
	defer func() {
		_ = container.Terminate(ctx)
	}()

	// The module sets the cluster's KV quota to the service minimum, which is
	// not enough for the three buckets the shipped config asks for. Raise it
	// rather than shrinking the config: the point is to exercise the real
	// bucket settings.
	if err := raiseMemoryQuota(ctx, container); err != nil {
		fmt.Printf("could not raise the cluster memory quota: %v\n", err)
		return 1
	}

	connectionString, err := container.ConnectionString(ctx)
	if err != nil {
		fmt.Printf("could not read the connection string: %v\n", err)
		return 1
	}
	// InitDB builds the scheme itself, so it wants host:port.
	endpoint := strings.TrimPrefix(connectionString, "couchbase://")

	os.Setenv("SCIM_COUCHBASE_URL", endpoint)
	os.Setenv("SCIM_ADMIN_USER", container.Username())
	os.Setenv("SCIM_ADMIN_PASSWORD", container.Password())
	// The container serves plain couchbase://, which is also what exercises
	// the SCIM_COUCHBASE_TLS escape hatch.
	os.Setenv("SCIM_COUCHBASE_TLS", "false")

	// Config lives at the repo root, two levels up from this package.
	scim.FolderBucketSetting = "../../config/bucketSettings/"

	if err := scim.InitDB(); err != nil {
		fmt.Printf("InitDB against the container failed: %v\n", err)
		return 1
	}
	defer func() { _ = scim.DB.Close() }()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	// ReadResourceType creates one bucket per resource type and its primary
	// index, so building the router is itself the startup test.
	if _, err := scim.NewRouter("../../config", engine,
		&scim.AnonymousAuthenticator{Roles: []string{"user", "admin", "superadmin", "role1", "role2"}}); err != nil {
		fmt.Printf("NewRouter against the container failed: %v\n", err)
		return 1
	}
	testRouter = engine

	return m.Run()
}

// raiseMemoryQuota gives the single-node cluster enough KV quota for every
// bucket the config declares.
func raiseMemoryQuota(ctx context.Context, container *tccouchbase.CouchbaseContainer) error {
	host, err := container.Host(ctx)
	if err != nil {
		return err
	}
	port, err := container.MappedPort(ctx, "8091/tcp")
	if err != nil {
		return err
	}

	form := url.Values{"memoryQuota": {"1024"}}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("http://%s:%s/pools/default", host, port.Port()),
		strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(container.Username(), container.Password())

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("the cluster answered %s: %s", response.Status, body)
	}
	return nil
}

// --- helpers ------------------------------------------------------------

func do(t *testing.T, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/scim+json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var out map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("not JSON (status %d): %v\nbody: %s", w.Code, err, w.Body.String())
	}
	return out
}

func requireStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("status = %d, want %d\nbody: %s", w.Code, want, w.Body.String())
	}
}

// createUser posts a user and returns its id.
func createUser(t *testing.T, userName string, extra map[string]interface{}) string {
	t.Helper()
	body := map[string]interface{}{
		"schemas":  []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		"userName": userName,
	}
	for k, v := range extra {
		body[k] = v
	}
	raw, _ := json.Marshal(body)
	w := do(t, http.MethodPost, usersPath, string(raw))
	requireStatus(t, w, http.StatusCreated)
	return decode(t, w)["id"].(string)
}

// search runs a query and returns the ListResponse.
func search(t *testing.T, query string) map[string]interface{} {
	t.Helper()
	w := do(t, http.MethodGet, usersPath+"?"+query, "")
	requireStatus(t, w, http.StatusOK)
	return decode(t, w)
}

func userNamesIn(t *testing.T, list map[string]interface{}) []string {
	t.Helper()
	resources, _ := list["Resources"].([]interface{})
	names := make([]string, 0, len(resources))
	for _, r := range resources {
		names = append(names, r.(map[string]interface{})["userName"].(string))
	}
	return names
}

// --- startup ------------------------------------------------------------

// Reaching any test at all means InitDB connected and NewRouter created the
// buckets; this makes the assertion explicit.
func TestStartupCreatedTheBuckets(t *testing.T) {
	for _, bucket := range []string{"User", "Group", "Element"} {
		if err := scim.DB.EnsureBucket(bucket); err != nil {
			t.Errorf("bucket %s is not usable: %v", bucket, err)
		}
	}
}

// Discovery has to work against a live server too: it is what an identity
// provider fetches before anything else.
func TestDiscoveryAgainstCouchbase(t *testing.T) {
	for _, path := range []string{"/ServiceProviderConfig", "/ResourceTypes", "/Schemas"} {
		w := do(t, http.MethodGet, scim.PREFIX+path, "")
		requireStatus(t, w, http.StatusOK)
	}

	w := do(t, http.MethodGet, scim.PREFIX+"/ResourceTypes", "")
	requireStatus(t, w, http.StatusOK)
	if total, _ := decode(t, w)["totalResults"].(float64); int(total) != 3 {
		t.Errorf("totalResults = %v, want 3", decode(t, w)["totalResults"])
	}
}

// --- CRUD against the real driver ---------------------------------------

func TestCRUDRoundTrip(t *testing.T) {
	id := createUser(t, "roundtrip@example.com", map[string]interface{}{
		"active": true,
		"emails": []interface{}{
			map[string]interface{}{"value": "roundtrip@example.com", "type": "work", "primary": true},
		},
	})

	w := do(t, http.MethodGet, usersPath+"/"+id, "")
	requireStatus(t, w, http.StatusOK)
	read := decode(t, w)
	if read["userName"] != "roundtrip@example.com" {
		t.Errorf("userName = %v", read["userName"])
	}
	emails, ok := read["emails"].([]interface{})
	if !ok || len(emails) != 1 {
		t.Fatalf("emails did not survive the round trip: %#v", read["emails"])
	}

	// PATCH
	patch := `{"schemas":["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
	           "Operations":[{"op":"replace","path":"active","value":false}]}`
	w = do(t, http.MethodPatch, usersPath+"/"+id, patch)
	requireStatus(t, w, http.StatusOK)
	if decode(t, w)["active"] != false {
		t.Error("the patch did not apply")
	}

	// PUT
	replacement := `{"schemas":["urn:ietf:params:scim:schemas:core:2.0:User"],"userName":"roundtrip2@example.com"}`
	w = do(t, http.MethodPut, usersPath+"/"+id, replacement)
	requireStatus(t, w, http.StatusOK)

	// DELETE, then gone
	w = do(t, http.MethodDelete, usersPath+"/"+id, "")
	requireStatus(t, w, http.StatusNoContent)
	w = do(t, http.MethodGet, usersPath+"/"+id, "")
	requireStatus(t, w, http.StatusNotFound)
}

func TestReadMissingResource(t *testing.T) {
	w := do(t, http.MethodGet, usersPath+"/no-such-id", "")
	requireStatus(t, w, http.StatusNotFound)
}

// userName is declared unique in the core schema. This is the check that
// stops a retried provisioning request from creating a second user.
func TestUniquenessAgainstCouchbase(t *testing.T) {
	userName := "unique-" + fmt.Sprint(time.Now().UnixNano())
	createUser(t, userName, nil)

	body := `{"schemas":["urn:ietf:params:scim:schemas:core:2.0:User"],"userName":"` + userName + `"}`
	w := do(t, http.MethodPost, usersPath, body)
	requireStatus(t, w, http.StatusConflict)
	if decode(t, w)["scimType"] != "uniqueness" {
		t.Errorf("scimType = %v", decode(t, w)["scimType"])
	}

	// A different value is accepted.
	createUser(t, userName+"-other", nil)
}

// Attribute names are case-insensitive, and the resource has to be stored with
// the declared spelling: a filter asks the schema for the path, so a document
// keyed "USERNAME" would be invisible to it.
func TestCaseInsensitiveAttributesAgainstCouchbase(t *testing.T) {
	userName := "case-" + fmt.Sprint(time.Now().UnixNano())
	body := `{"schemas":["URN:IETF:PARAMS:SCIM:SCHEMAS:CORE:2.0:USER"],"USERNAME":"` + userName + `"}`
	w := do(t, http.MethodPost, usersPath, body)
	requireStatus(t, w, http.StatusCreated)

	// Found by a filter that uses the declared spelling...
	list := search(t, "filter="+url.QueryEscape(`userName eq "`+userName+`"`))
	if total, _ := list["totalResults"].(float64); int(total) != 1 {
		t.Errorf("totalResults = %v, want 1", list["totalResults"])
	}

	// ...and by one that does not.
	list = search(t, "filter="+url.QueryEscape(`USERNAME eq "`+userName+`"`))
	if total, _ := list["totalResults"].(float64); int(total) != 1 {
		t.Errorf("a filter with different casing found %v", list["totalResults"])
	}
}

// --- the filter engine, against a real query engine ----------------------

// The comparison operators were crossed: gt produced >= and ge produced >.
// Unit tests pin the generated SQL; only a real query engine proves the
// boundary value falls on the right side.
func TestRangeOperatorBoundaries(t *testing.T) {
	prefix := "range-" + fmt.Sprint(time.Now().UnixNano())
	for _, n := range []int{1, 2, 3} {
		createUser(t, fmt.Sprintf("%s-%d", prefix, n), nil)
	}

	// nickName is a plain string attribute; use userName ordering instead by
	// filtering on the generated names, which sort naturally.
	cases := []struct {
		filter string
		want   []string
	}{
		{`userName gt "%[1]s-2"`, []string{prefix + "-3"}},
		{`userName ge "%[1]s-2"`, []string{prefix + "-2", prefix + "-3"}},
		{`userName lt "%[1]s-2"`, []string{prefix + "-1"}},
		{`userName le "%[1]s-2"`, []string{prefix + "-1", prefix + "-2"}},
	}

	for _, tc := range cases {
		filter := fmt.Sprintf(tc.filter, prefix)
		t.Run(filter, func(t *testing.T) {
			list := search(t, "filter="+url.QueryEscape(filter+` and userName sw "`+prefix+`"`)+"&sortBy=userName")
			got := userNamesIn(t, list)
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFilterOperatorsAgainstCouchbase(t *testing.T) {
	prefix := "ops-" + fmt.Sprint(time.Now().UnixNano())
	createUser(t, prefix+"-alpha@example.com", nil)
	createUser(t, prefix+"-beta@example.org", nil)

	cases := []struct {
		name   string
		filter string
		want   int
	}{
		{"eq", `userName eq "` + prefix + `-alpha@example.com"`, 1},
		{"ne", `userName sw "` + prefix + `" and userName ne "` + prefix + `-alpha@example.com"`, 1},
		{"co", `userName co "alpha" and userName sw "` + prefix + `"`, 1},
		{"sw", `userName sw "` + prefix + `"`, 2},
		{"ew", `userName ew ".org" and userName sw "` + prefix + `"`, 1},
		{"pr", `userName pr and userName sw "` + prefix + `"`, 2},
		{"and", `userName sw "` + prefix + `" and userName co "beta"`, 1},
		{"or", `userName eq "` + prefix + `-alpha@example.com" or userName eq "` + prefix + `-beta@example.org"`, 2},
		{"parens", `userName sw "` + prefix + `" and (userName co "alpha" or userName co "beta")`, 2},
		{"no match", `userName eq "` + prefix + `-nobody"`, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			list := search(t, "filter="+url.QueryEscape(tc.filter))
			total, ok := list["totalResults"].(float64)
			if !ok {
				t.Fatalf("totalResults missing from the ListResponse: %v", list)
			}
			if int(total) != tc.want {
				t.Errorf("totalResults = %d, want %d (filter %s)", int(total), tc.want, tc.filter)
			}
		})
	}
}

// A value ending in a backslash used to escape the closing quote of its own
// N1QL literal, swallowing the ORDER BY that follows. Against a real engine
// that is a query error, not a wrong result.
func TestFilterValuesWithBackslashes(t *testing.T) {
	userName := `backslash-DOMAIN\user`
	createUser(t, userName, nil)

	list := search(t, "filter="+url.QueryEscape(`userName eq "`+userName+`"`))
	total, ok := list["totalResults"].(float64)
	if !ok || int(total) != 1 {
		t.Errorf("totalResults = %v, want 1", list["totalResults"])
	}
}

// Filter values are bound, not concatenated, and the substring operators
// escape the client's own LIKE metacharacters. Searching for a literal percent
// sign used to hand the query engine a wildcard that matches everything.
func TestLikeWildcardsInValuesAreLiteral(t *testing.T) {
	prefix := "like-" + fmt.Sprint(time.Now().UnixNano())
	createUser(t, prefix+"-100%-discount", nil)
	createUser(t, prefix+"-plain", nil)
	createUser(t, prefix+"-a_b", nil)

	cases := []struct {
		name   string
		filter string
		want   int
	}{
		{"a literal percent matches only the one that has it", `userName co "100%-"`, 1},
		{"a percent is not a wildcard", `userName co "%plain%"`, 0},
		{"an underscore is not a single-character wildcard", `userName co "a_b"`, 1},
		{"an underscore does not match any character", `userName co "a_c"`, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			filter := tc.filter + ` and userName sw "` + prefix + `"`
			list := search(t, "filter="+url.QueryEscape(filter))
			total, ok := list["totalResults"].(float64)
			if !ok {
				t.Fatalf("totalResults missing: %v", list)
			}
			if int(total) != tc.want {
				t.Errorf("totalResults = %d, want %d (filter %s)", int(total), tc.want, filter)
			}
		})
	}
}

// A filter that does not parse must be refused before it reaches the cluster.
func TestMalformedFilterIsRefused(t *testing.T) {
	w := do(t, http.MethodGet, usersPath+"?filter="+url.QueryEscape(`userName eq "a" or 1=1`), "")
	requireStatus(t, w, http.StatusBadRequest)
	if decode(t, w)["scimType"] != "invalidFilter" {
		t.Errorf("scimType = %v", decode(t, w)["scimType"])
	}
}

// sortBy and pagination are translated into ORDER BY / OFFSET / LIMIT.
func TestSortingAndPaginationAgainstCouchbase(t *testing.T) {
	prefix := "page-" + fmt.Sprint(time.Now().UnixNano())
	for _, suffix := range []string{"a", "b", "c"} {
		createUser(t, prefix+"-"+suffix, nil)
	}
	filter := url.QueryEscape(`userName sw "` + prefix + `"`)

	ascending := userNamesIn(t, search(t, "filter="+filter+"&sortBy=userName"))
	if len(ascending) != 3 || !strings.HasSuffix(ascending[0], "-a") {
		t.Fatalf("ascending = %v", ascending)
	}

	descending := userNamesIn(t, search(t, "filter="+filter+"&sortBy=userName&sortOrder=descending"))
	if !strings.HasSuffix(descending[0], "-c") {
		t.Errorf("descending = %v", descending)
	}

	page := search(t, "filter="+filter+"&sortBy=userName&startIndex=2&count=1")
	if got := int(page["totalResults"].(float64)); got != 3 {
		t.Errorf("totalResults must count the whole set: %d", got)
	}
	if got := int(page["itemsPerPage"].(float64)); got != 1 {
		t.Errorf("itemsPerPage = %d", got)
	}
	if names := userNamesIn(t, page); len(names) != 1 || !strings.HasSuffix(names[0], "-b") {
		t.Errorf("page = %v", names)
	}
}

// An attribute the schema does not declare must not reach ORDER BY.
func TestUnknownSortByIsRefused(t *testing.T) {
	w := do(t, http.MethodGet, usersPath+"?sortBy="+url.QueryEscape("id` , (SELECT 1) x"), "")
	requireStatus(t, w, http.StatusBadRequest)
}
