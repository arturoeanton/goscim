// Generated from /home/r2t2/git/arturoeanton/goscim/ScimFilter.g4 by ANTLR 4.8
import org.antlr.v4.runtime.atn.*;
import org.antlr.v4.runtime.dfa.DFA;
import org.antlr.v4.runtime.*;
import org.antlr.v4.runtime.misc.*;
import org.antlr.v4.runtime.tree.*;
import java.util.List;
import java.util.Iterator;
import java.util.ArrayList;

@SuppressWarnings({"all", "warnings", "unchecked", "unused", "cast"})
public class ScimFilterParser extends Parser {
	static { RuntimeMetaData.checkVersion("4.8", RuntimeMetaData.VERSION); }

	protected static final DFA[] _decisionToDFA;
	protected static final PredictionContextCache _sharedContextCache =
		new PredictionContextCache();
	public static final int
		T__0=1, EQ=2, NE=3, CO=4, SW=5, EW=6, GT=7, LT=8, GE=9, LE=10, NOT=11, 
		AND=12, OR=13, PR=14, LPAREN=15, RPAREN=16, LBRAC=17, RBRAC=18, WS=19, 
		NUMBERS=20, BOOLEAN=21, ATTRNAME=22, ANY=23, EOL=24;
	public static final int
		RULE_start = 0, RULE_expression = 1, RULE_criteria = 2, RULE_criteriaValue = 3, 
		RULE_operator = 4;
	private static String[] makeRuleNames() {
		return new String[] {
			"start", "expression", "criteria", "criteriaValue", "operator"
		};
	}
	public static final String[] ruleNames = makeRuleNames();

	private static String[] makeLiteralNames() {
		return new String[] {
			null, "'\"'", null, null, null, null, null, null, null, null, null, null, 
			null, null, null, "'('", "')'", "'['", "']'", "' '"
		};
	}
	private static final String[] _LITERAL_NAMES = makeLiteralNames();
	private static String[] makeSymbolicNames() {
		return new String[] {
			null, null, "EQ", "NE", "CO", "SW", "EW", "GT", "LT", "GE", "LE", "NOT", 
			"AND", "OR", "PR", "LPAREN", "RPAREN", "LBRAC", "RBRAC", "WS", "NUMBERS", 
			"BOOLEAN", "ATTRNAME", "ANY", "EOL"
		};
	}
	private static final String[] _SYMBOLIC_NAMES = makeSymbolicNames();
	public static final Vocabulary VOCABULARY = new VocabularyImpl(_LITERAL_NAMES, _SYMBOLIC_NAMES);

	/**
	 * @deprecated Use {@link #VOCABULARY} instead.
	 */
	@Deprecated
	public static final String[] tokenNames;
	static {
		tokenNames = new String[_SYMBOLIC_NAMES.length];
		for (int i = 0; i < tokenNames.length; i++) {
			tokenNames[i] = VOCABULARY.getLiteralName(i);
			if (tokenNames[i] == null) {
				tokenNames[i] = VOCABULARY.getSymbolicName(i);
			}

			if (tokenNames[i] == null) {
				tokenNames[i] = "<INVALID>";
			}
		}
	}

	@Override
	@Deprecated
	public String[] getTokenNames() {
		return tokenNames;
	}

	@Override

	public Vocabulary getVocabulary() {
		return VOCABULARY;
	}

	@Override
	public String getGrammarFileName() { return "ScimFilter.g4"; }

	@Override
	public String[] getRuleNames() { return ruleNames; }

	@Override
	public String getSerializedATN() { return _serializedATN; }

	@Override
	public ATN getATN() { return _ATN; }

	public ScimFilterParser(TokenStream input) {
		super(input);
		_interp = new ParserATNSimulator(this,_ATN,_decisionToDFA,_sharedContextCache);
	}

	public static class StartContext extends ParserRuleContext {
		public TerminalNode EOF() { return getToken(ScimFilterParser.EOF, 0); }
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public StartContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_start; }
	}

	public final StartContext start() throws RecognitionException {
		StartContext _localctx = new StartContext(_ctx, getState());
		enterRule(_localctx, 0, RULE_start);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(13);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while ((((_la) & ~0x3f) == 0 && ((1L << _la) & ((1L << NOT) | (1L << LPAREN) | (1L << ATTRNAME))) != 0)) {
				{
				{
				setState(10);
				expression(0);
				}
				}
				setState(15);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(16);
			match(EOF);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	public static class ExpressionContext extends ParserRuleContext {
		public ExpressionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_expression; }
	 
		public ExpressionContext() { }
		public void copyFrom(ExpressionContext ctx) {
			super.copyFrom(ctx);
		}
	}
	public static class ATTR_PRContext extends ExpressionContext {
		public TerminalNode ATTRNAME() { return getToken(ScimFilterParser.ATTRNAME, 0); }
		public TerminalNode PR() { return getToken(ScimFilterParser.PR, 0); }
		public List<TerminalNode> WS() { return getTokens(ScimFilterParser.WS); }
		public TerminalNode WS(int i) {
			return getToken(ScimFilterParser.WS, i);
		}
		public ATTR_PRContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	public static class LBRAC_EXPR_RBRACContext extends ExpressionContext {
		public TerminalNode ATTRNAME() { return getToken(ScimFilterParser.ATTRNAME, 0); }
		public TerminalNode LBRAC() { return getToken(ScimFilterParser.LBRAC, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode RBRAC() { return getToken(ScimFilterParser.RBRAC, 0); }
		public List<TerminalNode> WS() { return getTokens(ScimFilterParser.WS); }
		public TerminalNode WS(int i) {
			return getToken(ScimFilterParser.WS, i);
		}
		public LBRAC_EXPR_RBRACContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	public static class ATTR_OPER_EXPRContext extends ExpressionContext {
		public TerminalNode ATTRNAME() { return getToken(ScimFilterParser.ATTRNAME, 0); }
		public OperatorContext operator() {
			return getRuleContext(OperatorContext.class,0);
		}
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public List<TerminalNode> WS() { return getTokens(ScimFilterParser.WS); }
		public TerminalNode WS(int i) {
			return getToken(ScimFilterParser.WS, i);
		}
		public ATTR_OPER_EXPRContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	public static class EXPR_OR_EXPRContext extends ExpressionContext {
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public TerminalNode OR() { return getToken(ScimFilterParser.OR, 0); }
		public List<TerminalNode> WS() { return getTokens(ScimFilterParser.WS); }
		public TerminalNode WS(int i) {
			return getToken(ScimFilterParser.WS, i);
		}
		public EXPR_OR_EXPRContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	public static class EXPR_OPER_EXPRContext extends ExpressionContext {
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public OperatorContext operator() {
			return getRuleContext(OperatorContext.class,0);
		}
		public List<TerminalNode> WS() { return getTokens(ScimFilterParser.WS); }
		public TerminalNode WS(int i) {
			return getToken(ScimFilterParser.WS, i);
		}
		public EXPR_OPER_EXPRContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	public static class NOT_EXPRContext extends ExpressionContext {
		public TerminalNode NOT() { return getToken(ScimFilterParser.NOT, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public List<TerminalNode> WS() { return getTokens(ScimFilterParser.WS); }
		public TerminalNode WS(int i) {
			return getToken(ScimFilterParser.WS, i);
		}
		public NOT_EXPRContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	public static class EXPR_AND_EXPRContext extends ExpressionContext {
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public TerminalNode AND() { return getToken(ScimFilterParser.AND, 0); }
		public List<TerminalNode> WS() { return getTokens(ScimFilterParser.WS); }
		public TerminalNode WS(int i) {
			return getToken(ScimFilterParser.WS, i);
		}
		public EXPR_AND_EXPRContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	public static class ATTR_OPER_VALUEContext extends ExpressionContext {
		public TerminalNode ATTRNAME() { return getToken(ScimFilterParser.ATTRNAME, 0); }
		public OperatorContext operator() {
			return getRuleContext(OperatorContext.class,0);
		}
		public CriteriaValueContext criteriaValue() {
			return getRuleContext(CriteriaValueContext.class,0);
		}
		public List<TerminalNode> WS() { return getTokens(ScimFilterParser.WS); }
		public TerminalNode WS(int i) {
			return getToken(ScimFilterParser.WS, i);
		}
		public ATTR_OPER_VALUEContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	public static class ATTR_OPER_CRITERIAContext extends ExpressionContext {
		public TerminalNode ATTRNAME() { return getToken(ScimFilterParser.ATTRNAME, 0); }
		public OperatorContext operator() {
			return getRuleContext(OperatorContext.class,0);
		}
		public CriteriaContext criteria() {
			return getRuleContext(CriteriaContext.class,0);
		}
		public List<TerminalNode> WS() { return getTokens(ScimFilterParser.WS); }
		public TerminalNode WS(int i) {
			return getToken(ScimFilterParser.WS, i);
		}
		public ATTR_OPER_CRITERIAContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	public static class LPAREN_EXPR_RPARENContext extends ExpressionContext {
		public TerminalNode LPAREN() { return getToken(ScimFilterParser.LPAREN, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode RPAREN() { return getToken(ScimFilterParser.RPAREN, 0); }
		public List<TerminalNode> WS() { return getTokens(ScimFilterParser.WS); }
		public TerminalNode WS(int i) {
			return getToken(ScimFilterParser.WS, i);
		}
		public LPAREN_EXPR_RPARENContext(ExpressionContext ctx) { copyFrom(ctx); }
	}

	public final ExpressionContext expression() throws RecognitionException {
		return expression(0);
	}

	private ExpressionContext expression(int _p) throws RecognitionException {
		ParserRuleContext _parentctx = _ctx;
		int _parentState = getState();
		ExpressionContext _localctx = new ExpressionContext(_ctx, _parentState);
		ExpressionContext _prevctx = _localctx;
		int _startState = 2;
		enterRecursionRule(_localctx, 2, RULE_expression, _p);
		int _la;
		try {
			int _alt;
			enterOuterAlt(_localctx, 1);
			{
			setState(108);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,13,_ctx) ) {
			case 1:
				{
				_localctx = new NOT_EXPRContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;

				setState(19);
				match(NOT);
				setState(21); 
				_errHandler.sync(this);
				_alt = 1+1;
				do {
					switch (_alt) {
					case 1+1:
						{
						{
						setState(20);
						match(WS);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(23); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,1,_ctx);
				} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				setState(25);
				expression(10);
				}
				break;
			case 2:
				{
				_localctx = new ATTR_PRContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(26);
				match(ATTRNAME);
				setState(28); 
				_errHandler.sync(this);
				_alt = 1+1;
				do {
					switch (_alt) {
					case 1+1:
						{
						{
						setState(27);
						match(WS);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(30); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,2,_ctx);
				} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				setState(32);
				match(PR);
				}
				break;
			case 3:
				{
				_localctx = new ATTR_OPER_EXPRContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(33);
				match(ATTRNAME);
				setState(35); 
				_errHandler.sync(this);
				_alt = 1+1;
				do {
					switch (_alt) {
					case 1+1:
						{
						{
						setState(34);
						match(WS);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(37); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,3,_ctx);
				} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				setState(39);
				operator();
				setState(41); 
				_errHandler.sync(this);
				_alt = 1+1;
				do {
					switch (_alt) {
					case 1+1:
						{
						{
						setState(40);
						match(WS);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(43); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,4,_ctx);
				} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				setState(45);
				expression(5);
				}
				break;
			case 4:
				{
				_localctx = new ATTR_OPER_CRITERIAContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(47);
				match(ATTRNAME);
				setState(49); 
				_errHandler.sync(this);
				_alt = 1+1;
				do {
					switch (_alt) {
					case 1+1:
						{
						{
						setState(48);
						match(WS);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(51); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,5,_ctx);
				} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				setState(53);
				operator();
				setState(55); 
				_errHandler.sync(this);
				_alt = 1+1;
				do {
					switch (_alt) {
					case 1+1:
						{
						{
						setState(54);
						match(WS);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(57); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,6,_ctx);
				} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				setState(59);
				criteria();
				}
				break;
			case 5:
				{
				_localctx = new ATTR_OPER_VALUEContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(61);
				match(ATTRNAME);
				setState(63); 
				_errHandler.sync(this);
				_alt = 1+1;
				do {
					switch (_alt) {
					case 1+1:
						{
						{
						setState(62);
						match(WS);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(65); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,7,_ctx);
				} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				setState(67);
				operator();
				setState(69); 
				_errHandler.sync(this);
				_alt = 1+1;
				do {
					switch (_alt) {
					case 1+1:
						{
						{
						setState(68);
						match(WS);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(71); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,8,_ctx);
				} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				setState(73);
				criteriaValue();
				}
				break;
			case 6:
				{
				_localctx = new LPAREN_EXPR_RPARENContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(75);
				match(LPAREN);
				setState(79);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,9,_ctx);
				while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
					if ( _alt==1+1 ) {
						{
						{
						setState(76);
						match(WS);
						}
						} 
					}
					setState(81);
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,9,_ctx);
				}
				setState(82);
				expression(0);
				setState(86);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,10,_ctx);
				while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
					if ( _alt==1+1 ) {
						{
						{
						setState(83);
						match(WS);
						}
						} 
					}
					setState(88);
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,10,_ctx);
				}
				setState(89);
				match(RPAREN);
				}
				break;
			case 7:
				{
				_localctx = new LBRAC_EXPR_RBRACContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(91);
				match(ATTRNAME);
				setState(92);
				match(LBRAC);
				setState(96);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,11,_ctx);
				while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
					if ( _alt==1+1 ) {
						{
						{
						setState(93);
						match(WS);
						}
						} 
					}
					setState(98);
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,11,_ctx);
				}
				setState(99);
				expression(0);
				setState(103);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,12,_ctx);
				while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
					if ( _alt==1+1 ) {
						{
						{
						setState(100);
						match(WS);
						}
						} 
					}
					setState(105);
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,12,_ctx);
				}
				setState(106);
				match(RBRAC);
				}
				break;
			}
			_ctx.stop = _input.LT(-1);
			setState(152);
			_errHandler.sync(this);
			_alt = getInterpreter().adaptivePredict(_input,21,_ctx);
			while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
				if ( _alt==1 ) {
					if ( _parseListeners!=null ) triggerExitRuleEvent();
					_prevctx = _localctx;
					{
					setState(150);
					_errHandler.sync(this);
					switch ( getInterpreter().adaptivePredict(_input,20,_ctx) ) {
					case 1:
						{
						_localctx = new EXPR_AND_EXPRContext(new ExpressionContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_expression);
						setState(110);
						if (!(precpred(_ctx, 9))) throw new FailedPredicateException(this, "precpred(_ctx, 9)");
						setState(112); 
						_errHandler.sync(this);
						_alt = 1+1;
						do {
							switch (_alt) {
							case 1+1:
								{
								{
								setState(111);
								match(WS);
								}
								}
								break;
							default:
								throw new NoViableAltException(this);
							}
							setState(114); 
							_errHandler.sync(this);
							_alt = getInterpreter().adaptivePredict(_input,14,_ctx);
						} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
						setState(116);
						match(AND);
						setState(118); 
						_errHandler.sync(this);
						_alt = 1+1;
						do {
							switch (_alt) {
							case 1+1:
								{
								{
								setState(117);
								match(WS);
								}
								}
								break;
							default:
								throw new NoViableAltException(this);
							}
							setState(120); 
							_errHandler.sync(this);
							_alt = getInterpreter().adaptivePredict(_input,15,_ctx);
						} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
						setState(122);
						expression(10);
						}
						break;
					case 2:
						{
						_localctx = new EXPR_OR_EXPRContext(new ExpressionContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_expression);
						setState(123);
						if (!(precpred(_ctx, 8))) throw new FailedPredicateException(this, "precpred(_ctx, 8)");
						setState(125); 
						_errHandler.sync(this);
						_alt = 1+1;
						do {
							switch (_alt) {
							case 1+1:
								{
								{
								setState(124);
								match(WS);
								}
								}
								break;
							default:
								throw new NoViableAltException(this);
							}
							setState(127); 
							_errHandler.sync(this);
							_alt = getInterpreter().adaptivePredict(_input,16,_ctx);
						} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
						setState(129);
						match(OR);
						setState(131); 
						_errHandler.sync(this);
						_la = _input.LA(1);
						do {
							{
							{
							setState(130);
							match(WS);
							}
							}
							setState(133); 
							_errHandler.sync(this);
							_la = _input.LA(1);
						} while ( _la==WS );
						setState(135);
						expression(9);
						}
						break;
					case 3:
						{
						_localctx = new EXPR_OPER_EXPRContext(new ExpressionContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_expression);
						setState(136);
						if (!(precpred(_ctx, 7))) throw new FailedPredicateException(this, "precpred(_ctx, 7)");
						setState(138); 
						_errHandler.sync(this);
						_alt = 1+1;
						do {
							switch (_alt) {
							case 1+1:
								{
								{
								setState(137);
								match(WS);
								}
								}
								break;
							default:
								throw new NoViableAltException(this);
							}
							setState(140); 
							_errHandler.sync(this);
							_alt = getInterpreter().adaptivePredict(_input,18,_ctx);
						} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
						setState(142);
						operator();
						setState(144); 
						_errHandler.sync(this);
						_alt = 1+1;
						do {
							switch (_alt) {
							case 1+1:
								{
								{
								setState(143);
								match(WS);
								}
								}
								break;
							default:
								throw new NoViableAltException(this);
							}
							setState(146); 
							_errHandler.sync(this);
							_alt = getInterpreter().adaptivePredict(_input,19,_ctx);
						} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
						setState(148);
						expression(8);
						}
						break;
					}
					} 
				}
				setState(154);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,21,_ctx);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			unrollRecursionContexts(_parentctx);
		}
		return _localctx;
	}

	public static class CriteriaContext extends ParserRuleContext {
		public CriteriaContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_criteria; }
	}

	public final CriteriaContext criteria() throws RecognitionException {
		CriteriaContext _localctx = new CriteriaContext(_ctx, getState());
		enterRule(_localctx, 4, RULE_criteria);
		try {
			int _alt;
			enterOuterAlt(_localctx, 1);
			{
			setState(155);
			match(T__0);
			setState(157); 
			_errHandler.sync(this);
			_alt = 1+1;
			do {
				switch (_alt) {
				case 1+1:
					{
					{
					setState(156);
					matchWildcard();
					}
					}
					break;
				default:
					throw new NoViableAltException(this);
				}
				setState(159); 
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,22,_ctx);
			} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
			setState(161);
			match(T__0);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	public static class CriteriaValueContext extends ParserRuleContext {
		public TerminalNode NUMBERS() { return getToken(ScimFilterParser.NUMBERS, 0); }
		public TerminalNode BOOLEAN() { return getToken(ScimFilterParser.BOOLEAN, 0); }
		public CriteriaValueContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_criteriaValue; }
	}

	public final CriteriaValueContext criteriaValue() throws RecognitionException {
		CriteriaValueContext _localctx = new CriteriaValueContext(_ctx, getState());
		enterRule(_localctx, 6, RULE_criteriaValue);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(163);
			_la = _input.LA(1);
			if ( !(_la==NUMBERS || _la==BOOLEAN) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	public static class OperatorContext extends ParserRuleContext {
		public TerminalNode EQ() { return getToken(ScimFilterParser.EQ, 0); }
		public TerminalNode NE() { return getToken(ScimFilterParser.NE, 0); }
		public TerminalNode CO() { return getToken(ScimFilterParser.CO, 0); }
		public TerminalNode SW() { return getToken(ScimFilterParser.SW, 0); }
		public TerminalNode EW() { return getToken(ScimFilterParser.EW, 0); }
		public TerminalNode GT() { return getToken(ScimFilterParser.GT, 0); }
		public TerminalNode LT() { return getToken(ScimFilterParser.LT, 0); }
		public TerminalNode GE() { return getToken(ScimFilterParser.GE, 0); }
		public TerminalNode LE() { return getToken(ScimFilterParser.LE, 0); }
		public OperatorContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_operator; }
	}

	public final OperatorContext operator() throws RecognitionException {
		OperatorContext _localctx = new OperatorContext(_ctx, getState());
		enterRule(_localctx, 8, RULE_operator);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(165);
			_la = _input.LA(1);
			if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & ((1L << EQ) | (1L << NE) | (1L << CO) | (1L << SW) | (1L << EW) | (1L << GT) | (1L << LT) | (1L << GE) | (1L << LE))) != 0)) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	public boolean sempred(RuleContext _localctx, int ruleIndex, int predIndex) {
		switch (ruleIndex) {
		case 1:
			return expression_sempred((ExpressionContext)_localctx, predIndex);
		}
		return true;
	}
	private boolean expression_sempred(ExpressionContext _localctx, int predIndex) {
		switch (predIndex) {
		case 0:
			return precpred(_ctx, 9);
		case 1:
			return precpred(_ctx, 8);
		case 2:
			return precpred(_ctx, 7);
		}
		return true;
	}

	public static final String _serializedATN =
		"\3\u608b\ua72a\u8133\ub9ed\u417c\u3be7\u7786\u5964\3\32\u00aa\4\2\t\2"+
		"\4\3\t\3\4\4\t\4\4\5\t\5\4\6\t\6\3\2\7\2\16\n\2\f\2\16\2\21\13\2\3\2\3"+
		"\2\3\3\3\3\3\3\6\3\30\n\3\r\3\16\3\31\3\3\3\3\3\3\6\3\37\n\3\r\3\16\3"+
		" \3\3\3\3\3\3\6\3&\n\3\r\3\16\3\'\3\3\3\3\6\3,\n\3\r\3\16\3-\3\3\3\3\3"+
		"\3\3\3\6\3\64\n\3\r\3\16\3\65\3\3\3\3\6\3:\n\3\r\3\16\3;\3\3\3\3\3\3\3"+
		"\3\6\3B\n\3\r\3\16\3C\3\3\3\3\6\3H\n\3\r\3\16\3I\3\3\3\3\3\3\3\3\7\3P"+
		"\n\3\f\3\16\3S\13\3\3\3\3\3\7\3W\n\3\f\3\16\3Z\13\3\3\3\3\3\3\3\3\3\3"+
		"\3\7\3a\n\3\f\3\16\3d\13\3\3\3\3\3\7\3h\n\3\f\3\16\3k\13\3\3\3\3\3\5\3"+
		"o\n\3\3\3\3\3\6\3s\n\3\r\3\16\3t\3\3\3\3\6\3y\n\3\r\3\16\3z\3\3\3\3\3"+
		"\3\6\3\u0080\n\3\r\3\16\3\u0081\3\3\3\3\6\3\u0086\n\3\r\3\16\3\u0087\3"+
		"\3\3\3\3\3\6\3\u008d\n\3\r\3\16\3\u008e\3\3\3\3\6\3\u0093\n\3\r\3\16\3"+
		"\u0094\3\3\3\3\7\3\u0099\n\3\f\3\16\3\u009c\13\3\3\4\3\4\6\4\u00a0\n\4"+
		"\r\4\16\4\u00a1\3\4\3\4\3\5\3\5\3\6\3\6\3\6\24\31 \'-\65;CIQXbitz\u0081"+
		"\u008e\u0094\u00a1\3\4\7\2\4\6\b\n\2\4\3\2\26\27\3\2\4\f\2\u00c1\2\17"+
		"\3\2\2\2\4n\3\2\2\2\6\u009d\3\2\2\2\b\u00a5\3\2\2\2\n\u00a7\3\2\2\2\f"+
		"\16\5\4\3\2\r\f\3\2\2\2\16\21\3\2\2\2\17\r\3\2\2\2\17\20\3\2\2\2\20\22"+
		"\3\2\2\2\21\17\3\2\2\2\22\23\7\2\2\3\23\3\3\2\2\2\24\25\b\3\1\2\25\27"+
		"\7\r\2\2\26\30\7\25\2\2\27\26\3\2\2\2\30\31\3\2\2\2\31\32\3\2\2\2\31\27"+
		"\3\2\2\2\32\33\3\2\2\2\33o\5\4\3\f\34\36\7\30\2\2\35\37\7\25\2\2\36\35"+
		"\3\2\2\2\37 \3\2\2\2 !\3\2\2\2 \36\3\2\2\2!\"\3\2\2\2\"o\7\20\2\2#%\7"+
		"\30\2\2$&\7\25\2\2%$\3\2\2\2&\'\3\2\2\2\'(\3\2\2\2\'%\3\2\2\2()\3\2\2"+
		"\2)+\5\n\6\2*,\7\25\2\2+*\3\2\2\2,-\3\2\2\2-.\3\2\2\2-+\3\2\2\2./\3\2"+
		"\2\2/\60\5\4\3\7\60o\3\2\2\2\61\63\7\30\2\2\62\64\7\25\2\2\63\62\3\2\2"+
		"\2\64\65\3\2\2\2\65\66\3\2\2\2\65\63\3\2\2\2\66\67\3\2\2\2\679\5\n\6\2"+
		"8:\7\25\2\298\3\2\2\2:;\3\2\2\2;<\3\2\2\2;9\3\2\2\2<=\3\2\2\2=>\5\6\4"+
		"\2>o\3\2\2\2?A\7\30\2\2@B\7\25\2\2A@\3\2\2\2BC\3\2\2\2CD\3\2\2\2CA\3\2"+
		"\2\2DE\3\2\2\2EG\5\n\6\2FH\7\25\2\2GF\3\2\2\2HI\3\2\2\2IJ\3\2\2\2IG\3"+
		"\2\2\2JK\3\2\2\2KL\5\b\5\2Lo\3\2\2\2MQ\7\21\2\2NP\7\25\2\2ON\3\2\2\2P"+
		"S\3\2\2\2QR\3\2\2\2QO\3\2\2\2RT\3\2\2\2SQ\3\2\2\2TX\5\4\3\2UW\7\25\2\2"+
		"VU\3\2\2\2WZ\3\2\2\2XY\3\2\2\2XV\3\2\2\2Y[\3\2\2\2ZX\3\2\2\2[\\\7\22\2"+
		"\2\\o\3\2\2\2]^\7\30\2\2^b\7\23\2\2_a\7\25\2\2`_\3\2\2\2ad\3\2\2\2bc\3"+
		"\2\2\2b`\3\2\2\2ce\3\2\2\2db\3\2\2\2ei\5\4\3\2fh\7\25\2\2gf\3\2\2\2hk"+
		"\3\2\2\2ij\3\2\2\2ig\3\2\2\2jl\3\2\2\2ki\3\2\2\2lm\7\24\2\2mo\3\2\2\2"+
		"n\24\3\2\2\2n\34\3\2\2\2n#\3\2\2\2n\61\3\2\2\2n?\3\2\2\2nM\3\2\2\2n]\3"+
		"\2\2\2o\u009a\3\2\2\2pr\f\13\2\2qs\7\25\2\2rq\3\2\2\2st\3\2\2\2tu\3\2"+
		"\2\2tr\3\2\2\2uv\3\2\2\2vx\7\16\2\2wy\7\25\2\2xw\3\2\2\2yz\3\2\2\2z{\3"+
		"\2\2\2zx\3\2\2\2{|\3\2\2\2|\u0099\5\4\3\f}\177\f\n\2\2~\u0080\7\25\2\2"+
		"\177~\3\2\2\2\u0080\u0081\3\2\2\2\u0081\u0082\3\2\2\2\u0081\177\3\2\2"+
		"\2\u0082\u0083\3\2\2\2\u0083\u0085\7\17\2\2\u0084\u0086\7\25\2\2\u0085"+
		"\u0084\3\2\2\2\u0086\u0087\3\2\2\2\u0087\u0085\3\2\2\2\u0087\u0088\3\2"+
		"\2\2\u0088\u0089\3\2\2\2\u0089\u0099\5\4\3\13\u008a\u008c\f\t\2\2\u008b"+
		"\u008d\7\25\2\2\u008c\u008b\3\2\2\2\u008d\u008e\3\2\2\2\u008e\u008f\3"+
		"\2\2\2\u008e\u008c\3\2\2\2\u008f\u0090\3\2\2\2\u0090\u0092\5\n\6\2\u0091"+
		"\u0093\7\25\2\2\u0092\u0091\3\2\2\2\u0093\u0094\3\2\2\2\u0094\u0095\3"+
		"\2\2\2\u0094\u0092\3\2\2\2\u0095\u0096\3\2\2\2\u0096\u0097\5\4\3\n\u0097"+
		"\u0099\3\2\2\2\u0098p\3\2\2\2\u0098}\3\2\2\2\u0098\u008a\3\2\2\2\u0099"+
		"\u009c\3\2\2\2\u009a\u0098\3\2\2\2\u009a\u009b\3\2\2\2\u009b\5\3\2\2\2"+
		"\u009c\u009a\3\2\2\2\u009d\u009f\7\3\2\2\u009e\u00a0\13\2\2\2\u009f\u009e"+
		"\3\2\2\2\u00a0\u00a1\3\2\2\2\u00a1\u00a2\3\2\2\2\u00a1\u009f\3\2\2\2\u00a2"+
		"\u00a3\3\2\2\2\u00a3\u00a4\7\3\2\2\u00a4\7\3\2\2\2\u00a5\u00a6\t\2\2\2"+
		"\u00a6\t\3\2\2\2\u00a7\u00a8\t\3\2\2\u00a8\13\3\2\2\2\31\17\31 \'-\65"+
		";CIQXbintz\u0081\u0087\u008e\u0094\u0098\u009a\u00a1";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}