package parser_test

import (
	"testing"

	"github.com/arturoeanton/goscim/scim/parser"
)

// Every comparison operator in RFC 7644 3.4.2.2 must map to the equivalent
// N1QL operator. gt/ge and lt/le used to be crossed with each other, so range
// queries silently returned the wrong set.
func TestComparisonOperators(t *testing.T) {
	cases := []struct {
		filter string
		where  string
	}{
		{`userName eq "bjensen"`, "`userName` = \"bjensen\""},
		{`userName ne "bjensen"`, "`userName` <> \"bjensen\""},
		{`age gt 10`, "`age` > 10"},
		{`age ge 10`, "`age` >= 10"},
		{`age lt 10`, "`age` < 10"},
		{`age le 10`, "`age` <= 10"},
		{`title pr`, "`title` IS NOT NULL"},
	}

	for _, tc := range cases {
		t.Run(tc.filter, func(t *testing.T) {
			page, count, _ := parser.FilterToN1QL("User", tc.filter)
			wantPage := "SELECT * FROM `User` WHERE " + tc.where
			if page != wantPage {
				t.Errorf("page query:\n  got:  %s\n  want: %s", page, wantPage)
			}
			wantCount := "SELECT count(*) as count FROM `User` WHERE " + tc.where
			if count != wantCount {
				t.Errorf("count query:\n  got:  %s\n  want: %s", count, wantCount)
			}
		})
	}
}

// gt/ge and lt/le must stay distinct from each other: this catches a future
// edit that re-crosses them or collapses both onto the same operator, even if
// the table above were updated without thinking.
func TestRangeOperatorsStayDistinct(t *testing.T) {
	gt, _, _ := parser.FilterToN1QL("User", "age gt 10")
	ge, _, _ := parser.FilterToN1QL("User", "age ge 10")
	lt, _, _ := parser.FilterToN1QL("User", "age lt 10")
	le, _, _ := parser.FilterToN1QL("User", "age le 10")

	if gt == ge {
		t.Error("gt and ge produce the same query")
	}
	if lt == le {
		t.Error("lt and le produce the same query")
	}
	// The strict operator must not carry the inclusive one's "=".
	if len(gt) >= len(ge) {
		t.Errorf("gt (%s) should be stricter than ge (%s)", gt, ge)
	}
	if len(lt) >= len(le) {
		t.Errorf("lt (%s) should be stricter than le (%s)", lt, le)
	}
}

// The substring operators wrap the value in wildcards according to the case.
func TestSubstringOperators(t *testing.T) {
	cases := []struct {
		filter string
		where  string
	}{
		{`emails co "example.com"`, "`emails` LIKE \"%example.com%\""},
		{`emails sw "example"`, "`emails` LIKE \"example%\""},
		{`emails ew ".com"`, "`emails` LIKE \"%.com\""},
	}
	for _, tc := range cases {
		t.Run(tc.filter, func(t *testing.T) {
			page, _, _ := parser.FilterToN1QL("User", tc.filter)
			want := "SELECT * FROM `User` WHERE " + tc.where
			if page != want {
				t.Errorf("\n  got:  %s\n  want: %s", page, want)
			}
		})
	}
}

// An empty filter must not produce a WHERE clause.
func TestEmptyFilter(t *testing.T) {
	page, count, _ := parser.FilterToN1QL("User", "")
	if page != "SELECT * FROM `User`" {
		t.Errorf("page = %s", page)
	}
	if count != "SELECT count(*)  as count FROM `User`" {
		t.Errorf("count = %s", count)
	}
}
