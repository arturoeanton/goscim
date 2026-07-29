package parser_test

import (
	"strings"
	"testing"

	"github.com/arturoeanton/goscim/scim/parser"
)

// ANTLR reports syntax errors and then recovers, so an unchecked parse tree
// still yields every token the client sent. The translator used to walk that
// tree regardless, concatenating the leftovers straight into the N1QL text.
// These are the payloads that got through before FilterToN1QL checked for
// syntax errors.
func TestMalformedFiltersAreRejected(t *testing.T) {
	cases := []struct {
		name   string
		filter string
	}{
		{"boolean tautology appended to a valid filter", `userName eq "a" or 1=1`},
		{"statement separator and comment", `userName eq "a" ,;-- garbage`},
		{"trailing subquery", "userName eq \"a\" ) union (select * from `User`"},
		{"operator without operand", `userName eq`},
		{"garbage tokens", `@@@@`},
		{"unbalanced parenthesis", `(userName eq "a"`},
		{"bare value", `"just a string"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			page, count, err := parser.FilterToN1QL("User", tc.filter)
			if err == nil {
				t.Fatalf("filter accepted, produced: %s", page)
			}
			var syntaxErr *parser.SyntaxError
			if !asSyntaxError(err, &syntaxErr) {
				t.Fatalf("error is %T, want *parser.SyntaxError", err)
			}
			if page != "" || count != "" {
				t.Errorf("a rejected filter must not yield queries: %q / %q", page, count)
			}
			if syntaxErr.Filter != tc.filter {
				t.Errorf("the error should carry the offending filter, carries %q", syntaxErr.Filter)
			}
			if len(syntaxErr.Problems) == 0 {
				t.Error("the error should explain what failed to parse")
			}
			if !strings.Contains(err.Error(), "invalid SCIM filter") {
				t.Errorf("unhelpful error message: %v", err)
			}
			if err := parser.Validate(tc.filter); err == nil {
				t.Error("Validate accepted a filter FilterToN1QL rejected")
			}
		})
	}
}

// The rejection must not be so eager that it breaks legitimate filters.
func TestWellFormedFiltersAreAccepted(t *testing.T) {
	cases := []string{
		``,
		`userName eq "bjensen"`,
		`name.familyName co "O"`,
		`title pr`,
		`title pr and userType eq "Employee"`,
		`title pr or userType eq "Intern"`,
		`userType eq "Employee" and (emails co "example.com" or emails co "example.org")`,
		`meta.lastModified gt "2011-05-13T04:42:34Z"`,
		`urn:ietf:params:scim:schemas:extension:one:2.0:Element.required ge 0`,
		`active eq true`,
	}

	for _, filter := range cases {
		t.Run(filter, func(t *testing.T) {
			page, count, err := parser.FilterToN1QL("User", filter)
			if err != nil {
				t.Fatalf("valid filter rejected: %v", err)
			}
			if page == "" || count == "" {
				t.Errorf("empty queries for %q", filter)
			}
			if err := parser.Validate(filter); err != nil {
				t.Errorf("Validate rejected a filter FilterToN1QL accepted: %v", err)
			}
		})
	}
}

// asSyntaxError is errors.As without pulling the import into every case.
func asSyntaxError(err error, target **parser.SyntaxError) bool {
	if e, ok := err.(*parser.SyntaxError); ok {
		*target = e
		return true
	}
	return false
}
