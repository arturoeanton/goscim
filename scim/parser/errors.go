package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

// SyntaxError reports that a SCIM filter could not be parsed. ANTLR recovers
// from syntax errors and keeps walking the tree, so without this the listener
// would happily emit the offending tokens straight into the generated N1QL.
type SyntaxError struct {
	Filter   string
	Problems []string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("invalid SCIM filter %q: %s", e.Filter, strings.Join(e.Problems, "; "))
}

// errorCollector records every syntax error reported by the lexer and parser.
type errorCollector struct {
	*antlr.DefaultErrorListener
	problems []string
}

func newErrorCollector() *errorCollector {
	return &errorCollector{DefaultErrorListener: antlr.NewDefaultErrorListener()}
}

func (c *errorCollector) SyntaxError(_ antlr.Recognizer, _ interface{}, line, column int, msg string, _ antlr.RecognitionException) {
	c.problems = append(c.problems, fmt.Sprintf("%d:%d %s", line, column, msg))
}

type validatorBracketScope struct {
	prefix   string
	attrNode antlr.TerminalNode
}

// treeValidator catches constructs that the permissive upstream grammar accepts
// without a syntax error, such as "--" matching [-.0-9]+, two adjacent
// expressions with no operator, or comparing an expression to an expression,
// while collecting full dotted paths for every referenced attribute (including
// sub-attributes inside attribute[...] value filters).
type treeValidator struct {
	*BaseScimFilterListener
	inCriteria bool
	problems   []string
	attributes []string
	scopes     []validatorBracketScope
}

func (v *treeValidator) EnterStart(ctx *StartContext) {
	if ctx != nil && len(ctx.AllExpression()) != 1 {
		v.problems = append(v.problems, "expected a single filter expression")
	}
}

func (v *treeValidator) EnterCriteria(*CriteriaContext) { v.inCriteria = true }
func (v *treeValidator) ExitCriteria(*CriteriaContext)  { v.inCriteria = false }

func (v *treeValidator) EnterLBRAC_EXPR_RBRAC(ctx *LBRAC_EXPR_RBRACContext) {
	if ctx == nil || ctx.ATTRNAME() == nil || ctx.Expression() == nil {
		v.problems = append(v.problems, "invalid complex value filter")
		return
	}
	rawAttr := ctx.ATTRNAME().GetText()
	fullPath := rawAttr
	if len(v.scopes) > 0 {
		fullPath = v.scopes[len(v.scopes)-1].prefix + "." + rawAttr
	}
	v.attributes = append(v.attributes, fullPath)
	v.scopes = append(v.scopes, validatorBracketScope{
		prefix:   fullPath,
		attrNode: ctx.ATTRNAME(),
	})
}

func (v *treeValidator) ExitLBRAC_EXPR_RBRAC(*LBRAC_EXPR_RBRACContext) {
	if len(v.scopes) > 0 {
		v.scopes = v.scopes[:len(v.scopes)-1]
	}
}

func (v *treeValidator) EnterATTR_OPER_EXPR(*ATTR_OPER_EXPRContext) {
	v.problems = append(v.problems, "right-hand side of comparison must be a literal value")
}

func (v *treeValidator) EnterEXPR_OPER_EXPR(*EXPR_OPER_EXPRContext) {
	v.problems = append(v.problems, "comparison operator requires an attribute path and a literal value")
}

func (v *treeValidator) VisitTerminal(node antlr.TerminalNode) {
	if v.inCriteria {
		return
	}
	switch node.GetSymbol().GetTokenType() {
	case ScimFilterParserATTRNAME:
		if len(v.scopes) > 0 {
			top := v.scopes[len(v.scopes)-1]
			if node == top.attrNode {
				return
			}
			v.attributes = append(v.attributes, top.prefix+"."+node.GetText())
			return
		}
		v.attributes = append(v.attributes, node.GetText())
	case ScimFilterParserNUMBERS:
		// The lexer rule [-.0-9]+ accepts "--" and "1.2.3". Reject anything
		// that is not a finite number so it can never fall through into N1QL.
		num, err := strconv.ParseFloat(node.GetText(), 64)
		if err != nil || math.IsNaN(num) || math.IsInf(num, 0) {
			v.problems = append(v.problems, fmt.Sprintf("invalid numeric literal %q", node.GetText()))
		}
	}
}

// parseFilter parses and validates a non-empty SCIM filter, returning its parse
// tree and the attribute paths it references.
func parseFilter(filter string) (IStartContext, []string, error) {
	normalized := rewriteBracketSubPaths(filter)
	p, collector := newFilterParser(normalized)
	tree := p.Start_()
	if len(collector.problems) > 0 {
		return nil, nil, &SyntaxError{Filter: filter, Problems: collector.problems}
	}
	validator := &treeValidator{}
	antlr.ParseTreeWalkerDefault.Walk(validator, tree)
	if len(validator.problems) > 0 {
		return nil, nil, &SyntaxError{Filter: filter, Problems: validator.problems}
	}
	return tree, validator.attributes, nil
}

// newFilterParser wires a parser for filter with the default error listeners
// replaced by a collector, so failures surface as values instead of stderr
// noise.
func newFilterParser(filter string) (*ScimFilterParser, *errorCollector) {
	collector := newErrorCollector()

	lexer := NewScimFilterLexer(antlr.NewInputStream(filter))
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(collector)

	p := NewScimFilterParser(antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel))
	p.RemoveErrorListeners()
	p.AddErrorListener(collector)

	return p, collector
}

// Validate reports whether filter is a syntactically valid SCIM filter
// expression. An empty filter is valid and selects everything.
func Validate(filter string) error {
	if filter == "" {
		return nil
	}
	_, _, err := parseFilter(filter)
	return err
}

// ReferencedAttributes parses filter and returns every attribute path it names.
func ReferencedAttributes(filter string) ([]string, error) {
	if filter == "" {
		return nil, nil
	}
	_, attrs, err := parseFilter(filter)
	return attrs, err
}

// rewriteBracketSubPaths normalises RFC 7644 3.4.2.2 valuePaths that carry a
// trailing sub-attribute after the closing bracket, e.g.
//
//	emails[type eq "work"].value co "@example.com"
//
// into the equivalent single-element conjunction
//
//	emails[(type eq "work") and value co "@example.com"]
//
// while leaving quoted string literals untouched.
func rewriteBracketSubPaths(input string) string {
	if !strings.Contains(input, "].") {
		return input
	}
	var out strings.Builder
	out.Grow(len(input) + 16)
	inQuote := false
	for i := 0; i < len(input); {
		ch := input[i]
		if ch == '"' {
			inQuote = !inQuote
			out.WriteByte(ch)
			i++
			continue
		}
		if !inQuote && ch == '[' {
			closeIdx := findMatchingBracket(input, i)
			if closeIdx < 0 {
				return input
			}
			inner := strings.TrimSpace(input[i+1 : closeIdx])
			if inner == "" {
				return input
			}
			innerRewritten := rewriteBracketSubPaths(inner)
			if closeIdx+1 < len(input) && input[closeIdx+1] == '.' {
				tailExpr, nextPos, ok := parseTrailingSubComparison(input, closeIdx+2)
				if !ok {
					return input
				}
				out.WriteString("[(")
				out.WriteString(innerRewritten)
				out.WriteString(") and ")
				out.WriteString(tailExpr)
				out.WriteByte(']')
				i = nextPos
				continue
			}
			out.WriteByte('[')
			out.WriteString(innerRewritten)
			out.WriteByte(']')
			i = closeIdx + 1
			continue
		}
		out.WriteByte(ch)
		i++
	}
	return out.String()
}

func findMatchingBracket(s string, openIdx int) int {
	depth := 0
	inQuote := false
	for i := openIdx; i < len(s); i++ {
		switch s[i] {
		case '"':
			inQuote = !inQuote
		case '[':
			if !inQuote {
				depth++
			}
		case ']':
			if !inQuote {
				depth--
				if depth == 0 {
					return i
				}
			}
		}
	}
	return -1
}

func parseTrailingSubComparison(s string, pos int) (expr string, nextPos int, ok bool) {
	subStart := pos
	for pos < len(s) && isAttrPathByte(s[pos]) {
		pos++
	}
	if pos == subStart {
		return "", 0, false
	}
	subPath := s[subStart:pos]

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
		return subPath + " " + op, pos, true
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
		return subPath + " " + op + " " + rhs, endQuote + 1, true
	}

	rhsStart := pos
	for pos < len(s) && !isFilterDelimiter(s[pos]) {
		pos++
	}
	if pos == rhsStart {
		return "", 0, false
	}
	rhs := s[rhsStart:pos]
	return subPath + " " + op + " " + rhs, pos, true
}

func isAttrPathByte(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') || b == '-' || b == '_' || b == '.' || b == ':' || b == '$'
}

func isAlphaByte(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isFilterDelimiter(b byte) bool {
	switch b {
	case ' ', '\t', '\r', '\n', ')', ']', '(', '[':
		return true
	default:
		return false
	}
}

func isComparisonOpWord(op string) bool {
	switch strings.ToLower(op) {
	case "eq", "ne", "co", "sw", "ew", "gt", "lt", "ge", "le":
		return true
	default:
		return false
	}
}
