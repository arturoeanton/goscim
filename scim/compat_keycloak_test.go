package scim

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestKeycloakRealmJWKSAndNestedRoles simulates a Keycloak realm JWKS endpoint
// (/realms/corp/protocol/openid-connect/certs) publishing mixed RSA-OAEP
// encryption keys alongside RS256, ES256, and EdDSA signing keys, and verifies
// both default realm_access.roles and SCIM_JWT_ROLES_CLAIM=resource_access.<client>.roles.
func TestKeycloakRealmJWKSAndNestedRoles(t *testing.T) {
	rsaEncKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	rsaSigKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	ecSigKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	coordLen := (elliptic.P256().Params().BitSize + 7) / 8
	padEC := func(b []byte) string {
		buf := make([]byte, coordLen)
		copy(buf[coordLen-len(b):], b)
		return base64.RawURLEncoding.EncodeToString(buf)
	}

	jwksSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []map[string]interface{}{
				{
					"kid": "kc-rsa-enc",
					"kty": "RSA",
					"alg": "RSA-OAEP",
					"use": "enc",
					"n":   base64.RawURLEncoding.EncodeToString(rsaEncKey.N.Bytes()),
					"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(rsaEncKey.E)).Bytes()),
				},
				{
					"kid": "kc-rs256-sig",
					"kty": "RSA",
					"alg": "RS256",
					"use": "sig",
					"n":   base64.RawURLEncoding.EncodeToString(rsaSigKey.N.Bytes()),
					"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(rsaSigKey.E)).Bytes()),
				},
				{
					"kid": "kc-es256-sig",
					"kty": "EC",
					"alg": "ES256",
					"use": "sig",
					"crv": "P-256",
					"x":   padEC(ecSigKey.X.Bytes()),
					"y":   padEC(ecSigKey.Y.Bytes()),
				},
				{
					"kid": "kc-eddsa-sig",
					"kty": "OKP",
					"alg": "EdDSA",
					"use": "sig",
					"crv": "Ed25519",
					"x":   base64.RawURLEncoding.EncodeToString(edPub),
				},
			},
		})
	}))
	t.Cleanup(jwksSrv.Close)

	kcIssuer := "https://sso.corp.example/realms/corp"
	auth, err := NewJWTAuthenticator(
		jwksSrv.URL,
		kcIssuer,
		"goscim-service",
		"resource_access.goscim-service.roles",
	)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := newTestServerAs(t, auth)

	baseClaims := func() jwt.MapClaims {
		return jwt.MapClaims{
			"iss":       kcIssuer,
			"aud":       "goscim-service",
			"azp":       "scim-sync-worker",
			"client_id": "scim-sync-worker",
			"typ":       "Bearer",
			"exp":       time.Now().Add(time.Hour).Unix(),
			"realm_access": map[string]interface{}{
				"roles": []string{"offline_access", "role1"},
			},
			"resource_access": map[string]interface{}{
				"goscim-service": map[string]interface{}{
					"roles": []string{"role2", "admin"},
				},
			},
		}
	}

	signToken := func(method jwt.SigningMethod, kid string, key interface{}) string {
		tok := jwt.NewWithClaims(method, baseClaims())
		tok.Header["kid"] = kid
		s, err := tok.SignedString(key)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}

	for _, tc := range []struct {
		name  string
		token string
	}{
		{"Keycloak RS256 service-account token", signToken(jwt.SigningMethodRS256, "kc-rs256-sig", rsaSigKey)},
		{"Keycloak ES256 service-account token", signToken(jwt.SigningMethodES256, "kc-es256-sig", ecSigKey)},
		{"Keycloak EdDSA service-account token", signToken(jwt.SigningMethodEdDSA, "kc-eddsa-sig", edPriv)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := newRequest(http.MethodPost, elementsPath, validElement("KC-"+tc.name, 1))
			req.Header.Set("Authorization", "Bearer "+tc.token)
			w := serve(r, req)
			requireStatus(t, w, http.StatusCreated)
			body := decode(t, w)
			// role1 comes from realm_access.roles; role2 comes from resource_access.goscim-service.roles!
			if body["name"] != "KC-"+tc.name || body["description"] == nil {
				t.Errorf("expected both realm_access.roles and resource_access roles to apply, got %v", body)
			}
		})
	}

	// Token signed with the RSA-OAEP encryption key MUST be rejected.
	encTok := signToken(jwt.SigningMethodRS256, "kc-rsa-enc", rsaEncKey)
	badReq := newRequest(http.MethodGet, elementsPath, "")
	badReq.Header.Set("Authorization", "Bearer "+encTok)
	requireSCIMError(t, serve(r, badReq), http.StatusUnauthorized)
}

// TestKeycloakOutboundSCIMProvisioningFlow exercises the RFC 7643/7644 client
// operations used by Keycloak outbound SCIM provisioning plugins (SCIM-SDK /
// scim-for-keycloak): Discovery, case-insensitive schema lookup, externalId
// reconciliation, members.value group filter, ETag/If-Match concurrency,
// POST /.search, and Bulk realm synchronisation.
func TestKeycloakOutboundSCIMProvisioningFlow(t *testing.T) {
	r, _ := newTestServer(t)

	// 1. Discovery including case-insensitive /Schemas/{urn} lookup and canonicalValues.
	w := do(t, r, http.MethodGet, "/scim/v2/Schemas/"+url.PathEscape("URN:IETF:PARAMS:SCIM:SCHEMAS:CORE:2.0:USER"), "")
	requireStatus(t, w, http.StatusOK)
	userSchemaDoc := decode(t, w)
	if userSchemaDoc["id"] != schemaUser {
		t.Fatalf("schema id = %v, want %s", userSchemaDoc["id"], schemaUser)
	}

	// 2. Bulk initial realm export: 2 users + 1 group referencing both via bulkId.
	w = do(t, r, http.MethodPost, "/scim/v2/Bulk", `{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:BulkRequest"],
		"Operations": [
			{
				"method": "POST",
				"path": "/Users",
				"bulkId": "kc-u1",
				"data": {
					"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
					"externalId": "kc-uuid-1111",
					"userName": "alice.kc",
					"active": true,
					"emails": [{"value": "alice@kc.example", "type": "work", "primary": true}]
				}
			},
			{
				"method": "POST",
				"path": "/Users",
				"bulkId": "kc-u2",
				"data": {
					"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
					"externalId": "kc-uuid-2222",
					"userName": "bob.kc",
					"active": true
				}
			},
			{
				"method": "POST",
				"path": "/Groups",
				"bulkId": "kc-g1",
				"data": {
					"schemas": ["urn:ietf:params:scim:schemas:core:2.0:Group"],
					"externalId": "kc-grp-realm-admins",
					"displayName": "realm-admins",
					"members": [
						{"value": "bulkId:kc-u1", "display": "alice.kc"},
						{"value": "bulkId:kc-u2", "display": "bob.kc"}
					]
				}
			}
		]
	}`)
	requireStatus(t, w, http.StatusOK)
	bulkOps := decode(t, w)["Operations"].([]interface{})
	if len(bulkOps) != 3 {
		t.Fatalf("expected 3 bulk operation results, got %d", len(bulkOps))
	}

	// 3. Reconcile user by Keycloak UUID (externalId) and core-URN prefixed userName.
	w = do(t, r, http.MethodGet, usersPath+"?filter="+url.QueryEscape(`externalId eq "kc-uuid-1111"`)+"&attributes=userName,externalId", "")
	requireStatus(t, w, http.StatusOK)
	found := decode(t, w)
	if found["totalResults"] != float64(1) {
		t.Fatalf("externalId filter returned %v, want 1", found["totalResults"])
	}
	alice := found["Resources"].([]interface{})[0].(map[string]interface{})
	aliceID := alice["id"].(string)
	if alice["externalId"] != "kc-uuid-1111" || alice["userName"] != "alice.kc" {
		t.Errorf("projected Alice = %v", alice)
	}

	w = do(t, r, http.MethodGet, usersPath+"?filter="+url.QueryEscape(
		`urn:ietf:params:scim:schemas:core:2.0:User:userName eq "ALICE.KC"`), "")
	requireStatus(t, w, http.StatusOK)
	if decode(t, w)["totalResults"] != float64(1) {
		t.Fatalf("core-URN prefixed userName filter returned %v", decode(t, w)["totalResults"])
	}

	// 4. Query groups containing Alice via dotted members.value filter.
	w = do(t, r, http.MethodGet, "/scim/v2/Groups?filter="+url.QueryEscape(`members.value eq "`+aliceID+`"`), "")
	requireStatus(t, w, http.StatusOK)
	if decode(t, w)["totalResults"] != float64(1) {
		t.Fatalf("members.value filter returned %v, want 1", decode(t, w)["totalResults"])
	}

	// 5. Optimistic concurrency via ETag / If-Match (SCIM-SDK behaviour).
	w = do(t, r, http.MethodGet, usersPath+"/"+aliceID, "")
	requireStatus(t, w, http.StatusOK)
	etag := w.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag header on GET /Users/{id}")
	}

	putReq := newRequest(http.MethodPut, usersPath+"/"+aliceID, `{
		"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
		"externalId": "kc-uuid-1111",
		"userName": "alice.kc",
		"displayName": "Alice Keycloak",
		"active": true
	}`)
	putReq.Header.Set("If-Match", etag)
	requireStatus(t, serve(r, putReq), http.StatusOK)

	// Reusing stale ETag must return 412 Precondition Failed.
	staleReq := newRequest(http.MethodPut, usersPath+"/"+aliceID, `{
		"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
		"userName": "alice.kc"
	}`)
	staleReq.Header.Set("If-Match", etag)
	requireStatus(t, serve(r, staleReq), http.StatusPreconditionFailed)

	// 6. POST /scim/v2/.search across Users and Groups by externalId prefix.
	w = do(t, r, http.MethodPost, "/scim/v2/.search", `{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:SearchRequest"],
		"filter": "externalId sw \"kc-\"",
		"attributes": ["externalId", "userName", "displayName"]
	}`)
	requireStatus(t, w, http.StatusOK)
	if decode(t, w)["totalResults"] != float64(3) {
		t.Fatalf("root /.search by externalId sw \"kc-\" returned %v, want 3", decode(t, w)["totalResults"])
	}
}
