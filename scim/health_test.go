package scim

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

type downStore struct {
	*MemoryStore
}

func (d *downStore) Ping(context.Context) error {
	return errors.New("couchbase://10.0.0.99:8093 query service unreachable")
}

func TestHealthAndReadinessEndpoints(t *testing.T) {
	// Use an authenticator that rejects unauthenticated requests so the test
	// also proves /healthz and /readyz stay reachable without credentials.
	r, mem := newTestServerAs(t, &BasicAuthenticator{users: map[string]basicUser{
		"admin": {password: "secret", roles: []string{"admin"}},
	}})

	t.Run("/healthz and /readyz answer 200 without authentication when store is healthy", func(t *testing.T) {
		w := do(t, r, http.MethodGet, "/healthz", "")
		requireStatus(t, w, http.StatusOK)
		if decode(t, w)["status"] != "ok" {
			t.Errorf("body = %s", w.Body.String())
		}

		w = do(t, r, http.MethodGet, "/readyz", "")
		requireStatus(t, w, http.StatusOK)
		if decode(t, w)["status"] != "ready" {
			t.Errorf("body = %s", w.Body.String())
		}
	})

	t.Run("/readyz answers 503 without leaking internal addresses when store fails", func(t *testing.T) {
		DB = &downStore{MemoryStore: mem}
		defer func() { DB = mem }()

		// Liveness stays 200 so orchestrators do not restart healthy processes.
		requireStatus(t, do(t, r, http.MethodGet, "/healthz", ""), http.StatusOK)

		w := do(t, r, http.MethodGet, "/readyz", "")
		requireStatus(t, w, http.StatusServiceUnavailable)
		if strings.Contains(w.Body.String(), "10.0.0.99") {
			t.Errorf("/readyz leaked internal error: %s", w.Body.String())
		}
	})
}
