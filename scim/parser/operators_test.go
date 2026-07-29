package parser_test

import (
	"strings"
	"testing"

	"github.com/arturoeanton/goscim/scim/parser"
)

// Every comparison operator in RFC 7644 3.4.2.2 must map to the equivalent
// N1QL operator, and the value must leave as a bound parameter rather than as
// text. gt/ge and lt/le used to be crossed with each other, so range queries
// silently returned the wrong set.
func TestComparisonOperators(t *testing.T) {
	cases := []struct {
		filter string
		where  string
		params []interface{}
	}{
		{`userName eq "bjensen"`, "`userName` = $1", []interface{}{"bjensen"}},
		{`userName ne "bjensen"`, "`userName` <> $1", []interface{}{"bjensen"}},
		{`age gt 10`, "`age` > $1", []interface{}{float64(10)}},
		{`age ge 10`, "`age` >= $1", []interface{}{float64(10)}},
		{`age lt 10`, "`age` < $1", []interface{}{float64(10)}},
		{`age le 10`, "`age` <= $1", []interface{}{float64(10)}},
		{`active eq true`, "`active` = $1", []interface{}{true}},
		{`active eq false`, "`active` = $1", []interface{}{false}},
		{`title pr`, "`title` IS NOT NULL", nil},
	}

	for _, tc := range cases {
		t.Run(tc.filter, func(t *testing.T) {
			query, err := parser.FilterToN1QL("User", tc.filter)
			if err != nil {
				t.Fatal(err)
			}
			wantPage := "SELECT * FROM `User` WHERE " + tc.where
			if query.Page != wantPage {
				t.Errorf("page:\n  got:  %s\n  want: %s", query.Page, wantPage)
			}
			wantCount := "SELECT count(*) as count FROM `User` WHERE " + tc.where
			if query.Count != wantCount {
				t.Errorf("count:\n  got:  %s\n  want: %s", query.Count, wantCount)
			}
			assertParams(t, query.Params, tc.params)
		})
	}
}

// gt/ge and lt/le must stay distinct from each other: this catches a future
// edit that re-crosses them or collapses both onto the same operator.
func TestRangeOperatorsStayDistinct(t *testing.T) {
	page := func(filter string) string {
		query, err := parser.FilterToN1QL("User", filter)
		if err != nil {
			t.Fatal(err)
		}
		return query.Page
	}
	gt, ge := page("age gt 10"), page("age ge 10")
	lt, le := page("age lt 10"), page("age le 10")

	if gt == ge || lt == le {
		t.Errorf("the strict and inclusive operators collapsed: %s / %s / %s / %s", gt, ge, lt, le)
	}
	if len(gt) >= len(ge) {
		t.Errorf("gt (%s) should be stricter than ge (%s)", gt, ge)
	}
	if len(lt) >= len(le) {
		t.Errorf("lt (%s) should be stricter than le (%s)", lt, le)
	}
}

// The substring operators become LIKE patterns. The wildcards belong to the
// pattern, so they go in the bound value, and the client's own % and _ are
// escaped with the ESCAPE clause the query carries.
func TestSubstringOperators(t *testing.T) {
	cases := []struct {
		filter string
		param  string
	}{
		{`emails co "example.com"`, "%example.com%"},
		{`emails sw "example"`, "example%"},
		{`emails ew ".com"`, "%.com"},
		{`emails co "50%"`, `%50\%%`},
		{`emails sw "a_b"`, `a\_b%`},
	}
	for _, tc := range cases {
		t.Run(tc.filter, func(t *testing.T) {
			query, err := parser.FilterToN1QL("User", tc.filter)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(query.Page, "LIKE $1") {
				t.Errorf("no bound LIKE: %s", query.Page)
			}
			if !strings.Contains(query.Page, `ESCAPE "\\"`) {
				t.Errorf("a LIKE pattern with escaping needs an ESCAPE clause: %s", query.Page)
			}
			assertParams(t, query.Params, []interface{}{tc.param})
		})
	}
}

// An empty filter must not produce a WHERE clause or any parameters.
func TestEmptyFilter(t *testing.T) {
	query, err := parser.FilterToN1QL("User", "")
	if err != nil {
		t.Fatal(err)
	}
	if query.Page != "SELECT * FROM `User`" {
		t.Errorf("page = %s", query.Page)
	}
	if query.Count != "SELECT count(*) as count FROM `User`" {
		t.Errorf("count = %s", query.Count)
	}
	if len(query.Params) != 0 {
		t.Errorf("params = %v", query.Params)
	}
}

// Parameters are numbered in the order they appear, and both queries share the
// same list.
func TestParameterNumbering(t *testing.T) {
	query, err := parser.FilterToN1QL("User",
		`userName eq "a" and (age gt 3 or title co "b")`)
	if err != nil {
		t.Fatal(err)
	}
	for _, placeholder := range []string{"$1", "$2", "$3"} {
		if !strings.Contains(query.Page, placeholder) {
			t.Errorf("%s missing from %s", placeholder, query.Page)
		}
	}
	assertParams(t, query.Params, []interface{}{"a", float64(3), "%b%"})
	if query.Count == "" || !strings.Contains(query.Count, "$3") {
		t.Errorf("the count query must carry the same placeholders: %s", query.Count)
	}
}

func assertParams(t *testing.T, got, want []interface{}) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("params = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("params[%d] = %#v (%T), want %#v (%T)", i, got[i], got[i], want[i], want[i])
		}
	}
}
