// Generated from ScimFilter.g4 by ANTLR 4.7.

package parser

import (
	"fmt"
	"unicode"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = unicode.IsLetter

var serializedLexerAtn = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 2, 26, 137,
	8, 1, 4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7,
	9, 7, 4, 8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 4, 11, 9, 11, 4, 12, 9, 12,
	4, 13, 9, 13, 4, 14, 9, 14, 4, 15, 9, 15, 4, 16, 9, 16, 4, 17, 9, 17, 4,
	18, 9, 18, 4, 19, 9, 19, 4, 20, 9, 20, 4, 21, 9, 21, 4, 22, 9, 22, 4, 23,
	9, 23, 4, 24, 9, 24, 4, 25, 9, 25, 3, 2, 3, 2, 3, 3, 3, 3, 3, 3, 3, 4,
	3, 4, 3, 4, 3, 5, 3, 5, 3, 5, 3, 6, 3, 6, 3, 6, 3, 7, 3, 7, 3, 7, 3, 8,
	3, 8, 3, 8, 3, 9, 3, 9, 3, 9, 3, 10, 3, 10, 3, 10, 3, 11, 3, 11, 3, 11,
	3, 12, 3, 12, 3, 12, 3, 12, 3, 13, 3, 13, 3, 13, 3, 13, 3, 14, 3, 14, 3,
	14, 3, 15, 3, 15, 3, 15, 3, 16, 3, 16, 3, 17, 3, 17, 3, 18, 3, 18, 3, 19,
	3, 19, 3, 20, 3, 20, 3, 21, 6, 21, 106, 10, 21, 13, 21, 14, 21, 107, 3,
	22, 3, 22, 3, 22, 3, 22, 3, 22, 3, 22, 3, 22, 3, 22, 3, 22, 5, 22, 119,
	10, 22, 3, 23, 5, 23, 122, 10, 23, 3, 23, 6, 23, 125, 10, 23, 13, 23, 14,
	23, 126, 3, 24, 3, 24, 3, 25, 6, 25, 132, 10, 25, 13, 25, 14, 25, 133,
	3, 25, 3, 25, 2, 2, 26, 3, 3, 5, 4, 7, 5, 9, 6, 11, 7, 13, 8, 15, 9, 17,
	10, 19, 11, 21, 12, 23, 13, 25, 14, 27, 15, 29, 16, 31, 17, 33, 18, 35,
	19, 37, 20, 39, 21, 41, 22, 43, 23, 45, 24, 47, 25, 49, 26, 3, 2, 20, 4,
	2, 71, 71, 103, 103, 4, 2, 83, 83, 115, 115, 4, 2, 80, 80, 112, 112, 4,
	2, 69, 69, 101, 101, 4, 2, 81, 81, 113, 113, 4, 2, 85, 85, 117, 117, 4,
	2, 89, 89, 121, 121, 4, 2, 73, 73, 105, 105, 4, 2, 86, 86, 118, 118, 4,
	2, 78, 78, 110, 110, 4, 2, 67, 67, 99, 99, 4, 2, 70, 70, 102, 102, 4, 2,
	84, 84, 116, 116, 4, 2, 82, 82, 114, 114, 4, 2, 47, 48, 50, 59, 7, 2, 47,
	48, 50, 60, 67, 92, 97, 97, 99, 124, 6, 2, 36, 36, 42, 43, 93, 93, 95,
	95, 4, 2, 11, 12, 14, 15, 2, 141, 2, 3, 3, 2, 2, 2, 2, 5, 3, 2, 2, 2, 2,
	7, 3, 2, 2, 2, 2, 9, 3, 2, 2, 2, 2, 11, 3, 2, 2, 2, 2, 13, 3, 2, 2, 2,
	2, 15, 3, 2, 2, 2, 2, 17, 3, 2, 2, 2, 2, 19, 3, 2, 2, 2, 2, 21, 3, 2, 2,
	2, 2, 23, 3, 2, 2, 2, 2, 25, 3, 2, 2, 2, 2, 27, 3, 2, 2, 2, 2, 29, 3, 2,
	2, 2, 2, 31, 3, 2, 2, 2, 2, 33, 3, 2, 2, 2, 2, 35, 3, 2, 2, 2, 2, 37, 3,
	2, 2, 2, 2, 39, 3, 2, 2, 2, 2, 41, 3, 2, 2, 2, 2, 43, 3, 2, 2, 2, 2, 45,
	3, 2, 2, 2, 2, 47, 3, 2, 2, 2, 2, 49, 3, 2, 2, 2, 3, 51, 3, 2, 2, 2, 5,
	53, 3, 2, 2, 2, 7, 56, 3, 2, 2, 2, 9, 59, 3, 2, 2, 2, 11, 62, 3, 2, 2,
	2, 13, 65, 3, 2, 2, 2, 15, 68, 3, 2, 2, 2, 17, 71, 3, 2, 2, 2, 19, 74,
	3, 2, 2, 2, 21, 77, 3, 2, 2, 2, 23, 80, 3, 2, 2, 2, 25, 84, 3, 2, 2, 2,
	27, 88, 3, 2, 2, 2, 29, 91, 3, 2, 2, 2, 31, 94, 3, 2, 2, 2, 33, 96, 3,
	2, 2, 2, 35, 98, 3, 2, 2, 2, 37, 100, 3, 2, 2, 2, 39, 102, 3, 2, 2, 2,
	41, 105, 3, 2, 2, 2, 43, 118, 3, 2, 2, 2, 45, 121, 3, 2, 2, 2, 47, 128,
	3, 2, 2, 2, 49, 131, 3, 2, 2, 2, 51, 52, 7, 36, 2, 2, 52, 4, 3, 2, 2, 2,
	53, 54, 9, 2, 2, 2, 54, 55, 9, 3, 2, 2, 55, 6, 3, 2, 2, 2, 56, 57, 9, 4,
	2, 2, 57, 58, 9, 2, 2, 2, 58, 8, 3, 2, 2, 2, 59, 60, 9, 5, 2, 2, 60, 61,
	9, 6, 2, 2, 61, 10, 3, 2, 2, 2, 62, 63, 9, 7, 2, 2, 63, 64, 9, 8, 2, 2,
	64, 12, 3, 2, 2, 2, 65, 66, 9, 2, 2, 2, 66, 67, 9, 8, 2, 2, 67, 14, 3,
	2, 2, 2, 68, 69, 9, 9, 2, 2, 69, 70, 9, 10, 2, 2, 70, 16, 3, 2, 2, 2, 71,
	72, 9, 11, 2, 2, 72, 73, 9, 10, 2, 2, 73, 18, 3, 2, 2, 2, 74, 75, 9, 9,
	2, 2, 75, 76, 9, 2, 2, 2, 76, 20, 3, 2, 2, 2, 77, 78, 9, 11, 2, 2, 78,
	79, 9, 2, 2, 2, 79, 22, 3, 2, 2, 2, 80, 81, 9, 4, 2, 2, 81, 82, 9, 6, 2,
	2, 82, 83, 9, 10, 2, 2, 83, 24, 3, 2, 2, 2, 84, 85, 9, 12, 2, 2, 85, 86,
	9, 4, 2, 2, 86, 87, 9, 13, 2, 2, 87, 26, 3, 2, 2, 2, 88, 89, 9, 6, 2, 2,
	89, 90, 9, 14, 2, 2, 90, 28, 3, 2, 2, 2, 91, 92, 9, 15, 2, 2, 92, 93, 9,
	14, 2, 2, 93, 30, 3, 2, 2, 2, 94, 95, 7, 42, 2, 2, 95, 32, 3, 2, 2, 2,
	96, 97, 7, 43, 2, 2, 97, 34, 3, 2, 2, 2, 98, 99, 7, 93, 2, 2, 99, 36, 3,
	2, 2, 2, 100, 101, 7, 95, 2, 2, 101, 38, 3, 2, 2, 2, 102, 103, 7, 34, 2,
	2, 103, 40, 3, 2, 2, 2, 104, 106, 9, 16, 2, 2, 105, 104, 3, 2, 2, 2, 106,
	107, 3, 2, 2, 2, 107, 105, 3, 2, 2, 2, 107, 108, 3, 2, 2, 2, 108, 42, 3,
	2, 2, 2, 109, 110, 7, 118, 2, 2, 110, 111, 7, 116, 2, 2, 111, 112, 7, 119,
	2, 2, 112, 119, 7, 103, 2, 2, 113, 114, 7, 104, 2, 2, 114, 115, 7, 99,
	2, 2, 115, 116, 7, 110, 2, 2, 116, 117, 7, 117, 2, 2, 117, 119, 7, 103,
	2, 2, 118, 109, 3, 2, 2, 2, 118, 113, 3, 2, 2, 2, 119, 44, 3, 2, 2, 2,
	120, 122, 7, 38, 2, 2, 121, 120, 3, 2, 2, 2, 121, 122, 3, 2, 2, 2, 122,
	124, 3, 2, 2, 2, 123, 125, 9, 17, 2, 2, 124, 123, 3, 2, 2, 2, 125, 126,
	3, 2, 2, 2, 126, 124, 3, 2, 2, 2, 126, 127, 3, 2, 2, 2, 127, 46, 3, 2,
	2, 2, 128, 129, 10, 18, 2, 2, 129, 48, 3, 2, 2, 2, 130, 132, 9, 19, 2,
	2, 131, 130, 3, 2, 2, 2, 132, 133, 3, 2, 2, 2, 133, 131, 3, 2, 2, 2, 133,
	134, 3, 2, 2, 2, 134, 135, 3, 2, 2, 2, 135, 136, 8, 25, 2, 2, 136, 50,
	3, 2, 2, 2, 8, 2, 107, 118, 121, 126, 133, 3, 8, 2, 2,
}

var lexerDeserializer = antlr.NewATNDeserializer(nil)
var lexerAtn = lexerDeserializer.DeserializeFromUInt16(serializedLexerAtn)

var lexerChannelNames = []string{
	"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
}

var lexerModeNames = []string{
	"DEFAULT_MODE",
}

var lexerLiteralNames = []string{
	"", "'\"'", "", "", "", "", "", "", "", "", "", "", "", "", "", "'('",
	"')'", "'['", "']'", "' '",
}

var lexerSymbolicNames = []string{
	"", "", "EQ", "NE", "CO", "SW", "EW", "GT", "LT", "GE", "LE", "NOT", "AND",
	"OR", "PR", "LPAREN", "RPAREN", "LBRAC", "RBRAC", "WS", "NUMBERS", "BOOLEAN",
	"ATTRNAME", "ANY", "EOL",
}

var lexerRuleNames = []string{
	"T__0", "EQ", "NE", "CO", "SW", "EW", "GT", "LT", "GE", "LE", "NOT", "AND",
	"OR", "PR", "LPAREN", "RPAREN", "LBRAC", "RBRAC", "WS", "NUMBERS", "BOOLEAN",
	"ATTRNAME", "ANY", "EOL",
}

type ScimFilterLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var lexerDecisionToDFA = make([]*antlr.DFA, len(lexerAtn.DecisionToState))

func init() {
	for index, ds := range lexerAtn.DecisionToState {
		lexerDecisionToDFA[index] = antlr.NewDFA(ds, index)
	}
}

func NewScimFilterLexer(input antlr.CharStream) *ScimFilterLexer {

	l := new(ScimFilterLexer)

	l.BaseLexer = antlr.NewBaseLexer(input)
	l.Interpreter = antlr.NewLexerATNSimulator(l, lexerAtn, lexerDecisionToDFA, antlr.NewPredictionContextCache())

	l.channelNames = lexerChannelNames
	l.modeNames = lexerModeNames
	l.RuleNames = lexerRuleNames
	l.LiteralNames = lexerLiteralNames
	l.SymbolicNames = lexerSymbolicNames
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
