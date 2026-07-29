package scim

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// principalContextKey is where the authenticated caller is kept for the
// duration of a request.
const principalContextKey = "scim.principal"

// Principal is the authenticated caller. Roles feed the $reader and $writer
// checks declared on schema attributes.
type Principal struct {
	Subject string
	Roles   []string
}

// HasRole reports whether the principal holds the given role.
func (p *Principal) HasRole(role string) bool {
	if p == nil {
		return false
	}
	for _, held := range p.Roles {
		if held == role {
			return true
		}
	}
	return false
}

// Authenticator turns a request into a Principal. Implementations must not
// write to the response: the middleware owns the error shape so that every
// scheme fails identically from the client's point of view.
type Authenticator interface {
	// Challenge is the WWW-Authenticate value sent with a 401.
	Challenge() string
	Authenticate(r *http.Request) (*Principal, error)
}

// ErrUnauthenticated is what an Authenticator returns when the request carries
// no usable credentials. The message reaching the client is deliberately
// uniform, so a caller cannot tell a bad password from an unknown user.
var ErrUnauthenticated = errors.New("valid credentials are required")

// Authenticate is the gin middleware that runs an Authenticator and publishes
// the resulting Principal for the handlers.
func Authenticate(authenticator Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, err := authenticator.Authenticate(c.Request)
		if err != nil || principal == nil {
			if challenge := authenticator.Challenge(); challenge != "" {
				c.Header("WWW-Authenticate", challenge)
			}
			MakeError(c, http.StatusUnauthorized, ErrUnauthenticated.Error())
			c.Abort()
			return
		}
		c.Set(principalContextKey, principal)
		c.Next()
	}
}

// PrincipalOf returns the authenticated caller, or nil when the request did not
// go through the authentication middleware.
func PrincipalOf(c *gin.Context) *Principal {
	value, exists := c.Get(principalContextKey)
	if !exists {
		return nil
	}
	principal, _ := value.(*Principal)
	return principal
}

// AnonymousAuthenticator accepts every request with a fixed set of roles. It
// exists so that "no authentication" is a deliberate, visible choice rather
// than the accidental default it used to be.
type AnonymousAuthenticator struct {
	Roles []string
}

// Challenge implements Authenticator. There is nothing to challenge for.
func (a *AnonymousAuthenticator) Challenge() string { return "" }

// Authenticate implements Authenticator.
func (a *AnonymousAuthenticator) Authenticate(*http.Request) (*Principal, error) {
	return &Principal{Subject: "anonymous", Roles: a.Roles}, nil
}

// NewAuthenticatorFromEnv builds the authenticator selected by SCIM_AUTH.
//
// There is no default: an unset SCIM_AUTH is an error rather than an open
// server. Running without authentication stays possible, but only by asking
// for it.
func NewAuthenticatorFromEnv() (Authenticator, error) {
	switch scheme := strings.ToLower(strings.TrimSpace(os.Getenv("SCIM_AUTH"))); scheme {
	case "jwt":
		return NewJWTAuthenticatorFromEnv()
	case "basic":
		return NewBasicAuthenticatorFromEnv()
	case "none":
		roles := splitAndTrim(os.Getenv("SCIM_ANONYMOUS_ROLES"))
		return &AnonymousAuthenticator{Roles: roles}, nil
	case "":
		return nil, errors.New(
			"SCIM_AUTH is not set: choose \"jwt\", \"basic\" or, to serve every request unauthenticated, \"none\"")
	default:
		return nil, fmt.Errorf("unknown SCIM_AUTH %q: expected \"jwt\", \"basic\" or \"none\"", scheme)
	}
}

// splitAndTrim parses a comma-separated environment value.
func splitAndTrim(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
