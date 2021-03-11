package parser

import (
	"fmt"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type ScimFilterListenerN1QL struct {
	*BaseScimFilterListener
	stack         []string
	query         string
	prevOperation string
}

func FilterToN1QL(resourceName string, filter string) string {
	is := antlr.NewInputStream(filter)
	// Create the Lexer
	lexer := NewScimFilterLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	// Create the Parser
	p := NewScimFilterParser(stream)
	// Finally parse the expression
	scimFilterListenerN1QL := ScimFilterListenerN1QL{query: "SELECT * FROM `" + resourceName + "` WHERE "}
	antlr.ParseTreeWalkerDefault.Walk(&scimFilterListenerN1QL, p.Start())
	return scimFilterListenerN1QL.query
}

func (l *ScimFilterListenerN1QL) push(i string) {
	l.stack = append(l.stack, i)
}

func (l *ScimFilterListenerN1QL) pop() string {
	if len(l.stack) < 1 {
		panic("stack is empty unable to pop")
	}

	// Get the last value from the stack.
	result := l.stack[len(l.stack)-1]

	// Remove the last element from the stack.
	l.stack = l.stack[:len(l.stack)-1]

	return result
}

// VisitTerminal is called when a terminal node is visited.
func (s *ScimFilterListenerN1QL) VisitTerminal(node antlr.TerminalNode) {
	value := node.GetText()
	switch node.GetSymbol().GetTokenType() {
	case ScimFilterParserATTRNAME:
		{
			payload, ok := node.GetParent().GetPayload().(*antlr.BaseParserRuleContext)
			if ok {
				_, ok := payload.BaseRuleContext.GetParent().(*ATTR_OPER_CRITERIAContext)
				if !ok {
					values := strings.Split(value, ".")
					for i, v := range values {
						fmt.Println(v)
						values[i] = "`" + v + "`"

						if i < len(values)-1 {
							next := values[i+1]
							if strings.Contains(next, ":") {
								values[i] = "`" + v
							}
						}

						if i > 0 {
							prev := values[i-1]
							if strings.Contains(prev, ":") {
								values[i] = v + "`"
							}
						}
					}
					value = strings.Join(values, ".")

					if s.prevOperation == "pr" {
						value = value + " IS NOT NULL"
						// IS NOT NULL - returns rows which contain a value (not NULL or missing).
						// IS NOT MISSING - returns rows which contain a value or null.
						// IS VALUED - synonym for IS NOT NULL
					}
				} else {
					if s.prevOperation == "co" {
						value = "%" + value + "%"
					}
					if s.prevOperation == "sw" {
						value = value + "%"
					}
					if s.prevOperation == "ew" {
						value = "%" + value
					}
				}
			}
			s.prevOperation = ""
		}
	case ScimFilterLexerEQ:
		{
			value = "="
			s.prevOperation = "eq"
		}
	case ScimFilterLexerNE:
		{
			value = "<>"
			s.prevOperation = "ne"
		}
	case ScimFilterLexerCO:
		{
			value = "LIKE"
			s.prevOperation = "co"
		}
	case ScimFilterLexerSW:
		{
			value = "LIKE"
			s.prevOperation = "sw"
		}
	case ScimFilterLexerEW:
		{
			value = "LIKE"
			s.prevOperation = "ew"
		}
	case ScimFilterLexerGE:
		{
			value = ">"
			s.prevOperation = "ge"
		}
	case ScimFilterLexerGT:
		{
			value = ">="
			s.prevOperation = "gt"
		}
	case ScimFilterLexerLE:
		{
			value = "<"
			s.prevOperation = "le"
		}
	case ScimFilterLexerLT:
		{
			value = "<="
			s.prevOperation = "lt"
		}

	case ScimFilterLexerPR:
		{
			s.prevOperation = "pr"
		}
	case ScimFilterParserEOF:
		{
			value = ""
		}
	}

	s.query = s.query + value
}
