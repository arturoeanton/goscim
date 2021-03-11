// Generated from ScimFilter.g4 by ANTLR 4.7.

package parser // ScimFilter

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = reflect.Copy
var _ = strconv.Itoa

var parserATN = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 3, 26, 170,
	4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 3, 2, 7, 2,
	14, 10, 2, 12, 2, 14, 2, 17, 11, 2, 3, 2, 3, 2, 3, 3, 3, 3, 3, 3, 6, 3,
	24, 10, 3, 13, 3, 14, 3, 25, 3, 3, 3, 3, 3, 3, 6, 3, 31, 10, 3, 13, 3,
	14, 3, 32, 3, 3, 3, 3, 3, 3, 6, 3, 38, 10, 3, 13, 3, 14, 3, 39, 3, 3, 3,
	3, 6, 3, 44, 10, 3, 13, 3, 14, 3, 45, 3, 3, 3, 3, 3, 3, 3, 3, 6, 3, 52,
	10, 3, 13, 3, 14, 3, 53, 3, 3, 3, 3, 6, 3, 58, 10, 3, 13, 3, 14, 3, 59,
	3, 3, 3, 3, 3, 3, 3, 3, 6, 3, 66, 10, 3, 13, 3, 14, 3, 67, 3, 3, 3, 3,
	6, 3, 72, 10, 3, 13, 3, 14, 3, 73, 3, 3, 3, 3, 3, 3, 3, 3, 7, 3, 80, 10,
	3, 12, 3, 14, 3, 83, 11, 3, 3, 3, 3, 3, 7, 3, 87, 10, 3, 12, 3, 14, 3,
	90, 11, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 7, 3, 97, 10, 3, 12, 3, 14, 3,
	100, 11, 3, 3, 3, 3, 3, 7, 3, 104, 10, 3, 12, 3, 14, 3, 107, 11, 3, 3,
	3, 3, 3, 5, 3, 111, 10, 3, 3, 3, 3, 3, 6, 3, 115, 10, 3, 13, 3, 14, 3,
	116, 3, 3, 3, 3, 6, 3, 121, 10, 3, 13, 3, 14, 3, 122, 3, 3, 3, 3, 3, 3,
	6, 3, 128, 10, 3, 13, 3, 14, 3, 129, 3, 3, 3, 3, 6, 3, 134, 10, 3, 13,
	3, 14, 3, 135, 3, 3, 3, 3, 3, 3, 6, 3, 141, 10, 3, 13, 3, 14, 3, 142, 3,
	3, 3, 3, 6, 3, 147, 10, 3, 13, 3, 14, 3, 148, 3, 3, 3, 3, 7, 3, 153, 10,
	3, 12, 3, 14, 3, 156, 11, 3, 3, 4, 3, 4, 6, 4, 160, 10, 4, 13, 4, 14, 4,
	161, 3, 4, 3, 4, 3, 5, 3, 5, 3, 6, 3, 6, 3, 6, 20, 25, 32, 39, 45, 53,
	59, 67, 73, 81, 88, 98, 105, 116, 122, 129, 142, 148, 161, 3, 4, 7, 2,
	4, 6, 8, 10, 2, 4, 3, 2, 22, 23, 3, 2, 4, 12, 2, 193, 2, 15, 3, 2, 2, 2,
	4, 110, 3, 2, 2, 2, 6, 157, 3, 2, 2, 2, 8, 165, 3, 2, 2, 2, 10, 167, 3,
	2, 2, 2, 12, 14, 5, 4, 3, 2, 13, 12, 3, 2, 2, 2, 14, 17, 3, 2, 2, 2, 15,
	13, 3, 2, 2, 2, 15, 16, 3, 2, 2, 2, 16, 18, 3, 2, 2, 2, 17, 15, 3, 2, 2,
	2, 18, 19, 7, 2, 2, 3, 19, 3, 3, 2, 2, 2, 20, 21, 8, 3, 1, 2, 21, 23, 7,
	13, 2, 2, 22, 24, 7, 21, 2, 2, 23, 22, 3, 2, 2, 2, 24, 25, 3, 2, 2, 2,
	25, 26, 3, 2, 2, 2, 25, 23, 3, 2, 2, 2, 26, 27, 3, 2, 2, 2, 27, 111, 5,
	4, 3, 12, 28, 30, 7, 24, 2, 2, 29, 31, 7, 21, 2, 2, 30, 29, 3, 2, 2, 2,
	31, 32, 3, 2, 2, 2, 32, 33, 3, 2, 2, 2, 32, 30, 3, 2, 2, 2, 33, 34, 3,
	2, 2, 2, 34, 111, 7, 16, 2, 2, 35, 37, 7, 24, 2, 2, 36, 38, 7, 21, 2, 2,
	37, 36, 3, 2, 2, 2, 38, 39, 3, 2, 2, 2, 39, 40, 3, 2, 2, 2, 39, 37, 3,
	2, 2, 2, 40, 41, 3, 2, 2, 2, 41, 43, 5, 10, 6, 2, 42, 44, 7, 21, 2, 2,
	43, 42, 3, 2, 2, 2, 44, 45, 3, 2, 2, 2, 45, 46, 3, 2, 2, 2, 45, 43, 3,
	2, 2, 2, 46, 47, 3, 2, 2, 2, 47, 48, 5, 4, 3, 7, 48, 111, 3, 2, 2, 2, 49,
	51, 7, 24, 2, 2, 50, 52, 7, 21, 2, 2, 51, 50, 3, 2, 2, 2, 52, 53, 3, 2,
	2, 2, 53, 54, 3, 2, 2, 2, 53, 51, 3, 2, 2, 2, 54, 55, 3, 2, 2, 2, 55, 57,
	5, 10, 6, 2, 56, 58, 7, 21, 2, 2, 57, 56, 3, 2, 2, 2, 58, 59, 3, 2, 2,
	2, 59, 60, 3, 2, 2, 2, 59, 57, 3, 2, 2, 2, 60, 61, 3, 2, 2, 2, 61, 62,
	5, 6, 4, 2, 62, 111, 3, 2, 2, 2, 63, 65, 7, 24, 2, 2, 64, 66, 7, 21, 2,
	2, 65, 64, 3, 2, 2, 2, 66, 67, 3, 2, 2, 2, 67, 68, 3, 2, 2, 2, 67, 65,
	3, 2, 2, 2, 68, 69, 3, 2, 2, 2, 69, 71, 5, 10, 6, 2, 70, 72, 7, 21, 2,
	2, 71, 70, 3, 2, 2, 2, 72, 73, 3, 2, 2, 2, 73, 74, 3, 2, 2, 2, 73, 71,
	3, 2, 2, 2, 74, 75, 3, 2, 2, 2, 75, 76, 5, 8, 5, 2, 76, 111, 3, 2, 2, 2,
	77, 81, 7, 17, 2, 2, 78, 80, 7, 21, 2, 2, 79, 78, 3, 2, 2, 2, 80, 83, 3,
	2, 2, 2, 81, 82, 3, 2, 2, 2, 81, 79, 3, 2, 2, 2, 82, 84, 3, 2, 2, 2, 83,
	81, 3, 2, 2, 2, 84, 88, 5, 4, 3, 2, 85, 87, 7, 21, 2, 2, 86, 85, 3, 2,
	2, 2, 87, 90, 3, 2, 2, 2, 88, 89, 3, 2, 2, 2, 88, 86, 3, 2, 2, 2, 89, 91,
	3, 2, 2, 2, 90, 88, 3, 2, 2, 2, 91, 92, 7, 18, 2, 2, 92, 111, 3, 2, 2,
	2, 93, 94, 7, 24, 2, 2, 94, 98, 7, 19, 2, 2, 95, 97, 7, 21, 2, 2, 96, 95,
	3, 2, 2, 2, 97, 100, 3, 2, 2, 2, 98, 99, 3, 2, 2, 2, 98, 96, 3, 2, 2, 2,
	99, 101, 3, 2, 2, 2, 100, 98, 3, 2, 2, 2, 101, 105, 5, 4, 3, 2, 102, 104,
	7, 21, 2, 2, 103, 102, 3, 2, 2, 2, 104, 107, 3, 2, 2, 2, 105, 106, 3, 2,
	2, 2, 105, 103, 3, 2, 2, 2, 106, 108, 3, 2, 2, 2, 107, 105, 3, 2, 2, 2,
	108, 109, 7, 20, 2, 2, 109, 111, 3, 2, 2, 2, 110, 20, 3, 2, 2, 2, 110,
	28, 3, 2, 2, 2, 110, 35, 3, 2, 2, 2, 110, 49, 3, 2, 2, 2, 110, 63, 3, 2,
	2, 2, 110, 77, 3, 2, 2, 2, 110, 93, 3, 2, 2, 2, 111, 154, 3, 2, 2, 2, 112,
	114, 12, 11, 2, 2, 113, 115, 7, 21, 2, 2, 114, 113, 3, 2, 2, 2, 115, 116,
	3, 2, 2, 2, 116, 117, 3, 2, 2, 2, 116, 114, 3, 2, 2, 2, 117, 118, 3, 2,
	2, 2, 118, 120, 7, 14, 2, 2, 119, 121, 7, 21, 2, 2, 120, 119, 3, 2, 2,
	2, 121, 122, 3, 2, 2, 2, 122, 123, 3, 2, 2, 2, 122, 120, 3, 2, 2, 2, 123,
	124, 3, 2, 2, 2, 124, 153, 5, 4, 3, 12, 125, 127, 12, 10, 2, 2, 126, 128,
	7, 21, 2, 2, 127, 126, 3, 2, 2, 2, 128, 129, 3, 2, 2, 2, 129, 130, 3, 2,
	2, 2, 129, 127, 3, 2, 2, 2, 130, 131, 3, 2, 2, 2, 131, 133, 7, 15, 2, 2,
	132, 134, 7, 21, 2, 2, 133, 132, 3, 2, 2, 2, 134, 135, 3, 2, 2, 2, 135,
	133, 3, 2, 2, 2, 135, 136, 3, 2, 2, 2, 136, 137, 3, 2, 2, 2, 137, 153,
	5, 4, 3, 11, 138, 140, 12, 9, 2, 2, 139, 141, 7, 21, 2, 2, 140, 139, 3,
	2, 2, 2, 141, 142, 3, 2, 2, 2, 142, 143, 3, 2, 2, 2, 142, 140, 3, 2, 2,
	2, 143, 144, 3, 2, 2, 2, 144, 146, 5, 10, 6, 2, 145, 147, 7, 21, 2, 2,
	146, 145, 3, 2, 2, 2, 147, 148, 3, 2, 2, 2, 148, 149, 3, 2, 2, 2, 148,
	146, 3, 2, 2, 2, 149, 150, 3, 2, 2, 2, 150, 151, 5, 4, 3, 10, 151, 153,
	3, 2, 2, 2, 152, 112, 3, 2, 2, 2, 152, 125, 3, 2, 2, 2, 152, 138, 3, 2,
	2, 2, 153, 156, 3, 2, 2, 2, 154, 152, 3, 2, 2, 2, 154, 155, 3, 2, 2, 2,
	155, 5, 3, 2, 2, 2, 156, 154, 3, 2, 2, 2, 157, 159, 7, 3, 2, 2, 158, 160,
	11, 2, 2, 2, 159, 158, 3, 2, 2, 2, 160, 161, 3, 2, 2, 2, 161, 162, 3, 2,
	2, 2, 161, 159, 3, 2, 2, 2, 162, 163, 3, 2, 2, 2, 163, 164, 7, 3, 2, 2,
	164, 7, 3, 2, 2, 2, 165, 166, 9, 2, 2, 2, 166, 9, 3, 2, 2, 2, 167, 168,
	9, 3, 2, 2, 168, 11, 3, 2, 2, 2, 25, 15, 25, 32, 39, 45, 53, 59, 67, 73,
	81, 88, 98, 105, 110, 116, 122, 129, 135, 142, 148, 152, 154, 161,
}
var deserializer = antlr.NewATNDeserializer(nil)
var deserializedATN = deserializer.DeserializeFromUInt16(parserATN)

var literalNames = []string{
	"", "'\"'", "", "", "", "", "", "", "", "", "", "", "", "", "", "'('",
	"')'", "'['", "']'", "' '",
}
var symbolicNames = []string{
	"", "", "EQ", "NE", "CO", "SW", "EW", "GT", "LT", "GE", "LE", "NOT", "AND",
	"OR", "PR", "LPAREN", "RPAREN", "LBRAC", "RBRAC", "WS", "NUMBERS", "BOOLEAN",
	"ATTRNAME", "ANY", "EOL",
}

var ruleNames = []string{
	"start", "expression", "criteria", "criteriaValue", "operator",
}
var decisionToDFA = make([]*antlr.DFA, len(deserializedATN.DecisionToState))

func init() {
	for index, ds := range deserializedATN.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(ds, index)
	}
}

type ScimFilterParser struct {
	*antlr.BaseParser
}

func NewScimFilterParser(input antlr.TokenStream) *ScimFilterParser {
	this := new(ScimFilterParser)

	this.BaseParser = antlr.NewBaseParser(input)

	this.Interpreter = antlr.NewParserATNSimulator(this, deserializedATN, decisionToDFA, antlr.NewPredictionContextCache())
	this.RuleNames = ruleNames
	this.LiteralNames = literalNames
	this.SymbolicNames = symbolicNames
	this.GrammarFileName = "ScimFilter.g4"

	return this
}

// ScimFilterParser tokens.
const (
	ScimFilterParserEOF      = antlr.TokenEOF
	ScimFilterParserT__0     = 1
	ScimFilterParserEQ       = 2
	ScimFilterParserNE       = 3
	ScimFilterParserCO       = 4
	ScimFilterParserSW       = 5
	ScimFilterParserEW       = 6
	ScimFilterParserGT       = 7
	ScimFilterParserLT       = 8
	ScimFilterParserGE       = 9
	ScimFilterParserLE       = 10
	ScimFilterParserNOT      = 11
	ScimFilterParserAND      = 12
	ScimFilterParserOR       = 13
	ScimFilterParserPR       = 14
	ScimFilterParserLPAREN   = 15
	ScimFilterParserRPAREN   = 16
	ScimFilterParserLBRAC    = 17
	ScimFilterParserRBRAC    = 18
	ScimFilterParserWS       = 19
	ScimFilterParserNUMBERS  = 20
	ScimFilterParserBOOLEAN  = 21
	ScimFilterParserATTRNAME = 22
	ScimFilterParserANY      = 23
	ScimFilterParserEOL      = 24
)

// ScimFilterParser rules.
const (
	ScimFilterParserRULE_start         = 0
	ScimFilterParserRULE_expression    = 1
	ScimFilterParserRULE_criteria      = 2
	ScimFilterParserRULE_criteriaValue = 3
	ScimFilterParserRULE_operator      = 4
)

// IStartContext is an interface to support dynamic dispatch.
type IStartContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsStartContext differentiates from other interfaces.
	IsStartContext()
}

type StartContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStartContext() *StartContext {
	var p = new(StartContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = ScimFilterParserRULE_start
	return p
}

func (*StartContext) IsStartContext() {}

func NewStartContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StartContext {
	var p = new(StartContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = ScimFilterParserRULE_start

	return p
}

func (s *StartContext) GetParser() antlr.Parser { return s.parser }

func (s *StartContext) EOF() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserEOF, 0)
}

func (s *StartContext) AllExpression() []IExpressionContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IExpressionContext)(nil)).Elem())
	var tst = make([]IExpressionContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IExpressionContext)
		}
	}

	return tst
}

func (s *StartContext) Expression(i int) IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *StartContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StartContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *StartContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.EnterStart(s)
	}
}

func (s *StartContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.ExitStart(s)
	}
}

func (p *ScimFilterParser) Start() (localctx IStartContext) {
	localctx = NewStartContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, ScimFilterParserRULE_start)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	p.SetState(13)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<ScimFilterParserNOT)|(1<<ScimFilterParserLPAREN)|(1<<ScimFilterParserATTRNAME))) != 0 {
		{
			p.SetState(10)
			p.expression(0)
		}

		p.SetState(15)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(16)
		p.Match(ScimFilterParserEOF)
	}

	return localctx
}

// IExpressionContext is an interface to support dynamic dispatch.
type IExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsExpressionContext differentiates from other interfaces.
	IsExpressionContext()
}

type ExpressionContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpressionContext() *ExpressionContext {
	var p = new(ExpressionContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = ScimFilterParserRULE_expression
	return p
}

func (*ExpressionContext) IsExpressionContext() {}

func NewExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionContext {
	var p = new(ExpressionContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = ScimFilterParserRULE_expression

	return p
}

func (s *ExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionContext) CopyFrom(ctx *ExpressionContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type ATTR_PRContext struct {
	*ExpressionContext
}

func NewATTR_PRContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ATTR_PRContext {
	var p = new(ATTR_PRContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *ATTR_PRContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ATTR_PRContext) ATTRNAME() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserATTRNAME, 0)
}

func (s *ATTR_PRContext) PR() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserPR, 0)
}

func (s *ATTR_PRContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(ScimFilterParserWS)
}

func (s *ATTR_PRContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(ScimFilterParserWS, i)
}

func (s *ATTR_PRContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.EnterATTR_PR(s)
	}
}

func (s *ATTR_PRContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.ExitATTR_PR(s)
	}
}

type LBRAC_EXPR_RBRACContext struct {
	*ExpressionContext
}

func NewLBRAC_EXPR_RBRACContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LBRAC_EXPR_RBRACContext {
	var p = new(LBRAC_EXPR_RBRACContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *LBRAC_EXPR_RBRACContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LBRAC_EXPR_RBRACContext) ATTRNAME() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserATTRNAME, 0)
}

func (s *LBRAC_EXPR_RBRACContext) LBRAC() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserLBRAC, 0)
}

func (s *LBRAC_EXPR_RBRACContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *LBRAC_EXPR_RBRACContext) RBRAC() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserRBRAC, 0)
}

func (s *LBRAC_EXPR_RBRACContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(ScimFilterParserWS)
}

func (s *LBRAC_EXPR_RBRACContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(ScimFilterParserWS, i)
}

func (s *LBRAC_EXPR_RBRACContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.EnterLBRAC_EXPR_RBRAC(s)
	}
}

func (s *LBRAC_EXPR_RBRACContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.ExitLBRAC_EXPR_RBRAC(s)
	}
}

type ATTR_OPER_EXPRContext struct {
	*ExpressionContext
}

func NewATTR_OPER_EXPRContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ATTR_OPER_EXPRContext {
	var p = new(ATTR_OPER_EXPRContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *ATTR_OPER_EXPRContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ATTR_OPER_EXPRContext) ATTRNAME() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserATTRNAME, 0)
}

func (s *ATTR_OPER_EXPRContext) Operator() IOperatorContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IOperatorContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IOperatorContext)
}

func (s *ATTR_OPER_EXPRContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ATTR_OPER_EXPRContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(ScimFilterParserWS)
}

func (s *ATTR_OPER_EXPRContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(ScimFilterParserWS, i)
}

func (s *ATTR_OPER_EXPRContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.EnterATTR_OPER_EXPR(s)
	}
}

func (s *ATTR_OPER_EXPRContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.ExitATTR_OPER_EXPR(s)
	}
}

type EXPR_OR_EXPRContext struct {
	*ExpressionContext
}

func NewEXPR_OR_EXPRContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *EXPR_OR_EXPRContext {
	var p = new(EXPR_OR_EXPRContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *EXPR_OR_EXPRContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *EXPR_OR_EXPRContext) AllExpression() []IExpressionContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IExpressionContext)(nil)).Elem())
	var tst = make([]IExpressionContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IExpressionContext)
		}
	}

	return tst
}

func (s *EXPR_OR_EXPRContext) Expression(i int) IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *EXPR_OR_EXPRContext) OR() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserOR, 0)
}

func (s *EXPR_OR_EXPRContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(ScimFilterParserWS)
}

func (s *EXPR_OR_EXPRContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(ScimFilterParserWS, i)
}

func (s *EXPR_OR_EXPRContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.EnterEXPR_OR_EXPR(s)
	}
}

func (s *EXPR_OR_EXPRContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.ExitEXPR_OR_EXPR(s)
	}
}

type EXPR_OPER_EXPRContext struct {
	*ExpressionContext
}

func NewEXPR_OPER_EXPRContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *EXPR_OPER_EXPRContext {
	var p = new(EXPR_OPER_EXPRContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *EXPR_OPER_EXPRContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *EXPR_OPER_EXPRContext) AllExpression() []IExpressionContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IExpressionContext)(nil)).Elem())
	var tst = make([]IExpressionContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IExpressionContext)
		}
	}

	return tst
}

func (s *EXPR_OPER_EXPRContext) Expression(i int) IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *EXPR_OPER_EXPRContext) Operator() IOperatorContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IOperatorContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IOperatorContext)
}

func (s *EXPR_OPER_EXPRContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(ScimFilterParserWS)
}

func (s *EXPR_OPER_EXPRContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(ScimFilterParserWS, i)
}

func (s *EXPR_OPER_EXPRContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.EnterEXPR_OPER_EXPR(s)
	}
}

func (s *EXPR_OPER_EXPRContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.ExitEXPR_OPER_EXPR(s)
	}
}

type NOT_EXPRContext struct {
	*ExpressionContext
}

func NewNOT_EXPRContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *NOT_EXPRContext {
	var p = new(NOT_EXPRContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *NOT_EXPRContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NOT_EXPRContext) NOT() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserNOT, 0)
}

func (s *NOT_EXPRContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *NOT_EXPRContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(ScimFilterParserWS)
}

func (s *NOT_EXPRContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(ScimFilterParserWS, i)
}

func (s *NOT_EXPRContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.EnterNOT_EXPR(s)
	}
}

func (s *NOT_EXPRContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.ExitNOT_EXPR(s)
	}
}

type EXPR_AND_EXPRContext struct {
	*ExpressionContext
}

func NewEXPR_AND_EXPRContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *EXPR_AND_EXPRContext {
	var p = new(EXPR_AND_EXPRContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *EXPR_AND_EXPRContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *EXPR_AND_EXPRContext) AllExpression() []IExpressionContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IExpressionContext)(nil)).Elem())
	var tst = make([]IExpressionContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IExpressionContext)
		}
	}

	return tst
}

func (s *EXPR_AND_EXPRContext) Expression(i int) IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *EXPR_AND_EXPRContext) AND() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserAND, 0)
}

func (s *EXPR_AND_EXPRContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(ScimFilterParserWS)
}

func (s *EXPR_AND_EXPRContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(ScimFilterParserWS, i)
}

func (s *EXPR_AND_EXPRContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.EnterEXPR_AND_EXPR(s)
	}
}

func (s *EXPR_AND_EXPRContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.ExitEXPR_AND_EXPR(s)
	}
}

type ATTR_OPER_VALUEContext struct {
	*ExpressionContext
}

func NewATTR_OPER_VALUEContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ATTR_OPER_VALUEContext {
	var p = new(ATTR_OPER_VALUEContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *ATTR_OPER_VALUEContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ATTR_OPER_VALUEContext) ATTRNAME() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserATTRNAME, 0)
}

func (s *ATTR_OPER_VALUEContext) Operator() IOperatorContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IOperatorContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IOperatorContext)
}

func (s *ATTR_OPER_VALUEContext) CriteriaValue() ICriteriaValueContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ICriteriaValueContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ICriteriaValueContext)
}

func (s *ATTR_OPER_VALUEContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(ScimFilterParserWS)
}

func (s *ATTR_OPER_VALUEContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(ScimFilterParserWS, i)
}

func (s *ATTR_OPER_VALUEContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.EnterATTR_OPER_VALUE(s)
	}
}

func (s *ATTR_OPER_VALUEContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.ExitATTR_OPER_VALUE(s)
	}
}

type ATTR_OPER_CRITERIAContext struct {
	*ExpressionContext
}

func NewATTR_OPER_CRITERIAContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ATTR_OPER_CRITERIAContext {
	var p = new(ATTR_OPER_CRITERIAContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *ATTR_OPER_CRITERIAContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ATTR_OPER_CRITERIAContext) ATTRNAME() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserATTRNAME, 0)
}

func (s *ATTR_OPER_CRITERIAContext) Operator() IOperatorContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IOperatorContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IOperatorContext)
}

func (s *ATTR_OPER_CRITERIAContext) Criteria() ICriteriaContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ICriteriaContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ICriteriaContext)
}

func (s *ATTR_OPER_CRITERIAContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(ScimFilterParserWS)
}

func (s *ATTR_OPER_CRITERIAContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(ScimFilterParserWS, i)
}

func (s *ATTR_OPER_CRITERIAContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.EnterATTR_OPER_CRITERIA(s)
	}
}

func (s *ATTR_OPER_CRITERIAContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.ExitATTR_OPER_CRITERIA(s)
	}
}

type LPAREN_EXPR_RPARENContext struct {
	*ExpressionContext
}

func NewLPAREN_EXPR_RPARENContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LPAREN_EXPR_RPARENContext {
	var p = new(LPAREN_EXPR_RPARENContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *LPAREN_EXPR_RPARENContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LPAREN_EXPR_RPARENContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserLPAREN, 0)
}

func (s *LPAREN_EXPR_RPARENContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *LPAREN_EXPR_RPARENContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserRPAREN, 0)
}

func (s *LPAREN_EXPR_RPARENContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(ScimFilterParserWS)
}

func (s *LPAREN_EXPR_RPARENContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(ScimFilterParserWS, i)
}

func (s *LPAREN_EXPR_RPARENContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.EnterLPAREN_EXPR_RPAREN(s)
	}
}

func (s *LPAREN_EXPR_RPARENContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.ExitLPAREN_EXPR_RPAREN(s)
	}
}

func (p *ScimFilterParser) Expression() (localctx IExpressionContext) {
	return p.expression(0)
}

func (p *ScimFilterParser) expression(_p int) (localctx IExpressionContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()
	_parentState := p.GetState()
	localctx = NewExpressionContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IExpressionContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 2
	p.EnterRecursionRule(localctx, 2, ScimFilterParserRULE_expression, _p)
	var _la int

	defer func() {
		p.UnrollRecursionContexts(_parentctx)
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(108)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 13, p.GetParserRuleContext()) {
	case 1:
		localctx = NewNOT_EXPRContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(19)
			p.Match(ScimFilterParserNOT)
		}
		p.SetState(21)
		p.GetErrorHandler().Sync(p)
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(20)
					p.Match(ScimFilterParserWS)
				}

			default:
				panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			}

			p.SetState(23)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 1, p.GetParserRuleContext())
		}
		{
			p.SetState(25)
			p.expression(10)
		}

	case 2:
		localctx = NewATTR_PRContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(26)
			p.Match(ScimFilterParserATTRNAME)
		}
		p.SetState(28)
		p.GetErrorHandler().Sync(p)
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(27)
					p.Match(ScimFilterParserWS)
				}

			default:
				panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			}

			p.SetState(30)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 2, p.GetParserRuleContext())
		}
		{
			p.SetState(32)
			p.Match(ScimFilterParserPR)
		}

	case 3:
		localctx = NewATTR_OPER_EXPRContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(33)
			p.Match(ScimFilterParserATTRNAME)
		}
		p.SetState(35)
		p.GetErrorHandler().Sync(p)
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(34)
					p.Match(ScimFilterParserWS)
				}

			default:
				panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			}

			p.SetState(37)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 3, p.GetParserRuleContext())
		}
		{
			p.SetState(39)
			p.Operator()
		}
		p.SetState(41)
		p.GetErrorHandler().Sync(p)
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(40)
					p.Match(ScimFilterParserWS)
				}

			default:
				panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			}

			p.SetState(43)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 4, p.GetParserRuleContext())
		}
		{
			p.SetState(45)
			p.expression(5)
		}

	case 4:
		localctx = NewATTR_OPER_CRITERIAContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(47)
			p.Match(ScimFilterParserATTRNAME)
		}
		p.SetState(49)
		p.GetErrorHandler().Sync(p)
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(48)
					p.Match(ScimFilterParserWS)
				}

			default:
				panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			}

			p.SetState(51)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 5, p.GetParserRuleContext())
		}
		{
			p.SetState(53)
			p.Operator()
		}
		p.SetState(55)
		p.GetErrorHandler().Sync(p)
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(54)
					p.Match(ScimFilterParserWS)
				}

			default:
				panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			}

			p.SetState(57)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 6, p.GetParserRuleContext())
		}
		{
			p.SetState(59)
			p.Criteria()
		}

	case 5:
		localctx = NewATTR_OPER_VALUEContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(61)
			p.Match(ScimFilterParserATTRNAME)
		}
		p.SetState(63)
		p.GetErrorHandler().Sync(p)
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(62)
					p.Match(ScimFilterParserWS)
				}

			default:
				panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			}

			p.SetState(65)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 7, p.GetParserRuleContext())
		}
		{
			p.SetState(67)
			p.Operator()
		}
		p.SetState(69)
		p.GetErrorHandler().Sync(p)
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(68)
					p.Match(ScimFilterParserWS)
				}

			default:
				panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			}

			p.SetState(71)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 8, p.GetParserRuleContext())
		}
		{
			p.SetState(73)
			p.CriteriaValue()
		}

	case 6:
		localctx = NewLPAREN_EXPR_RPARENContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(75)
			p.Match(ScimFilterParserLPAREN)
		}
		p.SetState(79)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 9, p.GetParserRuleContext())

		for _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1+1 {
				{
					p.SetState(76)
					p.Match(ScimFilterParserWS)
				}

			}
			p.SetState(81)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 9, p.GetParserRuleContext())
		}
		{
			p.SetState(82)
			p.expression(0)
		}
		p.SetState(86)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 10, p.GetParserRuleContext())

		for _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1+1 {
				{
					p.SetState(83)
					p.Match(ScimFilterParserWS)
				}

			}
			p.SetState(88)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 10, p.GetParserRuleContext())
		}
		{
			p.SetState(89)
			p.Match(ScimFilterParserRPAREN)
		}

	case 7:
		localctx = NewLBRAC_EXPR_RBRACContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(91)
			p.Match(ScimFilterParserATTRNAME)
		}
		{
			p.SetState(92)
			p.Match(ScimFilterParserLBRAC)
		}
		p.SetState(96)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 11, p.GetParserRuleContext())

		for _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1+1 {
				{
					p.SetState(93)
					p.Match(ScimFilterParserWS)
				}

			}
			p.SetState(98)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 11, p.GetParserRuleContext())
		}
		{
			p.SetState(99)
			p.expression(0)
		}
		p.SetState(103)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 12, p.GetParserRuleContext())

		for _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1+1 {
				{
					p.SetState(100)
					p.Match(ScimFilterParserWS)
				}

			}
			p.SetState(105)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 12, p.GetParserRuleContext())
		}
		{
			p.SetState(106)
			p.Match(ScimFilterParserRBRAC)
		}

	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(152)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 21, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(150)
			p.GetErrorHandler().Sync(p)
			switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 20, p.GetParserRuleContext()) {
			case 1:
				localctx = NewEXPR_AND_EXPRContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ScimFilterParserRULE_expression)
				p.SetState(110)

				if !(p.Precpred(p.GetParserRuleContext(), 9)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 9)", ""))
				}
				p.SetState(112)
				p.GetErrorHandler().Sync(p)
				_alt = 1 + 1
				for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
					switch _alt {
					case 1 + 1:
						{
							p.SetState(111)
							p.Match(ScimFilterParserWS)
						}

					default:
						panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
					}

					p.SetState(114)
					p.GetErrorHandler().Sync(p)
					_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 14, p.GetParserRuleContext())
				}
				{
					p.SetState(116)
					p.Match(ScimFilterParserAND)
				}
				p.SetState(118)
				p.GetErrorHandler().Sync(p)
				_alt = 1 + 1
				for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
					switch _alt {
					case 1 + 1:
						{
							p.SetState(117)
							p.Match(ScimFilterParserWS)
						}

					default:
						panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
					}

					p.SetState(120)
					p.GetErrorHandler().Sync(p)
					_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 15, p.GetParserRuleContext())
				}
				{
					p.SetState(122)
					p.expression(10)
				}

			case 2:
				localctx = NewEXPR_OR_EXPRContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ScimFilterParserRULE_expression)
				p.SetState(123)

				if !(p.Precpred(p.GetParserRuleContext(), 8)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 8)", ""))
				}
				p.SetState(125)
				p.GetErrorHandler().Sync(p)
				_alt = 1 + 1
				for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
					switch _alt {
					case 1 + 1:
						{
							p.SetState(124)
							p.Match(ScimFilterParserWS)
						}

					default:
						panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
					}

					p.SetState(127)
					p.GetErrorHandler().Sync(p)
					_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 16, p.GetParserRuleContext())
				}
				{
					p.SetState(129)
					p.Match(ScimFilterParserOR)
				}
				p.SetState(131)
				p.GetErrorHandler().Sync(p)
				_la = p.GetTokenStream().LA(1)

				for ok := true; ok; ok = _la == ScimFilterParserWS {
					{
						p.SetState(130)
						p.Match(ScimFilterParserWS)
					}

					p.SetState(133)
					p.GetErrorHandler().Sync(p)
					_la = p.GetTokenStream().LA(1)
				}
				{
					p.SetState(135)
					p.expression(9)
				}

			case 3:
				localctx = NewEXPR_OPER_EXPRContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ScimFilterParserRULE_expression)
				p.SetState(136)

				if !(p.Precpred(p.GetParserRuleContext(), 7)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 7)", ""))
				}
				p.SetState(138)
				p.GetErrorHandler().Sync(p)
				_alt = 1 + 1
				for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
					switch _alt {
					case 1 + 1:
						{
							p.SetState(137)
							p.Match(ScimFilterParserWS)
						}

					default:
						panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
					}

					p.SetState(140)
					p.GetErrorHandler().Sync(p)
					_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 18, p.GetParserRuleContext())
				}
				{
					p.SetState(142)
					p.Operator()
				}
				p.SetState(144)
				p.GetErrorHandler().Sync(p)
				_alt = 1 + 1
				for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
					switch _alt {
					case 1 + 1:
						{
							p.SetState(143)
							p.Match(ScimFilterParserWS)
						}

					default:
						panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
					}

					p.SetState(146)
					p.GetErrorHandler().Sync(p)
					_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 19, p.GetParserRuleContext())
				}
				{
					p.SetState(148)
					p.expression(8)
				}

			}

		}
		p.SetState(154)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 21, p.GetParserRuleContext())
	}

	return localctx
}

// ICriteriaContext is an interface to support dynamic dispatch.
type ICriteriaContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsCriteriaContext differentiates from other interfaces.
	IsCriteriaContext()
}

type CriteriaContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCriteriaContext() *CriteriaContext {
	var p = new(CriteriaContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = ScimFilterParserRULE_criteria
	return p
}

func (*CriteriaContext) IsCriteriaContext() {}

func NewCriteriaContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CriteriaContext {
	var p = new(CriteriaContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = ScimFilterParserRULE_criteria

	return p
}

func (s *CriteriaContext) GetParser() antlr.Parser { return s.parser }
func (s *CriteriaContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CriteriaContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CriteriaContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.EnterCriteria(s)
	}
}

func (s *CriteriaContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.ExitCriteria(s)
	}
}

func (p *ScimFilterParser) Criteria() (localctx ICriteriaContext) {
	localctx = NewCriteriaContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, ScimFilterParserRULE_criteria)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(155)
		p.Match(ScimFilterParserT__0)
	}
	p.SetState(157)
	p.GetErrorHandler().Sync(p)
	_alt = 1 + 1
	for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
		switch _alt {
		case 1 + 1:
			p.SetState(156)
			p.MatchWildcard()

		default:
			panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		}

		p.SetState(159)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 22, p.GetParserRuleContext())
	}
	{
		p.SetState(161)
		p.Match(ScimFilterParserT__0)
	}

	return localctx
}

// ICriteriaValueContext is an interface to support dynamic dispatch.
type ICriteriaValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsCriteriaValueContext differentiates from other interfaces.
	IsCriteriaValueContext()
}

type CriteriaValueContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCriteriaValueContext() *CriteriaValueContext {
	var p = new(CriteriaValueContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = ScimFilterParserRULE_criteriaValue
	return p
}

func (*CriteriaValueContext) IsCriteriaValueContext() {}

func NewCriteriaValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CriteriaValueContext {
	var p = new(CriteriaValueContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = ScimFilterParserRULE_criteriaValue

	return p
}

func (s *CriteriaValueContext) GetParser() antlr.Parser { return s.parser }

func (s *CriteriaValueContext) NUMBERS() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserNUMBERS, 0)
}

func (s *CriteriaValueContext) BOOLEAN() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserBOOLEAN, 0)
}

func (s *CriteriaValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CriteriaValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CriteriaValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.EnterCriteriaValue(s)
	}
}

func (s *CriteriaValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.ExitCriteriaValue(s)
	}
}

func (p *ScimFilterParser) CriteriaValue() (localctx ICriteriaValueContext) {
	localctx = NewCriteriaValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, ScimFilterParserRULE_criteriaValue)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	p.SetState(163)
	_la = p.GetTokenStream().LA(1)

	if !(_la == ScimFilterParserNUMBERS || _la == ScimFilterParserBOOLEAN) {
		p.GetErrorHandler().RecoverInline(p)
	} else {
		p.GetErrorHandler().ReportMatch(p)
		p.Consume()
	}

	return localctx
}

// IOperatorContext is an interface to support dynamic dispatch.
type IOperatorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsOperatorContext differentiates from other interfaces.
	IsOperatorContext()
}

type OperatorContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOperatorContext() *OperatorContext {
	var p = new(OperatorContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = ScimFilterParserRULE_operator
	return p
}

func (*OperatorContext) IsOperatorContext() {}

func NewOperatorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OperatorContext {
	var p = new(OperatorContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = ScimFilterParserRULE_operator

	return p
}

func (s *OperatorContext) GetParser() antlr.Parser { return s.parser }

func (s *OperatorContext) EQ() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserEQ, 0)
}

func (s *OperatorContext) NE() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserNE, 0)
}

func (s *OperatorContext) CO() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserCO, 0)
}

func (s *OperatorContext) SW() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserSW, 0)
}

func (s *OperatorContext) EW() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserEW, 0)
}

func (s *OperatorContext) GT() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserGT, 0)
}

func (s *OperatorContext) LT() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserLT, 0)
}

func (s *OperatorContext) GE() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserGE, 0)
}

func (s *OperatorContext) LE() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserLE, 0)
}

func (s *OperatorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OperatorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OperatorContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.EnterOperator(s)
	}
}

func (s *OperatorContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ScimFilterListener); ok {
		listenerT.ExitOperator(s)
	}
}

func (p *ScimFilterParser) Operator() (localctx IOperatorContext) {
	localctx = NewOperatorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, ScimFilterParserRULE_operator)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	p.SetState(165)
	_la = p.GetTokenStream().LA(1)

	if !(((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<ScimFilterParserEQ)|(1<<ScimFilterParserNE)|(1<<ScimFilterParserCO)|(1<<ScimFilterParserSW)|(1<<ScimFilterParserEW)|(1<<ScimFilterParserGT)|(1<<ScimFilterParserLT)|(1<<ScimFilterParserGE)|(1<<ScimFilterParserLE))) != 0) {
		p.GetErrorHandler().RecoverInline(p)
	} else {
		p.GetErrorHandler().ReportMatch(p)
		p.Consume()
	}

	return localctx
}

func (p *ScimFilterParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 1:
		var t *ExpressionContext = nil
		if localctx != nil {
			t = localctx.(*ExpressionContext)
		}
		return p.Expression_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *ScimFilterParser) Expression_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 9)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 8)

	case 2:
		return p.Precpred(p.GetParserRuleContext(), 7)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
