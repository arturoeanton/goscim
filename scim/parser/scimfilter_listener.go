// Generated from ScimFilter.g4 by ANTLR 4.7.

package parser // ScimFilter

import "github.com/antlr/antlr4/runtime/Go/antlr"

// ScimFilterListener is a complete listener for a parse tree produced by ScimFilterParser.
type ScimFilterListener interface {
	antlr.ParseTreeListener

	// EnterStart is called when entering the start production.
	EnterStart(c *StartContext)

	// EnterATTR_PR is called when entering the ATTR_PR production.
	EnterATTR_PR(c *ATTR_PRContext)

	// EnterLBRAC_EXPR_RBRAC is called when entering the LBRAC_EXPR_RBRAC production.
	EnterLBRAC_EXPR_RBRAC(c *LBRAC_EXPR_RBRACContext)

	// EnterATTR_OPER_EXPR is called when entering the ATTR_OPER_EXPR production.
	EnterATTR_OPER_EXPR(c *ATTR_OPER_EXPRContext)

	// EnterEXPR_OR_EXPR is called when entering the EXPR_OR_EXPR production.
	EnterEXPR_OR_EXPR(c *EXPR_OR_EXPRContext)

	// EnterEXPR_OPER_EXPR is called when entering the EXPR_OPER_EXPR production.
	EnterEXPR_OPER_EXPR(c *EXPR_OPER_EXPRContext)

	// EnterNOT_EXPR is called when entering the NOT_EXPR production.
	EnterNOT_EXPR(c *NOT_EXPRContext)

	// EnterEXPR_AND_EXPR is called when entering the EXPR_AND_EXPR production.
	EnterEXPR_AND_EXPR(c *EXPR_AND_EXPRContext)

	// EnterATTR_OPER_VALUE is called when entering the ATTR_OPER_VALUE production.
	EnterATTR_OPER_VALUE(c *ATTR_OPER_VALUEContext)

	// EnterATTR_OPER_CRITERIA is called when entering the ATTR_OPER_CRITERIA production.
	EnterATTR_OPER_CRITERIA(c *ATTR_OPER_CRITERIAContext)

	// EnterLPAREN_EXPR_RPAREN is called when entering the LPAREN_EXPR_RPAREN production.
	EnterLPAREN_EXPR_RPAREN(c *LPAREN_EXPR_RPARENContext)

	// EnterCriteria is called when entering the criteria production.
	EnterCriteria(c *CriteriaContext)

	// EnterCriteriaValue is called when entering the criteriaValue production.
	EnterCriteriaValue(c *CriteriaValueContext)

	// EnterOperator is called when entering the operator production.
	EnterOperator(c *OperatorContext)

	// ExitStart is called when exiting the start production.
	ExitStart(c *StartContext)

	// ExitATTR_PR is called when exiting the ATTR_PR production.
	ExitATTR_PR(c *ATTR_PRContext)

	// ExitLBRAC_EXPR_RBRAC is called when exiting the LBRAC_EXPR_RBRAC production.
	ExitLBRAC_EXPR_RBRAC(c *LBRAC_EXPR_RBRACContext)

	// ExitATTR_OPER_EXPR is called when exiting the ATTR_OPER_EXPR production.
	ExitATTR_OPER_EXPR(c *ATTR_OPER_EXPRContext)

	// ExitEXPR_OR_EXPR is called when exiting the EXPR_OR_EXPR production.
	ExitEXPR_OR_EXPR(c *EXPR_OR_EXPRContext)

	// ExitEXPR_OPER_EXPR is called when exiting the EXPR_OPER_EXPR production.
	ExitEXPR_OPER_EXPR(c *EXPR_OPER_EXPRContext)

	// ExitNOT_EXPR is called when exiting the NOT_EXPR production.
	ExitNOT_EXPR(c *NOT_EXPRContext)

	// ExitEXPR_AND_EXPR is called when exiting the EXPR_AND_EXPR production.
	ExitEXPR_AND_EXPR(c *EXPR_AND_EXPRContext)

	// ExitATTR_OPER_VALUE is called when exiting the ATTR_OPER_VALUE production.
	ExitATTR_OPER_VALUE(c *ATTR_OPER_VALUEContext)

	// ExitATTR_OPER_CRITERIA is called when exiting the ATTR_OPER_CRITERIA production.
	ExitATTR_OPER_CRITERIA(c *ATTR_OPER_CRITERIAContext)

	// ExitLPAREN_EXPR_RPAREN is called when exiting the LPAREN_EXPR_RPAREN production.
	ExitLPAREN_EXPR_RPAREN(c *LPAREN_EXPR_RPARENContext)

	// ExitCriteria is called when exiting the criteria production.
	ExitCriteria(c *CriteriaContext)

	// ExitCriteriaValue is called when exiting the criteriaValue production.
	ExitCriteriaValue(c *CriteriaValueContext)

	// ExitOperator is called when exiting the operator production.
	ExitOperator(c *OperatorContext)
}
