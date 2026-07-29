package parser

import (
	"regexp"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

// ScimFilterListenerN1QL is
type ScimFilterListenerN1QL struct {
	*BaseScimFilterListener
	query         string
	prevOperation string
	inCriteria    bool
}

// EnterCriteria and ExitCriteria track whether the terminals being visited are
// the contents of a quoted value, which have to be escaped rather than emitted
// verbatim.
func (l *ScimFilterListenerN1QL) EnterCriteria(c *CriteriaContext) { l.inCriteria = true }

// ExitCriteria is called when the parser exits a quoted value.
func (l *ScimFilterListenerN1QL) ExitCriteria(c *CriteriaContext) { l.inCriteria = false }

// escapeStringLiteral makes a value safe to sit inside a double-quoted N1QL
// string literal. Only the backslash needs handling: the grammar's ANY token
// excludes the double quote so a value cannot contain one, but a trailing
// backslash would escape the literal's own closing quote and swallow the
// ORDER BY / OFFSET / LIMIT that gets appended after the filter.
func escapeStringLiteral(value string) string {
	return strings.ReplaceAll(value, `\`, `\\`)
}

// FilterToN1QL translates a SCIM filter into the page and count N1QL queries
// for resourceName. A filter that does not parse is rejected with a
// *SyntaxError rather than translated: ANTLR recovers from syntax errors and
// keeps walking, so an unchecked tree lets arbitrary tokens through into the
// query.
func FilterToN1QL(resourceName string, filter string) (string, string, error) {
	query := resourceName + "`"
	if filter == "" {
		return "SELECT * FROM `" + resourceName + "`", "SELECT count(*)  as count FROM `" + resourceName + "`", nil
	}
	p, collector := newFilterParser(filter)
	tree := p.Start_()
	if len(collector.problems) > 0 {
		return "", "", &SyntaxError{Filter: filter, Problems: collector.problems}
	}
	scimFilterListenerN1QL := ScimFilterListenerN1QL{query: query + " WHERE "}
	antlr.ParseTreeWalkerDefault.Walk(&scimFilterListenerN1QL, tree)
	return "SELECT * FROM `" + scimFilterListenerN1QL.query, "SELECT count(*) as count FROM `" + scimFilterListenerN1QL.query, nil
}

// VisitTerminal is called when a terminal node is visited.
func (l *ScimFilterListenerN1QL) VisitTerminal(node antlr.TerminalNode) {
	value := node.GetText()
	if l.inCriteria && value != `"` {
		value = escapeStringLiteral(value)
	}
	switch node.GetSymbol().GetTokenType() {
	case ScimFilterParserATTRNAME:
		{
			// The same token type carries both attribute paths and the text
			// inside a quoted value, told apart by where in the tree it sits.
			// EnterCriteria/ExitCriteria already track that, which is both
			// clearer and sturdier than reaching through the node's parent
			// payload as this used to.
			if l.inCriteria {
				switch l.prevOperation {
				case "co":
					value = "%" + value + "%"
				case "sw":
					value = value + "%"
				case "ew":
					value = "%" + value
				}
			} else {
				value = AddQuote(value)
			}
			l.prevOperation = ""
		}
	case ScimFilterLexerEQ:
		{
			value = "="
			l.prevOperation = "eq"
		}
	case ScimFilterLexerNE:
		{
			value = "<>"
			l.prevOperation = "ne"
		}
	case ScimFilterLexerCO:
		{
			value = "LIKE"
			l.prevOperation = "co"
		}
	case ScimFilterLexerSW:
		{
			value = "LIKE"
			l.prevOperation = "sw"
		}
	case ScimFilterLexerEW:
		{
			value = "LIKE"
			l.prevOperation = "ew"
		}
	case ScimFilterLexerGE:
		{
			value = ">="
			l.prevOperation = "ge"
		}
	case ScimFilterLexerGT:
		{
			value = ">"
			l.prevOperation = "gt"
		}
	case ScimFilterLexerLE:
		{
			value = "<="
			l.prevOperation = "le"
		}
	case ScimFilterLexerLT:
		{
			value = "<"
			l.prevOperation = "lt"
		}

	case ScimFilterLexerPR:
		{
			// The grammar requires at least one WS token before PR, which is
			// emitted verbatim, so this value must not add its own leading space.
			value = "IS NOT NULL"
			// IS NOT NULL - returns rows which contain a value (not NULL or missing).
			// IS NOT MISSING - returns rows which contain a value or null.
			// IS VALUED - synonym for IS NOT NULL
			l.prevOperation = "pr"
		}
	case ScimFilterParserEOF:
		{
			value = ""
		}
	}
	l.query = l.query + value
}

// urnPathPattern splits "urn:...:Schema.attr.sub" into its schema URN and the
// attribute path that follows it. Compiled once: it sits on the request path.
var urnPathPattern = regexp.MustCompile(`^(urn[:\w\.\_]*)(:-*)?(:[\w]*)(\.)(.*)$`)

// SplitURNPath separates the schema URN prefix, when there is one, from the
// attribute path. It lives here because both the N1QL translation and the
// scim package's path handling need exactly this split, and scim already
// imports this package.
func SplitURNPath(path string) (urn string, attrPath string) {
	if urnPathPattern.MatchString(path) {
		return urnPathPattern.ReplaceAllString(path, `${1}${2}${3}`),
			urnPathPattern.ReplaceAllString(path, `${5}`)
	}
	return "", path
}

// AddQuote turns a SCIM attribute path into a backtick-quoted N1QL identifier,
// splitting off a schema URN prefix when there is one.
//
// Every segment is escaped: a backtick inside an identifier is written twice,
// so a value carrying one cannot close the quoting and append its own N1QL.
// Callers are still expected to validate the path against the schema before
// getting here — this is the second line of defence, not the first.
func AddQuote(value string) string {
	schemaURN, path := SplitURNPath(value)
	urn := ""
	if schemaURN != "" {
		urn = "`" + escapeIdentifier(schemaURN) + "`."
	}
	segments := strings.Split(path, ".")
	for i, segment := range segments {
		segments[i] = escapeIdentifier(segment)
	}
	return urn + "`" + strings.Join(segments, "`.`") + "`"
}

// escapeIdentifier doubles backticks so they are literal inside a quoted N1QL
// identifier.
func escapeIdentifier(value string) string {
	return strings.ReplaceAll(value, "`", "``")
}
