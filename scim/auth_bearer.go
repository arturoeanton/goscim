package scim

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// BearerTokenEntry binds one static opaque bearer token to a Principal.
type BearerTokenEntry struct {
	Subject string
	Token   string
	Roles   []string
}

// BearerAuthenticator validates static opaque Bearer tokens (RFC 6750) for
// identity providers whose SCIM 2.0 connector authenticates with a shared
// secret token rather than OAuth 2.0 JWTs (such as Microsoft Entra ID's
// "Secret Token" credential mode and Okta's "API Token" header mode).
type BearerAuthenticator struct {
	entries []BearerTokenEntry
}

// NewBearerAuthenticator constructs a BearerAuthenticator from pre-validated entries.
func NewBearerAuthenticator(entries []BearerTokenEntry) (*BearerAuthenticator, error) {
	cleaned := make([]BearerTokenEntry, 0, len(entries))
	for _, e := range entries {
		sub := strings.TrimSpace(e.Subject)
		if sub == "" {
			sub = "bearer-client"
		}
		if e.Token == "" {
			return nil, errors.New("bearer token must not be empty")
		}
		cleaned = append(cleaned, BearerTokenEntry{
			Subject: sub,
			Token:   e.Token,
			Roles:   e.Roles,
		})
	}
	if len(cleaned) == 0 {
		return nil, errors.New("at least one bearer token is required")
	}
	return &BearerAuthenticator{entries: cleaned}, nil
}

// NewBearerAuthenticatorFromEnv reads SCIM_BEARER_TOKENS
// ("subject:token:role1,role2;other:token2:role3") and/or the single-token
// shorthand SCIM_BEARER_TOKEN (+ optional SCIM_BEARER_SUBJECT and SCIM_BEARER_ROLES).
func NewBearerAuthenticatorFromEnv() (*BearerAuthenticator, error) {
	var entries []BearerTokenEntry

	if rawList := strings.TrimSpace(os.Getenv("SCIM_BEARER_TOKENS")); rawList != "" {
		for _, item := range strings.Split(rawList, ";") {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			fields := strings.SplitN(item, ":", 3)
			if len(fields) < 2 || strings.TrimSpace(fields[0]) == "" || fields[1] == "" {
				return nil, fmt.Errorf("SCIM_BEARER_TOKENS entry %q is not \"subject:token[:roles]\"", item)
			}
			entry := BearerTokenEntry{
				Subject: strings.TrimSpace(fields[0]),
				Token:   fields[1],
			}
			if len(fields) == 3 {
				entry.Roles = splitAndTrim(fields[2])
			}
			entries = append(entries, entry)
		}
	}

	if singleToken := os.Getenv("SCIM_BEARER_TOKEN"); singleToken != "" {
		subject := strings.TrimSpace(os.Getenv("SCIM_BEARER_SUBJECT"))
		if subject == "" {
			subject = "bearer-client"
		}
		entries = append(entries, BearerTokenEntry{
			Subject: subject,
			Token:   singleToken,
			Roles:   splitAndTrim(os.Getenv("SCIM_BEARER_ROLES")),
		})
	}

	if len(entries) == 0 {
		return nil, errors.New(`SCIM_AUTH="bearer" needs SCIM_BEARER_TOKEN or SCIM_BEARER_TOKENS`)
	}
	return NewBearerAuthenticator(entries)
}

// Challenge implements Authenticator.
func (a *BearerAuthenticator) Challenge() string { return `Bearer realm="SCIM"` }

// AuthenticationSchemes implements Authenticator.
func (a *BearerAuthenticator) AuthenticationSchemes() []AuthenticationScheme {
	return []AuthenticationScheme{{
		Type:        "oauthbearertoken",
		Name:        "OAuth Bearer Token",
		Description: "Authentication scheme using the OAuth Bearer Token Standard",
		SpecURI:     "http://www.rfc-editor.org/info/rfc6750",
		Primary:     true,
	}}
}

// Authenticate implements Authenticator, checking every configured token in
// constant time without early exit so response latency never leaks which token
// matched or how many tokens are configured.
func (a *BearerAuthenticator) Authenticate(r *http.Request) (*Principal, error) {
	const prefix = "bearer "
	header := r.Header.Get("Authorization")
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return nil, ErrUnauthenticated
	}
	presented := strings.TrimSpace(header[len(prefix):])
	if presented == "" || len(a.entries) == 0 {
		return nil, ErrUnauthenticated
	}

	matchedIdx := -1
	for i, entry := range a.entries {
		if constantTimeEqual(presented, entry.Token) && matchedIdx == -1 {
			matchedIdx = i
		}
	}
	if matchedIdx < 0 {
		return nil, ErrUnauthenticated
	}
	matched := a.entries[matchedIdx]
	return &Principal{Subject: matched.Subject, Roles: matched.Roles}, nil
}
