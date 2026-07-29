package scim

import (
	"net/http"
	"strings"
	"testing"
)

// The three discovery endpoints answered 501 and hung off the server root
// rather than /scim/v2, and the shipped service provider config was never
// read. They are the first thing an identity provider asks for.

func TestServiceProviderConfig(t *testing.T) {
	r, _ := newTestServer(t)

	w := do(t, r, http.MethodGet, PREFIX+"/ServiceProviderConfig", "")
	requireStatus(t, w, http.StatusOK)
	body := decode(t, w)

	schemas := body["schemas"].([]interface{})
	if len(schemas) != 1 || schemas[0] != "urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig" {
		t.Errorf("schemas = %v", schemas)
	}

	// The document must describe what this server does, not what the file
	// used to claim. The shipped config advertised bulk support, which does
	// not exist.
	capabilities := map[string]bool{
		"patch":          true,
		"bulk":           false,
		"filter":         true,
		"changePassword": false,
		"sort":           true,
		"etag":           true,
	}
	for name, want := range capabilities {
		section, ok := body[name].(map[string]interface{})
		if !ok {
			t.Errorf("%s is missing", name)
			continue
		}
		if section["supported"] != want {
			t.Errorf("%s.supported = %v, want %v", name, section["supported"], want)
		}
	}

	filter := body["filter"].(map[string]interface{})
	if filter["maxResults"] != float64(maxSearchCount) {
		t.Errorf("filter.maxResults = %v, want %d", filter["maxResults"], maxSearchCount)
	}
}

// The advertised authentication schemes come from the authenticator actually
// in force, so the two cannot drift apart.
func TestServiceProviderConfigReportsTheActiveAuthentication(t *testing.T) {
	cases := []struct {
		name          string
		authenticator Authenticator
		wantType      string
	}{
		{"basic", &BasicAuthenticator{}, "httpbasic"},
		{"anonymous", &AnonymousAuthenticator{}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := newTestServerAs(t, tc.authenticator)
			w := do(t, r, http.MethodGet, PREFIX+"/ServiceProviderConfig", "")
			requireStatus(t, w, http.StatusOK)

			schemes, ok := decode(t, w)["authenticationSchemes"].([]interface{})
			if !ok {
				t.Fatalf("authenticationSchemes missing: %s", w.Body.String())
			}
			if tc.wantType == "" {
				if len(schemes) != 0 {
					t.Errorf("an unauthenticated server advertised %v", schemes)
				}
				return
			}
			if len(schemes) != 1 || schemes[0].(map[string]interface{})["type"] != tc.wantType {
				t.Errorf("authenticationSchemes = %v, want type %s", schemes, tc.wantType)
			}
		})
	}
}

func TestResourceTypesDiscovery(t *testing.T) {
	r, _ := newTestServer(t)

	w := do(t, r, http.MethodGet, PREFIX+"/ResourceTypes", "")
	requireStatus(t, w, http.StatusOK)
	body := decode(t, w)

	if body["totalResults"] != float64(3) {
		t.Errorf("totalResults = %v, want 3 (User, Group, Element)", body["totalResults"])
	}
	names := map[string]map[string]interface{}{}
	for _, raw := range body["Resources"].([]interface{}) {
		item := raw.(map[string]interface{})
		names[item["name"].(string)] = item
	}
	for _, want := range []string{"User", "Group", "Element"} {
		item, present := names[want]
		if !present {
			t.Fatalf("%s is missing from /ResourceTypes", want)
		}
		if item["endpoint"] == "" || item["schema"] == "" {
			t.Errorf("%s is incomplete: %v", want, item)
		}
		meta := item["meta"].(map[string]interface{})
		if meta["resourceType"] != "ResourceType" {
			t.Errorf("%s meta.resourceType = %v", want, meta["resourceType"])
		}
		if !strings.HasPrefix(meta["location"].(string), PREFIX+"/ResourceTypes/") {
			t.Errorf("%s meta.location = %v", want, meta["location"])
		}
	}

	// And individually.
	w = do(t, r, http.MethodGet, PREFIX+"/ResourceTypes/User", "")
	requireStatus(t, w, http.StatusOK)
	if decode(t, w)["endpoint"] != "/Users" {
		t.Errorf("the User resource type is wrong: %s", w.Body.String())
	}

	w = do(t, r, http.MethodGet, PREFIX+"/ResourceTypes/NoSuchThing", "")
	requireStatus(t, w, http.StatusNotFound)
}

func TestSchemasDiscovery(t *testing.T) {
	r, _ := newTestServer(t)

	w := do(t, r, http.MethodGet, PREFIX+"/Schemas", "")
	requireStatus(t, w, http.StatusOK)
	body := decode(t, w)

	if body["totalResults"] != float64(len(Schemas)) {
		t.Errorf("totalResults = %v, want %d", body["totalResults"], len(Schemas))
	}

	var user map[string]interface{}
	for _, raw := range body["Resources"].([]interface{}) {
		item := raw.(map[string]interface{})
		if item["id"] == schemaUser {
			user = item
		}
	}
	if user == nil {
		t.Fatalf("the core User schema is missing from /Schemas")
	}
	attributes := user["attributes"].([]interface{})
	if len(attributes) == 0 {
		t.Fatal("the User schema has no attributes")
	}

	// Individually, by URN.
	w = do(t, r, http.MethodGet, PREFIX+"/Schemas/"+schemaUser, "")
	requireStatus(t, w, http.StatusOK)
	if decode(t, w)["id"] != schemaUser {
		t.Errorf("wrong schema: %s", w.Body.String())
	}

	w = do(t, r, http.MethodGet, PREFIX+"/Schemas/urn:made:up", "")
	requireStatus(t, w, http.StatusNotFound)
}

// $reader and $writer are this project's authorization extension, not part of
// the schema a client should consume.
func TestSchemaDiscoveryHidesTheAuthorizationExtension(t *testing.T) {
	r, _ := newTestServer(t)

	// The Element core schema is the one that declares them.
	w := do(t, r, http.MethodGet, PREFIX+"/Schemas/"+schemaCore, "")
	requireStatus(t, w, http.StatusOK)

	if body := w.Body.String(); strings.Contains(body, "$reader") || strings.Contains(body, "$writer") {
		t.Errorf("the authorization extension leaked into /Schemas: %s", body)
	}
}

// Discovery has to be reachable without credentials: a client needs it to
// find out how to authenticate (RFC 7644 2).
func TestDiscoveryDoesNotRequireCredentials(t *testing.T) {
	r, _ := newTestServerAs(t, &BasicAuthenticator{})
	for _, path := range []string{"/ServiceProviderConfig", "/ResourceTypes", "/Schemas"} {
		w := do(t, r, http.MethodGet, PREFIX+path, "")
		if w.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200 without credentials", path, w.Code)
		}
	}
}
