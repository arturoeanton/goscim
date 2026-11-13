package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

// Query is a translated SCIM filter: N1QL text with positional placeholders,
// plus the values to bind to them.
//
// Values are never written into the text. Concatenating them is what made this
// translation an injection surface to begin with, and no amount of escaping is
// as reliable as not doing it: a bound value cannot become syntax, whatever it
// contains.
type Query struct {
	Page   string
	Count  string
	Params []interface{}
}

// NameResolver maps an attribute path to the spelling its schema declares.
// RFC 7643 2.1 makes attribute names case-insensitive, but N1QL identifiers are
// not, so a filter naming "USERNAME" has to become `userName` or it matches a
// key no document has.
type NameResolver func(path string) string

type n1qlArrayScope struct {
	attrNode      antlr.TerminalNode
	lbracNode     antlr.TerminalNode
	rbracNode     antlr.TerminalNode
	fullPath      string
	resolvedFull  string
	varName       string
	trimLeadingWS bool
}

// ScimFilterListenerN1QL walks the parse tree building the N1QL text.
type ScimFilterListenerN1QL struct {
	*BaseScimFilterListener
	query         []byte
	params        []interface{}
	prevOperation string
	resolve       NameResolver

	inCriteria bool
	criteria   strings.Builder

	scopes  []*n1qlArrayScope
	nextVar int
}

func (l *ScimFilterListenerN1QL) write(s string) {
	l.query = append(l.query, s...)
}

func (l *ScimFilterListenerN1QL) trimTrailingSpaces() {
	for len(l.query) > 0 && l.query[len(l.query)-1] == ' ' {
		l.query = l.query[:len(l.query)-1]
	}
}

// EnterLBRAC_EXPR_RBRAC translates attribute[subFilter] into a N1QL quantified
// array predicate:
//
//	(ANY `_v0` IN `attribute` SATISFIES ... END)
func (l *ScimFilterListenerN1QL) EnterLBRAC_EXPR_RBRAC(c *LBRAC_EXPR_RBRACContext) {
	if len(l.scopes) > 0 {
		l.scopes[len(l.scopes)-1].trimLeadingWS = false
	}

	rawAttr := c.ATTRNAME().GetText()
	fullPath := rawAttr
	if len(l.scopes) > 0 {
		fullPath = l.scopes[len(l.scopes)-1].fullPath + "." + rawAttr
	}
	resolvedFull := fullPath
	if l.resolve != nil {
		resolvedFull = l.resolve(fullPath)
	}

	var collectionExpr string
	if len(l.scopes) == 0 {
		collectionExpr = AddQuote(resolvedFull)
	} else {
		parent := l.scopes[len(l.scopes)-1]
		rel := stripParentPrefix(resolvedFull, parent.resolvedFull, rawAttr)
		collectionExpr = "`" + parent.varName + "`." + AddQuote(rel)
	}

	varName := "_v" + strconv.Itoa(l.nextVar)
	l.nextVar++
	l.prevOperation = ""
	l.write("(ANY `" + varName + "` IN " + collectionExpr + " SATISFIES ")

	l.scopes = append(l.scopes, &n1qlArrayScope{
		attrNode:      c.ATTRNAME(),
		lbracNode:     c.LBRAC(),
		rbracNode:     c.RBRAC(),
		fullPath:      fullPath,
		resolvedFull:  resolvedFull,
		varName:       varName,
		trimLeadingWS: true,
	})
}

// ExitLBRAC_EXPR_RBRAC closes the N1QL quantified array predicate.
func (l *ScimFilterListenerN1QL) ExitLBRAC_EXPR_RBRAC(*LBRAC_EXPR_RBRACContext) {
	if len(l.scopes) > 0 {
		l.scopes = l.scopes[:len(l.scopes)-1]
	}
	l.trimTrailingSpaces()
	l.write(" END)")
}

func stripParentPrefix(resolvedFull, parentResolved, fallback string) string {
	prefix := parentResolved + "."
	if len(resolvedFull) > len(prefix) && strings.EqualFold(resolvedFull[:len(prefix)], prefix) {
		return resolvedFull[len(prefix):]
	}
	return fallback
}

// EnterCriteria starts collecting the text of a quoted value. Its terminals are
// buffered rather than written, so the value leaves as a parameter.
func (l *ScimFilterListenerN1QL) EnterCriteria(c *CriteriaContext) {
	if len(l.scopes) > 0 {
		l.scopes[len(l.scopes)-1].trimLeadingWS = false
	}
	l.inCriteria = true
	l.criteria.Reset()
}

// ExitCriteria binds the collected value and writes its placeholder.
func (l *ScimFilterListenerN1QL) ExitCriteria(c *CriteriaContext) {
	l.inCriteria = false
	value, isPattern := l.likePattern(l.criteria.String())
	l.bind(value)
	if isPattern {
		// Makes the escaping in escapeLikeWildcards effective.
		l.write(` ESCAPE "\\"`)
	}
}

// likePattern shapes a value for the operator it belongs to. For the substring
// operators the value becomes a LIKE pattern, which means its own % and _ have
// to be escaped first: a client searching for a literal percent sign was
// otherwise handing the query engine a wildcard that matches everything.
func (l *ScimFilterListenerN1QL) likePattern(value string) (string, bool) {
	switch l.prevOperation {
	case "co":
		return "%" + escapeLikeWildcards(value) + "%", true
	case "sw":
		return escapeLikeWildcards(value) + "%", true
	case "ew":
		return "%" + escapeLikeWildcards(value), true
	default:
		return value, false
	}
}

// escapeLikeWildcards neutralises the LIKE metacharacters, matching the ESCAPE
// clause the translation emits.
func escapeLikeWildcards(value string) string {
	return strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(value)
}

// bind records a value and writes its positional placeholder.
func (l *ScimFilterListenerN1QL) bind(value interface{}) {
	l.params = append(l.params, value)
	l.write("$" + strconv.Itoa(len(l.params)))
	l.prevOperation = ""
}

// FilterToN1QL translates a SCIM filter into the page and count queries for
// resourceName, together with the values to bind.
//
// A filter that does not parse is rejected rather than translated: ANTLR
// recovers from syntax errors and keeps walking, so an unchecked tree lets
// arbitrary tokens through into the query.
func FilterToN1QL(resourceName string, filter string, resolve NameResolver) (Query, error) {
	from := "FROM `" + escapeIdentifier(resourceName) + "`"
	if filter == "" {
		return Query{
			Page:  "SELECT * " + from,
			Count: "SELECT count(*) as count " + from,
		}, nil
	}

	tree, _, err := parseFilter(filter)
	if err != nil {
		return Query{}, err
	}

	listener := &ScimFilterListenerN1QL{resolve: resolve}
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	where := " WHERE " + string(listener.query)
	return Query{
		Page:   "SELECT * " + from + where,
		Count:  "SELECT count(*) as count " + from + where,
		Params: listener.params,
	}, nil
}

// VisitTerminal is called when a terminal node is visited.
func (l *ScimFilterListenerN1QL) VisitTerminal(node antlr.TerminalNode) {
	value := node.GetText()

	// Inside a quoted value nothing reaches the query: the text is collected
	// and bound once the value ends. The quotes themselves are delimiters.
	if l.inCriteria {
		if value != `"` {
			l.criteria.WriteString(value)
		}
		return
	}

	tokenType := node.GetSymbol().GetTokenType()
	if len(l.scopes) > 0 {
		top := l.scopes[len(l.scopes)-1]
		if node == top.attrNode || node == top.lbracNode || node == top.rbracNode {
			return
		}
		if top.trimLeadingWS && tokenType == ScimFilterParserWS {
			return
		}
		top.trimLeadingWS = false
	}

	switch tokenType {
	case ScimFilterParserATTRNAME:
		if len(l.scopes) > 0 {
			top := l.scopes[len(l.scopes)-1]
			fullPath := top.fullPath + "." + value
			resolvedFull := fullPath
			if l.resolve != nil {
				resolvedFull = l.resolve(fullPath)
			}
			rel := stripParentPrefix(resolvedFull, top.resolvedFull, value)
			value = "`" + top.varName + "`." + AddQuote(rel)
		} else {
			if l.resolve != nil {
				value = l.resolve(value)
			}
			value = AddQuote(value)
		}
		l.prevOperation = ""

	case ScimFilterParserNUMBERS:
		// Bound as a number, so the comparison stays numeric. parseFilter has
		// already verified that the token parses as a finite float64.
		number, err := strconv.ParseFloat(value, 64)
		if err == nil {
			l.bind(number)
		}
		return

	case ScimFilterParserBOOLEAN:
		l.bind(value == "true")
		return

	case ScimFilterLexerEQ:
		value = "="
		l.prevOperation = "eq"
	case ScimFilterLexerNE:
		value = "<>"
		l.prevOperation = "ne"
	case ScimFilterLexerCO:
		value = "LIKE"
		l.prevOperation = "co"
	case ScimFilterLexerSW:
		value = "LIKE"
		l.prevOperation = "sw"
	case ScimFilterLexerEW:
		value = "LIKE"
		l.prevOperation = "ew"
	case ScimFilterLexerGE:
		value = ">="
		l.prevOperation = "ge"
	case ScimFilterLexerGT:
		value = ">"
		l.prevOperation = "gt"
	case ScimFilterLexerLE:
		value = "<="
		l.prevOperation = "le"
	case ScimFilterLexerLT:
		value = "<"
		l.prevOperation = "lt"

	case ScimFilterLexerPR:
		// The grammar requires at least one WS token before PR, which is
		// emitted verbatim, so this value must not add its own leading space.
		value = "IS NOT NULL"
		l.prevOperation = "pr"

	case ScimFilterParserEOF:
		value = ""
	}

	l.write(value)
}

// urnPathPattern splits "urn:...:Schema.attr.sub" into its schema URN and the
// attribute path that follows it. Compiled once: it sits on the request path.
//
// Matched case-insensitively, because RFC 7643 2.1 makes schema URNs
// case-insensitive and a client writing "URN:IETF:..." is entitled to be
// understood. The captured text keeps the client's spelling; canonicalising it
// is the scim package's job.
var urnPathPattern = regexp.MustCompile(`(?i)^(urn[:\w\.\_]*)(:-*)?(:[\w]*)(\.)(.*)$`)

// SplitURNPath separates the schema URN prefix, when there is one, from the
// attribute path. It lives here because both the N1QL translation and the scim
// package's path handling need exactly this split, and scim already imports
// this package.
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
// getting here -- this is the second line of defence, not the first.
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
