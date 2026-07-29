package parser_test

import (
	"strings"
	"testing"

	"github.com/arturoeanton/goscim/scim/parser"
)

// Filter values are concatenated into the generated N1QL as double-quoted
// string literals. A value ending in a backslash used to escape the literal's
// own closing quote, so everything appended afterwards -- the ORDER BY, OFFSET
// and LIMIT that Search adds -- was swallowed into the string.
func TestFilterValuesCannotEscapeTheirLiteral(t *testing.T) {
	cases := []struct {
		name   string
		filter string
		want   string
	}{
		{"trailing backslash", `name eq "a\"`, "`name` = \"a\\\\\""},
		{"embedded backslash", `name eq "a\b"`, "`name` = \"a\\\\b\""},
		{"double backslash", `name eq "a\\"`, "`name` = \"a\\\\\\\\\""},
		{"windows-style value", `name eq "DOMAIN\user"`, "`name` = \"DOMAIN\\\\user\""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			page, _, err := parser.FilterToN1QL("Element", tc.filter)
			if err != nil {
				t.Fatalf("filter rejected: %v", err)
			}
			want := "SELECT * FROM `Element` WHERE " + tc.want
			if page != want {
				t.Errorf("\n  got:  %s\n  want: %s", page, want)
			}
			assertLiteralsAreClosed(t, page)
		})
	}
}

// Values without backslashes must come through untouched.
func TestOrdinaryFilterValuesAreUnchanged(t *testing.T) {
	for _, filter := range []string{
		`name eq "bjensen"`,
		`name eq "O'Malley"`,
		`name co "example.com"`,
		`meta.lastModified gt "2011-05-13T04:42:34Z"`,
	} {
		t.Run(filter, func(t *testing.T) {
			page, _, err := parser.FilterToN1QL("Element", filter)
			if err != nil {
				t.Fatalf("filter rejected: %v", err)
			}
			if strings.Contains(page, `\`) {
				t.Errorf("a value with no backslash gained one: %s", page)
			}
			assertLiteralsAreClosed(t, page)
		})
	}
}

// assertLiteralsAreClosed checks that every double quote in the generated query
// opens or closes a literal, i.e. that none of them was neutralised by a
// preceding backslash.
func assertLiteralsAreClosed(t *testing.T, query string) {
	t.Helper()
	open := false
	for i := 0; i < len(query); i++ {
		switch query[i] {
		case '\\':
			if open {
				i++ // the next byte is escaped, whatever it is
			}
		case '"':
			open = !open
		}
	}
	if open {
		t.Errorf("the query ends inside an unterminated string literal: %s", query)
	}
}
