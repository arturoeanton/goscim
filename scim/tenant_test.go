package scim

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestMultiTenantURLAndBearerTokenIsolation(t *testing.T) {
	t.Setenv("SCIM_BEARER_TOKENS",
		"acme-prov:tok-acme:admin,superadmin,user,role1@acme;globex-prov:tok-globex:admin,superadmin,user,role1@globex;global-admin:tok-root:admin,superadmin,user,role1")
	auth, err := NewBearerAuthenticatorFromEnv()
	if err != nil {
		t.Fatalf("NewBearerAuthenticatorFromEnv: %v", err)
	}

	r, _ := newTestServerAs(t, auth)

	withToken := func(method, path, token, body string) *httptest.ResponseRecorder {
		req := newRequest(method, path, body)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		return serve(r, req)
	}

	userPayload := `{"schemas":["urn:ietf:params:scim:schemas:core:2.0:User"],"userName":"adele@shared.example","active":true}`

	// 1. Create identical userName in acme, globex, and default root namespace.
	wAcme := withToken(http.MethodPost, "/tenants/acme/scim/v2/Users", "tok-acme", userPayload)
	requireStatus(t, wAcme, http.StatusCreated)
	acmeUser := decode(t, wAcme)
	acmeID := acmeUser["id"].(string)
	if loc := wAcme.Header().Get("Location"); loc == "" || loc[len(loc)-len("/tenants/acme/scim/v2/Users/"+acmeID):] != "/tenants/acme/scim/v2/Users/"+acmeID {
		t.Fatalf("unexpected acme Location header: %q", loc)
	}

	wGlobex := withToken(http.MethodPost, "/tenants/globex/scim/v2/Users", "tok-globex", userPayload)
	requireStatus(t, wGlobex, http.StatusCreated)

	wRoot := withToken(http.MethodPost, "/scim/v2/Users", "tok-root", userPayload)
	requireStatus(t, wRoot, http.StatusCreated)

	// 2. Duplicate inside acme must still fail with 409 uniqueness.
	wDup := withToken(http.MethodPost, "/tenants/acme/scim/v2/Users", "tok-acme", userPayload)
	errObj := requireSCIMError(t, wDup, http.StatusConflict)
	if errObj["scimType"] != "uniqueness" {
		t.Fatalf("scimType = %v, want uniqueness", errObj["scimType"])
	}

	// 3. Cross-tenant token use on /tenants/globex/scim/v2 with tok-acme must fail 403.
	wCross := withToken(http.MethodGet, "/tenants/globex/scim/v2/Users", "tok-acme", "")
	requireStatus(t, wCross, http.StatusForbidden)

	// 4. Tenant-bound token calling root /scim/v2/Users is automatically scoped to acme.
	wAutoScoped := withToken(http.MethodGet, "/scim/v2/Users", "tok-acme", "")
	requireStatus(t, wAutoScoped, http.StatusOK)
	listAcme := decode(t, wAutoScoped)
	if int(listAcme["totalResults"].(float64)) != 1 {
		t.Fatalf("auto-scoped totalResults = %v, want 1", listAcme["totalResults"])
	}
	resources := listAcme["Resources"].([]interface{})
	if resources[0].(map[string]interface{})["id"] != acmeID {
		t.Fatalf("auto-scoped user id mismatch")
	}

	// 5. Group.members -> User.groups sync stays strictly inside tenant acme.
	wGrp := withToken(http.MethodPost, "/tenants/acme/scim/v2/Groups", "tok-acme",
		`{"schemas":["urn:ietf:params:scim:schemas:core:2.0:Group"],"displayName":"AcmeEng","members":[{"value":"`+acmeID+`"}]}`)
	requireStatus(t, wGrp, http.StatusCreated)

	wReadAcme := withToken(http.MethodGet, "/tenants/acme/scim/v2/Users/"+acmeID, "tok-acme", "")
	requireStatus(t, wReadAcme, http.StatusOK)
	groups, _ := decode(t, wReadAcme)["groups"].([]interface{})
	if len(groups) != 1 {
		t.Fatalf("acme User.groups = %v, want 1 group", groups)
	}
	ref, _ := groups[0].(map[string]interface{})["$ref"].(string)
	if ref == "" || ref[:len("/tenants/acme/scim/v2/Groups/")] != "/tenants/acme/scim/v2/Groups/" {
		t.Fatalf("expected tenant-scoped group $ref, got %q", ref)
	}

	// 6. Invalid tenant slug is rejected with 400.
	wBadTenant := withToken(http.MethodGet, "/tenants/bad$slug/scim/v2/Users", "tok-root", "")
	requireStatus(t, wBadTenant, http.StatusBadRequest)
}

func TestMultiTenantJWTClaimIsolation(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	jwksPayload, _ := json.Marshal(map[string]interface{}{
		"keys": []map[string]interface{}{{
			"kty": "RSA",
			"use": "sig",
			"alg": "RS256",
			"kid": "k1",
			"n":   base64.RawURLEncoding.EncodeToString(priv.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(priv.E)).Bytes()),
		}},
	})

	jwtAuth, err := NewJWTAuthenticator("http://jwks.internal/keys", "https://issuer.example", "goscim", "")
	if err != nil {
		t.Fatalf("NewJWTAuthenticator: %v", err)
	}
	jwtAuth.SetTenantClaim("tid")
	jwtAuth.keys.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()
			rec.Header().Set("Content-Type", "application/json")
			_, _ = rec.Write(jwksPayload)
			return rec.Result(), nil
		}),
	}

	gin.SetMode(gin.TestMode)
	r, _ := newTestServerAs(t, jwtAuth)

	mint := func(tid string) string {
		tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"iss":   "https://issuer.example",
			"aud":   "goscim",
			"sub":   "jwt-client",
			"exp":   time.Now().Add(time.Hour).Unix(),
			"roles": []string{"admin", "superadmin", "user"},
			"tid":   tid,
		})
		tok.Header["kid"] = "k1"
		signed, err := tok.SignedString(priv)
		if err != nil {
			t.Fatalf("SignedString: %v", err)
		}
		return signed
	}

	tokAcme := mint("acme")
	reqCreate := newRequest(http.MethodPost, "/scim/v2/Users",
		`{"schemas":["urn:ietf:params:scim:schemas:core:2.0:User"],"userName":"jwt.user@acme.io"}`)
	reqCreate.Header.Set("Authorization", "Bearer "+tokAcme)
	requireStatus(t, serve(r, reqCreate), http.StatusCreated)

	// Root /scim/v2/Users with no tid token (e.g. globex token) sees 0 users.
	tokGlobex := mint("globex")
	reqListGlobex := newRequest(http.MethodGet, "/scim/v2/Users", "")
	reqListGlobex.Header.Set("Authorization", "Bearer "+tokGlobex)
	wGlobex := serve(r, reqListGlobex)
	requireStatus(t, wGlobex, http.StatusOK)
	if int(decode(t, wGlobex)["totalResults"].(float64)) != 0 {
		t.Fatalf("globex JWT saw acme user")
	}

	// Using acme JWT on /tenants/globex/scim/v2/Users must return 403.
	reqCross := newRequest(http.MethodGet, "/tenants/globex/scim/v2/Users", "")
	reqCross.Header.Set("Authorization", "Bearer "+tokAcme)
	requireStatus(t, serve(r, reqCross), http.StatusForbidden)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
