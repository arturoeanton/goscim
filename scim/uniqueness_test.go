package scim

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
)

// The core User schema declares userName with "uniqueness": "server", and
// nothing enforced it. An identity provider retrying a provisioning request
// that timed out would quietly end up with two users of the same name, and
// every later lookup by that name would be ambiguous.
func TestUniqueAttributesAreEnforcedOnCreate(t *testing.T) {
	r, _ := newTestServer(t)

	body := `{"schemas":["` + schemaUser + `"],"userName":"jane.doe"}`
	requireStatus(t, do(t, r, http.MethodPost, usersPath, body), http.StatusCreated)

	// The same request again, as a retry would send it.
	w := do(t, r, http.MethodPost, usersPath, body)
	out := requireSCIMError(t, w, http.StatusConflict)
	if out["scimType"] != "uniqueness" {
		t.Errorf("scimType = %v, want uniqueness", out["scimType"])
	}
}

// A different value is fine, and so is a resource keeping its own.
func TestUniquenessAllowsDistinctValues(t *testing.T) {
	r, _ := newTestServer(t)

	requireStatus(t, do(t, r, http.MethodPost, usersPath,
		`{"schemas":["`+schemaUser+`"],"userName":"jane.doe"}`), http.StatusCreated)
	requireStatus(t, do(t, r, http.MethodPost, usersPath,
		`{"schemas":["`+schemaUser+`"],"userName":"john.doe"}`), http.StatusCreated)
}

func TestUniquenessOnReplace(t *testing.T) {
	r, _ := newTestServer(t)

	w := do(t, r, http.MethodPost, usersPath, `{"schemas":["`+schemaUser+`"],"userName":"jane.doe"}`)
	requireStatus(t, w, http.StatusCreated)
	janeID := decode(t, w)["id"].(string)

	requireStatus(t, do(t, r, http.MethodPost, usersPath,
		`{"schemas":["`+schemaUser+`"],"userName":"john.doe"}`), http.StatusCreated)

	// Jane keeping her own userName must not collide with herself.
	w = do(t, r, http.MethodPut, usersPath+"/"+janeID,
		`{"schemas":["`+schemaUser+`"],"userName":"jane.doe","displayName":"Jane"}`)
	requireStatus(t, w, http.StatusOK)

	// Taking John's is a conflict.
	w = do(t, r, http.MethodPut, usersPath+"/"+janeID,
		`{"schemas":["`+schemaUser+`"],"userName":"john.doe"}`)
	requireSCIMError(t, w, http.StatusConflict)
}

// A patch funnels into replace, so the same rule has to hold.
func TestUniquenessOnPatch(t *testing.T) {
	r, _ := newTestServer(t)

	w := do(t, r, http.MethodPost, usersPath, `{"schemas":["`+schemaUser+`"],"userName":"jane.doe"}`)
	requireStatus(t, w, http.StatusCreated)
	janeID := decode(t, w)["id"].(string)

	requireStatus(t, do(t, r, http.MethodPost, usersPath,
		`{"schemas":["`+schemaUser+`"],"userName":"john.doe"}`), http.StatusCreated)

	patch := patchBody(map[string]interface{}{"op": "replace", "path": "userName", "value": "john.doe"})
	requireSCIMError(t, do(t, r, http.MethodPatch, usersPath+"/"+janeID, patch), http.StatusConflict)
}

// A deleted resource frees its value.
func TestUniquenessReleasedOnDelete(t *testing.T) {
	r, _ := newTestServer(t)

	w := do(t, r, http.MethodPost, usersPath, `{"schemas":["`+schemaUser+`"],"userName":"jane.doe"}`)
	requireStatus(t, w, http.StatusCreated)
	id := decode(t, w)["id"].(string)

	requireStatus(t, do(t, r, http.MethodDelete, usersPath+"/"+id, ""), http.StatusNoContent)
	requireStatus(t, do(t, r, http.MethodPost, usersPath,
		`{"schemas":["`+schemaUser+`"],"userName":"jane.doe"}`), http.StatusCreated)
}

// Uniqueness applies inside extensions too, where the value is stored nested
// under the schema URN.
func TestUniquenessInsideAnExtension(t *testing.T) {
	r, _ := newTestServer(t)

	// The Element extension declares "required" without uniqueness, so tighten
	// it: the handlers read Schemas per request.
	schema := Schemas[schemaExt]
	attributes := append([]Attribute(nil), schema.Attributes...)
	for i := range attributes {
		if attributes[i].Name == "required" {
			attributes[i].Uniqueness = "server"
		}
	}
	schema.Attributes = attributes
	Schemas[schemaExt] = schema

	requireStatus(t, do(t, r, http.MethodPost, elementsPath, validElement("Element1", 7)), http.StatusCreated)

	w := do(t, r, http.MethodPost, elementsPath, validElement("Element2", 7))
	requireSCIMError(t, w, http.StatusConflict)

	// A different value in the extension is accepted.
	requireStatus(t, do(t, r, http.MethodPost, elementsPath, validElement("Element3", 8)), http.StatusCreated)
}

func TestMemoryStoreFindIDByAttribute(t *testing.T) {
	s := NewMemoryStore()
	doc := map[string]interface{}{"id": "1", "userName": "jane", "urn:x": map[string]interface{}{"code": "abc"}}
	if err := s.Upsert("User", "1", doc); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		path  string
		value interface{}
		want  string
	}{
		{"userName", "jane", "1"},
		{"userName", "john", ""},
		{"urn:x.code", "abc", "1"},
		{"missing", "x", ""},
	}
	for _, tc := range cases {
		got, err := s.FindIDByAttribute("User", tc.path, tc.value)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Errorf("FindIDByAttribute(%q, %v) = %q, want %q", tc.path, tc.value, got, tc.want)
		}
	}
}

func TestUniquenessIsCaseInsensitiveWhenCaseExactFalse(t *testing.T) {
	r, _ := newTestServer(t)

	w := do(t, r, http.MethodPost, usersPath, `{"schemas":["`+schemaUser+`"],"userName":"Jane.Doe"}`)
	requireStatus(t, w, http.StatusCreated)
	janeID := decode(t, w)["id"].(string)

	// Different casing of the same caseExact:false userName must collide.
	w = do(t, r, http.MethodPost, usersPath, `{"schemas":["`+schemaUser+`"],"userName":"jane.doe"}`)
	out := requireSCIMError(t, w, http.StatusConflict)
	if out["scimType"] != "uniqueness" {
		t.Errorf("scimType = %v, want uniqueness", out["scimType"])
	}

	// Jane changing the casing of her own userName must succeed.
	w = do(t, r, http.MethodPut, usersPath+"/"+janeID, `{"schemas":["`+schemaUser+`"],"userName":"jane.doe"}`)
	requireStatus(t, w, http.StatusOK)

	// Renaming Jane releases "jane.doe" immediately for another user.
	w = do(t, r, http.MethodPut, usersPath+"/"+janeID, `{"schemas":["`+schemaUser+`"],"userName":"alicia.doe"}`)
	requireStatus(t, w, http.StatusOK)

	requireStatus(t, do(t, r, http.MethodPost, usersPath,
		`{"schemas":["`+schemaUser+`"],"userName":"JANE.DOE"}`), http.StatusCreated)
}

func TestConcurrentUniquenessAtomicUnderRace(t *testing.T) {
	r, _ := newTestServer(t)

	const workers = 30
	spellings := []string{"concurrent.user", "Concurrent.User", "CONCURRENT.USER"}

	var (
		start     = make(chan struct{})
		wg        sync.WaitGroup
		created   atomic.Int32
		conflicts atomic.Int32
	)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		name := spellings[i%len(spellings)]
		go func(u string) {
			defer wg.Done()
			<-start
			w := do(t, r, http.MethodPost, usersPath, `{"schemas":["`+schemaUser+`"],"userName":"`+u+`"}`)
			switch w.Code {
			case http.StatusCreated:
				created.Add(1)
			case http.StatusConflict:
				conflicts.Add(1)
			default:
				t.Errorf("unexpected status %d: %s", w.Code, w.Body.String())
			}
		}(name)
	}
	close(start)
	wg.Wait()

	if created.Load() != 1 || conflicts.Load() != workers-1 {
		t.Fatalf("concurrent POST: created=%d conflicts=%d, want 1 and %d",
			created.Load(), conflicts.Load(), workers-1)
	}

	// Concurrent PUT race: 20 distinct users simultaneously attempt to claim the same new userName.
	const putWorkers = 20
	ids := make([]string, putWorkers)
	for i := 0; i < putWorkers; i++ {
		w := do(t, r, http.MethodPost, usersPath,
			fmt.Sprintf(`{"schemas":[%q],"userName":"seed-%d"}`, schemaUser, i))
		requireStatus(t, w, http.StatusCreated)
		ids[i] = decode(t, w)["id"].(string)
	}

	var (
		putStart     = make(chan struct{})
		putWG        sync.WaitGroup
		putOK        atomic.Int32
		putConflicts atomic.Int32
	)
	for i := 0; i < putWorkers; i++ {
		putWG.Add(1)
		id := ids[i]
		go func(userID string) {
			defer putWG.Done()
			<-putStart
			w := do(t, r, http.MethodPut, usersPath+"/"+userID,
				`{"schemas":["`+schemaUser+`"],"userName":"contested.name"}`)
			switch w.Code {
			case http.StatusOK:
				putOK.Add(1)
			case http.StatusConflict:
				putConflicts.Add(1)
			default:
				t.Errorf("unexpected PUT status %d: %s", w.Code, w.Body.String())
			}
		}(id)
	}
	close(putStart)
	putWG.Wait()

	if putOK.Load() != 1 || putConflicts.Load() != putWorkers-1 {
		t.Fatalf("concurrent PUT: ok=%d conflicts=%d, want 1 and %d",
			putOK.Load(), putConflicts.Load(), putWorkers-1)
	}
}

var _ = json.Marshal
