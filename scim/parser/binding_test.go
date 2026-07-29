package parser_test

import (
	"strings"
	"testing"

	"github.com/arturoeanton/goscim/scim/parser"
)

// The property that makes injection impossible rather than merely difficult:
// a filter value never appears in the query text. Escaping can be got wrong
// one case at a time; this cannot be got wrong at all.
func TestFilterValuesNeverReachTheQueryText(t *testing.T) {
	// Values chosen to be recognisable in the output, and to include every
	// character that previously mattered: quotes cannot appear (the grammar
	// excludes them), but backslashes, backticks and SQL punctuation can.
	values := []string{
		"bjensen",
		`DOMAIN\user`,
		"trailing-backslash\\",
		"back`tick",
		"semi;colon",
		"comma,separated",
		"percent%sign",
		"under_score",
		"dash-dash--comment",
		"parens)and(",
	}

	for _, value := range values {
		t.Run(value, func(t *testing.T) {
			for _, operator := range []string{"eq", "ne", "co", "sw", "ew"} {
				filter := `userName ` + operator + ` "` + value + `"`
				query, err := parser.FilterToN1QL("User", filter)
				if err != nil {
					// Some punctuation is not valid filter syntax outside a
					// value; a rejection is a fine outcome, silence is not.
					continue
				}
				for _, text := range []string{query.Page, query.Count} {
					if strings.Contains(text, value) {
						t.Errorf("%s put the value into the query text: %s", operator, text)
					}
				}
				if len(query.Params) != 1 {
					t.Fatalf("%s bound %d parameters, want 1", operator, len(query.Params))
				}
				bound, _ := query.Params[0].(string)
				if !strings.Contains(bound, strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(value)) &&
					bound != value {
					t.Errorf("%s bound %q, which does not carry the value %q", operator, bound, value)
				}
			}
		})
	}
}

// The count and page queries must agree on their placeholders, since they are
// executed with the same parameter list.
func TestBothQueriesShareTheParameters(t *testing.T) {
	query, err := parser.FilterToN1QL("User", `a eq "x" and b co "y" and c gt 1`)
	if err != nil {
		t.Fatal(err)
	}
	if len(query.Params) != 3 {
		t.Fatalf("params = %#v", query.Params)
	}
	for i := 1; i <= 3; i++ {
		placeholder := "$" + string(rune('0'+i))
		if !strings.Contains(query.Page, placeholder) {
			t.Errorf("page is missing %s: %s", placeholder, query.Page)
		}
		if !strings.Contains(query.Count, placeholder) {
			t.Errorf("count is missing %s: %s", placeholder, query.Count)
		}
	}
}
