package scim

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/arturoeanton/goscim/scim/parser"
)

// sortBy reaches the query as an identifier, not as a bound parameter, so it
// has to be constrained to attributes the schema declares. Anything else is an
// ORDER BY clause under the client's control.
func TestSearchRejectsUnknownSortBy(t *testing.T) {
	cases := []struct {
		name   string
		sortBy string
	}{
		{"identifier break-out", "id` , (SELECT * FROM `Element`) x"},
		{"undeclared attribute", "doesNotExist"},
		{"undeclared nested attribute", "name.somethingElse"},
		{"schema the resource type does not use", "urn:ietf:params:scim:schemas:core:2.0:User.userName"},
		{"unknown schema urn", "urn:made:up:2.0:Thing.attr"},
		{"empty path in a list", "name,"},
		{"statement terminator", "name; drop"},
		{"comment", "name-- x"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := newTestServer(t)
			createElement(t, r, "AAA", 1)

			target := elementsPath + "?sortBy=" + url.QueryEscape(tc.sortBy)
			w := do(t, r, http.MethodGet, target, "")

			body := requireSCIMError(t, w, http.StatusBadRequest)
			if body["scimType"] != "invalidValue" {
				t.Errorf("scimType = %v, want invalidValue", body["scimType"])
			}
		})
	}
}

// The check must not reject the values a real client sends.
func TestSearchAcceptsDeclaredSortBy(t *testing.T) {
	cases := []string{
		"name",
		"description",
		"$ref",
		"id",
		"meta.lastModified",
		"name,description",
		" name , description ",
		"urn:ietf:params:scim:schemas:extension:one:2.0:Element.required",
	}

	for _, sortBy := range cases {
		t.Run(sortBy, func(t *testing.T) {
			r, _ := newTestServer(t)
			createElement(t, r, "AAA", 1)

			target := elementsPath + "?sortBy=" + url.QueryEscape(sortBy)
			w := do(t, r, http.MethodGet, target, "")
			if w.Code != http.StatusOK {
				t.Errorf("declared attribute rejected: status %d, body %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestNormalizeSortBy(t *testing.T) {
	newTestServer(t) // loads the schemas into the package globals
	elements := Resources["/Elements"]

	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"", "", false},
		{"   ", "", false},
		{"name", "name", false},
		{" name , description ", "name,description", false},
		{"id", "id", false},
		{"meta.version", "meta.version", false},
		{"nope", "", true},
		{"name,", "", true},
		{"`x`", "", true},
	}

	for _, tc := range cases {
		got, err := NormalizeSortBy(elements, tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("NormalizeSortBy(%q) accepted the value, got %q", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("NormalizeSortBy(%q) = %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("NormalizeSortBy(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// Second line of defence: even handed a path with a backtick, the quoting must
// not let it close the identifier.
func TestAddQuoteEscapesBackticks(t *testing.T) {
	got := parser.AddQuote("id` , (SELECT * FROM `Element`) x")
	if strings.HasPrefix(got, "`id` ") {
		t.Errorf("the identifier is still closed by the injected backtick: %s", got)
	}
	if !strings.HasPrefix(got, "`id``") {
		t.Errorf("backtick not doubled: %s", got)
	}
	if !strings.HasSuffix(got, "`") {
		t.Errorf("identifier not closed: %s", got)
	}

	// Ordinary paths are untouched.
	if q := parser.AddQuote("name.familyName"); q != "`name`.`familyName`" {
		t.Errorf("AddQuote(name.familyName) = %s", q)
	}
}
