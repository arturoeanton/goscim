package scim

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Until now every SCIM endpoint was public. These tests hold the line.
func TestEndpointsRequireAuthentication(t *testing.T) {
	r, _ := newTestServerAs(t, &BasicAuthenticator{})

	cases := []struct {
		method string
		target string
	}{
		{http.MethodPost, elementsPath},
		{http.MethodGet, elementsPath},
		{http.MethodGet, elementsPath + "/some-id"},
		{http.MethodPut, elementsPath + "/some-id"},
		{http.MethodPatch, elementsPath + "/some-id"},
		{http.MethodDelete, elementsPath + "/some-id"},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.target, func(t *testing.T) {
			w := do(t, r, tc.method, tc.target, validElement("Element1", 1))
			requireSCIMError(t, w, http.StatusUnauthorized)
			if w.Header().Get("WWW-Authenticate") == "" {
				t.Error("a 401 must say how to authenticate")
			}
		})
	}
}

func TestBasicAuthentication(t *testing.T) {
	auth := &BasicAuthenticator{users: map[string]basicUser{
		"alice": {password: "correct-horse", roles: []string{"admin", "role1"}},
	}}
	r, _ := newTestServerAs(t, auth)

	t.Run("valid credentials are accepted", func(t *testing.T) {
		req := newRequest(http.MethodPost, elementsPath, validElement("Element1", 1))
		req.SetBasicAuth("alice", "correct-horse")
		w := serve(r, req)
		requireStatus(t, w, http.StatusCreated)
	})

	t.Run("a wrong password is refused", func(t *testing.T) {
		req := newRequest(http.MethodGet, elementsPath, "")
		req.SetBasicAuth("alice", "wrong")
		requireStatus(t, serve(r, req), http.StatusUnauthorized)
	})

	t.Run("an unknown user is refused", func(t *testing.T) {
		req := newRequest(http.MethodGet, elementsPath, "")
		req.SetBasicAuth("mallory", "correct-horse")
		requireStatus(t, serve(r, req), http.StatusUnauthorized)
	})

	t.Run("an empty password is refused", func(t *testing.T) {
		req := newRequest(http.MethodGet, elementsPath, "")
		req.SetBasicAuth("alice", "")
		requireStatus(t, serve(r, req), http.StatusUnauthorized)
	})
}

// The caller's roles decide what the read filter returns, so the token really
// does drive authorization rather than just gating the door.
func TestRolesFromTheCallerDriveTheReadFilter(t *testing.T) {
	// description declares $reader ["role2","role3"].
	t.Run("a caller holding the role sees the attribute", func(t *testing.T) {
		r, _ := newTestServerAs(t, &AnonymousAuthenticator{Roles: []string{"role1", "role2"}})
		created := createElement(t, r, "Element1", 1)
		if created["description"] != "description of Element1" {
			t.Errorf("description = %v", created["description"])
		}
	})

	t.Run("a caller without it does not", func(t *testing.T) {
		r, _ := newTestServerAs(t, &AnonymousAuthenticator{Roles: []string{"role1"}})
		created := createElement(t, r, "Element1", 1)
		if _, present := created["description"]; present {
			t.Errorf("description leaked to a caller without the role: %v", created["description"])
		}
	})

	t.Run("a caller with no roles at all sees neither", func(t *testing.T) {
		r, _ := newTestServerAs(t, &AnonymousAuthenticator{})
		created := createElement(t, r, "Element1", 1)
		// name declares $reader ["role2","role1"], description ["role2","role3"].
		for _, attribute := range []string{"name", "description"} {
			if _, present := created[attribute]; present {
				t.Errorf("%s leaked to a caller with no roles", attribute)
			}
		}
		// Attributes without a $reader stay visible.
		if created["$ref"] != "/Element1" {
			t.Errorf("$ref = %v", created["$ref"])
		}
	})
}

// --- JWT ---------------------------------------------------------------

type testIssuer struct {
	key    *rsa.PrivateKey
	kid    string
	server *httptest.Server
}

// newTestIssuer publishes a JWKS with a freshly generated RSA key.
func newTestIssuer(t *testing.T) *testIssuer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	issuer := &testIssuer{key: key, kid: "test-key"}
	issuer.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		document := map[string]interface{}{"keys": []map[string]string{{
			"kty": "RSA",
			"kid": issuer.kid,
			"n":   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		}}}
		_ = json.NewEncoder(w).Encode(document)
	}))
	t.Cleanup(issuer.server.Close)
	return issuer
}

func (i *testIssuer) sign(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = i.kid
	signed, err := token.SignedString(i.key)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func validClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"iss":   "https://issuer.example",
		"aud":   "scim-api",
		"sub":   "user-123",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"roles": []interface{}{"admin", "role1"},
	}
}

func newJWTServer(t *testing.T, issuer *testIssuer) *gin.Engine {
	t.Helper()
	authenticator, err := NewJWTAuthenticator(issuer.server.URL, "https://issuer.example", "scim-api", "")
	if err != nil {
		t.Fatal(err)
	}
	r, _ := newTestServerAs(t, authenticator)
	return r
}

func TestJWTAuthentication(t *testing.T) {
	issuer := newTestIssuer(t)
	r := newJWTServer(t, issuer)

	t.Run("a valid token is accepted", func(t *testing.T) {
		req := newRequest(http.MethodPost, elementsPath, validElement("Element1", 1))
		req.Header.Set("Authorization", "Bearer "+issuer.sign(t, validClaims()))
		w := serve(r, req)
		requireStatus(t, w, http.StatusCreated)
	})

	rejected := []struct {
		name   string
		mutate func(jwt.MapClaims)
	}{
		{"a token from another issuer", func(c jwt.MapClaims) { c["iss"] = "https://evil.example" }},
		{"a token for another audience", func(c jwt.MapClaims) { c["aud"] = "some-other-api" }},
		{"an expired token", func(c jwt.MapClaims) { c["exp"] = time.Now().Add(-time.Hour).Unix() }},
		{"a token with no expiry", func(c jwt.MapClaims) { delete(c, "exp") }},
		{"a token with no issuer", func(c jwt.MapClaims) { delete(c, "iss") }},
		{"a token with no audience", func(c jwt.MapClaims) { delete(c, "aud") }},
	}
	for _, tc := range rejected {
		t.Run(tc.name+" is refused", func(t *testing.T) {
			claims := validClaims()
			tc.mutate(claims)
			req := newRequest(http.MethodGet, elementsPath, "")
			req.Header.Set("Authorization", "Bearer "+issuer.sign(t, claims))
			requireStatus(t, serve(r, req), http.StatusUnauthorized)
		})
	}

	t.Run("a token signed by someone else is refused", func(t *testing.T) {
		other := newTestIssuer(t)
		other.kid = issuer.kid // same key id, different key
		req := newRequest(http.MethodGet, elementsPath, "")
		req.Header.Set("Authorization", "Bearer "+other.sign(t, validClaims()))
		requireStatus(t, serve(r, req), http.StatusUnauthorized)
	})

	// Algorithm confusion: an HS256 token whose secret is the issuer's public
	// key. Restricting the accepted algorithms to RSA is what stops it.
	t.Run("a symmetric token is refused", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims())
		token.Header["kid"] = issuer.kid
		signed, err := token.SignedString(issuer.key.N.Bytes())
		if err != nil {
			t.Fatal(err)
		}
		req := newRequest(http.MethodGet, elementsPath, "")
		req.Header.Set("Authorization", "Bearer "+signed)
		requireStatus(t, serve(r, req), http.StatusUnauthorized)
	})

	// An algorithm the server does not list, but which verifies against the
	// very same RSA public key. Only the explicit allow-list rejects this one:
	// the key type matches, so nothing else would notice.
	t.Run("an algorithm outside the allow-list is refused", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodPS256, validClaims())
		token.Header["kid"] = issuer.kid
		signed, err := token.SignedString(issuer.key)
		if err != nil {
			t.Fatal(err)
		}
		req := newRequest(http.MethodGet, elementsPath, "")
		req.Header.Set("Authorization", "Bearer "+signed)
		requireStatus(t, serve(r, req), http.StatusUnauthorized)
	})

	t.Run("an unsigned token is refused", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodNone, validClaims())
		signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if err != nil {
			t.Fatal(err)
		}
		req := newRequest(http.MethodGet, elementsPath, "")
		req.Header.Set("Authorization", "Bearer "+signed)
		requireStatus(t, serve(r, req), http.StatusUnauthorized)
	})

	malformed := []string{"", "Bearer", "Bearer ", "Bearer not-a-token", "Basic abc", "Bearer a.b.c"}
	for _, header := range malformed {
		t.Run(fmt.Sprintf("a malformed Authorization header %q is refused", header), func(t *testing.T) {
			req := newRequest(http.MethodGet, elementsPath, "")
			if header != "" {
				req.Header.Set("Authorization", header)
			}
			requireStatus(t, serve(r, req), http.StatusUnauthorized)
		})
	}
}

// The roles in the token are the ones the read filter uses.
func TestJWTRolesReachTheReadFilter(t *testing.T) {
	issuer := newTestIssuer(t)
	r := newJWTServer(t, issuer)

	claims := validClaims()
	claims["roles"] = []interface{}{"role2"} // grants description, not name
	req := newRequest(http.MethodPost, elementsPath, validElement("Element1", 1))
	req.Header.Set("Authorization", "Bearer "+issuer.sign(t, claims))
	w := serve(r, req)
	requireStatus(t, w, http.StatusCreated)

	body := decode(t, w)
	if body["description"] != "description of Element1" {
		t.Errorf("role2 grants description, got %v", body["description"])
	}
	if body["name"] != "Element1" {
		t.Errorf("role2 also grants name, got %v", body["name"])
	}
}

func TestJWTAuthenticatorRequiresIssuerAndAudience(t *testing.T) {
	cases := []struct {
		name                  string
		url, issuer, audience string
	}{
		{"no jwks url", "", "iss", "aud"},
		{"no issuer", "https://x/jwks", "", "aud"},
		{"no audience", "https://x/jwks", "iss", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewJWTAuthenticator(tc.url, tc.issuer, tc.audience, ""); err == nil {
				t.Error("the configuration was accepted")
			}
		})
	}
}

// --- claims ------------------------------------------------------------

// RFC 9068 2.2.3.1 reuses SCIM's own roles/groups/entitlements attributes, and
// those are multi-valued, so all three shapes turn up in real tokens.
func TestRolesFromClaims(t *testing.T) {
	cases := []struct {
		name   string
		claims map[string]interface{}
		extra  string
		want   []string
	}{
		{
			"a list of strings",
			map[string]interface{}{"roles": []interface{}{"admin", "user"}},
			"", []string{"admin", "user"},
		},
		{
			"SCIM multi-valued objects",
			map[string]interface{}{"roles": []interface{}{
				map[string]interface{}{"value": "admin", "type": "direct"},
				map[string]interface{}{"value": "user"},
			}},
			"", []string{"admin", "user"},
		},
		{
			"a space-delimited scope",
			map[string]interface{}{"scope": "scim:read scim:write"},
			"", []string{"scim:read", "scim:write"},
		},
		{
			"roles, groups and entitlements together",
			map[string]interface{}{
				"roles":        []interface{}{"admin"},
				"groups":       []interface{}{"engineering"},
				"entitlements": []interface{}{"seat"},
			},
			"", []string{"admin", "engineering", "seat"},
		},
		{
			"duplicates collapse",
			map[string]interface{}{
				"roles":  []interface{}{"admin", "admin"},
				"groups": []interface{}{"admin"},
			},
			"", []string{"admin"},
		},
		{
			"a deployment-specific claim",
			map[string]interface{}{"https://example.com/roles": []interface{}{"custom"}},
			"https://example.com/roles", []string{"custom"},
		},
		{
			"nothing at all",
			map[string]interface{}{"sub": "user-1"},
			"", []string{},
		},
		{
			"values of unusable shapes are ignored",
			map[string]interface{}{"roles": []interface{}{42, nil, map[string]interface{}{"nope": 1}, "ok"}},
			"", []string{"ok"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := rolesFromClaims(tc.claims, tc.extra)
			if len(got) != len(tc.want) {
				t.Fatalf("roles = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("roles = %v, want %v", got, tc.want)
					break
				}
			}
		})
	}
}

// --- configuration -----------------------------------------------------

// An unset SCIM_AUTH must not quietly mean "no authentication".
func TestAuthenticatorFromEnv(t *testing.T) {
	t.Run("unset is an error", func(t *testing.T) {
		t.Setenv("SCIM_AUTH", "")
		if _, err := NewAuthenticatorFromEnv(); err == nil {
			t.Error("an unset SCIM_AUTH was accepted")
		}
	})

	t.Run("an unknown scheme is an error", func(t *testing.T) {
		t.Setenv("SCIM_AUTH", "magic")
		if _, err := NewAuthenticatorFromEnv(); err == nil {
			t.Error("an unknown scheme was accepted")
		}
	})

	t.Run("none is explicit", func(t *testing.T) {
		t.Setenv("SCIM_AUTH", "none")
		t.Setenv("SCIM_ANONYMOUS_ROLES", "role1, role2")
		authenticator, err := NewAuthenticatorFromEnv()
		if err != nil {
			t.Fatal(err)
		}
		anonymous, ok := authenticator.(*AnonymousAuthenticator)
		if !ok {
			t.Fatalf("authenticator = %T", authenticator)
		}
		if len(anonymous.Roles) != 2 || anonymous.Roles[0] != "role1" || anonymous.Roles[1] != "role2" {
			t.Errorf("roles = %v", anonymous.Roles)
		}
	})

	t.Run("basic needs a user table", func(t *testing.T) {
		t.Setenv("SCIM_AUTH", "basic")
		t.Setenv("SCIM_BASIC_USERS", "")
		if _, err := NewAuthenticatorFromEnv(); err == nil {
			t.Error("basic without users was accepted")
		}

		t.Setenv("SCIM_BASIC_USERS", "alice:secret:admin,role1;bob:hunter2")
		authenticator, err := NewAuthenticatorFromEnv()
		if err != nil {
			t.Fatal(err)
		}
		basic := authenticator.(*BasicAuthenticator)
		if len(basic.users) != 2 {
			t.Fatalf("users = %v", basic.users)
		}
		if got := basic.users["alice"].roles; len(got) != 2 || got[0] != "admin" {
			t.Errorf("alice roles = %v", got)
		}
		if len(basic.users["bob"].roles) != 0 {
			t.Errorf("bob should have no roles: %v", basic.users["bob"].roles)
		}
	})

	t.Run("jwt needs its issuer settings", func(t *testing.T) {
		t.Setenv("SCIM_AUTH", "jwt")
		t.Setenv("SCIM_JWT_JWKS_URL", "")
		if _, err := NewAuthenticatorFromEnv(); err == nil {
			t.Error("jwt without a JWKS URL was accepted")
		}

		t.Setenv("SCIM_JWT_JWKS_URL", "https://issuer.example/jwks")
		t.Setenv("SCIM_JWT_ISSUER", "https://issuer.example")
		t.Setenv("SCIM_JWT_AUDIENCE", "scim-api")
		if _, err := NewAuthenticatorFromEnv(); err != nil {
			t.Errorf("a complete jwt configuration was refused: %v", err)
		}
	})
}

// NewRouter must not be talked into building an unauthenticated server by
// omission.
func TestRouterRequiresAnAuthenticator(t *testing.T) {
	DB = NewMemoryStore()
	if _, err := NewRouter("../config", nil, nil); err == nil {
		t.Error("NewRouter accepted a nil Authenticator")
	}
}

// The discovery endpoints are allowed to be anonymous (RFC 7644 2).
func TestDiscoveryStaysUnauthenticated(t *testing.T) {
	r, _ := newTestServerAs(t, &BasicAuthenticator{})
	for _, path := range []string{"/ServiceProviderConfig", "/ResourceTypes", "/Schemas"} {
		w := do(t, r, http.MethodGet, path, "")
		if w.Code == http.StatusUnauthorized {
			t.Errorf("%s should not require authentication", path)
		}
	}
}
