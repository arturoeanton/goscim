// Generated from ScimFilter.g4 by ANTLR 4.7.

package parser // ScimFilter

import "github.com/antlr/antlr4/runtime/Go/antlr"

// BaseScimFilterListener is a complete listener for a parse tree produced by ScimFilterParser.
type BaseScimFilterListener struct{}

var _ ScimFilterListener = &BaseScimFilterListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseScimFilterListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseScimFilterListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseScimFilterListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseScimFilterListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterStart is called when production start is entered.
func (s *BaseScimFilterListener) EnterStart(ctx *StartContext) {}

// ExitStart is called when production start is exited.
func (s *BaseScimFilterListener) ExitStart(ctx *StartContext) {}

// EnterATTR_PR is called when production ATTR_PR is entered.
func (s *BaseScimFilterListener) EnterATTR_PR(ctx *ATTR_PRContext) {}

// ExitATTR_PR is called when production ATTR_PR is exited.
func (s *BaseScimFilterListener) ExitATTR_PR(ctx *ATTR_PRContext) {}

// EnterLBRAC_EXPR_RBRAC is called when production LBRAC_EXPR_RBRAC is entered.
func (s *BaseScimFilterListener) EnterLBRAC_EXPR_RBRAC(ctx *LBRAC_EXPR_RBRACContext) {}

// ExitLBRAC_EXPR_RBRAC is called when production LBRAC_EXPR_RBRAC is exited.
func (s *BaseScimFilterListener) ExitLBRAC_EXPR_RBRAC(ctx *LBRAC_EXPR_RBRACContext) {}

// EnterATTR_OPER_EXPR is called when production ATTR_OPER_EXPR is entered.
func (s *BaseScimFilterListener) EnterATTR_OPER_EXPR(ctx *ATTR_OPER_EXPRContext) {}

// ExitATTR_OPER_EXPR is called when production ATTR_OPER_EXPR is exited.
func (s *BaseScimFilterListener) ExitATTR_OPER_EXPR(ctx *ATTR_OPER_EXPRContext) {}

// EnterEXPR_OR_EXPR is called when production EXPR_OR_EXPR is entered.
func (s *BaseScimFilterListener) EnterEXPR_OR_EXPR(ctx *EXPR_OR_EXPRContext) {}

// ExitEXPR_OR_EXPR is called when production EXPR_OR_EXPR is exited.
func (s *BaseScimFilterListener) ExitEXPR_OR_EXPR(ctx *EXPR_OR_EXPRContext) {}

// EnterEXPR_OPER_EXPR is called when production EXPR_OPER_EXPR is entered.
func (s *BaseScimFilterListener) EnterEXPR_OPER_EXPR(ctx *EXPR_OPER_EXPRContext) {}

// ExitEXPR_OPER_EXPR is called when production EXPR_OPER_EXPR is exited.
func (s *BaseScimFilterListener) ExitEXPR_OPER_EXPR(ctx *EXPR_OPER_EXPRContext) {}

// EnterNOT_EXPR is called when production NOT_EXPR is entered.
func (s *BaseScimFilterListener) EnterNOT_EXPR(ctx *NOT_EXPRContext) {}

// ExitNOT_EXPR is called when production NOT_EXPR is exited.
func (s *BaseScimFilterListener) ExitNOT_EXPR(ctx *NOT_EXPRContext) {}

// EnterEXPR_AND_EXPR is called when production EXPR_AND_EXPR is entered.
func (s *BaseScimFilterListener) EnterEXPR_AND_EXPR(ctx *EXPR_AND_EXPRContext) {}

// ExitEXPR_AND_EXPR is called when production EXPR_AND_EXPR is exited.
func (s *BaseScimFilterListener) ExitEXPR_AND_EXPR(ctx *EXPR_AND_EXPRContext) {}

// EnterATTR_OPER_VALUE is called when production ATTR_OPER_VALUE is entered.
func (s *BaseScimFilterListener) EnterATTR_OPER_VALUE(ctx *ATTR_OPER_VALUEContext) {}

// ExitATTR_OPER_VALUE is called when production ATTR_OPER_VALUE is exited.
func (s *BaseScimFilterListener) ExitATTR_OPER_VALUE(ctx *ATTR_OPER_VALUEContext) {}

// EnterATTR_OPER_CRITERIA is called when production ATTR_OPER_CRITERIA is entered.
func (s *BaseScimFilterListener) EnterATTR_OPER_CRITERIA(ctx *ATTR_OPER_CRITERIAContext) {}

// ExitATTR_OPER_CRITERIA is called when production ATTR_OPER_CRITERIA is exited.
func (s *BaseScimFilterListener) ExitATTR_OPER_CRITERIA(ctx *ATTR_OPER_CRITERIAContext) {}

// EnterLPAREN_EXPR_RPAREN is called when production LPAREN_EXPR_RPAREN is entered.
func (s *BaseScimFilterListener) EnterLPAREN_EXPR_RPAREN(ctx *LPAREN_EXPR_RPARENContext) {}

// ExitLPAREN_EXPR_RPAREN is called when production LPAREN_EXPR_RPAREN is exited.
func (s *BaseScimFilterListener) ExitLPAREN_EXPR_RPAREN(ctx *LPAREN_EXPR_RPARENContext) {}

// EnterCriteria is called when production criteria is entered.
func (s *BaseScimFilterListener) EnterCriteria(ctx *CriteriaContext) {}

// ExitCriteria is called when production criteria is exited.
func (s *BaseScimFilterListener) ExitCriteria(ctx *CriteriaContext) {}

// EnterCriteriaValue is called when production criteriaValue is entered.
func (s *BaseScimFilterListener) EnterCriteriaValue(ctx *CriteriaValueContext) {}

// ExitCriteriaValue is called when production criteriaValue is exited.
func (s *BaseScimFilterListener) ExitCriteriaValue(ctx *CriteriaValueContext) {}

// EnterOperator is called when production operator is entered.
func (s *BaseScimFilterListener) EnterOperator(ctx *OperatorContext) {}

// ExitOperator is called when production operator is exited.
func (s *BaseScimFilterListener) ExitOperator(ctx *OperatorContext) {}
