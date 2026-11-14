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

// FilterOptions supplies optional schema callbacks to FilterToN1QLWithOptions.
type FilterOptions struct {
	Resolve       NameResolver
	IsCaseExact   func(path string) bool
	IsMultiValued func(path string) bool
}

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
	query          []byte
	params         []interface{}
	prevOperation  string
	resolve        NameResolver
	isCaseExact    func(path string) bool
	foldNextString bool

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

// EnterATTR_OPER_CRITERIA checks whether a string literal comparison targets a
// case-insensitive attribute so both sides can be folded with LOWER(...).
func (l *ScimFilterListenerN1QL) EnterATTR_OPER_CRITERIA(c *ATTR_OPER_CRITERIAContext) {
	l.foldNextString = false
	if l.isCaseExact == nil || c == nil || c.ATTRNAME() == nil || c.Operator() == nil {
		return
	}
	op := strings.ToLower(c.Operator().GetText())
	switch op {
	case "eq", "ne", "co", "sw", "ew":
		rawAttr := c.ATTRNAME().GetText()
		fullPath := rawAttr
		if len(l.scopes) > 0 {
			fullPath = l.scopes[len(l.scopes)-1].fullPath + "." + rawAttr
		}
		if !l.isCaseExact(fullPath) {
			l.foldNextString = true
		}
	}
}

// ExitATTR_OPER_CRITERIA resets case-folding state.
func (l *ScimFilterListenerN1QL) ExitATTR_OPER_CRITERIA(*ATTR_OPER_CRITERIAContext) {
	l.foldNextString = false
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
	l.bindString(value, l.foldNextString)
	l.foldNextString = false
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

func (l *ScimFilterListenerN1QL) bindString(value string, lower bool) {
	l.params = append(l.params, value)
	placeholder := "$" + strconv.Itoa(len(l.params))
	if lower {
		l.write("LOWER(" + placeholder + ")")
	} else {
		l.write(placeholder)
	}
	l.prevOperation = ""
}

// FilterToN1QL translates a SCIM filter into the page and count queries for
// resourceName, together with the values to bind.
//
// A filter that does not parse is rejected rather than translated: ANTLR
// recovers from syntax errors and keeps walking, so an unchecked tree lets
// arbitrary tokens through into the query.
func FilterToN1QL(resourceName string, filter string, resolve NameResolver) (Query, error) {
	return FilterToN1QLWithOptions(resourceName, filter, FilterOptions{Resolve: resolve})
}

// FilterToN1QLWithOptions translates a SCIM filter into N1QL using schema
// metadata for attribute canonicalisation, case-insensitivity, and multi-valued
// dotted sub-attribute expansion (e.g. emails.value eq "x").
func FilterToN1QLWithOptions(resourceName string, filter string, opts FilterOptions) (Query, error) {
	from := "FROM `" + escapeIdentifier(resourceName) + "`"
	if filter == "" {
		return Query{
			Page:  "SELECT * " + from,
			Count: "SELECT count(*) as count " + from,
		}, nil
	}

	expanded := filter
	if opts.IsMultiValued != nil {
		expanded = expandMultiValuedDottedPaths(filter, opts.IsMultiValued)
	}

	tree, _, err := parseFilter(expanded)
	if err != nil {
		return Query{}, err
	}

	listener := &ScimFilterListenerN1QL{
		resolve:     opts.Resolve,
		isCaseExact: opts.IsCaseExact,
	}
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	where := " WHERE " + string(listener.query)
	return Query{
		Page:   "SELECT * " + from + where,
		Count:  "SELECT count(*) as count " + from + where,
		Params: listener.params,
	}, nil
}

// expandMultiValuedDottedPaths rewrites top-level dotted comparisons on
// multi-valued complex attributes (such as `emails.value eq "x"` or
// `members.value pr`) into bracketed value filters (`emails[value eq "x"]`) so
// N1QL evaluates them with ANY ... SATISFIES ... END over the JSON array.
func expandMultiValuedDottedPaths(input string, isMultiValued func(path string) bool) string {
	if !strings.Contains(input, ".") {
		return input
	}
	var out strings.Builder
	out.Grow(len(input) + 16)
	inQuote := false
	bracketDepth := 0

	for i := 0; i < len(input); {
		ch := input[i]
		if ch == '"' {
			inQuote = !inQuote
			out.WriteByte(ch)
			i++
			continue
		}
		if !inQuote {
			if ch == '[' {
				bracketDepth++
				out.WriteByte(ch)
				i++
				continue
			}
			if ch == ']' {
				if bracketDepth > 0 {
					bracketDepth--
				}
				out.WriteByte(ch)
				i++
				continue
			}
			if bracketDepth == 0 && isAttrStartByte(ch) && (i == 0 || isFilterBoundary(input[i-1])) {
				rewritten, nextPos, ok := tryRewriteMultiValuedClause(input, i, isMultiValued)
				if ok {
					out.WriteString(rewritten)
					i = nextPos
					continue
				}
			}
		}
		out.WriteByte(ch)
		i++
	}
	return out.String()
}

func isAttrStartByte(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '$'
}

func isFilterBoundary(b byte) bool {
	switch b {
	case ' ', '\t', '\r', '\n', '(':
		return true
	default:
		return false
	}
}

func tryRewriteMultiValuedClause(s string, pos int, isMultiValued func(path string) bool) (string, int, bool) {
	attrStart := pos
	for pos < len(s) && isAttrPathByte(s[pos]) {
		pos++
	}
	attrToken := s[attrStart:pos]
	if pos < len(s) && s[pos] == '[' {
		return "", 0, false
	}
	lowerTok := strings.ToLower(attrToken)
	if lowerTok == "and" || lowerTok == "or" || lowerTok == "not" {
		return "", 0, false
	}

	mvAttr, subAttr, hasMV := splitMultiValuedAttrToken(attrToken, isMultiValued)
	if !hasMV {
		return "", 0, false
	}

	wsStart := pos
	for pos < len(s) && s[pos] == ' ' {
		pos++
	}
	if pos == wsStart {
		return "", 0, false
	}

	opStart := pos
	for pos < len(s) && isAlphaByte(s[pos]) {
		pos++
	}
	if pos == opStart {
		return "", 0, false
	}
	op := s[opStart:pos]

	if strings.EqualFold(op, "pr") {
		if pos < len(s) && !isFilterDelimiter(s[pos]) {
			return "", 0, false
		}
		return mvAttr + "[" + subAttr + " " + op + "]", pos, true
	}
	if !isComparisonOpWord(op) {
		return "", 0, false
	}

	ws2Start := pos
	for pos < len(s) && s[pos] == ' ' {
		pos++
	}
	if pos == ws2Start || pos >= len(s) {
		return "", 0, false
	}

	if s[pos] == '"' {
		endQuote := pos + 1
		for endQuote < len(s) && s[endQuote] != '"' {
			endQuote++
		}
		if endQuote >= len(s) {
			return "", 0, false
		}
		rhs := s[pos : endQuote+1]
		return mvAttr + "[" + subAttr + " " + op + " " + rhs + "]", endQuote + 1, true
	}

	rhsStart := pos
	for pos < len(s) && !isFilterDelimiter(s[pos]) {
		pos++
	}
	if pos == rhsStart {
		return "", 0, false
	}
	rhs := s[rhsStart:pos]
	return mvAttr + "[" + subAttr + " " + op + " " + rhs + "]", pos, true
}

func splitMultiValuedAttrToken(token string, isMultiValued func(path string) bool) (mvAttr, subAttr string, ok bool) {
	urn, attrPath := SplitURNPath(token)
	dotIdx := strings.IndexByte(attrPath, '.')
	if dotIdx <= 0 || dotIdx >= len(attrPath)-1 {
		return "", "", false
	}
	head := attrPath[:dotIdx]
	tail := attrPath[dotIdx+1:]
	candidate := head
	if urn != "" {
		candidate = urn + ":" + head
	}
	if !isMultiValued(candidate) {
		return "", "", false
	}
	if urn != "" {
		return urn + "." + head, tail, true
	}
	return head, tail, true
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
		if l.foldNextString {
			value = "LOWER(" + value + ")"
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

// urnVersionSchemaPattern splits RFC 7643/7644 schema-prefixed paths using
// either ":" (RFC 7644 3.4.2.2 / 3.10) or "." as the delimiter after the
// <version>:<SchemaName> suffix, e.g.
//
//	urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:employeeNumber
//	urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager.value
//	urn:ietf:params:scim:schemas:extension:one:2.0:Element.boolean
var urnVersionSchemaPattern = regexp.MustCompile(`(?i)^(urn:[a-z0-9_.\-]+(?::[a-z0-9_.\-]+)*:[0-9]+(?:\.[0-9]+)*:[a-z0-9_\-]+)[:\.](.+)$`)

// urnDotFallbackPattern handles unversioned URN prefixes separated by ".".
var urnDotFallbackPattern = regexp.MustCompile(`(?i)^(urn[:\w\.\_-]*)(:-*)?(:[\w]*)(\.)([^:]+)$`)

// SplitURNPath separates the schema URN prefix, when there is one, from the
// attribute path. It lives here because both the N1QL translation and the scim
// package's path handling need exactly this split, and scim already imports
// this package.
func SplitURNPath(path string) (urn string, attrPath string) {
	if !strings.HasPrefix(strings.ToLower(path), "urn:") {
		return "", path
	}
	if m := urnVersionSchemaPattern.FindStringSubmatch(path); len(m) == 3 {
		return m[1], m[2]
	}
	if urnDotFallbackPattern.MatchString(path) {
		return urnDotFallbackPattern.ReplaceAllString(path, `${1}${2}${3}`),
			urnDotFallbackPattern.ReplaceAllString(path, `${5}`)
	}
	if lastColon := strings.LastIndexByte(path, ':'); lastColon > 4 && lastColon < len(path)-1 {
		return path[:lastColon], path[lastColon+1:]
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
