package parser

import (
	"regexp"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// ScimFilterListenerN1QL is
type ScimFilterListenerN1QL struct {
	*BaseScimFilterListener
	query         string
	prevOperation string
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
	tree := p.Start()
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
	switch node.GetSymbol().GetTokenType() {
	case ScimFilterParserATTRNAME:
		{
			payload, ok := node.GetParent().GetPayload().(*antlr.BaseParserRuleContext)
			if ok {
				_, ok := payload.BaseRuleContext.GetParent().(*ATTR_OPER_CRITERIAContext)
				if !ok {
					value = AddQuote(value)
				} else {
					if l.prevOperation == "co" {
						value = "%" + value + "%"
					}
					if l.prevOperation == "sw" {
						value = value + "%"
					}
					if l.prevOperation == "ew" {
						value = "%" + value
					}
				}
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

// AddQuote turns a SCIM attribute path into a backtick-quoted N1QL identifier,
// splitting off a schema URN prefix when there is one.
//
// Every segment is escaped: a backtick inside an identifier is written twice,
// so a value carrying one cannot close the quoting and append its own N1QL.
// Callers are still expected to validate the path against the schema before
// getting here — this is the second line of defence, not the first.
func AddQuote(value string) string {
	re := regexp.MustCompile(`^(urn[:\w\.\_]*)(:-*)?(:[\w]*)(\.)(.*)$`)
	urn := ""
	if re.MatchString(value) {
		urn = "`" + escapeIdentifier(re.ReplaceAllString(value, `${1}${2}${3}`)) + "`."
	}
	path := re.ReplaceAllString(value, `${5}`)
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
