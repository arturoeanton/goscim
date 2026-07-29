package parser

import (
	"fmt"
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
	p, collector := newFilterParser(filter)
	p.Start_()
	if len(collector.problems) > 0 {
		return &SyntaxError{Filter: filter, Problems: collector.problems}
	}
	return nil
}
