// Code generated from ScimFilter.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"sync"
	"unicode"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = sync.Once{}
var _ = unicode.IsLetter

type ScimFilterLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var ScimFilterLexerLexerStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	ChannelNames           []string
	ModeNames              []string
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func scimfilterlexerLexerInit() {
	staticData := &ScimFilterLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE",
	}
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
		"T__0", "EQ", "NE", "CO", "SW", "EW", "GT", "LT", "GE", "LE", "NOT",
		"AND", "OR", "PR", "LPAREN", "RPAREN", "LBRAC", "RBRAC", "WS", "NUMBERS",
		"BOOLEAN", "ATTRNAME", "ANY", "EOL",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 24, 135, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15,
		7, 15, 2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7,
		20, 2, 21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 1, 0, 1, 0, 1, 1, 1, 1, 1,
		1, 1, 2, 1, 2, 1, 2, 1, 3, 1, 3, 1, 3, 1, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1,
		5, 1, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 1, 8, 1, 8, 1, 8, 1, 9, 1, 9, 1,
		9, 1, 10, 1, 10, 1, 10, 1, 10, 1, 11, 1, 11, 1, 11, 1, 11, 1, 12, 1, 12,
		1, 12, 1, 13, 1, 13, 1, 13, 1, 14, 1, 14, 1, 15, 1, 15, 1, 16, 1, 16, 1,
		17, 1, 17, 1, 18, 1, 18, 1, 19, 4, 19, 104, 8, 19, 11, 19, 12, 19, 105,
		1, 20, 1, 20, 1, 20, 1, 20, 1, 20, 1, 20, 1, 20, 1, 20, 1, 20, 3, 20, 117,
		8, 20, 1, 21, 3, 21, 120, 8, 21, 1, 21, 4, 21, 123, 8, 21, 11, 21, 12,
		21, 124, 1, 22, 1, 22, 1, 23, 4, 23, 130, 8, 23, 11, 23, 12, 23, 131, 1,
		23, 1, 23, 0, 0, 24, 1, 1, 3, 2, 5, 3, 7, 4, 9, 5, 11, 6, 13, 7, 15, 8,
		17, 9, 19, 10, 21, 11, 23, 12, 25, 13, 27, 14, 29, 15, 31, 16, 33, 17,
		35, 18, 37, 19, 39, 20, 41, 21, 43, 22, 45, 23, 47, 24, 1, 0, 18, 2, 0,
		69, 69, 101, 101, 2, 0, 81, 81, 113, 113, 2, 0, 78, 78, 110, 110, 2, 0,
		67, 67, 99, 99, 2, 0, 79, 79, 111, 111, 2, 0, 83, 83, 115, 115, 2, 0, 87,
		87, 119, 119, 2, 0, 71, 71, 103, 103, 2, 0, 84, 84, 116, 116, 2, 0, 76,
		76, 108, 108, 2, 0, 65, 65, 97, 97, 2, 0, 68, 68, 100, 100, 2, 0, 82, 82,
		114, 114, 2, 0, 80, 80, 112, 112, 2, 0, 45, 46, 48, 57, 5, 0, 45, 46, 48,
		58, 65, 90, 95, 95, 97, 122, 4, 0, 34, 34, 40, 41, 91, 91, 93, 93, 2, 0,
		9, 10, 12, 13, 139, 0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0, 0, 0, 5, 1, 0, 0,
		0, 0, 7, 1, 0, 0, 0, 0, 9, 1, 0, 0, 0, 0, 11, 1, 0, 0, 0, 0, 13, 1, 0,
		0, 0, 0, 15, 1, 0, 0, 0, 0, 17, 1, 0, 0, 0, 0, 19, 1, 0, 0, 0, 0, 21, 1,
		0, 0, 0, 0, 23, 1, 0, 0, 0, 0, 25, 1, 0, 0, 0, 0, 27, 1, 0, 0, 0, 0, 29,
		1, 0, 0, 0, 0, 31, 1, 0, 0, 0, 0, 33, 1, 0, 0, 0, 0, 35, 1, 0, 0, 0, 0,
		37, 1, 0, 0, 0, 0, 39, 1, 0, 0, 0, 0, 41, 1, 0, 0, 0, 0, 43, 1, 0, 0, 0,
		0, 45, 1, 0, 0, 0, 0, 47, 1, 0, 0, 0, 1, 49, 1, 0, 0, 0, 3, 51, 1, 0, 0,
		0, 5, 54, 1, 0, 0, 0, 7, 57, 1, 0, 0, 0, 9, 60, 1, 0, 0, 0, 11, 63, 1,
		0, 0, 0, 13, 66, 1, 0, 0, 0, 15, 69, 1, 0, 0, 0, 17, 72, 1, 0, 0, 0, 19,
		75, 1, 0, 0, 0, 21, 78, 1, 0, 0, 0, 23, 82, 1, 0, 0, 0, 25, 86, 1, 0, 0,
		0, 27, 89, 1, 0, 0, 0, 29, 92, 1, 0, 0, 0, 31, 94, 1, 0, 0, 0, 33, 96,
		1, 0, 0, 0, 35, 98, 1, 0, 0, 0, 37, 100, 1, 0, 0, 0, 39, 103, 1, 0, 0,
		0, 41, 116, 1, 0, 0, 0, 43, 119, 1, 0, 0, 0, 45, 126, 1, 0, 0, 0, 47, 129,
		1, 0, 0, 0, 49, 50, 5, 34, 0, 0, 50, 2, 1, 0, 0, 0, 51, 52, 7, 0, 0, 0,
		52, 53, 7, 1, 0, 0, 53, 4, 1, 0, 0, 0, 54, 55, 7, 2, 0, 0, 55, 56, 7, 0,
		0, 0, 56, 6, 1, 0, 0, 0, 57, 58, 7, 3, 0, 0, 58, 59, 7, 4, 0, 0, 59, 8,
		1, 0, 0, 0, 60, 61, 7, 5, 0, 0, 61, 62, 7, 6, 0, 0, 62, 10, 1, 0, 0, 0,
		63, 64, 7, 0, 0, 0, 64, 65, 7, 6, 0, 0, 65, 12, 1, 0, 0, 0, 66, 67, 7,
		7, 0, 0, 67, 68, 7, 8, 0, 0, 68, 14, 1, 0, 0, 0, 69, 70, 7, 9, 0, 0, 70,
		71, 7, 8, 0, 0, 71, 16, 1, 0, 0, 0, 72, 73, 7, 7, 0, 0, 73, 74, 7, 0, 0,
		0, 74, 18, 1, 0, 0, 0, 75, 76, 7, 9, 0, 0, 76, 77, 7, 0, 0, 0, 77, 20,
		1, 0, 0, 0, 78, 79, 7, 2, 0, 0, 79, 80, 7, 4, 0, 0, 80, 81, 7, 8, 0, 0,
		81, 22, 1, 0, 0, 0, 82, 83, 7, 10, 0, 0, 83, 84, 7, 2, 0, 0, 84, 85, 7,
		11, 0, 0, 85, 24, 1, 0, 0, 0, 86, 87, 7, 4, 0, 0, 87, 88, 7, 12, 0, 0,
		88, 26, 1, 0, 0, 0, 89, 90, 7, 13, 0, 0, 90, 91, 7, 12, 0, 0, 91, 28, 1,
		0, 0, 0, 92, 93, 5, 40, 0, 0, 93, 30, 1, 0, 0, 0, 94, 95, 5, 41, 0, 0,
		95, 32, 1, 0, 0, 0, 96, 97, 5, 91, 0, 0, 97, 34, 1, 0, 0, 0, 98, 99, 5,
		93, 0, 0, 99, 36, 1, 0, 0, 0, 100, 101, 5, 32, 0, 0, 101, 38, 1, 0, 0,
		0, 102, 104, 7, 14, 0, 0, 103, 102, 1, 0, 0, 0, 104, 105, 1, 0, 0, 0, 105,
		103, 1, 0, 0, 0, 105, 106, 1, 0, 0, 0, 106, 40, 1, 0, 0, 0, 107, 108, 5,
		116, 0, 0, 108, 109, 5, 114, 0, 0, 109, 110, 5, 117, 0, 0, 110, 117, 5,
		101, 0, 0, 111, 112, 5, 102, 0, 0, 112, 113, 5, 97, 0, 0, 113, 114, 5,
		108, 0, 0, 114, 115, 5, 115, 0, 0, 115, 117, 5, 101, 0, 0, 116, 107, 1,
		0, 0, 0, 116, 111, 1, 0, 0, 0, 117, 42, 1, 0, 0, 0, 118, 120, 5, 36, 0,
		0, 119, 118, 1, 0, 0, 0, 119, 120, 1, 0, 0, 0, 120, 122, 1, 0, 0, 0, 121,
		123, 7, 15, 0, 0, 122, 121, 1, 0, 0, 0, 123, 124, 1, 0, 0, 0, 124, 122,
		1, 0, 0, 0, 124, 125, 1, 0, 0, 0, 125, 44, 1, 0, 0, 0, 126, 127, 8, 16,
		0, 0, 127, 46, 1, 0, 0, 0, 128, 130, 7, 17, 0, 0, 129, 128, 1, 0, 0, 0,
		130, 131, 1, 0, 0, 0, 131, 129, 1, 0, 0, 0, 131, 132, 1, 0, 0, 0, 132,
		133, 1, 0, 0, 0, 133, 134, 6, 23, 0, 0, 134, 48, 1, 0, 0, 0, 6, 0, 105,
		116, 119, 124, 131, 1, 6, 0, 0,
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

// ScimFilterLexerInit initializes any static state used to implement ScimFilterLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewScimFilterLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func ScimFilterLexerInit() {
	staticData := &ScimFilterLexerLexerStaticData
	staticData.once.Do(scimfilterlexerLexerInit)
}

// NewScimFilterLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewScimFilterLexer(input antlr.CharStream) *ScimFilterLexer {
	ScimFilterLexerInit()
	l := new(ScimFilterLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &ScimFilterLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "ScimFilter.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// ScimFilterLexer tokens.
const (
	ScimFilterLexerT__0     = 1
	ScimFilterLexerEQ       = 2
	ScimFilterLexerNE       = 3
	ScimFilterLexerCO       = 4
	ScimFilterLexerSW       = 5
	ScimFilterLexerEW       = 6
	ScimFilterLexerGT       = 7
	ScimFilterLexerLT       = 8
	ScimFilterLexerGE       = 9
	ScimFilterLexerLE       = 10
	ScimFilterLexerNOT      = 11
	ScimFilterLexerAND      = 12
	ScimFilterLexerOR       = 13
	ScimFilterLexerPR       = 14
	ScimFilterLexerLPAREN   = 15
	ScimFilterLexerRPAREN   = 16
	ScimFilterLexerLBRAC    = 17
	ScimFilterLexerRBRAC    = 18
	ScimFilterLexerWS       = 19
	ScimFilterLexerNUMBERS  = 20
	ScimFilterLexerBOOLEAN  = 21
	ScimFilterLexerATTRNAME = 22
	ScimFilterLexerANY      = 23
	ScimFilterLexerEOL      = 24
)
