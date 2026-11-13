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
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func padBytes(b []byte, size int) []byte {
	if len(b) >= size {
		return b
	}
	out := make([]byte, size)
	copy(out[size-len(b):], b)
	return out
}

func ecJWK(kid string, key *ecdsa.PrivateKey, crv string, size int) map[string]interface{} {
	return map[string]interface{}{
		"kty": "EC",
		"use": "sig",
		"kid": kid,
		"crv": crv,
		"x":   base64.RawURLEncoding.EncodeToString(padBytes(key.X.Bytes(), size)),
		"y":   base64.RawURLEncoding.EncodeToString(padBytes(key.Y.Bytes(), size)),
	}
}

func TestJWTECDSAAndEdDSAAuthentication(t *testing.T) {
	p256Key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	p384Key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	p521Key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		doc := map[string]interface{}{
			"keys": []map[string]interface{}{
				ecJWK("ec-256", p256Key, "P-256", 32),
				ecJWK("ec-384", p384Key, "P-384", 48),
				ecJWK("ec-521", p521Key, "P-521", 66),
				{
					"kty": "OKP",
					"use": "sig",
					"kid": "okp-ed25519",
					"crv": "Ed25519",
					"x":   base64.RawURLEncoding.EncodeToString(edPub),
				},
				{
					"kty": "OKP",
					"use": "enc",
					"kid": "okp-x25519",
					"crv": "X25519",
					"x":   base64.RawURLEncoding.EncodeToString(edPub),
				},
				{
					"kty": "RSA",
					"use": "sig",
					"alg": "RS384",
					"kid": "rsa-pinned-384",
					"n":   base64.RawURLEncoding.EncodeToString(rsaKey.N.Bytes()),
					"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(rsaKey.E)).Bytes()),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(doc)
	}))
	t.Cleanup(jwksServer.Close)

	auth, err := NewJWTAuthenticator(jwksServer.URL, "https://issuer.example", "scim-api", "")
	if err != nil {
		t.Fatal(err)
	}
	r, _ := newTestServerAs(t, auth)

	signWith := func(method jwt.SigningMethod, kid string, priv interface{}) string {
		tok := jwt.NewWithClaims(method, validClaims())
		tok.Header["kid"] = kid
		signed, err := tok.SignedString(priv)
		if err != nil {
			t.Fatal(err)
		}
		return signed
	}

	validCases := []struct {
		name   string
		method jwt.SigningMethod
		kid    string
		priv   interface{}
	}{
		{"ES256 with P-256 JWK", jwt.SigningMethodES256, "ec-256", p256Key},
		{"ES384 with P-384 JWK", jwt.SigningMethodES384, "ec-384", p384Key},
		{"ES512 with P-521 JWK", jwt.SigningMethodES512, "ec-521", p521Key},
		{"EdDSA with Ed25519 OKP JWK", jwt.SigningMethodEdDSA, "okp-ed25519", edPriv},
		{"RS384 with alg-pinned RSA JWK", jwt.SigningMethodRS384, "rsa-pinned-384", rsaKey},
	}

	for _, tc := range validCases {
		t.Run(tc.name+" is accepted", func(t *testing.T) {
			req := newRequest(http.MethodGet, elementsPath, "")
			req.Header.Set("Authorization", "Bearer "+signWith(tc.method, tc.kid, tc.priv))
			requireStatus(t, serve(r, req), http.StatusOK)
		})
	}

	t.Run("cross-curve mismatch (ES256 header naming a P-384 JWK) is refused", func(t *testing.T) {
		req := newRequest(http.MethodGet, elementsPath, "")
		req.Header.Set("Authorization", "Bearer "+signWith(jwt.SigningMethodES256, "ec-384", p256Key))
		requireStatus(t, serve(r, req), http.StatusUnauthorized)
	})

	t.Run("JWK alg restriction prevents RS256 token against RS384-pinned JWK", func(t *testing.T) {
		req := newRequest(http.MethodGet, elementsPath, "")
		req.Header.Set("Authorization", "Bearer "+signWith(jwt.SigningMethodRS256, "rsa-pinned-384", rsaKey))
		requireStatus(t, serve(r, req), http.StatusUnauthorized)
	})

	t.Run("X25519 encryption OKP key is never used for EdDSA verification", func(t *testing.T) {
		req := newRequest(http.MethodGet, elementsPath, "")
		req.Header.Set("Authorization", "Bearer "+signWith(jwt.SigningMethodEdDSA, "okp-x25519", edPriv))
		requireStatus(t, serve(r, req), http.StatusUnauthorized)
	})
}

func TestECAndOKPKeyValidationRejectsInvalidKeyMaterial(t *testing.T) {
	// Off-curve P-256 point (x=1, y=1 is not on y^2 = x^3 - 3x + b).
	one32 := base64.RawURLEncoding.EncodeToString(padBytes([]byte{1}, 32))
	if _, err := ecPublicKey("P-256", one32, one32); err == nil {
		t.Error("accepted an off-curve P-256 point")
	}

	// Point at infinity (all zeros).
	zero32 := base64.RawURLEncoding.EncodeToString(make([]byte, 32))
	if _, err := ecPublicKey("P-256", zero32, zero32); err == nil {
		t.Error("accepted an all-zero EC point")
	}

	// Truncated coordinate length.
	short := base64.RawURLEncoding.EncodeToString([]byte{1, 2, 3})
	if _, err := ecPublicKey("P-256", short, short); err == nil {
		t.Error("accepted truncated EC coordinates")
	}

	// Unsupported EC curve.
	if _, err := ecPublicKey("secp256k1", one32, one32); err == nil {
		t.Error("accepted unsupported EC curve secp256k1")
	}

	// Unsupported OKP curve and bad key length.
	if _, err := okpPublicKey("X25519", zero32); err == nil {
		t.Error("accepted X25519 as a signing OKP key")
	}
	if _, err := okpPublicKey("Ed25519", short); err == nil {
		t.Error("accepted Ed25519 key of wrong length")
	}
}
