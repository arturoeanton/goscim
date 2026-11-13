package scim

import (
	"bytes"
	"crypto"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
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

// acceptedJWTAlgorithms is the strict allow-list of asymmetric JWS algorithms
// accepted for OAuth 2.0 access tokens: RSA PKCS#1 v1.5, NIST-curve ECDSA, and
// Ed25519 EdDSA. Symmetric algorithms (HS*) and "none" are never allowed.
var acceptedJWTAlgorithms = []string{
	"RS256", "RS384", "RS512",
	"ES256", "ES384", "ES512",
	"EdDSA",
}

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
			keys:       map[string]*jwkEntry{},
		},
		rolesClaim: rolesClaim,
		parser: jwt.NewParser(
			jwt.WithValidMethods(acceptedJWTAlgorithms),
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
	entry, err := a.keys.key(kid)
	if err != nil {
		return nil, err
	}
	alg := token.Method.Alg()
	if !keyMatchesAlg(entry.publicKey, alg, entry.alg) {
		return nil, fmt.Errorf("signing key %q does not match algorithm %s", kid, alg)
	}
	return entry.publicKey, nil
}

type jwkEntry struct {
	publicKey crypto.PublicKey
	alg       string
}

type rawJWK struct {
	Kty    string   `json:"kty"`
	Use    string   `json:"use"`
	KeyOps []string `json:"key_ops"`
	Alg    string   `json:"alg"`
	Kid    string   `json:"kid"`
	Crv    string   `json:"crv"`
	N      string   `json:"n"`
	E      string   `json:"e"`
	X      string   `json:"x"`
	Y      string   `json:"y"`
}

// jwksCache holds the issuer's signing keys, refetching them when a token
// arrives signed by a key it has not seen.
type jwksCache struct {
	url        string
	client     *http.Client
	minRefresh time.Duration

	refreshMu   sync.Mutex
	mu          sync.RWMutex
	keys        map[string]*jwkEntry
	singleKey   *jwkEntry
	lastAttempt time.Time
}

// key returns the signing key with the given id, refreshing the cache once if
// it is unknown.
func (c *jwksCache) key(kid string) (*jwkEntry, error) {
	if entry := c.cached(kid); entry != nil {
		return entry, nil
	}
	// An unknown id usually means the issuer rotated its keys. Refetching is
	// rate limited so that a stream of tokens naming keys that do not exist
	// cannot be turned into a flood of requests at the issuer.
	if err := c.refresh(); err != nil {
		return nil, err
	}
	if entry := c.cached(kid); entry != nil {
		return entry, nil
	}
	return nil, fmt.Errorf("no signing key %q in the JWKS", kid)
}

func (c *jwksCache) cached(kid string) *jwkEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if kid == "" && c.singleKey != nil {
		// A JWKS with a single usable signing key is commonly published without
		// an id, and tokens from it carry no kid either.
		return c.singleKey
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
		Keys []rawJWK `json:"keys"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, maxJWKSBytes)).Decode(&document); err != nil {
		return fmt.Errorf("parsing the JWKS: %w", err)
	}

	keys := make(map[string]*jwkEntry)
	var usable []*jwkEntry
	for _, raw := range document.Keys {
		entry, ok := parseSigningJWK(raw)
		if !ok {
			continue
		}
		keys[raw.Kid] = entry
		usable = append(usable, entry)
	}
	if len(usable) == 0 {
		return errors.New("the JWKS contains no usable signing key")
	}

	var single *jwkEntry
	if len(usable) == 1 {
		single = usable[0]
	}

	c.mu.Lock()
	c.keys = keys
	c.singleKey = single
	c.mu.Unlock()
	return nil
}

func parseSigningJWK(raw rawJWK) (*jwkEntry, bool) {
	// Encryption keys ("use": "enc") and keys restricted to non-verification
	// operations must never verify tokens (RFC 7517 4.2-4.3).
	if raw.Use != "" && raw.Use != "sig" {
		return nil, false
	}
	if len(raw.KeyOps) > 0 && !containsOp(raw.KeyOps, "verify") {
		return nil, false
	}
	if raw.Alg != "" && !isAcceptedJWTAlgorithm(raw.Alg) {
		return nil, false
	}

	var (
		pub crypto.PublicKey
		err error
	)
	switch raw.Kty {
	case "RSA":
		pub, err = rsaPublicKey(raw.N, raw.E)
	case "EC":
		pub, err = ecPublicKey(raw.Crv, raw.X, raw.Y)
	case "OKP":
		pub, err = okpPublicKey(raw.Crv, raw.X)
	default:
		return nil, false
	}
	if err != nil {
		return nil, false
	}
	if raw.Alg != "" && !keyMatchesAlg(pub, raw.Alg, raw.Alg) {
		return nil, false
	}
	return &jwkEntry{publicKey: pub, alg: raw.Alg}, true
}

func isAcceptedJWTAlgorithm(alg string) bool {
	for _, accepted := range acceptedJWTAlgorithms {
		if alg == accepted {
			return true
		}
	}
	return false
}

func keyMatchesAlg(pub crypto.PublicKey, tokenAlg, jwkAlg string) bool {
	if jwkAlg != "" && jwkAlg != tokenAlg {
		return false
	}
	switch tokenAlg {
	case "RS256", "RS384", "RS512":
		_, ok := pub.(*rsa.PublicKey)
		return ok
	case "ES256":
		ec, ok := pub.(*ecdsa.PublicKey)
		return ok && ec.Curve == elliptic.P256()
	case "ES384":
		ec, ok := pub.(*ecdsa.PublicKey)
		return ok && ec.Curve == elliptic.P384()
	case "ES512":
		ec, ok := pub.(*ecdsa.PublicKey)
		return ok && ec.Curve == elliptic.P521()
	case "EdDSA":
		ed, ok := pub.(ed25519.PublicKey)
		return ok && len(ed) == ed25519.PublicKeySize
	default:
		return false
	}
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

// ecPublicKey rebuilds an ECDSA public key from a JWK (RFC 7518 6.2.1),
// enforcing exact coordinate sizes and full NIST SP 800-56A point validation
// via crypto/ecdh so off-curve or infinity points are rejected.
func ecPublicKey(crv, xCoord, yCoord string) (*ecdsa.PublicKey, error) {
	var (
		stdCurve  elliptic.Curve
		ecdhCurve ecdh.Curve
		coordLen  int
	)
	switch crv {
	case "P-256":
		stdCurve = elliptic.P256()
		ecdhCurve = ecdh.P256()
		coordLen = 32
	case "P-384":
		stdCurve = elliptic.P384()
		ecdhCurve = ecdh.P384()
		coordLen = 48
	case "P-521":
		stdCurve = elliptic.P521()
		ecdhCurve = ecdh.P521()
		coordLen = 66
	default:
		return nil, fmt.Errorf("unsupported EC curve %q", crv)
	}

	xBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(xCoord, "="))
	if err != nil {
		return nil, fmt.Errorf("decoding EC x coordinate: %w", err)
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(yCoord, "="))
	if err != nil {
		return nil, fmt.Errorf("decoding EC y coordinate: %w", err)
	}
	if len(xBytes) != coordLen || len(yBytes) != coordLen {
		return nil, errors.New("EC coordinate length does not match curve")
	}

	uncompressed := make([]byte, 1+2*coordLen)
	uncompressed[0] = 0x04
	copy(uncompressed[1:1+coordLen], xBytes)
	copy(uncompressed[1+coordLen:], yBytes)
	if _, err := ecdhCurve.NewPublicKey(uncompressed); err != nil {
		return nil, fmt.Errorf("invalid EC public key point: %w", err)
	}

	return &ecdsa.PublicKey{
		Curve: stdCurve,
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}

// okpPublicKey rebuilds an Ed25519 public key from an OKP JWK (RFC 8037 2).
func okpPublicKey(crv, xCoord string) (ed25519.PublicKey, error) {
	if crv != "Ed25519" {
		return nil, fmt.Errorf("unsupported OKP curve %q", crv)
	}
	xBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(xCoord, "="))
	if err != nil {
		return nil, fmt.Errorf("decoding Ed25519 public key: %w", err)
	}
	if len(xBytes) != ed25519.PublicKeySize {
		return nil, errors.New("invalid Ed25519 public key length")
	}
	return ed25519.PublicKey(xBytes), nil
}
