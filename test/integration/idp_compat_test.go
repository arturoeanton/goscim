//go:build integration

package integration_test

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arturoeanton/goscim/scim"
	"github.com/gin-gonic/gin"
)

func newAuthenticatedIntegrationRouter(t *testing.T, auth scim.Authenticator) *gin.Engine {
	t.Helper()
	engine := gin.New()
	router, err := scim.NewRouter("../../config", engine, auth)
	if err != nil {
		t.Fatalf("NewRouter failed: %v", err)
	}
	t.Cleanup(func() {
		_, _ = scim.NewRouter("../../config", gin.New(),
			&scim.AnonymousAuthenticator{Roles: []string{"user", "admin", "superadmin", "role1", "role2"}})
	})
	return router
}

func doAuth(t *testing.T, router *gin.Engine, token, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	return doAuthHeaders(t, router, token, method, target, body, nil)
}

func doAuthHeaders(t *testing.T, router *gin.Engine, token, method, target, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/scim+json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func signJWS(t *testing.T, alg, kid string, claims map[string]interface{}, key interface{}) string {
	t.Helper()
	headerBytes, err := json.Marshal(map[string]string{
		"alg": alg,
		"kid": kid,
		"typ": "at+jwt",
	})
	if err != nil {
		t.Fatalf("marshal JWS header: %v", err)
	}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal JWS claims: %v", err)
	}
	signingInput := base64.RawURLEncoding.EncodeToString(headerBytes) + "." +
		base64.RawURLEncoding.EncodeToString(payloadBytes)

	var sig []byte
	switch k := key.(type) {
	case *rsa.PrivateKey:
		digest := sha256.Sum256([]byte(signingInput))
		sig, err = rsa.SignPKCS1v15(rand.Reader, k, crypto.SHA256, digest[:])
	case *ecdsa.PrivateKey:
		digest := sha256.Sum256([]byte(signingInput))
		r, s, signErr := ecdsa.Sign(rand.Reader, k, digest[:])
		err = signErr
		if err == nil {
			sig = make([]byte, 64)
			r.FillBytes(sig[:32])
			s.FillBytes(sig[32:])
		}
	case ed25519.PrivateKey:
		sig = ed25519.Sign(k, []byte(signingInput))
	default:
		t.Fatalf("unsupported signing key type %T", key)
	}
	if err != nil {
		t.Fatalf("sign JWS (%s): %v", alg, err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func rsaPublicJWK(kid, use, alg string, pub *rsa.PublicKey) map[string]interface{} {
	return map[string]interface{}{
		"kty": "RSA",
		"kid": kid,
		"use": use,
		"alg": alg,
		"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
}

func ecP256PublicJWK(kid string, pub *ecdsa.PublicKey) map[string]interface{} {
	x := make([]byte, 32)
	y := make([]byte, 32)
	pub.X.FillBytes(x)
	pub.Y.FillBytes(y)
	return map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"kid": kid,
		"use": "sig",
		"alg": "ES256",
		"x":   base64.RawURLEncoding.EncodeToString(x),
		"y":   base64.RawURLEncoding.EncodeToString(y),
	}
}

func ed25519PublicJWK(kid string, pub ed25519.PublicKey) map[string]interface{} {
	return map[string]interface{}{
		"kty": "OKP",
		"crv": "Ed25519",
		"kid": kid,
		"use": "sig",
		"alg": "EdDSA",
		"x":   base64.RawURLEncoding.EncodeToString(pub),
	}
}

// TestMicrosoftEntraIDLifecycleAgainstCouchbase exercises Microsoft Entra ID's
// SCIM 2.0 provisioning wire protocol end-to-end against Couchbase KV, atomic
// uniqueness sentinels, and N1QL queries.
func TestMicrosoftEntraIDLifecycleAgainstCouchbase(t *testing.T) {
	const secretToken = "entra-secret-token-couchbase-2026"
	bearerAuth, err := scim.NewBearerAuthenticator([]scim.BearerTokenEntry{{
		Subject: "entra-provisioning-service",
		Token:   secretToken,
		Roles:   []string{"admin"},
	}})
	if err != nil {
		t.Fatalf("NewBearerAuthenticator: %v", err)
	}
	router := newAuthenticatedIntegrationRouter(t, bearerAuth)

	// Discovery advertises oauthbearertoken; unauthenticated /Users returns 401.
	wSPC := doAuth(t, router, "", http.MethodGet, scim.PREFIX+"/ServiceProviderConfig", "")
	requireStatus(t, wSPC, http.StatusOK)
	schemes, _ := decode(t, wSPC)["authenticationSchemes"].([]interface{})
	if len(schemes) != 1 || schemes[0].(map[string]interface{})["type"] != "oauthbearertoken" {
		t.Fatalf("ServiceProviderConfig authenticationSchemes = %#v, want oauthbearertoken", schemes)
	}
	wUnauth := doAuth(t, router, "", http.MethodGet, usersPath, "")
	requireStatus(t, wUnauth, http.StatusUnauthorized)

	prefix := fmt.Sprintf("entra-%d", time.Now().UnixNano())
	upn := "AdeleV." + prefix + "@contoso.onmicrosoft.com"
	extID := "EntraObj-" + prefix + "-CaseExact"
	empNo := "EMP-" + prefix

	// Pre-create N1QL lookup with uppercase UPN returns totalResults = 0.
	qPre := "filter=" + url.QueryEscape(fmt.Sprintf("userName eq %q", strings.ToUpper(upn)))
	wPre := doAuth(t, router, secretToken, http.MethodGet, usersPath+"?"+qPre, "")
	requireStatus(t, wPre, http.StatusOK)
	if total := int(decode(t, wPre)["totalResults"].(float64)); total != 0 {
		t.Fatalf("pre-create totalResults = %d, want 0", total)
	}

	// Provision manager and Adele with Entra gallery payload (externalId,
	// addresses.primary, EnterpriseUser extension).
	mgrRaw, _ := json.Marshal(map[string]interface{}{
		"schemas":  []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		"userName": "MeganB." + prefix + "@contoso.onmicrosoft.com",
		"active":   true,
	})
	wMgr := doAuth(t, router, secretToken, http.MethodPost, usersPath, string(mgrRaw))
	requireStatus(t, wMgr, http.StatusCreated)
	managerID := decode(t, wMgr)["id"].(string)

	adeleRaw, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{
			"urn:ietf:params:scim:schemas:core:2.0:User",
			"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
		},
		"externalId":  extID,
		"userName":    upn,
		"active":      true,
		"displayName": "Adele Vance",
		"name": map[string]interface{}{
			"givenName":  "Adele",
			"familyName": "Vance",
		},
		"emails": []map[string]interface{}{
			{"value": strings.ToLower(upn), "type": "work", "primary": true},
		},
		"addresses": []map[string]interface{}{
			{
				"type":          "work",
				"streetAddress": "One Microsoft Way",
				"locality":      "Redmond",
				"region":        "WA",
				"postalCode":    "98052",
				"country":       "US",
				"primary":       true,
			},
		},
		"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]interface{}{
			"employeeNumber": empNo,
			"department":     "Product Marketing",
			"costCenter":     "CC-100",
		},
	})
	wCreate := doAuth(t, router, secretToken, http.MethodPost, usersPath, string(adeleRaw))
	requireStatus(t, wCreate, http.StatusCreated)
	adeleID := decode(t, wCreate)["id"].(string)

	var groupID string
	t.Cleanup(func() {
		if groupID != "" {
			doAuth(t, router, secretToken, http.MethodDelete, scim.PREFIX+"/Groups/"+groupID, "")
		}
		doAuth(t, router, secretToken, http.MethodDelete, usersPath+"/"+adeleID, "")
		doAuth(t, router, secretToken, http.MethodDelete, usersPath+"/"+managerID, "")
	})

	// 1. N1QL case-insensitive userName filter (userName is caseExact: false).
	qLower := "filter=" + url.QueryEscape(fmt.Sprintf("userName eq %q", strings.ToLower(upn)))
	wLower := doAuth(t, router, secretToken, http.MethodGet, usersPath+"?"+qLower, "")
	requireStatus(t, wLower, http.StatusOK)
	if got := userNamesIn(t, decode(t, wLower)); len(got) != 1 || got[0] != upn {
		t.Fatalf("N1QL case-insensitive userName search = %v, want [%s]", got, upn)
	}

	// 2. N1QL case-exact externalId filter (externalId is caseExact: true per RFC 7643 3.1).
	qExtExact := "filter=" + url.QueryEscape(fmt.Sprintf("externalId eq %q", extID))
	wExtExact := doAuth(t, router, secretToken, http.MethodGet, usersPath+"?"+qExtExact, "")
	requireStatus(t, wExtExact, http.StatusOK)
	if total := int(decode(t, wExtExact)["totalResults"].(float64)); total != 1 {
		t.Fatalf("N1QL exact externalId totalResults = %d, want 1", total)
	}

	qExtWrongCase := "filter=" + url.QueryEscape(fmt.Sprintf("externalId eq %q", strings.ToUpper(extID)))
	wExtWrong := doAuth(t, router, secretToken, http.MethodGet, usersPath+"?"+qExtWrongCase, "")
	requireStatus(t, wExtWrong, http.StatusOK)
	if total := int(decode(t, wExtWrong)["totalResults"].(float64)); total != 0 {
		t.Fatalf("N1QL mismatched-case externalId totalResults = %d, want 0", total)
	}

	// 3. N1QL colon-delimited EnterpriseUser URN filter.
	qEmp := "filter=" + url.QueryEscape(
		fmt.Sprintf("urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:employeeNumber eq %q", empNo))
	wEmp := doAuth(t, router, secretToken, http.MethodGet, usersPath+"?"+qEmp, "")
	requireStatus(t, wEmp, http.StatusOK)
	if total := int(decode(t, wEmp)["totalResults"].(float64)); total != 1 {
		t.Fatalf("N1QL colon-URN employeeNumber totalResults = %d, want 1", total)
	}

	// 4. Entra multi-op PATCH: bracket filter, pathless colon-URN keys,
	// stringified boolean ("False"), and stringified EnterpriseUser.manager ID.
	newEmail := "adele.updated." + prefix + "@contoso.com"
	patchUserRaw, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		"Operations": []map[string]interface{}{
			{
				"op":    "Replace",
				"path":  `emails[type eq "work"].value`,
				"value": newEmail,
			},
			{
				"op": "Add",
				"value": map[string]interface{}{
					"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department": "Cloud Engineering",
					"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager":    managerID,
					"name.givenName": "Adele Miriam",
					"active":         "False",
				},
			},
		},
	})
	wPatch := doAuth(t, router, secretToken, http.MethodPatch, usersPath+"/"+adeleID, string(patchUserRaw))
	requireStatus(t, wPatch, http.StatusOK)
	patched := decode(t, wPatch)
	if activeVal, ok := patched["active"].(bool); !ok || activeVal {
		t.Fatalf("patched active = %#v, want boolean false", patched["active"])
	}

	// Verify normalized boolean and normalized manager.value are queryable in Couchbase N1QL.
	qMgr := "filter=" + url.QueryEscape(fmt.Sprintf(
		`id eq %q and active eq false and urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager.value eq %q`,
		adeleID, managerID))
	wMgrSearch := doAuth(t, router, secretToken, http.MethodGet, usersPath+"?"+qMgr, "")
	requireStatus(t, wMgrSearch, http.StatusOK)
	if total := int(decode(t, wMgrSearch)["totalResults"].(float64)); total != 1 {
		t.Fatalf("N1QL post-PATCH active+manager.value totalResults = %d, want 1", total)
	}

	// 5. Entra Group lifecycle: ?excludedAttributes=members, Add members,
	// legacy Remove with value array (preserving other members), and bracket Remove.
	grpRaw, _ := json.Marshal(map[string]interface{}{
		"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:Group"},
		"externalId":  "EntraGrp-" + prefix,
		"displayName": "Security-" + prefix,
	})
	wGrp := doAuth(t, router, secretToken, http.MethodPost, scim.PREFIX+"/Groups?excludedAttributes=members", string(grpRaw))
	requireStatus(t, wGrp, http.StatusCreated)
	groupID = decode(t, wGrp)["id"].(string)

	addMembersRaw, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		"Operations": []map[string]interface{}{
			{
				"op":   "Add",
				"path": "members",
				"value": []map[string]interface{}{
					{"value": adeleID},
					{"value": managerID},
				},
			},
		},
	})
	requireStatus(t, doAuth(t, router, secretToken, http.MethodPatch, scim.PREFIX+"/Groups/"+groupID, string(addMembersRaw)), http.StatusOK)

	legacyRemoveRaw, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		"Operations": []map[string]interface{}{
			{
				"op":   "Remove",
				"path": "members",
				"value": []map[string]interface{}{
					{"value": adeleID},
				},
			},
		},
	})
	wLegacyRem := doAuth(t, router, secretToken, http.MethodPatch, scim.PREFIX+"/Groups/"+groupID, string(legacyRemoveRaw))
	requireStatus(t, wLegacyRem, http.StatusOK)
	remMembers, _ := decode(t, wLegacyRem)["members"].([]interface{})
	if len(remMembers) != 1 || remMembers[0].(map[string]interface{})["value"] != managerID {
		t.Fatalf("members after legacy Entra Remove = %#v, want only %s", remMembers, managerID)
	}

	// Confirm User.groups sync in Couchbase: Adele removed, Manager retained.
	wAdeleAfter := doAuth(t, router, secretToken, http.MethodGet, usersPath+"/"+adeleID+"?attributes=groups", "")
	requireStatus(t, wAdeleAfter, http.StatusOK)
	if g, has := decode(t, wAdeleAfter)["groups"].([]interface{}); has && len(g) != 0 {
		t.Fatalf("Adele groups after removal = %#v, want empty", g)
	}
	wMgrAfter := doAuth(t, router, secretToken, http.MethodGet, usersPath+"/"+managerID+"?attributes=groups", "")
	requireStatus(t, wMgrAfter, http.StatusOK)
	if g, _ := decode(t, wMgrAfter)["groups"].([]interface{}); len(g) != 1 {
		t.Fatalf("Manager groups after Adele removal = %#v, want 1 entry", g)
	}

	// Modern ?aadOptscim062020 bracket-filter Remove for remaining member.
	modernRemoveRaw, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		"Operations": []map[string]interface{}{
			{
				"op":   "Remove",
				"path": fmt.Sprintf("members[value eq %q]", managerID),
			},
		},
	})
	wModRem := doAuth(t, router, secretToken, http.MethodPatch, scim.PREFIX+"/Groups/"+groupID, string(modernRemoveRaw))
	requireStatus(t, wModRem, http.StatusOK)
	if m, has := decode(t, wModRem)["members"].([]interface{}); has && len(m) != 0 {
		t.Fatalf("members after modern Entra Remove = %#v, want empty", m)
	}
}

// TestOktaRunscopeLifecycleAgainstCouchbase verifies Okta Lifecycle Management,
// ES256 OAuth 2.0 JWT authentication, Runscope conformance flows, dotted
// multi-valued N1QL filters, atomic case-insensitive uniqueness, full PUT
// readOnly/writeOnly/null semantics, and Group Push against Couchbase.
func TestOktaRunscopeLifecycleAgainstCouchbase(t *testing.T) {
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey P-256: %v", err)
	}
	const (
		issuer   = "https://acme.okta.com/oauth2/aus-scim"
		audience = "https://scim.acme.internal/scim/v2"
		kid      = "okta-es256-2026"
	)
	jwksSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []interface{}{ecP256PublicJWK(kid, &ecKey.PublicKey)},
		})
	}))
	defer jwksSrv.Close()

	jwtAuth, err := scim.NewJWTAuthenticator(jwksSrv.URL, issuer, audience, "")
	if err != nil {
		t.Fatalf("NewJWTAuthenticator: %v", err)
	}
	router := newAuthenticatedIntegrationRouter(t, jwtAuth)

	token := signJWS(t, "ES256", kid, map[string]interface{}{
		"iss": issuer,
		"aud": audience,
		"cid": "0oa-okta-scim-client",
		"scp": []string{"scim:admin", "admin"},
		"exp": time.Now().Add(10 * time.Minute).Unix(),
	}, ecKey)

	// Case-insensitive schema discovery + canonicalValues/referenceTypes exposure.
	wSchema := doAuth(t, router, "", http.MethodGet, scim.PREFIX+"/Schemas/urn:ietf:params:scim:schemas:core:2.0:user", "")
	requireStatus(t, wSchema, http.StatusOK)
	if !strings.Contains(wSchema.Body.String(), `"canonicalValues"`) || !strings.Contains(wSchema.Body.String(), `"referenceTypes"`) {
		t.Fatalf("Schemas/User missing canonicalValues or referenceTypes metadata")
	}

	prefix := fmt.Sprintf("okta-%d", time.Now().UnixNano())
	userName := "okta.user." + prefix + "@example.com"
	secondaryEmail := "okta.alt." + prefix + "@example.com"
	extID := "00u-" + prefix

	// 1. Okta User creation with externalId, password, and explicit null optional fields.
	createRaw, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{
			"urn:ietf:params:scim:schemas:core:2.0:User",
			"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
		},
		"externalId":  extID,
		"userName":    userName,
		"password":    "OktaInitialSecret#2026",
		"active":      true,
		"displayName": "Okta Test User",
		"title":       "Staff Identity Engineer",
		"nickName":    nil,
		"profileUrl":  nil,
		"name": map[string]interface{}{
			"givenName":  "Okta",
			"familyName": "User",
			"middleName": nil,
		},
		"emails": []map[string]interface{}{
			{"value": userName, "type": "work", "primary": true},
			{"value": secondaryEmail, "type": "home", "primary": false},
		},
		"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]interface{}{
			"department": "Identity Engineering",
			"division":   nil,
		},
	})
	wCreate := doAuth(t, router, token, http.MethodPost, usersPath, string(createRaw))
	requireStatus(t, wCreate, http.StatusCreated)
	created := decode(t, wCreate)
	userID := created["id"].(string)
	if _, leaked := created["password"]; leaked {
		t.Fatalf("password must never be returned on HTTP response")
	}
	if _, present := created["nickName"]; present {
		t.Fatalf("null nickName must be stripped, got %#v", created["nickName"])
	}

	var groupID string
	t.Cleanup(func() {
		if groupID != "" {
			doAuth(t, router, token, http.MethodDelete, scim.PREFIX+"/Groups/"+groupID, "")
		}
		doAuth(t, router, token, http.MethodDelete, usersPath+"/"+userID, "")
	})

	// Confirm Couchbase stored an Argon2id password hash rather than plaintext.
	rawDoc, err := scim.DB.Get("User", userID)
	if err != nil {
		t.Fatalf("DB.Get User: %v", err)
	}
	storedHash, _ := rawDoc["password"].(string)
	if !strings.HasPrefix(storedHash, "$argon2id$") {
		t.Fatalf("stored password in Couchbase = %q, want $argon2id$ hash", storedHash)
	}

	// 2. N1QL case-insensitive userName lookup + dotted emails.value lookup.
	qUser := "filter=" + url.QueryEscape(fmt.Sprintf("userName eq %q", strings.ToUpper(userName)))
	wUserSearch := doAuth(t, router, token, http.MethodGet, usersPath+"?"+qUser, "")
	requireStatus(t, wUserSearch, http.StatusOK)
	if total := int(decode(t, wUserSearch)["totalResults"].(float64)); total != 1 {
		t.Fatalf("N1QL uppercase userName lookup totalResults = %d, want 1", total)
	}

	qEmail := "filter=" + url.QueryEscape(fmt.Sprintf("emails.value eq %q", strings.ToUpper(secondaryEmail)))
	wEmailSearch := doAuth(t, router, token, http.MethodGet, usersPath+"?"+qEmail, "")
	requireStatus(t, wEmailSearch, http.StatusOK)
	if total := int(decode(t, wEmailSearch)["totalResults"].(float64)); total != 1 {
		t.Fatalf("N1QL dotted emails.value lookup totalResults = %d, want 1", total)
	}

	// 3. Duplicate userName with different casing rejected with 409 uniqueness.
	dupRaw, _ := json.Marshal(map[string]interface{}{
		"schemas":  []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		"userName": strings.ToUpper(userName),
	})
	wDup := doAuth(t, router, token, http.MethodPost, usersPath, string(dupRaw))
	requireStatus(t, wDup, http.StatusConflict)
	if decode(t, wDup)["scimType"] != "uniqueness" {
		t.Fatalf("duplicate scimType = %v, want uniqueness", decode(t, wDup)["scimType"])
	}

	// 4. Okta Group Push creates group and populates readOnly User.groups in Couchbase.
	grpRaw, _ := json.Marshal(map[string]interface{}{
		"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:Group"},
		"externalId":  "00g-" + prefix,
		"displayName": "Okta-Push-" + prefix,
		"members": []map[string]interface{}{
			{"value": userID, "display": "Okta Test User"},
		},
	})
	wGrp := doAuth(t, router, token, http.MethodPost, scim.PREFIX+"/Groups", string(grpRaw))
	requireStatus(t, wGrp, http.StatusCreated)
	groupID = decode(t, wGrp)["id"].(string)

	wGetBeforePut := doAuth(t, router, token, http.MethodGet, usersPath+"/"+userID, "")
	requireStatus(t, wGetBeforePut, http.StatusOK)
	beforePut := decode(t, wGetBeforePut)

	// 5. Okta Full PUT: echoes readOnly groups + meta, omits password, clears
	// title and EnterpriseUser via explicit JSON null.
	putRaw, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{
			"urn:ietf:params:scim:schemas:core:2.0:User",
			"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
		},
		"id":          userID,
		"externalId":  extID,
		"userName":    userName,
		"active":      true,
		"displayName": "Okta Updated User",
		"title":       nil,
		"groups":      beforePut["groups"],
		"meta":        beforePut["meta"],
		"name": map[string]interface{}{
			"givenName":  "Okta",
			"familyName": "Updated",
		},
		"emails": []map[string]interface{}{
			{"value": userName, "type": "work", "primary": true},
		},
		"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": nil,
	})
	wPut := doAuth(t, router, token, http.MethodPut, usersPath+"/"+userID, string(putRaw))
	requireStatus(t, wPut, http.StatusOK)
	afterPut := decode(t, wPut)
	if _, hasTitle := afterPut["title"]; hasTitle {
		t.Fatalf("PUT with title:null must clear title, got %#v", afterPut["title"])
	}
	if _, hasEnt := afterPut["urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"]; hasEnt {
		t.Fatalf("PUT with EnterpriseUser:null must clear extension")
	}
	if groupsAfter, _ := afterPut["groups"].([]interface{}); len(groupsAfter) != 1 {
		t.Fatalf("PUT must preserve computed User.groups, got %#v", afterPut["groups"])
	}

	// Stored Argon2id password hash must survive PUT that omits writeOnly password.
	rawDocAfterPut, err := scim.DB.Get("User", userID)
	if err != nil {
		t.Fatalf("DB.Get User after PUT: %v", err)
	}
	if rawDocAfterPut["password"] != storedHash {
		t.Fatalf("password hash changed/lost across PUT without password")
	}

	// 6. Okta Deactivation via pathless PATCH.
	deactRaw, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		"Operations": []map[string]interface{}{
			{"op": "replace", "value": map[string]interface{}{"active": false}},
		},
	})
	wDeact := doAuth(t, router, token, http.MethodPatch, usersPath+"/"+userID, string(deactRaw))
	requireStatus(t, wDeact, http.StatusOK)
	if decode(t, wDeact)["active"] != false {
		t.Fatalf("deactivated user active = %v, want false", decode(t, wDeact)["active"])
	}
}

// TestKeycloakOIDCAndSCIMAgainstCouchbase verifies Keycloak realm JWKS
// verification (mixed sig + enc keys across RS256, ES256, EdDSA), nested role
// claims (realm_access.roles and resource_access.<client>.roles), /Bulk realm
// export, ETag/If-Match optimistic concurrency, and N1QL reconciliation queries
// against Couchbase.
func TestKeycloakOIDCAndSCIMAgainstCouchbase(t *testing.T) {
	rsaSigKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey sig: %v", err)
	}
	rsaEncKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey enc: %v", err)
	}
	ecSigKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ecdsa.GenerateKey: %v", err)
	}
	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("ed25519.GenerateKey: %v", err)
	}

	const (
		issuer   = "https://sso.corp.internal/realms/platform"
		audience = "goscim-service"
	)
	jwksSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []interface{}{
				rsaPublicJWK("kc-rsa-enc", "enc", "RSA-OAEP", &rsaEncKey.PublicKey),
				rsaPublicJWK("kc-rs256-sig", "sig", "RS256", &rsaSigKey.PublicKey),
				ecP256PublicJWK("kc-es256-sig", &ecSigKey.PublicKey),
				ed25519PublicJWK("kc-eddsa-sig", edPub),
			},
		})
	}))
	defer jwksSrv.Close()

	jwtAuth, err := scim.NewJWTAuthenticator(
		jwksSrv.URL,
		issuer,
		audience,
		"resource_access.goscim-service.roles",
	)
	if err != nil {
		t.Fatalf("NewJWTAuthenticator: %v", err)
	}
	router := newAuthenticatedIntegrationRouter(t, jwtAuth)

	exp := time.Now().Add(10 * time.Minute).Unix()

	// Token signed with Keycloak's RSA-OAEP encryption key must be rejected.
	encToken := signJWS(t, "RS256", "kc-rsa-enc", map[string]interface{}{
		"iss": issuer,
		"aud": audience,
		"sub": "attacker",
		"exp": exp,
	}, rsaEncKey)
	requireStatus(t, doAuth(t, router, encToken, http.MethodGet, usersPath, ""), http.StatusUnauthorized)

	// RS256 token carrying default Keycloak realm_access.roles.
	rsToken := signJWS(t, "RS256", "kc-rs256-sig", map[string]interface{}{
		"iss":          issuer,
		"aud":          []string{"account", audience},
		"sub":          "f:kc-realm:service-account-scim",
		"realm_access": map[string]interface{}{"roles": []string{"offline_access", "admin"}},
		"exp":          exp,
	}, rsaSigKey)

	// ES256 token carrying Keycloak client roles in resource_access.goscim-service.roles.
	esToken := signJWS(t, "ES256", "kc-es256-sig", map[string]interface{}{
		"iss": issuer,
		"aud": audience,
		"azp": "keycloak-scim-sync",
		"resource_access": map[string]interface{}{
			audience: map[string]interface{}{"roles": []string{"admin"}},
		},
		"exp": exp,
	}, ecSigKey)

	// EdDSA token carrying azp without sub.
	edToken := signJWS(t, "EdDSA", "kc-eddsa-sig", map[string]interface{}{
		"iss":          issuer,
		"aud":          audience,
		"azp":          "keycloak-eddsa-sync",
		"realm_access": map[string]interface{}{"roles": []string{"admin"}},
		"exp":          exp,
	}, edPriv)

	prefix := fmt.Sprintf("kc-%d", time.Now().UnixNano())
	kcUUID1 := "kc-uuid-1-" + prefix
	kcUUID2 := "kc-uuid-2-" + prefix
	uName1 := "jdoe-" + prefix
	uName2 := "msmith-" + prefix

	// 1. Keycloak initial realm sync via /Bulk authenticated with RS256 token.
	bulkBody, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:BulkRequest"},
		"Operations": []map[string]interface{}{
			{
				"method": "POST",
				"bulkId": "u1",
				"path":   "/Users",
				"data": map[string]interface{}{
					"schemas":    []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
					"externalId": kcUUID1,
					"userName":   uName1,
					"active":     true,
					"emails": []map[string]interface{}{
						{"value": uName1 + "@corp.internal", "type": "work", "primary": true},
					},
				},
			},
			{
				"method": "POST",
				"bulkId": "u2",
				"path":   "/Users",
				"data": map[string]interface{}{
					"schemas":    []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
					"externalId": kcUUID2,
					"userName":   uName2,
					"active":     true,
				},
			},
			{
				"method": "POST",
				"bulkId": "g1",
				"path":   "/Groups",
				"data": map[string]interface{}{
					"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:Group"},
					"externalId":  "kc-grp-" + prefix,
					"displayName": "realm-admins-" + prefix,
					"members": []map[string]interface{}{
						{"value": "bulkId:u1", "type": "User"},
						{"value": "bulkId:u2", "type": "User"},
					},
				},
			},
		},
	})
	wBulk := doAuth(t, router, rsToken, http.MethodPost, scim.PREFIX+"/Bulk", string(bulkBody))
	requireStatus(t, wBulk, http.StatusOK)
	ops := decode(t, wBulk)["Operations"].([]interface{})
	idAt := func(i int) string {
		loc := ops[i].(map[string]interface{})["location"].(string)
		return loc[strings.LastIndex(loc, "/")+1:]
	}
	u1ID := idAt(0)
	u2ID := idAt(1)
	g1ID := idAt(2)

	t.Cleanup(func() {
		doAuth(t, router, rsToken, http.MethodDelete, scim.PREFIX+"/Groups/"+g1ID, "")
		doAuth(t, router, rsToken, http.MethodDelete, usersPath+"/"+u1ID, "")
		doAuth(t, router, rsToken, http.MethodDelete, usersPath+"/"+u2ID, "")
	})

	// 2. Reconciliation N1QL queries authenticated with ES256 and EdDSA tokens:
	// - externalId eq "<kc-uuid>"
	// - core URN-prefixed userName eq "<username>"
	// - Group members.value eq "<userId>"
	qExt := "filter=" + url.QueryEscape(fmt.Sprintf("externalId eq %q", kcUUID1))
	wExt := doAuth(t, router, esToken, http.MethodGet, usersPath+"?"+qExt, "")
	requireStatus(t, wExt, http.StatusOK)
	if total := int(decode(t, wExt)["totalResults"].(float64)); total != 1 {
		t.Fatalf("Keycloak N1QL externalId lookup totalResults = %d, want 1", total)
	}

	qCoreURN := "filter=" + url.QueryEscape(
		fmt.Sprintf("urn:ietf:params:scim:schemas:core:2.0:User:userName eq %q", strings.ToUpper(uName1)))
	wCoreURN := doAuth(t, router, edToken, http.MethodGet, usersPath+"?"+qCoreURN, "")
	requireStatus(t, wCoreURN, http.StatusOK)
	if total := int(decode(t, wCoreURN)["totalResults"].(float64)); total != 1 {
		t.Fatalf("Keycloak N1QL core-URN userName lookup totalResults = %d, want 1", total)
	}

	qGrpMember := "filter=" + url.QueryEscape(fmt.Sprintf("members.value eq %q", u1ID))
	wGrpMember := doAuth(t, router, edToken, http.MethodGet, scim.PREFIX+"/Groups?"+qGrpMember, "")
	requireStatus(t, wGrpMember, http.StatusOK)
	if total := int(decode(t, wGrpMember)["totalResults"].(float64)); total != 1 {
		t.Fatalf("Keycloak N1QL Groups members.value lookup totalResults = %d, want 1", total)
	}

	// 3. Optimistic concurrency control (ETag + If-Match) against Couchbase.
	wGetU1 := doAuth(t, router, esToken, http.MethodGet, usersPath+"/"+u1ID, "")
	requireStatus(t, wGetU1, http.StatusOK)
	etag := wGetU1.Header().Get("ETag")
	if etag == "" {
		t.Fatalf("expected ETag header on GET /Users/%s", u1ID)
	}

	patchTitle, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		"Operations": []map[string]interface{}{
			{"op": "replace", "path": "title", "value": "Principal Architect"},
		},
	})
	wStale := doAuthHeaders(t, router, esToken, http.MethodPatch, usersPath+"/"+u1ID, string(patchTitle),
		map[string]string{"If-Match": `W/"stale-etag"`})
	requireStatus(t, wStale, http.StatusPreconditionFailed)

	wFresh := doAuthHeaders(t, router, esToken, http.MethodPatch, usersPath+"/"+u1ID, string(patchTitle),
		map[string]string{"If-Match": etag})
	requireStatus(t, wFresh, http.StatusOK)

	// 4. Root POST /.search across resources in Couchbase.
	rootSearchBody, _ := json.Marshal(map[string]interface{}{
		"schemas":    []string{"urn:ietf:params:scim:api:messages:2.0:SearchRequest"},
		"filter":     fmt.Sprintf("externalId eq %q", kcUUID1),
		"attributes": []string{"userName", "externalId"},
	})
	wRootSearch := doAuth(t, router, edToken, http.MethodPost, scim.PREFIX+"/.search", string(rootSearchBody))
	requireStatus(t, wRootSearch, http.StatusOK)
	if total := int(decode(t, wRootSearch)["totalResults"].(float64)); total != 1 {
		t.Fatalf("root POST /.search externalId totalResults = %d, want 1", total)
	}
}
