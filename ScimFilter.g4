/*
 * antlr-scim-filter is available under the MIT License (2008). See http://opensource.org/licenses/MIT for full text.
 *
 * Copyright (c) 2016, Gluu
 *
 * Author: Val Pecaoco
 */
grammar ScimFilter;

options
{
  language = Java;
}

start
 : expression* EOF
 ;

expression
 : NOT WS+? expression                        # NOT_EXPR
 | expression WS+? AND WS+? expression        # EXPR_AND_EXPR
 | expression WS+? OR WS+ expression          # EXPR_OR_EXPR
 | expression WS+? operator WS+? expression   # EXPR_OPER_EXPR
 | ATTRNAME WS+? PR                           # ATTR_PR
 | ATTRNAME WS+? operator WS+? expression     # ATTR_OPER_EXPR
 | ATTRNAME WS+? operator WS+? criteria       # ATTR_OPER_CRITERIA
 | ATTRNAME WS+? operator WS+? criteriaValue  # ATTR_OPER_VALUE
 | LPAREN WS*? expression WS*? RPAREN         # LPAREN_EXPR_RPAREN
 | ATTRNAME LBRAC WS*? expression WS*? RBRAC  # LBRAC_EXPR_RBRAC
 ;


criteria : '"' .+? '"';
criteriaValue : 
  NUMBERS | BOOLEAN
  ;


operator
 : EQ | NE | CO | SW | EW | GT | LT | GE | LE
 ;

EQ : [eE][qQ];
NE : [nN][eE];
CO : [cC][oO];
SW : [sS][wW];
EW : [eE][wW];
GT : [gG][tT];
LT : [lL][tT];
GE : [gG][eE];
LE : [lL][eE];

NOT : [nN][oO][tT];

AND : [aA][nN][dD];
OR  : [oO][rR];

PR : [pP][rR];

LPAREN : '(';
RPAREN : ')';

LBRAC : '[';
RBRAC : ']';

WS : ' ';

NUMBERS:  [-.0-9]+;
BOOLEAN:  'true'|'false';
ATTRNAME : '$'? [-_.:a-zA-Z0-9]+;


ANY : ~('"' | '(' | ')' | '[' | ']');

EOL : [\t\r\n\u000C]+ -> skip;