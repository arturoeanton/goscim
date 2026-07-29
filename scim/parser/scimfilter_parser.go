// Code generated from ScimFilter.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // ScimFilter

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type ScimFilterParser struct {
	*antlr.BaseParser
}

var ScimFilterParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func scimfilterParserInit() {
	staticData := &ScimFilterParserStaticData
	staticData.LiteralNames = []string{
		"", "'\"'", "", "", "", "", "", "", "", "", "", "", "", "", "", "'('",
		"')'", "'['", "']'", "' '",
	}
	staticData.SymbolicNames = []string{
		"", "", "EQ", "NE", "CO", "SW", "EW", "GT", "LT", "GE", "LE", "NOT",
		"AND", "OR", "PR", "LPAREN", "RPAREN", "LBRAC", "RBRAC", "WS", "NUMBERS",
		"BOOLEAN", "ATTRNAME", "ANY", "EOL",
	}
	staticData.RuleNames = []string{
		"start", "expression", "criteria", "criteriaValue", "operator",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 24, 168, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 1, 0, 5, 0, 12, 8, 0, 10, 0, 12, 0, 15, 9, 0, 1, 0, 1, 0, 1, 1, 1, 1,
		1, 1, 4, 1, 22, 8, 1, 11, 1, 12, 1, 23, 1, 1, 1, 1, 1, 1, 4, 1, 29, 8,
		1, 11, 1, 12, 1, 30, 1, 1, 1, 1, 1, 1, 4, 1, 36, 8, 1, 11, 1, 12, 1, 37,
		1, 1, 1, 1, 4, 1, 42, 8, 1, 11, 1, 12, 1, 43, 1, 1, 1, 1, 1, 1, 1, 1, 4,
		1, 50, 8, 1, 11, 1, 12, 1, 51, 1, 1, 1, 1, 4, 1, 56, 8, 1, 11, 1, 12, 1,
		57, 1, 1, 1, 1, 1, 1, 1, 1, 4, 1, 64, 8, 1, 11, 1, 12, 1, 65, 1, 1, 1,
		1, 4, 1, 70, 8, 1, 11, 1, 12, 1, 71, 1, 1, 1, 1, 1, 1, 1, 1, 5, 1, 78,
		8, 1, 10, 1, 12, 1, 81, 9, 1, 1, 1, 1, 1, 5, 1, 85, 8, 1, 10, 1, 12, 1,
		88, 9, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 5, 1, 95, 8, 1, 10, 1, 12, 1, 98,
		9, 1, 1, 1, 1, 1, 5, 1, 102, 8, 1, 10, 1, 12, 1, 105, 9, 1, 1, 1, 1, 1,
		3, 1, 109, 8, 1, 1, 1, 1, 1, 4, 1, 113, 8, 1, 11, 1, 12, 1, 114, 1, 1,
		1, 1, 4, 1, 119, 8, 1, 11, 1, 12, 1, 120, 1, 1, 1, 1, 1, 1, 4, 1, 126,
		8, 1, 11, 1, 12, 1, 127, 1, 1, 1, 1, 4, 1, 132, 8, 1, 11, 1, 12, 1, 133,
		1, 1, 1, 1, 1, 1, 4, 1, 139, 8, 1, 11, 1, 12, 1, 140, 1, 1, 1, 1, 4, 1,
		145, 8, 1, 11, 1, 12, 1, 146, 1, 1, 1, 1, 5, 1, 151, 8, 1, 10, 1, 12, 1,
		154, 9, 1, 1, 2, 1, 2, 4, 2, 158, 8, 2, 11, 2, 12, 2, 159, 1, 2, 1, 2,
		1, 3, 1, 3, 1, 4, 1, 4, 1, 4, 18, 23, 30, 37, 43, 51, 57, 65, 71, 79, 86,
		96, 103, 114, 120, 127, 140, 146, 159, 1, 2, 5, 0, 2, 4, 6, 8, 0, 2, 1,
		0, 20, 21, 1, 0, 2, 10, 191, 0, 13, 1, 0, 0, 0, 2, 108, 1, 0, 0, 0, 4,
		155, 1, 0, 0, 0, 6, 163, 1, 0, 0, 0, 8, 165, 1, 0, 0, 0, 10, 12, 3, 2,
		1, 0, 11, 10, 1, 0, 0, 0, 12, 15, 1, 0, 0, 0, 13, 11, 1, 0, 0, 0, 13, 14,
		1, 0, 0, 0, 14, 16, 1, 0, 0, 0, 15, 13, 1, 0, 0, 0, 16, 17, 5, 0, 0, 1,
		17, 1, 1, 0, 0, 0, 18, 19, 6, 1, -1, 0, 19, 21, 5, 11, 0, 0, 20, 22, 5,
		19, 0, 0, 21, 20, 1, 0, 0, 0, 22, 23, 1, 0, 0, 0, 23, 24, 1, 0, 0, 0, 23,
		21, 1, 0, 0, 0, 24, 25, 1, 0, 0, 0, 25, 109, 3, 2, 1, 10, 26, 28, 5, 22,
		0, 0, 27, 29, 5, 19, 0, 0, 28, 27, 1, 0, 0, 0, 29, 30, 1, 0, 0, 0, 30,
		31, 1, 0, 0, 0, 30, 28, 1, 0, 0, 0, 31, 32, 1, 0, 0, 0, 32, 109, 5, 14,
		0, 0, 33, 35, 5, 22, 0, 0, 34, 36, 5, 19, 0, 0, 35, 34, 1, 0, 0, 0, 36,
		37, 1, 0, 0, 0, 37, 38, 1, 0, 0, 0, 37, 35, 1, 0, 0, 0, 38, 39, 1, 0, 0,
		0, 39, 41, 3, 8, 4, 0, 40, 42, 5, 19, 0, 0, 41, 40, 1, 0, 0, 0, 42, 43,
		1, 0, 0, 0, 43, 44, 1, 0, 0, 0, 43, 41, 1, 0, 0, 0, 44, 45, 1, 0, 0, 0,
		45, 46, 3, 2, 1, 5, 46, 109, 1, 0, 0, 0, 47, 49, 5, 22, 0, 0, 48, 50, 5,
		19, 0, 0, 49, 48, 1, 0, 0, 0, 50, 51, 1, 0, 0, 0, 51, 52, 1, 0, 0, 0, 51,
		49, 1, 0, 0, 0, 52, 53, 1, 0, 0, 0, 53, 55, 3, 8, 4, 0, 54, 56, 5, 19,
		0, 0, 55, 54, 1, 0, 0, 0, 56, 57, 1, 0, 0, 0, 57, 58, 1, 0, 0, 0, 57, 55,
		1, 0, 0, 0, 58, 59, 1, 0, 0, 0, 59, 60, 3, 4, 2, 0, 60, 109, 1, 0, 0, 0,
		61, 63, 5, 22, 0, 0, 62, 64, 5, 19, 0, 0, 63, 62, 1, 0, 0, 0, 64, 65, 1,
		0, 0, 0, 65, 66, 1, 0, 0, 0, 65, 63, 1, 0, 0, 0, 66, 67, 1, 0, 0, 0, 67,
		69, 3, 8, 4, 0, 68, 70, 5, 19, 0, 0, 69, 68, 1, 0, 0, 0, 70, 71, 1, 0,
		0, 0, 71, 72, 1, 0, 0, 0, 71, 69, 1, 0, 0, 0, 72, 73, 1, 0, 0, 0, 73, 74,
		3, 6, 3, 0, 74, 109, 1, 0, 0, 0, 75, 79, 5, 15, 0, 0, 76, 78, 5, 19, 0,
		0, 77, 76, 1, 0, 0, 0, 78, 81, 1, 0, 0, 0, 79, 80, 1, 0, 0, 0, 79, 77,
		1, 0, 0, 0, 80, 82, 1, 0, 0, 0, 81, 79, 1, 0, 0, 0, 82, 86, 3, 2, 1, 0,
		83, 85, 5, 19, 0, 0, 84, 83, 1, 0, 0, 0, 85, 88, 1, 0, 0, 0, 86, 87, 1,
		0, 0, 0, 86, 84, 1, 0, 0, 0, 87, 89, 1, 0, 0, 0, 88, 86, 1, 0, 0, 0, 89,
		90, 5, 16, 0, 0, 90, 109, 1, 0, 0, 0, 91, 92, 5, 22, 0, 0, 92, 96, 5, 17,
		0, 0, 93, 95, 5, 19, 0, 0, 94, 93, 1, 0, 0, 0, 95, 98, 1, 0, 0, 0, 96,
		97, 1, 0, 0, 0, 96, 94, 1, 0, 0, 0, 97, 99, 1, 0, 0, 0, 98, 96, 1, 0, 0,
		0, 99, 103, 3, 2, 1, 0, 100, 102, 5, 19, 0, 0, 101, 100, 1, 0, 0, 0, 102,
		105, 1, 0, 0, 0, 103, 104, 1, 0, 0, 0, 103, 101, 1, 0, 0, 0, 104, 106,
		1, 0, 0, 0, 105, 103, 1, 0, 0, 0, 106, 107, 5, 18, 0, 0, 107, 109, 1, 0,
		0, 0, 108, 18, 1, 0, 0, 0, 108, 26, 1, 0, 0, 0, 108, 33, 1, 0, 0, 0, 108,
		47, 1, 0, 0, 0, 108, 61, 1, 0, 0, 0, 108, 75, 1, 0, 0, 0, 108, 91, 1, 0,
		0, 0, 109, 152, 1, 0, 0, 0, 110, 112, 10, 9, 0, 0, 111, 113, 5, 19, 0,
		0, 112, 111, 1, 0, 0, 0, 113, 114, 1, 0, 0, 0, 114, 115, 1, 0, 0, 0, 114,
		112, 1, 0, 0, 0, 115, 116, 1, 0, 0, 0, 116, 118, 5, 12, 0, 0, 117, 119,
		5, 19, 0, 0, 118, 117, 1, 0, 0, 0, 119, 120, 1, 0, 0, 0, 120, 121, 1, 0,
		0, 0, 120, 118, 1, 0, 0, 0, 121, 122, 1, 0, 0, 0, 122, 151, 3, 2, 1, 10,
		123, 125, 10, 8, 0, 0, 124, 126, 5, 19, 0, 0, 125, 124, 1, 0, 0, 0, 126,
		127, 1, 0, 0, 0, 127, 128, 1, 0, 0, 0, 127, 125, 1, 0, 0, 0, 128, 129,
		1, 0, 0, 0, 129, 131, 5, 13, 0, 0, 130, 132, 5, 19, 0, 0, 131, 130, 1,
		0, 0, 0, 132, 133, 1, 0, 0, 0, 133, 131, 1, 0, 0, 0, 133, 134, 1, 0, 0,
		0, 134, 135, 1, 0, 0, 0, 135, 151, 3, 2, 1, 9, 136, 138, 10, 7, 0, 0, 137,
		139, 5, 19, 0, 0, 138, 137, 1, 0, 0, 0, 139, 140, 1, 0, 0, 0, 140, 141,
		1, 0, 0, 0, 140, 138, 1, 0, 0, 0, 141, 142, 1, 0, 0, 0, 142, 144, 3, 8,
		4, 0, 143, 145, 5, 19, 0, 0, 144, 143, 1, 0, 0, 0, 145, 146, 1, 0, 0, 0,
		146, 147, 1, 0, 0, 0, 146, 144, 1, 0, 0, 0, 147, 148, 1, 0, 0, 0, 148,
		149, 3, 2, 1, 8, 149, 151, 1, 0, 0, 0, 150, 110, 1, 0, 0, 0, 150, 123,
		1, 0, 0, 0, 150, 136, 1, 0, 0, 0, 151, 154, 1, 0, 0, 0, 152, 150, 1, 0,
		0, 0, 152, 153, 1, 0, 0, 0, 153, 3, 1, 0, 0, 0, 154, 152, 1, 0, 0, 0, 155,
		157, 5, 1, 0, 0, 156, 158, 9, 0, 0, 0, 157, 156, 1, 0, 0, 0, 158, 159,
		1, 0, 0, 0, 159, 160, 1, 0, 0, 0, 159, 157, 1, 0, 0, 0, 160, 161, 1, 0,
		0, 0, 161, 162, 5, 1, 0, 0, 162, 5, 1, 0, 0, 0, 163, 164, 7, 0, 0, 0, 164,
		7, 1, 0, 0, 0, 165, 166, 7, 1, 0, 0, 166, 9, 1, 0, 0, 0, 23, 13, 23, 30,
		37, 43, 51, 57, 65, 71, 79, 86, 96, 103, 108, 114, 120, 127, 133, 140,
		146, 150, 152, 159,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// ScimFilterParserInit initializes any static state used to implement ScimFilterParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewScimFilterParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func ScimFilterParserInit() {
	staticData := &ScimFilterParserStaticData
	staticData.once.Do(scimfilterParserInit)
}

// NewScimFilterParser produces a new parser instance for the optional input antlr.TokenStream.
func NewScimFilterParser(input antlr.TokenStream) *ScimFilterParser {
	ScimFilterParserInit()
	this := new(ScimFilterParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &ScimFilterParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
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

	// Getter signatures
	EOF() antlr.TerminalNode
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext

	// IsStartContext differentiates from other interfaces.
	IsStartContext()
}

type StartContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStartContext() *StartContext {
	var p = new(StartContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ScimFilterParserRULE_start
	return p
}

func InitEmptyStartContext(p *StartContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ScimFilterParserRULE_start
}

func (*StartContext) IsStartContext() {}

func NewStartContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StartContext {
	var p = new(StartContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ScimFilterParserRULE_start

	return p
}

func (s *StartContext) GetParser() antlr.Parser { return s.parser }

func (s *StartContext) EOF() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserEOF, 0)
}

func (s *StartContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *StartContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

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

func (p *ScimFilterParser) Start_() (localctx IStartContext) {
	localctx = NewStartContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, ScimFilterParserRULE_start)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(13)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&4229120) != 0 {
		{
			p.SetState(10)
			p.expression(0)
		}

		p.SetState(15)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(16)
		p.Match(ScimFilterParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
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
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpressionContext() *ExpressionContext {
	var p = new(ExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ScimFilterParserRULE_expression
	return p
}

func InitEmptyExpressionContext(p *ExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ScimFilterParserRULE_expression
}

func (*ExpressionContext) IsExpressionContext() {}

func NewExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionContext {
	var p = new(ExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ScimFilterParserRULE_expression

	return p
}

func (s *ExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionContext) CopyAll(ctx *ExpressionContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type ATTR_PRContext struct {
	ExpressionContext
}

func NewATTR_PRContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ATTR_PRContext {
	var p = new(ATTR_PRContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

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
	ExpressionContext
}

func NewLBRAC_EXPR_RBRACContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LBRAC_EXPR_RBRACContext {
	var p = new(LBRAC_EXPR_RBRACContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

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
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

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
	ExpressionContext
}

func NewATTR_OPER_EXPRContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ATTR_OPER_EXPRContext {
	var p = new(ATTR_OPER_EXPRContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *ATTR_OPER_EXPRContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ATTR_OPER_EXPRContext) ATTRNAME() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserATTRNAME, 0)
}

func (s *ATTR_OPER_EXPRContext) Operator() IOperatorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOperatorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOperatorContext)
}

func (s *ATTR_OPER_EXPRContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

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
	ExpressionContext
}

func NewEXPR_OR_EXPRContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *EXPR_OR_EXPRContext {
	var p = new(EXPR_OR_EXPRContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *EXPR_OR_EXPRContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *EXPR_OR_EXPRContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *EXPR_OR_EXPRContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

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
	ExpressionContext
}

func NewEXPR_OPER_EXPRContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *EXPR_OPER_EXPRContext {
	var p = new(EXPR_OPER_EXPRContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *EXPR_OPER_EXPRContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *EXPR_OPER_EXPRContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *EXPR_OPER_EXPRContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *EXPR_OPER_EXPRContext) Operator() IOperatorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOperatorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

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
	ExpressionContext
}

func NewNOT_EXPRContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *NOT_EXPRContext {
	var p = new(NOT_EXPRContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *NOT_EXPRContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NOT_EXPRContext) NOT() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserNOT, 0)
}

func (s *NOT_EXPRContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

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
	ExpressionContext
}

func NewEXPR_AND_EXPRContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *EXPR_AND_EXPRContext {
	var p = new(EXPR_AND_EXPRContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *EXPR_AND_EXPRContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *EXPR_AND_EXPRContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *EXPR_AND_EXPRContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

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
	ExpressionContext
}

func NewATTR_OPER_VALUEContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ATTR_OPER_VALUEContext {
	var p = new(ATTR_OPER_VALUEContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *ATTR_OPER_VALUEContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ATTR_OPER_VALUEContext) ATTRNAME() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserATTRNAME, 0)
}

func (s *ATTR_OPER_VALUEContext) Operator() IOperatorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOperatorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOperatorContext)
}

func (s *ATTR_OPER_VALUEContext) CriteriaValue() ICriteriaValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICriteriaValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

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
	ExpressionContext
}

func NewATTR_OPER_CRITERIAContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ATTR_OPER_CRITERIAContext {
	var p = new(ATTR_OPER_CRITERIAContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *ATTR_OPER_CRITERIAContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ATTR_OPER_CRITERIAContext) ATTRNAME() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserATTRNAME, 0)
}

func (s *ATTR_OPER_CRITERIAContext) Operator() IOperatorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOperatorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOperatorContext)
}

func (s *ATTR_OPER_CRITERIAContext) Criteria() ICriteriaContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICriteriaContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

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
	ExpressionContext
}

func NewLPAREN_EXPR_RPARENContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LPAREN_EXPR_RPARENContext {
	var p = new(LPAREN_EXPR_RPARENContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *LPAREN_EXPR_RPARENContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LPAREN_EXPR_RPARENContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(ScimFilterParserLPAREN, 0)
}

func (s *LPAREN_EXPR_RPARENContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

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

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(108)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 13, p.GetParserRuleContext()) {
	case 1:
		localctx = NewNOT_EXPRContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(19)
			p.Match(ScimFilterParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(21)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(20)
					p.Match(ScimFilterParserWS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			default:
				p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
				goto errorExit
			}

			p.SetState(23)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
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
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(28)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(27)
					p.Match(ScimFilterParserWS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			default:
				p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
				goto errorExit
			}

			p.SetState(30)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		{
			p.SetState(32)
			p.Match(ScimFilterParserPR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		localctx = NewATTR_OPER_EXPRContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(33)
			p.Match(ScimFilterParserATTRNAME)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(35)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(34)
					p.Match(ScimFilterParserWS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			default:
				p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
				goto errorExit
			}

			p.SetState(37)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		{
			p.SetState(39)
			p.Operator()
		}
		p.SetState(41)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(40)
					p.Match(ScimFilterParserWS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			default:
				p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
				goto errorExit
			}

			p.SetState(43)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
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
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(49)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(48)
					p.Match(ScimFilterParserWS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			default:
				p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
				goto errorExit
			}

			p.SetState(51)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		{
			p.SetState(53)
			p.Operator()
		}
		p.SetState(55)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(54)
					p.Match(ScimFilterParserWS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			default:
				p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
				goto errorExit
			}

			p.SetState(57)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 6, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
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
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(63)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(62)
					p.Match(ScimFilterParserWS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			default:
				p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
				goto errorExit
			}

			p.SetState(65)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 7, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		{
			p.SetState(67)
			p.Operator()
		}
		p.SetState(69)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = 1 + 1
		for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1 + 1:
				{
					p.SetState(68)
					p.Match(ScimFilterParserWS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			default:
				p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
				goto errorExit
			}

			p.SetState(71)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 8, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
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
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(79)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 9, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1+1 {
				{
					p.SetState(76)
					p.Match(ScimFilterParserWS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(81)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 9, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		{
			p.SetState(82)
			p.expression(0)
		}
		p.SetState(86)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 10, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1+1 {
				{
					p.SetState(83)
					p.Match(ScimFilterParserWS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(88)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 10, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		{
			p.SetState(89)
			p.Match(ScimFilterParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 7:
		localctx = NewLBRAC_EXPR_RBRACContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(91)
			p.Match(ScimFilterParserATTRNAME)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(92)
			p.Match(ScimFilterParserLBRAC)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(96)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 11, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1+1 {
				{
					p.SetState(93)
					p.Match(ScimFilterParserWS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(98)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 11, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		{
			p.SetState(99)
			p.expression(0)
		}
		p.SetState(103)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 12, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1+1 {
				{
					p.SetState(100)
					p.Match(ScimFilterParserWS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(105)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 12, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		{
			p.SetState(106)
			p.Match(ScimFilterParserRBRAC)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(152)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 21, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(150)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 20, p.GetParserRuleContext()) {
			case 1:
				localctx = NewEXPR_AND_EXPRContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ScimFilterParserRULE_expression)
				p.SetState(110)

				if !(p.Precpred(p.GetParserRuleContext(), 9)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 9)", ""))
					goto errorExit
				}
				p.SetState(112)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_alt = 1 + 1
				for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
					switch _alt {
					case 1 + 1:
						{
							p.SetState(111)
							p.Match(ScimFilterParserWS)
							if p.HasError() {
								// Recognition error - abort rule
								goto errorExit
							}
						}

					default:
						p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
						goto errorExit
					}

					p.SetState(114)
					p.GetErrorHandler().Sync(p)
					_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 14, p.GetParserRuleContext())
					if p.HasError() {
						goto errorExit
					}
				}
				{
					p.SetState(116)
					p.Match(ScimFilterParserAND)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				p.SetState(118)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_alt = 1 + 1
				for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
					switch _alt {
					case 1 + 1:
						{
							p.SetState(117)
							p.Match(ScimFilterParserWS)
							if p.HasError() {
								// Recognition error - abort rule
								goto errorExit
							}
						}

					default:
						p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
						goto errorExit
					}

					p.SetState(120)
					p.GetErrorHandler().Sync(p)
					_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 15, p.GetParserRuleContext())
					if p.HasError() {
						goto errorExit
					}
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
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 8)", ""))
					goto errorExit
				}
				p.SetState(125)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_alt = 1 + 1
				for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
					switch _alt {
					case 1 + 1:
						{
							p.SetState(124)
							p.Match(ScimFilterParserWS)
							if p.HasError() {
								// Recognition error - abort rule
								goto errorExit
							}
						}

					default:
						p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
						goto errorExit
					}

					p.SetState(127)
					p.GetErrorHandler().Sync(p)
					_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 16, p.GetParserRuleContext())
					if p.HasError() {
						goto errorExit
					}
				}
				{
					p.SetState(129)
					p.Match(ScimFilterParserOR)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				p.SetState(131)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)

				for ok := true; ok; ok = _la == ScimFilterParserWS {
					{
						p.SetState(130)
						p.Match(ScimFilterParserWS)
						if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
						}
					}

					p.SetState(133)
					p.GetErrorHandler().Sync(p)
					if p.HasError() {
						goto errorExit
					}
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
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 7)", ""))
					goto errorExit
				}
				p.SetState(138)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_alt = 1 + 1
				for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
					switch _alt {
					case 1 + 1:
						{
							p.SetState(137)
							p.Match(ScimFilterParserWS)
							if p.HasError() {
								// Recognition error - abort rule
								goto errorExit
							}
						}

					default:
						p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
						goto errorExit
					}

					p.SetState(140)
					p.GetErrorHandler().Sync(p)
					_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 18, p.GetParserRuleContext())
					if p.HasError() {
						goto errorExit
					}
				}
				{
					p.SetState(142)
					p.Operator()
				}
				p.SetState(144)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_alt = 1 + 1
				for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
					switch _alt {
					case 1 + 1:
						{
							p.SetState(143)
							p.Match(ScimFilterParserWS)
							if p.HasError() {
								// Recognition error - abort rule
								goto errorExit
							}
						}

					default:
						p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
						goto errorExit
					}

					p.SetState(146)
					p.GetErrorHandler().Sync(p)
					_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 19, p.GetParserRuleContext())
					if p.HasError() {
						goto errorExit
					}
				}
				{
					p.SetState(148)
					p.expression(8)
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(154)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 21, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
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
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCriteriaContext() *CriteriaContext {
	var p = new(CriteriaContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ScimFilterParserRULE_criteria
	return p
}

func InitEmptyCriteriaContext(p *CriteriaContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ScimFilterParserRULE_criteria
}

func (*CriteriaContext) IsCriteriaContext() {}

func NewCriteriaContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CriteriaContext {
	var p = new(CriteriaContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

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
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(155)
		p.Match(ScimFilterParserT__0)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(157)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = 1 + 1
	for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
		switch _alt {
		case 1 + 1:
			p.SetState(156)
			p.MatchWildcard()

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(159)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 22, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	{
		p.SetState(161)
		p.Match(ScimFilterParserT__0)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICriteriaValueContext is an interface to support dynamic dispatch.
type ICriteriaValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NUMBERS() antlr.TerminalNode
	BOOLEAN() antlr.TerminalNode

	// IsCriteriaValueContext differentiates from other interfaces.
	IsCriteriaValueContext()
}

type CriteriaValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCriteriaValueContext() *CriteriaValueContext {
	var p = new(CriteriaValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ScimFilterParserRULE_criteriaValue
	return p
}

func InitEmptyCriteriaValueContext(p *CriteriaValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ScimFilterParserRULE_criteriaValue
}

func (*CriteriaValueContext) IsCriteriaValueContext() {}

func NewCriteriaValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CriteriaValueContext {
	var p = new(CriteriaValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

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

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(163)
		_la = p.GetTokenStream().LA(1)

		if !(_la == ScimFilterParserNUMBERS || _la == ScimFilterParserBOOLEAN) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IOperatorContext is an interface to support dynamic dispatch.
type IOperatorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EQ() antlr.TerminalNode
	NE() antlr.TerminalNode
	CO() antlr.TerminalNode
	SW() antlr.TerminalNode
	EW() antlr.TerminalNode
	GT() antlr.TerminalNode
	LT() antlr.TerminalNode
	GE() antlr.TerminalNode
	LE() antlr.TerminalNode

	// IsOperatorContext differentiates from other interfaces.
	IsOperatorContext()
}

type OperatorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOperatorContext() *OperatorContext {
	var p = new(OperatorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ScimFilterParserRULE_operator
	return p
}

func InitEmptyOperatorContext(p *OperatorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ScimFilterParserRULE_operator
}

func (*OperatorContext) IsOperatorContext() {}

func NewOperatorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OperatorContext {
	var p = new(OperatorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

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

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(165)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2044) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
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
