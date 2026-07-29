package scim

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// The production store must reject a malformed filter as ErrInvalidFilter
// before it builds or runs anything. A nil cluster is enough to prove it never
// gets that far: if the rejection were removed this would panic instead of
// returning.
func TestCouchbaseStoreRejectsMalformedFiltersBeforeQuerying(t *testing.T) {
	store := NewCouchbaseStore(nil)

	for _, filter := range []string{`userName eq "a" or 1=1`, `@@@@`, `name eq`} {
		t.Run(filter, func(t *testing.T) {
			total, resources, err := store.Search(SearchQuery{Bucket: "Element", Filter: filter})
			if !errors.Is(err, ErrInvalidFilter) {
				t.Fatalf("err = %v, want ErrInvalidFilter", err)
			}
			if total != 0 || resources != nil {
				t.Errorf("a rejected filter must not yield results: %d, %v", total, resources)
			}
		})
	}
}

// A filter the parser cannot understand is a client error, not a server error,
// and must never be translated. Before this was enforced, ANTLR recovered from
// the syntax error and the leftover tokens were concatenated into the
// generated N1QL.
func TestSearchRejectsMalformedFilters(t *testing.T) {
	cases := []struct {
		name   string
		filter string
	}{
		{"boolean tautology", `userName eq "a" or 1=1`},
		{"statement separator", `userName eq "a" ,;-- garbage`},
		{"trailing subquery", "userName eq \"a\" ) union (select * from `Element`"},
		{"garbage tokens", `@@@@`},
		{"operator without operand", `name eq`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := newTestServer(t)
			createElement(t, r, "AAA", 1)

			target := elementsPath + "?filter=" + url.QueryEscape(tc.filter)
			w := do(t, r, http.MethodGet, target, "")

			body := requireSCIMError(t, w, http.StatusBadRequest)
			if body["scimType"] != "invalidFilter" {
				t.Errorf("scimType = %v, want invalidFilter", body["scimType"])
			}
			if body["status"] != "400" {
				t.Errorf("status in the body = %v", body["status"])
			}
			// A rejected filter must not come back with a query attached.
			if detail, _ := body["detail"].(string); strings.Contains(detail, "SELECT") {
				t.Errorf("the error leaks the generated query: %s", detail)
			}
		})
	}
}

// The rejection must not be over-eager: a well-formed filter has to get past
// validation. The in-memory store does not evaluate filters, so it answers 500
// here — the point is only that the request is not turned away as malformed.
func TestSearchAcceptsWellFormedFilters(t *testing.T) {
	cases := []string{
		`name eq "AAA"`,
		`name co "AA"`,
		`name pr`,
		`name eq "AAA" and $ref pr`,
		`urn:ietf:params:scim:schemas:extension:one:2.0:Element.required ge 0`,
	}

	for _, filter := range cases {
		t.Run(filter, func(t *testing.T) {
			r, _ := newTestServer(t)
			createElement(t, r, "AAA", 1)

			target := elementsPath + "?filter=" + url.QueryEscape(filter)
			w := do(t, r, http.MethodGet, target, "")

			if w.Code == http.StatusBadRequest {
				t.Errorf("well-formed filter rejected as malformed: %s", w.Body.String())
			}
		})
	}
}
