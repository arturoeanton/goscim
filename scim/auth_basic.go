package scim

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// BasicAuthenticator checks HTTP Basic credentials against a static table.
//
// It is meant for development and for the smoke tests an operator runs by
// hand: the credentials live in an environment variable in the clear, and
// there is no lockout or rate limiting. Production deployments want the JWT
// authenticator.
type BasicAuthenticator struct {
	// users maps a username to its password and roles.
	users map[string]basicUser
}

type basicUser struct {
	password string
	roles    []string
}

// NewBasicAuthenticator builds an authenticator from a username to
// password/roles table.
func NewBasicAuthenticator(users map[string][]string, passwords map[string]string) *BasicAuthenticator {
	table := make(map[string]basicUser, len(passwords))
	for name, password := range passwords {
		table[name] = basicUser{password: password, roles: users[name]}
	}
	return &BasicAuthenticator{users: table}
}

// NewBasicAuthenticatorFromEnv reads SCIM_BASIC_USERS, formatted as
// "user:password:role1,role2;other:password:role3".
func NewBasicAuthenticatorFromEnv() (*BasicAuthenticator, error) {
	raw := strings.TrimSpace(os.Getenv("SCIM_BASIC_USERS"))
	if raw == "" {
		return nil, errors.New(`SCIM_AUTH="basic" needs SCIM_BASIC_USERS, formatted as "user:password:role1,role2;other:password:role3"`)
	}

	table := make(map[string]basicUser)
	for _, entry := range strings.Split(raw, ";") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		fields := strings.SplitN(entry, ":", 3)
		if len(fields) < 2 || strings.TrimSpace(fields[0]) == "" || fields[1] == "" {
			return nil, fmt.Errorf("SCIM_BASIC_USERS entry %q is not \"user:password:roles\"", entry)
		}
		user := basicUser{password: fields[1]}
		if len(fields) == 3 {
			user.roles = splitAndTrim(fields[2])
		}
		table[strings.TrimSpace(fields[0])] = user
	}
	if len(table) == 0 {
		return nil, errors.New("SCIM_BASIC_USERS defines no users")
	}
	return &BasicAuthenticator{users: table}, nil
}

// Challenge implements Authenticator.
func (a *BasicAuthenticator) Challenge() string { return `Basic realm="SCIM"` }

// AuthenticationSchemes implements Authenticator.
func (a *BasicAuthenticator) AuthenticationSchemes() []AuthenticationScheme {
	return []AuthenticationScheme{{
		Type:        "httpbasic",
		Name:        "HTTP Basic",
		Description: "Authentication scheme using the HTTP Basic Standard",
		SpecURI:     "http://www.rfc-editor.org/info/rfc2617",
		Primary:     true,
	}}
}

// Authenticate implements Authenticator.
func (a *BasicAuthenticator) Authenticate(r *http.Request) (*Principal, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, ErrUnauthenticated
	}
	user, known := a.users[username]

	// Compare even when the user is unknown, against a value that cannot
	// match, so that the time taken does not reveal which usernames exist.
	expected := user.password
	if !known {
		expected = "\x00unknown-user"
	}
	if !constantTimeEqual(password, expected) || !known {
		return nil, ErrUnauthenticated
	}
	return &Principal{Subject: username, Roles: user.roles}, nil
}

// constantTimeEqual compares two secrets without leaking their length, by
// comparing digests of fixed size.
func constantTimeEqual(a, b string) bool {
	left := sha256.Sum256([]byte(a))
	right := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(left[:], right[:]) == 1
}
