package scim

import (
	"net/http"
	"strings"
	"testing"
)

func TestPasswordModes(t *testing.T) {
	t.Run("default argon2id hashes on POST, preserves on PUT without password, and rehashes on PATCH", func(t *testing.T) {
		r, store := newTestServer(t)

		w := do(t, r, http.MethodPost, usersPath,
			`{"schemas":["`+schemaUser+`"],"userName":"alice","password":"initial-password"}`)
		requireStatus(t, w, http.StatusCreated)
		id := decode(t, w)["id"].(string)

		stored, _ := store.Get("User", id)
		firstHash, _ := stored["password"].(string)
		if !strings.HasPrefix(firstHash, "$argon2id$v=19$") {
			t.Fatalf("password = %q, want argon2id PHC string", firstHash)
		}
		if ok, err := VerifyPasswordHash(firstHash, "initial-password"); err != nil || !ok {
			t.Fatalf("VerifyPasswordHash failed: %v, %v", ok, err)
		}

		// Ordinary read-modify-write PUT omits password; stored hash must survive untouched.
		w = do(t, r, http.MethodPut, usersPath+"/"+id,
			`{"schemas":["`+schemaUser+`"],"userName":"alice","displayName":"Alice"}`)
		requireStatus(t, w, http.StatusOK)

		stored, _ = store.Get("User", id)
		if stored["password"] != firstHash {
			t.Fatalf("PUT without password altered or dropped the hash: got %v, want %q", stored["password"], firstHash)
		}

		// PATCH on an unrelated attribute must not re-hash the stored hash.
		w = do(t, r, http.MethodPatch, usersPath+"/"+id,
			patchBody(map[string]interface{}{"op": "replace", "path": "displayName", "value": "Alice Smith"}))
		requireStatus(t, w, http.StatusOK)

		stored, _ = store.Get("User", id)
		if stored["password"] != firstHash {
			t.Fatalf("unrelated PATCH altered the hash: got %v, want %q", stored["password"], firstHash)
		}

		// PATCH replacing password hashes the new value.
		w = do(t, r, http.MethodPatch, usersPath+"/"+id,
			patchBody(map[string]interface{}{"op": "replace", "path": "PASSWORD", "value": "rotated-password"}))
		requireStatus(t, w, http.StatusOK)

		stored, _ = store.Get("User", id)
		secondHash, _ := stored["password"].(string)
		if secondHash == firstHash {
			t.Fatal("rotating password kept the old hash")
		}
		if ok, err := VerifyPasswordHash(secondHash, "rotated-password"); err != nil || !ok {
			t.Fatalf("rotated password did not verify: %v, %v", ok, err)
		}
		if ok, _ := VerifyPasswordHash(secondHash, "initial-password"); ok {
			t.Fatal("old password still verified after rotation")
		}
	})

	t.Run("bcrypt hashes passwords and refuses >72 bytes with 400 invalidValue", func(t *testing.T) {
		t.Setenv("SCIM_PASSWORD_MODE", "bcrypt")
		r, store := newTestServer(t)

		w := do(t, r, http.MethodPost, usersPath,
			`{"schemas":["`+schemaUser+`"],"userName":"bob","password":"horse-staple-battery"}`)
		requireStatus(t, w, http.StatusCreated)
		id := decode(t, w)["id"].(string)

		stored, _ := store.Get("User", id)
		hash, _ := stored["password"].(string)
		if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
			t.Fatalf("password = %q, want bcrypt hash", hash)
		}
		if ok, err := VerifyPasswordHash(hash, "horse-staple-battery"); err != nil || !ok {
			t.Fatalf("bcrypt hash did not verify: %v, %v", ok, err)
		}

		tooLong := strings.Repeat("x", 73)
		w = do(t, r, http.MethodPost, usersPath,
			`{"schemas":["`+schemaUser+`"],"userName":"bob2","password":"`+tooLong+`"}`)
		out := requireSCIMError(t, w, http.StatusBadRequest)
		if out["scimType"] != "invalidValue" {
			t.Errorf("scimType = %v, want invalidValue", out["scimType"])
		}
	})

	t.Run("reject refuses password writes while allowing passwordless users", func(t *testing.T) {
		t.Setenv("SCIM_PASSWORD_MODE", "reject")
		r, _ := newTestServer(t)

		w := do(t, r, http.MethodPost, usersPath,
			`{"schemas":["`+schemaUser+`"],"userName":"carol","password":"secret"}`)
		out := requireSCIMError(t, w, http.StatusBadRequest)
		if out["scimType"] != "invalidValue" {
			t.Errorf("scimType = %v, want invalidValue", out["scimType"])
		}

		w = do(t, r, http.MethodPost, usersPath,
			`{"schemas":["`+schemaUser+`"],"userName":"carol"}`)
		requireStatus(t, w, http.StatusCreated)
	})

	t.Run("plain stores cleartext when explicitly configured", func(t *testing.T) {
		t.Setenv("SCIM_PASSWORD_MODE", "plain")
		r, store := newTestServer(t)

		w := do(t, r, http.MethodPost, usersPath,
			`{"schemas":["`+schemaUser+`"],"userName":"dave","password":"raw-value"}`)
		requireStatus(t, w, http.StatusCreated)
		id := decode(t, w)["id"].(string)

		stored, _ := store.Get("User", id)
		if stored["password"] != "raw-value" {
			t.Errorf("password = %v, want raw-value", stored["password"])
		}
	})

	t.Run("empty password and unknown SCIM_PASSWORD_MODE are rejected", func(t *testing.T) {
		r, _ := newTestServer(t)
		w := do(t, r, http.MethodPost, usersPath,
			`{"schemas":["`+schemaUser+`"],"userName":"empty","password":""}`)
		out := requireSCIMError(t, w, http.StatusBadRequest)
		if out["scimType"] != "invalidValue" {
			t.Errorf("scimType = %v, want invalidValue", out["scimType"])
		}

		t.Setenv("SCIM_PASSWORD_MODE", "md5")
		if _, err := PasswordModeFromEnv(); err == nil {
			t.Error("PasswordModeFromEnv accepted md5")
		}
	})
}
