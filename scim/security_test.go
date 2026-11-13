package scim

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestCaseInsensitiveWriteAccessCannotBeBypassed verifies that sending
// attribute names or extension URNs in a different case cannot slip past
// $writer, readOnly or immutable enforcement before ValidateSchemas renames the
// keys to their canonical spelling.
func TestCaseInsensitiveWriteAccessCannotBeBypassed(t *testing.T) {
	t.Run("$writer on a core attribute", func(t *testing.T) {
		r, store := newTestServerAs(t, &AnonymousAuthenticator{Roles: []string{"user"}})
		tightenWriterOnDescription(t, "privileged")

		body := `{"schemas":["` + schemaCore + `","` + schemaExt + `"],
		          "name":"Element1","DESCRIPTION":"smuggled","$ref":"/Element1",
		          "` + schemaExt + `":{"required":1}}`
		w := do(t, r, http.MethodPost, elementsPath, body)
		requireSCIMError(t, w, http.StatusForbidden)

		total, _, _ := store.Search(SearchQuery{Bucket: "Element", Limit: 10})
		if total != 0 {
			t.Fatalf("bypass persisted %d resources", total)
		}
	})

	t.Run("$writer inside an extension with an uppercase URN", func(t *testing.T) {
		r, _ := newTestServerAs(t, &AnonymousAuthenticator{Roles: []string{"user"}})
		schema := Schemas[schemaExt]
		attrs := append([]Attribute(nil), schema.Attributes...)
		for i := range attrs {
			if attrs[i].Name == "numberInteger" {
				attrs[i].Writer = strs("privileged")
			}
		}
		schema.Attributes = attrs
		Schemas[schemaExt] = schema

		upperExt := strings.ToUpper(schemaExt)
		body := `{"schemas":["` + schemaCore + `","` + upperExt + `"],
		          "name":"Element1","$ref":"/Element1",
		          "` + upperExt + `":{"required":1,"NUMBERINTEGER":99}}`
		w := do(t, r, http.MethodPost, elementsPath, body)
		requireSCIMError(t, w, http.StatusForbidden)
	})

	t.Run("readOnly attribute in uppercase is still stripped", func(t *testing.T) {
		r, store := newTestServer(t)
		body := `{"schemas":["` + schemaUser + `"],"userName":"jane","GROUPS":[{"value":"admins"}]}`
		w := do(t, r, http.MethodPost, usersPath, body)
		requireStatus(t, w, http.StatusCreated)

		stored, _ := store.Get("User", decode(t, w)["id"].(string))
		if _, present := stored["groups"]; present {
			t.Errorf("uppercase GROUPS bypassed readOnly: %v", stored)
		}
		if _, present := stored["GROUPS"]; present {
			t.Errorf("uppercase GROUPS survived: %v", stored)
		}
	})

	t.Run("duplicate case-colliding keys are rejected", func(t *testing.T) {
		r, _ := newTestServer(t)
		body := `{"schemas":["` + schemaUser + `"],"userName":"jane","groups":[{"value":"a"}],"GROUPS":[{"value":"b"}]}`
		w := do(t, r, http.MethodPost, usersPath, body)
		out := requireSCIMError(t, w, http.StatusBadRequest)
		if out["scimType"] != "invalidSyntax" {
			t.Errorf("scimType = %v, want invalidSyntax", out["scimType"])
		}
	})

	t.Run("immutable attribute in uppercase cannot be changed", func(t *testing.T) {
		newTestServer(t)
		rt := writerSchemas(t)
		c, w := testContext(nil)
		existing := map[string]interface{}{"identifier": "abc"}
		if EnforceWriteAccess(c, rt, map[string]interface{}{"IDENTIFIER": "different"}, existing) {
			t.Fatal("uppercase IDENTIFIER bypassed immutability")
		}
		requireStatus(t, w, http.StatusBadRequest)
	})
}

// TestImmutableAndWriterRulesHoldOnRemovalAndUpdate covers the update and
// removal paths: removing an immutable attribute or a $writer-protected
// attribute must be refused, while patching an unrelated attribute on a
// resource that carries a $writer-protected attribute must still succeed.
func TestImmutableAndWriterRulesHoldOnRemovalAndUpdate(t *testing.T) {
	t.Run("immutable attributes cannot be removed to be rewritten later", func(t *testing.T) {
		newTestServer(t)
		rt := writerSchemas(t)
		c, w := testContext(nil)
		existing := map[string]interface{}{"identifier": "abc"}
		if EnforceWriteAccess(c, rt, map[string]interface{}{"open": "x"}, existing) {
			t.Fatal("omitting an immutable attribute was allowed to wipe it")
		}
		requireStatus(t, w, http.StatusBadRequest)
		if decode(t, w)["scimType"] != "mutability" {
			t.Errorf("scimType = %v, want mutability", decode(t, w)["scimType"])
		}
	})

	t.Run("PATCH remove on a $writer-restricted attribute is 403", func(t *testing.T) {
		adminRouter, store := newTestServerAs(t, &AnonymousAuthenticator{Roles: []string{"privileged", "role2"}})
		tightenWriterOnDescription(t, "privileged")
		created := createElement(t, adminRouter, "Element1", 1)
		id := created["id"].(string)

		// Rebuild router over the same store as an unprivileged caller.
		userRouter := ginRouterWithStore(t, store, &AnonymousAuthenticator{Roles: []string{"role1"}})
		tightenWriterOnDescription(t, "privileged")

		w := do(t, userRouter, http.MethodPatch, elementsPath+"/"+id,
			patchBody(map[string]interface{}{"op": "remove", "path": "description"}))
		requireSCIMError(t, w, http.StatusForbidden)

		stored, _ := store.Get("Element", id)
		if stored["description"] != "description of Element1" {
			t.Errorf("restricted attribute was removed: %v", stored)
		}
	})

	t.Run("PATCH on an unrestricted attribute preserves untouched $writer attributes", func(t *testing.T) {
		adminRouter, store := newTestServerAs(t, &AnonymousAuthenticator{Roles: []string{"privileged", "role2"}})
		tightenWriterOnDescription(t, "privileged")
		created := createElement(t, adminRouter, "Element1", 1)
		id := created["id"].(string)

		userRouter := ginRouterWithStore(t, store, &AnonymousAuthenticator{Roles: []string{"role1"}})
		tightenWriterOnDescription(t, "privileged")

		w := do(t, userRouter, http.MethodPatch, elementsPath+"/"+id,
			patchBody(map[string]interface{}{"op": "replace", "path": "name", "value": "Element1-renamed"}))
		requireStatus(t, w, http.StatusOK)

		stored, _ := store.Get("Element", id)
		if stored["name"] != "Element1-renamed" {
			t.Errorf("name = %v", stored["name"])
		}
		if stored["description"] != "description of Element1" {
			t.Errorf("untouched restricted attribute was lost: %v", stored["description"])
		}
	})

	t.Run("PUT by non-reader preserves hidden $writer attributes and gives no value oracle", func(t *testing.T) {
		adminRouter, store := newTestServerAs(t, &AnonymousAuthenticator{Roles: []string{"privileged", "role2"}})
		tightenWriterOnDescription(t, "privileged")
		created := createElement(t, adminRouter, "Element1", 1)
		id := created["id"].(string)

		// role1 cannot read description ($reader: role2, role3) and cannot write it ($writer: privileged).
		userRouter := ginRouterWithStore(t, store, &AnonymousAuthenticator{Roles: []string{"role1"}})
		tightenWriterOnDescription(t, "privileged")

		// Even guessing the exact stored value is 403, so there is no oracle.
		guessBody := `{"schemas":["` + schemaCore + `","` + schemaExt + `"],
		               "name":"Element1","description":"description of Element1","$ref":"/Element1",
		               "` + schemaExt + `":{"required":1}}`
		requireSCIMError(t, do(t, userRouter, http.MethodPut, elementsPath+"/"+id, guessBody), http.StatusForbidden)

		// A normal read-modify-write PUT (omitting the hidden attribute) updates
		// the writable fields while keeping the hidden attribute intact.
		cleanBody := `{"schemas":["` + schemaCore + `","` + schemaExt + `"],
		               "name":"Element1-updated","$ref":"/Element1",
		               "` + schemaExt + `":{"required":2}}`
		requireStatus(t, do(t, userRouter, http.MethodPut, elementsPath+"/"+id, cleanBody), http.StatusOK)

		stored, _ := store.Get("Element", id)
		if stored["description"] != "description of Element1" {
			t.Errorf("PUT wiped a hidden restricted attribute: %v", stored)
		}
	})
}

// TestDeleteEnforcesWriterPermissions ensures a caller who is not allowed to
// write a resource's protected attributes cannot delete the resource either.
func TestDeleteEnforcesWriterPermissions(t *testing.T) {
	adminRouter, store := newTestServerAs(t, &AnonymousAuthenticator{Roles: []string{"privileged", "role2"}})
	tightenWriterOnDescription(t, "privileged")
	created := createElement(t, adminRouter, "Element1", 1)
	id := created["id"].(string)

	userRouter := ginRouterWithStore(t, store, &AnonymousAuthenticator{Roles: []string{"role1"}})
	tightenWriterOnDescription(t, "privileged")

	requireSCIMError(t, do(t, userRouter, http.MethodDelete, elementsPath+"/"+id, ""), http.StatusForbidden)
	if _, err := store.Get("Element", id); err != nil {
		t.Fatal("forbidden DELETE removed the resource")
	}

	// Privileged caller can still delete it.
	requireStatus(t, do(t, adminRouter, http.MethodDelete, elementsPath+"/"+id, ""), http.StatusNoContent)
}

// TestFilterAndSortCannotExfiltrateSecrets blocks blind extraction of writeOnly
// secrets (User.password) and $reader-hidden attributes through filter/sortBy.
func TestFilterAndSortCannotExfiltrateSecrets(t *testing.T) {
	r, _ := newTestServerAs(t, &AnonymousAuthenticator{Roles: []string{"role1"}})
	do(t, r, http.MethodPost, usersPath, `{"schemas":["`+schemaUser+`"],"userName":"jane","password":"secret-value"}`)
	createElement(t, r, "Element1", 1)

	t.Run("filtering on password is refused", func(t *testing.T) {
		for _, f := range []string{`password pr`, `password sw "s"`, `PASSWORD eq "secret-value"`} {
			w := do(t, r, http.MethodGet, usersPath+"?filter="+url.QueryEscape(f), "")
			out := requireSCIMError(t, w, http.StatusBadRequest)
			if out["scimType"] != "invalidFilter" {
				t.Errorf("filter %q: scimType = %v, want invalidFilter", f, out["scimType"])
			}
		}
	})

	t.Run("sorting by password is refused", func(t *testing.T) {
		for _, s := range []string{"password", "PASSWORD"} {
			w := do(t, r, http.MethodGet, usersPath+"?sortBy="+s, "")
			out := requireSCIMError(t, w, http.StatusBadRequest)
			if out["scimType"] != "invalidValue" {
				t.Errorf("sortBy %q: scimType = %v, want invalidValue", s, out["scimType"])
			}
		}
	})

	t.Run("filtering on an attribute hidden by $reader is refused", func(t *testing.T) {
		// role1 cannot read Element.description ($reader: role2, role3).
		w := do(t, r, http.MethodGet, elementsPath+"?filter="+url.QueryEscape(`description co "Element1"`), "")
		out := requireSCIMError(t, w, http.StatusBadRequest)
		if out["scimType"] != "invalidFilter" {
			t.Errorf("scimType = %v, want invalidFilter", out["scimType"])
		}
	})
}

// TestJWKSHardening exercises key-use filtering, minimum RSA strength and
// single-flight / rate-limited JWKS refreshes.
func TestJWKSHardening(t *testing.T) {
	t.Run("encryption JWKs (use=enc) are never used to verify tokens", func(t *testing.T) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err)
		}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"keys": []map[string]interface{}{{
					"kty": "RSA",
					"use": "enc",
					"kid": "enc-key",
					"n":   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
					"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
				}},
			})
		}))
		defer srv.Close()

		issuer := &testIssuer{key: key, kid: "enc-key", server: srv}
		r := newJWTServer(t, issuer)

		req := newRequest(http.MethodGet, elementsPath, "")
		req.Header.Set("Authorization", "Bearer "+issuer.sign(t, validClaims()))
		requireStatus(t, serve(r, req), http.StatusUnauthorized)
	})

	t.Run("weak RSA parameters are rejected", func(t *testing.T) {
		smallKey, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			t.Fatal(err)
		}
		n1024 := base64.RawURLEncoding.EncodeToString(smallKey.N.Bytes())
		e65537 := base64.RawURLEncoding.EncodeToString(big.NewInt(65537).Bytes())
		if _, err := rsaPublicKey(n1024, e65537); err == nil {
			t.Error("accepted a 1024-bit RSA modulus")
		}

		bigKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		n2048 := base64.RawURLEncoding.EncodeToString(bigKey.N.Bytes())
		for _, badExp := range []int64{1, 2, 65536} {
			e := base64.RawURLEncoding.EncodeToString(big.NewInt(badExp).Bytes())
			if _, err := rsaPublicKey(n2048, e); err == nil {
				t.Errorf("accepted RSA exponent %d", badExp)
			}
		}
	})

	t.Run("concurrent unknown-kid burst triggers only one JWKS fetch even on error", func(t *testing.T) {
		var fetches int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			atomic.AddInt32(&fetches, 1)
			time.Sleep(20 * time.Millisecond)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		auth, err := NewJWTAuthenticator(srv.URL, "https://issuer.example", "scim-api", "")
		if err != nil {
			t.Fatal(err)
		}

		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = auth.keys.key("unknown-kid")
			}()
		}
		wg.Wait()

		if got := atomic.LoadInt32(&fetches); got != 1 {
			t.Errorf("JWKS fetched %d times, want 1", got)
		}
	})
}

type failingStore struct {
	*MemoryStore
}

func (f *failingStore) Get(string, string) (map[string]interface{}, error) {
	return nil, errors.New("couchbase://10.0.0.42:11210 bucket `User` timeout")
}

// TestInputAndErrorHygiene covers malformed JSON handling, 500 error redaction,
// password log redaction, Location header hardening and 64-bit integer bounds.
func TestInputAndErrorHygiene(t *testing.T) {
	t.Run("malformed JSON and null payloads return 400 invalidSyntax", func(t *testing.T) {
		r, _ := newTestServer(t)
		created := createElement(t, r, "Element1", 1)
		id := created["id"].(string)

		for _, bad := range []string{`{`, `null`, `[1,2,3]`, `"string"`} {
			out := requireSCIMError(t, do(t, r, http.MethodPost, elementsPath, bad), http.StatusBadRequest)
			if out["scimType"] != "invalidSyntax" {
				t.Errorf("POST %q: scimType = %v, want invalidSyntax", bad, out["scimType"])
			}
			out = requireSCIMError(t, do(t, r, http.MethodPut, elementsPath+"/"+id, bad), http.StatusBadRequest)
			if out["scimType"] != "invalidSyntax" {
				t.Errorf("PUT %q: scimType = %v, want invalidSyntax", bad, out["scimType"])
			}
		}
	})

	t.Run("internal store errors do not leak driver details in 500 responses", func(t *testing.T) {
		r, mem := newTestServer(t)
		DB = &failingStore{MemoryStore: mem}

		w := do(t, r, http.MethodGet, usersPath+"/any-id", "")
		out := requireSCIMError(t, w, http.StatusInternalServerError)
		if strings.Contains(out["detail"].(string), "10.0.0.42") || strings.Contains(out["detail"].(string), "couchbase") {
			t.Errorf("500 response leaked internal error: %v", out["detail"])
		}
	})

	t.Run("hidepassword never leaks characters of short or long passwords", func(t *testing.T) {
		for _, pw := range []string{"x", "ab", "admin!"} {
			masked := hidepassword(pw)
			if strings.Contains(masked, pw[:1]) && masked != "[REDACTED]" {
				t.Errorf("hidepassword(%q) leaked password characters: %q", pw, masked)
			}
		}
		if got := hidepassword(""); got != "(empty)" {
			t.Errorf("hidepassword(\"\") = %q, want (empty)", got)
		}
	})

	t.Run("SCIM_PUBLIC_URL and Host sanitisation protect Location", func(t *testing.T) {
		r, _ := newTestServer(t)
		t.Setenv("SCIM_PUBLIC_URL", "https://scim.example.org/")
		w := do(t, r, http.MethodPost, elementsPath, validElement("ElementPub", 1))
		requireStatus(t, w, http.StatusCreated)
		if loc := w.Header().Get("Location"); !strings.HasPrefix(loc, "https://scim.example.org/scim/v2/Elements/") {
			t.Errorf("Location = %q", loc)
		}

		t.Setenv("SCIM_PUBLIC_URL", "")
		req := newRequest(http.MethodPost, elementsPath, validElement("ElementHost", 2))
		req.Host = "attacker@evil.example/redirect"
		w = serve(r, req)
		requireStatus(t, w, http.StatusCreated)
		if loc := w.Header().Get("Location"); strings.Contains(loc, "evil.example") {
			t.Errorf("poisoned Host reached Location: %q", loc)
		}
		if w.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Errorf("X-Content-Type-Options = %q, want nosniff", w.Header().Get("X-Content-Type-Options"))
		}
	})

	t.Run("integers outside int64 range are rejected on every architecture", func(t *testing.T) {
		r, _ := newTestServer(t)
		// 9223372036854775808 == 2^63, which is exact in float64 and saturated
		// int64 on arm64 so float64(int64(v)) == v used to pass.
		body := `{"schemas":["` + schemaCore + `","` + schemaExt + `"],
		          "name":"ElementOverflow","$ref":"/o",
		          "` + schemaExt + `":{"required":9223372036854775808}}`
		requireSCIMError(t, do(t, r, http.MethodPost, elementsPath, body), http.StatusBadRequest)
	})
}

func tightenWriterOnDescription(t *testing.T, role string) {
	t.Helper()
	schema := Schemas[schemaCore]
	attrs := append([]Attribute(nil), schema.Attributes...)
	for i := range attrs {
		if attrs[i].Name == "description" {
			attrs[i].Writer = strs(role)
		}
	}
	schema.Attributes = attrs
	Schemas[schemaCore] = schema
}

func ginRouterWithStore(t *testing.T, store *MemoryStore, auth Authenticator) http.Handler {
	t.Helper()
	DB = store
	r, err := NewRouter("../config", nil, auth)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	return r
}
