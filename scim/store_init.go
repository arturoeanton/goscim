package scim

import (
	"fmt"
	"os"
	"strings"
)

// StoreBackend identifies the configured persistence driver.
type StoreBackend string

const (
	StoreBackendCouchbase StoreBackend = "couchbase"
	StoreBackendSQLite    StoreBackend = "sqlite"
	StoreBackendMemory    StoreBackend = "memory"
)

// StoreBackendFromEnv resolves SCIM_STORE, defaulting to couchbase when unset
// so existing deployments preserve their behaviour.
func StoreBackendFromEnv() (StoreBackend, error) {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("SCIM_STORE")))
	switch StoreBackend(raw) {
	case "", StoreBackendCouchbase:
		return StoreBackendCouchbase, nil
	case StoreBackendSQLite:
		return StoreBackendSQLite, nil
	case StoreBackendMemory, "mem":
		return StoreBackendMemory, nil
	default:
		return "", fmt.Errorf("unknown SCIM_STORE=%q (expected \"couchbase\", \"sqlite\", or \"memory\")", raw)
	}
}

// InitStoreFromEnv initialises the Store selected by SCIM_STORE and installs it
// in DB.
func InitStoreFromEnv() (StoreBackend, error) {
	backend, err := StoreBackendFromEnv()
	if err != nil {
		return "", err
	}
	switch backend {
	case StoreBackendMemory:
		DB = NewMemoryStore()
		return StoreBackendMemory, nil
	case StoreBackendSQLite:
		path := strings.TrimSpace(os.Getenv("SCIM_SQLITE_PATH"))
		if path == "" {
			path = "goscim.db"
		}
		store, err := NewSQLiteStore(path)
		if err != nil {
			return "", err
		}
		DB = store
		return StoreBackendSQLite, nil
	default:
		if err := InitDB(); err != nil {
			return "", err
		}
		return StoreBackendCouchbase, nil
	}
}

// ActiveStoreBackend returns the backend type of DB.
func ActiveStoreBackend() string {
	switch s := DB.(type) {
	case *CouchbaseStore:
		return string(StoreBackendCouchbase)
	case *SQLiteStore:
		return fmt.Sprintf("sqlite (%s)", s.Path())
	case *MemoryStore:
		return string(StoreBackendMemory)
	default:
		if DB == nil {
			return "uninitialised"
		}
		return "custom"
	}
}
