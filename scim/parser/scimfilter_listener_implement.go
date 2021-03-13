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

// FilterToN1QL is ...
func FilterToN1QL(resourceName string, filter string) string {
	query := "SELECT * FROM `" + resourceName + "`"
	if filter == "" {
		return query
	}
	is := antlr.NewInputStream(filter)
	lexer := NewScimFilterLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := NewScimFilterParser(stream)
	scimFilterListenerN1QL := ScimFilterListenerN1QL{query: query + " WHERE "}
	antlr.ParseTreeWalkerDefault.Walk(&scimFilterListenerN1QL, p.Start())
	return scimFilterListenerN1QL.query
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
					re := regexp.MustCompile(`^(urn[:\w\.\_]*)(:-*)?(:[\w]*)(\.)(.*)$`)
					urn := ""
					if re.MatchString(value) {
						urn = "`" + re.ReplaceAllString(value, `${1}${2}${3}`) + "`."
					}
					path := re.ReplaceAllString(value, `${5}`)
					path = urn + "`" + strings.Join(strings.Split(path, "."), "`.`") + "`"
					value = path
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
			value = ">"
			l.prevOperation = "ge"
		}
	case ScimFilterLexerGT:
		{
			value = ">="
			l.prevOperation = "gt"
		}
	case ScimFilterLexerLE:
		{
			value = "<"
			l.prevOperation = "le"
		}
	case ScimFilterLexerLT:
		{
			value = "<="
			l.prevOperation = "lt"
		}

	case ScimFilterLexerPR:
		{
			value = " IS NOT NULL"
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
