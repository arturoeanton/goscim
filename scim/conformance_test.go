package scim

import (
	"net/http"
	"strings"
	"testing"
)

// RFC 7644 3.3: a successful create answers 201 with the URI of the new
// resource in Location. It used to answer 200 with no Location at all.
func TestCreateAnswers201WithLocation(t *testing.T) {
	r, _ := newTestServer(t)

	w := do(t, r, http.MethodPost, elementsPath, validElement("Element1", 1))
	requireStatus(t, w, http.StatusCreated)

	id := decode(t, w)["id"].(string)
	location := w.Header().Get("Location")
	if location == "" {
		t.Fatal("no Location header")
	}
	if !strings.HasSuffix(location, PREFIX+"/Elements/"+id) {
		t.Errorf("Location = %s, should point at the new resource", location)
	}
	if !strings.HasPrefix(location, "http://") {
		t.Errorf("Location should be absolute: %s", location)
	}
}

// RFC 7644 3.1: SCIM payloads use the application/scim+json media type.
func TestResponsesUseTheSCIMMediaType(t *testing.T) {
	r, _ := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)

	cases := []struct {
		name   string
		method string
		target string
		body   string
	}{
		{"create", http.MethodPost, elementsPath, validElement("Element2", 2)},
		{"read", http.MethodGet, elementsPath + "/" + id, ""},
		{"search", http.MethodGet, elementsPath, ""},
		{"replace", http.MethodPut, elementsPath + "/" + id, validElement("Element1", 3)},
		{"error", http.MethodGet, elementsPath + "/does-not-exist", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := do(t, r, tc.method, tc.target, tc.body)
			contentType := w.Header().Get("Content-Type")
			if !strings.HasPrefix(contentType, "application/scim+json") {
				t.Errorf("Content-Type = %q", contentType)
			}
		})
	}
}

// meta.location has to be resolvable, which means carrying the /scim/v2 prefix
// the routes are actually mounted under. It used to be "/Elements/<id>".
func TestMetaLocationCarriesTheEndpointPrefix(t *testing.T) {
	r, _ := newTestServer(t)
	created := createElement(t, r, "Element1", 1)
	id := created["id"].(string)

	want := PREFIX + "/Elements/" + id
	meta := created["meta"].(map[string]interface{})
	if meta["location"] != want {
		t.Errorf("create meta.location = %v, want %v", meta["location"], want)
	}

	w := do(t, r, http.MethodGet, elementsPath+"/"+id, "")
	requireStatus(t, w, http.StatusOK)
	meta = decode(t, w)["meta"].(map[string]interface{})
	if meta["location"] != want {
		t.Errorf("read meta.location = %v, want %v", meta["location"], want)
	}
}

// Every write stamps a fresh meta.version but nothing compared it, so two
// concurrent updates silently overwrote each other. The version is now
// published as an ETag and honoured on If-Match.
func TestConcurrencyControl(t *testing.T) {
	t.Run("responses carry the version as an ETag", func(t *testing.T) {
		r, _ := newTestServer(t)
		w := do(t, r, http.MethodPost, elementsPath, validElement("Element1", 1))
		requireStatus(t, w, http.StatusCreated)

		version := decode(t, w)["meta"].(map[string]interface{})["version"].(string)
		if etag := w.Header().Get("ETag"); etag != `W/"`+version+`"` {
			t.Errorf("ETag = %q, want the weak form of %q", etag, version)
		}

		id := decode(t, w)["id"].(string)
		w = do(t, r, http.MethodGet, elementsPath+"/"+id, "")
		if w.Header().Get("ETag") == "" {
			t.Error("read carries no ETag")
		}
	})

	t.Run("a stale If-Match is refused", func(t *testing.T) {
		r, store := newTestServer(t)
		created := createElement(t, r, "Element1", 1)
		id := created["id"].(string)
		staleVersion := created["meta"].(map[string]interface{})["version"].(string)

		// Someone else writes first.
		w := do(t, r, http.MethodPut, elementsPath+"/"+id, validElement("Element1-by-someone-else", 2))
		requireStatus(t, w, http.StatusOK)

		// Our update still carries the version we read.
		req := newRequest(http.MethodPut, elementsPath+"/"+id, validElement("Element1-mine", 3))
		req.Header.Set("If-Match", `W/"`+staleVersion+`"`)
		w = serve(r, req)
		requireStatus(t, w, http.StatusPreconditionFailed)

		stored, _ := store.Get("Element", id)
		if stored["name"] != "Element1-by-someone-else" {
			t.Errorf("the stale write went through: %v", stored["name"])
		}
	})

	t.Run("a current If-Match is accepted in every form", func(t *testing.T) {
		for _, form := range []func(string) string{
			func(v string) string { return v },             // the bare meta.version
			func(v string) string { return `"` + v + `"` }, // quoted
			func(v string) string { return `W/"` + v + `"` },
			func(string) string { return "*" },
		} {
			r, _ := newTestServer(t)
			created := createElement(t, r, "Element1", 1)
			id := created["id"].(string)
			version := created["meta"].(map[string]interface{})["version"].(string)

			req := newRequest(http.MethodPut, elementsPath+"/"+id, validElement("Element1-updated", 2))
			req.Header.Set("If-Match", form(version))
			w := serve(r, req)
			requireStatus(t, w, http.StatusOK)
		}
	})

	t.Run("If-Match guards delete too", func(t *testing.T) {
		r, store := newTestServer(t)
		created := createElement(t, r, "Element1", 1)
		id := created["id"].(string)

		req := newRequest(http.MethodDelete, elementsPath+"/"+id, "")
		req.Header.Set("If-Match", `W/"not-the-current-version"`)
		w := serve(r, req)
		requireStatus(t, w, http.StatusPreconditionFailed)

		if _, err := store.Get("Element", id); err != nil {
			t.Error("the resource was deleted despite the failed precondition")
		}
	})

	t.Run("If-None-Match answers 304", func(t *testing.T) {
		r, _ := newTestServer(t)
		created := createElement(t, r, "Element1", 1)
		id := created["id"].(string)
		version := created["meta"].(map[string]interface{})["version"].(string)

		req := newRequest(http.MethodGet, elementsPath+"/"+id, "")
		req.Header.Set("If-None-Match", `W/"`+version+`"`)
		w := serve(r, req)
		requireStatus(t, w, http.StatusNotModified)
		if w.Body.Len() != 0 {
			t.Errorf("304 with a body: %s", w.Body.String())
		}

		// A different version still returns the resource.
		req = newRequest(http.MethodGet, elementsPath+"/"+id, "")
		req.Header.Set("If-None-Match", `W/"something-else"`)
		w = serve(r, req)
		requireStatus(t, w, http.StatusOK)
	})
}

// RFC 7644 3.4.2.4: itemsPerPage is how many resources the page carries. It
// used to echo the requested count regardless of what came back.
func TestItemsPerPageReflectsThePage(t *testing.T) {
	r, _ := newTestServer(t)
	createElement(t, r, "AAA", 1)
	createElement(t, r, "BBB", 2)
	createElement(t, r, "CCC", 3)

	cases := []struct {
		query string
		want  float64
	}{
		{"?sortBy=name&count=2", 2},
		{"?sortBy=name&count=100", 3},
		{"?sortBy=name&startIndex=3&count=100", 1},
	}

	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			w := do(t, r, http.MethodGet, elementsPath+tc.query, "")
			requireStatus(t, w, http.StatusOK)
			body := decode(t, w)

			if body["itemsPerPage"] != tc.want {
				t.Errorf("itemsPerPage = %v, want %v", body["itemsPerPage"], tc.want)
			}
			if body["totalResults"] != float64(3) {
				t.Errorf("totalResults = %v, want 3", body["totalResults"])
			}
			if got := float64(len(body["Resources"].([]interface{}))); got != tc.want {
				t.Errorf("itemsPerPage says %v but %v resources came back", body["itemsPerPage"], got)
			}
		})
	}
}

// RFC 7644 3.4.2 makes totalResults required. A search that matched nothing
// used to answer without it, and without Resources, because both fields were
// omitempty -- leaving a client unable to tell an empty result from a
// malformed response.
func TestEmptySearchStillCarriesTheListFields(t *testing.T) {
	r, _ := newTestServer(t)

	w := do(t, r, http.MethodGet, elementsPath, "")
	requireStatus(t, w, http.StatusOK)
	body := decode(t, w)

	for _, field := range []string{"schemas", "totalResults", "itemsPerPage", "startIndex", "Resources"} {
		if _, present := body[field]; !present {
			t.Errorf("%s is missing from an empty ListResponse: %s", field, w.Body.String())
		}
	}
	if body["totalResults"] != float64(0) {
		t.Errorf("totalResults = %v", body["totalResults"])
	}
	resources, ok := body["Resources"].([]interface{})
	if !ok || len(resources) != 0 {
		t.Errorf("Resources = %#v, want an empty array", body["Resources"])
	}
}
