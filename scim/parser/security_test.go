package parser_test

import (
	"strings"
	"testing"

	"github.com/arturoeanton/goscim/scim/parser"
)

// TestPermissiveGrammarConstructsAreRejected verifies that constructs accepted
// by the raw ANTLR grammar without a lexer/parser error -- such as "--" matching
// [-.0-9]+, adjacent expressions without a boolean operator, expression-to-
// expression comparisons and unsupported bracketed value filters -- are refused
// before they can emit broken or unescaped N1QL.
func TestPermissiveGrammarConstructsAreRejected(t *testing.T) {
	cases := []struct {
		name   string
		filter string
	}{
		{"SQL comment lexed as NUMBERS", `age eq --`},
		{"multiple decimal points lexed as NUMBERS", `age eq 1.2.3`},
		{"lone minus sign lexed as NUMBERS", `age eq -`},
		{"two adjacent expressions without AND/OR", `(userName eq "a")(active eq true)`},
		{"attribute compared to boolean expression", `userName eq title pr`},
		{"expression compared to expression", `(userName eq "a") eq (active eq true)`},
		{"empty complex value filter", `emails[]`},
		{"malformed complex value filter", `emails[type eq]`},
		{"SQL comment inside complex value filter", `emails[type eq --]`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := parser.FilterToN1QL("User", tc.filter, nil)
			if err == nil {
				t.Fatalf("filter %q was accepted and emitted: %s", tc.filter, query.Page)
			}
			if strings.Contains(query.Page, "--") {
				t.Fatalf("SQL comment reached query text: %s", query.Page)
			}
			if err := parser.Validate(tc.filter); err == nil {
				t.Errorf("Validate(%q) accepted a filter FilterToN1QL rejected", tc.filter)
			}
		})
	}
}

func TestReferencedAttributesExtractsAllPaths(t *testing.T) {
	attrs, err := parser.ReferencedAttributes(`userName eq "jane" and (name.familyName co "Doe" or active eq true)`)
	if err != nil {
		t.Fatal(err)
	}
	want := "userName,name.familyName,active"
	if got := strings.Join(attrs, ","); got != want {
		t.Errorf("ReferencedAttributes = %q, want %q", got, want)
	}
}
