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

// treeValidator catches constructs that the permissive upstream grammar accepts
// without a syntax error, but that either are not valid SCIM filters (such as
// "--" matching [-.0-9]+, two adjacent expressions with no operator, or
// comparing an expression to an expression) or have no N1QL translation yet
// (complex value filters like emails[type eq "work"]).
type treeValidator struct {
	*BaseScimFilterListener
	inCriteria bool
	problems   []string
	attributes []string
}

func (v *treeValidator) EnterStart(ctx *StartContext) {
	if ctx != nil && len(ctx.AllExpression()) != 1 {
		v.problems = append(v.problems, "expected a single filter expression")
	}
}

func (v *treeValidator) EnterCriteria(*CriteriaContext) { v.inCriteria = true }
func (v *treeValidator) ExitCriteria(*CriteriaContext)  { v.inCriteria = false }

func (v *treeValidator) EnterLBRAC_EXPR_RBRAC(*LBRAC_EXPR_RBRACContext) {
	v.problems = append(v.problems, "complex value filters (attribute[...]) are not supported")
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
	p, collector := newFilterParser(filter)
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
