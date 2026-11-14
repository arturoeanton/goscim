package scim

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestOktaOAuth2JWTAuthentication verifies Okta Custom Authorization Server
// access tokens (ES256, "cid" client identifier without "sub", and "scp" as a
// JSON array of scope strings).
func TestOktaOAuth2JWTAuthentication(t *testing.T) {
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	coordLen := (elliptic.P256().Params().BitSize + 7) / 8
	pad := func(b []byte) string {
		buf := make([]byte, coordLen)
		copy(buf[coordLen-len(b):], b)
		return base64.RawURLEncoding.EncodeToString(buf)
	}

	jwksSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []map[string]interface{}{{
				"kty": "EC",
				"use": "sig",
				"crv": "P-256",
				"alg": "ES256",
				"kid": "okta-es256-key",
				"x":   pad(ecKey.X.Bytes()),
				"y":   pad(ecKey.Y.Bytes()),
			}},
		})
	}))
	t.Cleanup(jwksSrv.Close)

	oktaIssuer := "https://dev-123456.okta.com/oauth2/aus123456"
	auth, err := NewJWTAuthenticator(jwksSrv.URL, oktaIssuer, "https://scim.example.com", "")
	if err != nil {
		t.Fatal(err)
	}
	r, _ := newTestServerAs(t, auth)

	tok := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": oktaIssuer,
		"aud": "https://scim.example.com",
		"cid": "0oa1oktaClientApp",
		"scp": []string{"role1", "role2", "scim:provision"},
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tok.Header["kid"] = "okta-es256-key"
	signed, err := tok.SignedString(ecKey)
	if err != nil {
		t.Fatal(err)
	}

	req := newRequest(http.MethodPost, elementsPath, validElement("OktaElem", 1))
	req.Header.Set("Authorization", "Bearer "+signed)
	w := serve(r, req)
	requireStatus(t, w, http.StatusCreated)
	if decode(t, w)["description"] != "description of OktaElem" {
		t.Errorf("expected Okta scp array to grant role2 read access")
	}
}

// TestOktaRunscopeLifecycleConformance executes the full Okta SCIM 2.0
// Lifecycle Management & Runscope test sequence: user lookup, create with
// externalId and explicit JSON null optional fields, case-insensitive userName
// search, emails.value search, uniqueness 409, Group Push, full PUT echoing
// readOnly groups + null clearing, deactivation/reactivation via pathless PATCH,
// and group deletion.
func TestOktaRunscopeLifecycleConformance(t *testing.T) {
	t.Setenv("SCIM_AUTH", "bearer")
	t.Setenv("SCIM_BEARER_TOKENS", "okta-lcm:okta-api-token-xyz:admin,role1,role2")
	auth, err := NewAuthenticatorFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	r, store := newTestServerAs(t, auth)

	oktaReq := func(method, target, body string) *httptest.ResponseRecorder {
		req := newRequest(method, target, body)
		req.Header.Set("Authorization", "Bearer okta-api-token-xyz")
		return serve(r, req)
	}

	// 1. Existence check before create: must return 200 ListResponse with totalResults: 0.
	w := oktaReq(http.MethodGet, usersPath+"?filter="+url.QueryEscape(`userName eq "Runscope.User@example.com"`)+"&startIndex=1&count=100", "")
	requireStatus(t, w, http.StatusOK)
	list := decode(t, w)
	if list["totalResults"] != float64(0) {
		t.Fatalf("totalResults = %v, want 0", list["totalResults"])
	}

	// 2. Create User with externalId, password, addresses.primary, and explicit JSON nulls.
	createPayload := `{
		"schemas": [
			"urn:ietf:params:scim:schemas:core:2.0:User",
			"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
		],
		"externalId": "00u1oktaExternalId",
		"userName": "Runscope.User@example.com",
		"password": "OktaInitialPassword!99",
		"active": true,
		"displayName": "Runscope User",
		"nickName": null,
		"profileUrl": null,
		"title": "Staff Engineer",
		"name": {
			"givenName": "Runscope",
			"middleName": null,
			"familyName": "User"
		},
		"emails": [
			{"primary": true, "value": "Runscope.User@example.com", "type": "work"}
		],
		"addresses": [
			{"primary": true, "type": "work", "locality": "San Francisco", "region": "CA", "country": "US"}
		],
		"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
			"department": "Identity",
			"division": null
		}
	}`
	w = oktaReq(http.MethodPost, usersPath, createPayload)
	requireStatus(t, w, http.StatusCreated)
	created := decode(t, w)
	userID := created["id"].(string)
	if created["externalId"] != "00u1oktaExternalId" {
		t.Errorf("externalId = %v, want 00u1oktaExternalId", created["externalId"])
	}
	if _, hasNick := created["nickName"]; hasNick {
		t.Errorf("null nickName should be omitted, got %v", created["nickName"])
	}
	if _, hasPwd := created["password"]; hasPwd {
		t.Error("password must never be returned")
	}
	storedDoc, _ := store.Get("User", userID)
	storedHash, _ := storedDoc["password"].(string)
	if !strings.HasPrefix(storedHash, "$argon2id$") && !strings.HasPrefix(storedHash, "$2") {
		t.Errorf("expected hashed password in store, got %q", storedHash)
	}

	// 3. Case-insensitive userName filter and dotted emails.value filter.
	w = oktaReq(http.MethodGet, usersPath+"?filter="+url.QueryEscape(`userName eq "runscope.user@example.com"`), "")
	requireStatus(t, w, http.StatusOK)
	if decode(t, w)["totalResults"] != float64(1) {
		t.Fatalf("Okta case-insensitive userName filter failed: %s", w.Body.String())
	}

	w = oktaReq(http.MethodGet, usersPath+"?filter="+url.QueryEscape(`emails.value eq "runscope.user@example.com"`), "")
	requireStatus(t, w, http.StatusOK)
	if decode(t, w)["totalResults"] != float64(1) {
		t.Fatalf("Okta emails.value filter failed: %s", w.Body.String())
	}

	// 4. Case-insensitive uniqueness conflict (409 uniqueness).
	w = oktaReq(http.MethodPost, usersPath, `{
		"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
		"userName": "RUNSCOPE.USER@EXAMPLE.COM"
	}`)
	errBody := requireSCIMError(t, w, http.StatusConflict)
	if errBody["scimType"] != "uniqueness" {
		t.Errorf("scimType = %v, want uniqueness", errBody["scimType"])
	}

	// 5. Okta Group Push: create group with user as member.
	w = oktaReq(http.MethodPost, "/scim/v2/Groups", `{
		"schemas": ["urn:ietf:params:scim:schemas:core:2.0:Group"],
		"externalId": "00g1oktaGroup",
		"displayName": "Okta Engineering",
		"members": [
			{"value": "`+userID+`", "display": "Runscope User"}
		]
	}`)
	requireStatus(t, w, http.StatusCreated)
	grpID := decode(t, w)["id"].(string)

	// 6. Read user (now carrying server-computed readOnly groups) and perform an
	// Okta full-profile PUT that echoes groups, clears title via null, and clears
	// the optional EnterpriseUser extension via null.
	w = oktaReq(http.MethodGet, usersPath+"/"+userID, "")
	requireStatus(t, w, http.StatusOK)
	fetched := decode(t, w)
	userGroups, _ := fetched["groups"].([]interface{})
	if len(userGroups) != 1 {
		t.Fatalf("expected 1 group on user before PUT, got %v", fetched["groups"])
	}

	fetched["displayName"] = "Runscope Updated"
	fetched["title"] = nil
	fetched[schemaUserEnt] = nil
	putRaw, _ := json.Marshal(fetched)

	w = oktaReq(http.MethodPut, usersPath+"/"+userID, string(putRaw))
	requireStatus(t, w, http.StatusOK)
	replaced := decode(t, w)
	if replaced["displayName"] != "Runscope Updated" {
		t.Errorf("displayName = %v", replaced["displayName"])
	}
	if _, hasTitle := replaced["title"]; hasTitle {
		t.Errorf("title set to null on PUT should be cleared, got %v", replaced["title"])
	}
	if _, hasEnt := replaced[schemaUserEnt]; hasEnt {
		t.Errorf("optional extension set to null on PUT should be cleared, got %v", replaced[schemaUserEnt])
	}
	// Server-owned User.groups and bcrypt password hash must remain intact!
	if rg, _ := replaced["groups"].([]interface{}); len(rg) != 1 {
		t.Errorf("echoing readOnly groups on Okta PUT must preserve membership, got %v", replaced["groups"])
	}
	afterPutDoc, _ := store.Get("User", userID)
	if afterPutDoc["password"] != storedHash {
		t.Errorf("omitting password on Okta PUT must preserve existing hash")
	}

	// 7. Deactivate user via Okta pathless PATCH, then reactivate.
	w = oktaReq(http.MethodPatch, usersPath+"/"+userID, `{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
		"Operations": [{"op": "replace", "value": {"active": false}}]
	}`)
	requireStatus(t, w, http.StatusOK)
	if decode(t, w)["active"] != false {
		t.Errorf("expected active=false after deactivation PATCH")
	}

	// 8. Delete group and verify User.groups is cleaned up.
	w = oktaReq(http.MethodDelete, "/scim/v2/Groups/"+grpID, "")
	requireStatus(t, w, http.StatusNoContent)

	w = oktaReq(http.MethodGet, usersPath+"/"+userID, "")
	requireStatus(t, w, http.StatusOK)
	if g, has := decode(t, w)["groups"].([]interface{}); has && len(g) > 0 {
		t.Errorf("expected User.groups to be empty after Okta Group delete, got %v", g)
	}
}
