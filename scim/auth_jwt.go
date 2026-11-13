package scim

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// maxJWKSBytes caps how much of a JWKS response is read into memory.
const maxJWKSBytes = 1 << 20 // 1 MiB

// minRSAModulusBits is the minimum RSA key size RFC 7518 3.3 permits for
// RS256, RS384 and RS512.
const minRSAModulusBits = 2048

// JWTAuthenticator validates OAuth 2.0 access tokens as described by RFC 9068,
// with signing keys fetched from the issuer's JWKS endpoint.
type JWTAuthenticator struct {
	keys       *jwksCache
	rolesClaim string
	parser     *jwt.Parser
}

// NewJWTAuthenticator builds a validator for tokens from a single issuer.
//
// issuer and audience are required. A token is only meaningful as an
// authorization decision if the server checks that it was minted by the
// issuer it trusts and intended for this service; skipping either turns any
// token the client can obtain anywhere into a valid credential here.
//
// rolesClaim is optional and names a deployment-specific claim to read in
// addition to the standard ones.
func NewJWTAuthenticator(jwksURL, issuer, audience, rolesClaim string) (*JWTAuthenticator, error) {
	if jwksURL == "" {
		return nil, errors.New("a JWKS URL is required")
	}
	if issuer == "" {
		return nil, errors.New("an expected issuer is required")
	}
	if audience == "" {
		return nil, errors.New("an expected audience is required")
	}

	return &JWTAuthenticator{
		keys: &jwksCache{
			url:        jwksURL,
			client:     &http.Client{Timeout: 10 * time.Second},
			minRefresh: time.Minute,
			keys:       map[string]*rsa.PublicKey{},
		},
		rolesClaim: rolesClaim,
		parser: jwt.NewParser(
			// Only asymmetric signatures: allowing HS256 here would let anyone
			// holding the public key mint tokens, and "none" needs no key at all.
			jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}),
			jwt.WithIssuer(issuer),
			jwt.WithAudience(audience),
			jwt.WithExpirationRequired(),
		),
	}, nil
}

// NewJWTAuthenticatorFromEnv reads the JWT configuration from the environment.
func NewJWTAuthenticatorFromEnv() (*JWTAuthenticator, error) {
	authenticator, err := NewJWTAuthenticator(
		os.Getenv("SCIM_JWT_JWKS_URL"),
		os.Getenv("SCIM_JWT_ISSUER"),
		os.Getenv("SCIM_JWT_AUDIENCE"),
		os.Getenv("SCIM_JWT_ROLES_CLAIM"),
	)
	if err != nil {
		return nil, fmt.Errorf(`SCIM_AUTH="jwt": %w (set SCIM_JWT_JWKS_URL, SCIM_JWT_ISSUER and SCIM_JWT_AUDIENCE)`, err)
	}
	return authenticator, nil
}

// Challenge implements Authenticator.
func (a *JWTAuthenticator) Challenge() string { return `Bearer realm="SCIM"` }

// AuthenticationSchemes implements Authenticator.
func (a *JWTAuthenticator) AuthenticationSchemes() []AuthenticationScheme {
	return []AuthenticationScheme{{
		Type:        "oauthbearertoken",
		Name:        "OAuth Bearer Token",
		Description: "Authentication scheme using the OAuth Bearer Token Standard",
		SpecURI:     "http://www.rfc-editor.org/info/rfc6750",
		Primary:     true,
	}}
}

// Authenticate implements Authenticator.
func (a *JWTAuthenticator) Authenticate(r *http.Request) (*Principal, error) {
	const prefix = "bearer "
	header := r.Header.Get("Authorization")
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return nil, ErrUnauthenticated
	}

	claims := jwt.MapClaims{}
	if _, err := a.parser.ParseWithClaims(strings.TrimSpace(header[len(prefix):]), claims, a.keyfunc); err != nil {
		return nil, ErrUnauthenticated
	}

	subject, _ := claims["sub"].(string)
	return &Principal{Subject: subject, Roles: rolesFromClaims(claims, a.rolesClaim)}, nil
}

func (a *JWTAuthenticator) keyfunc(token *jwt.Token) (interface{}, error) {
	kid, _ := token.Header["kid"].(string)
	return a.keys.key(kid)
}

// jwksCache holds the issuer's signing keys, refetching them when a token
// arrives signed by a key it has not seen.
type jwksCache struct {
	url        string
	client     *http.Client
	minRefresh time.Duration

	refreshMu   sync.Mutex
	mu          sync.RWMutex
	keys        map[string]*rsa.PublicKey
	lastAttempt time.Time
}

// key returns the signing key with the given id, refreshing the cache once if
// it is unknown.
func (c *jwksCache) key(kid string) (*rsa.PublicKey, error) {
	if key := c.cached(kid); key != nil {
		return key, nil
	}
	// An unknown id usually means the issuer rotated its keys. Refetching is
	// rate limited so that a stream of tokens naming keys that do not exist
	// cannot be turned into a flood of requests at the issuer.
	if err := c.refresh(); err != nil {
		return nil, err
	}
	if key := c.cached(kid); key != nil {
		return key, nil
	}
	return nil, fmt.Errorf("no signing key %q in the JWKS", kid)
}

func (c *jwksCache) cached(kid string) *rsa.PublicKey {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if kid == "" && len(c.keys) == 1 {
		// A JWKS with a single key is commonly published without an id, and
		// tokens from it carry no kid either.
		for _, key := range c.keys {
			return key
		}
	}
	return c.keys[kid]
}

func (c *jwksCache) refresh() error {
	// Serialise refreshes so a burst of tokens carrying an unknown kid does not
	// turn into an equal burst of concurrent outbound fetches, and record the
	// attempt before calling out so a failing JWKS endpoint is rate limited too.
	c.refreshMu.Lock()
	defer c.refreshMu.Unlock()

	c.mu.RLock()
	attemptedRecently := time.Since(c.lastAttempt) < c.minRefresh
	c.mu.RUnlock()
	if attemptedRecently {
		return nil
	}
	c.mu.Lock()
	c.lastAttempt = time.Now()
	c.mu.Unlock()

	response, err := c.client.Get(c.url)
	if err != nil {
		return fmt.Errorf("fetching the JWKS: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("fetching the JWKS: the issuer answered %s", response.Status)
	}

	var document struct {
		Keys []struct {
			Kty    string   `json:"kty"`
			Use    string   `json:"use"`
			KeyOps []string `json:"key_ops"`
			Kid    string   `json:"kid"`
			N      string   `json:"n"`
			E      string   `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, maxJWKSBytes)).Decode(&document); err != nil {
		return fmt.Errorf("parsing the JWKS: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey)
	for _, entry := range document.Keys {
		// Only RSA signing keys are understood, matching the algorithms the
		// parser accepts. Encryption keys ("use": "enc") and keys restricted to
		// non-verification operations must never verify tokens (RFC 7517 4.2-4.3).
		if entry.Kty != "RSA" {
			continue
		}
		if entry.Use != "" && entry.Use != "sig" {
			continue
		}
		if len(entry.KeyOps) > 0 && !containsOp(entry.KeyOps, "verify") {
			continue
		}
		key, err := rsaPublicKey(entry.N, entry.E)
		if err != nil {
			continue
		}
		keys[entry.Kid] = key
	}
	if len(keys) == 0 {
		return errors.New("the JWKS contains no usable RSA signing key")
	}

	c.mu.Lock()
	c.keys = keys
	c.mu.Unlock()
	return nil
}

func containsOp(ops []string, wanted string) bool {
	for _, op := range ops {
		if op == wanted {
			return true
		}
	}
	return false
}

// rsaPublicKey rebuilds a public key from the base64url modulus and exponent
// of a JWK (RFC 7518 6.3.1), rejecting keys weaker than RFC 7518 3.3 permits.
func rsaPublicKey(modulus, exponent string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(modulus, "="))
	if err != nil {
		return nil, fmt.Errorf("decoding the modulus: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(exponent, "="))
	if err != nil {
		return nil, fmt.Errorf("decoding the exponent: %w", err)
	}
	eBytes = bytes.TrimLeft(eBytes, "\x00")
	if len(nBytes) == 0 || len(eBytes) == 0 || len(eBytes) > 4 {
		return nil, errors.New("malformed RSA key material")
	}

	value := 0
	for _, b := range eBytes {
		value = value<<8 | int(b)
	}
	if value < 3 || value%2 == 0 {
		return nil, errors.New("malformed RSA exponent")
	}
	n := new(big.Int).SetBytes(nBytes)
	if n.BitLen() < minRSAModulusBits {
		return nil, errors.New("RSA modulus shorter than 2048 bits")
	}
	return &rsa.PublicKey{N: n, E: value}, nil
}
