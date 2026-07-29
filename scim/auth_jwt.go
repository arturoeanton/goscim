package scim

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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

	mu          sync.RWMutex
	keys        map[string]*rsa.PublicKey
	lastFetched time.Time
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
	c.mu.Lock()
	fetchedRecently := time.Since(c.lastFetched) < c.minRefresh
	c.mu.Unlock()
	if fetchedRecently {
		return nil
	}

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
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(response.Body).Decode(&document); err != nil {
		return fmt.Errorf("parsing the JWKS: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey)
	for _, entry := range document.Keys {
		// Only RSA keys are understood, matching the signing algorithms the
		// parser accepts. Anything else is skipped rather than guessed at.
		if entry.Kty != "RSA" {
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
	c.lastFetched = time.Now()
	c.mu.Unlock()
	return nil
}

// rsaPublicKey rebuilds a public key from the base64url modulus and exponent
// of a JWK (RFC 7518 6.3.1).
func rsaPublicKey(modulus, exponent string) (*rsa.PublicKey, error) {
	n, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(modulus, "="))
	if err != nil {
		return nil, fmt.Errorf("decoding the modulus: %w", err)
	}
	e, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(exponent, "="))
	if err != nil {
		return nil, fmt.Errorf("decoding the exponent: %w", err)
	}
	if len(n) == 0 || len(e) == 0 || len(e) > 8 {
		return nil, errors.New("malformed RSA key material")
	}

	value := 0
	for _, b := range e {
		value = value<<8 | int(b)
	}
	if value <= 0 {
		return nil, errors.New("malformed RSA exponent")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: value}, nil
}
