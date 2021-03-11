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
		ATTRNAME=20, ANY=21, EOL=22;
	public static final int
		RULE_start = 0, RULE_expression = 1, RULE_criteria = 2, RULE_operator = 3;
	private static String[] makeRuleNames() {
		return new String[] {
			"start", "expression", "criteria", "operator"
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
			"AND", "OR", "PR", "LPAREN", "RPAREN", "LBRAC", "RBRAC", "WS", "ATTRNAME", 
			"ANY", "EOL"
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
			setState(11);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while ((((_la) & ~0x3f) == 0 && ((1L << _la) & ((1L << NOT) | (1L << LPAREN) | (1L << ATTRNAME))) != 0)) {
				{
				{
				setState(8);
				expression(0);
				}
				}
				setState(13);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(14);
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
			setState(92);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,11,_ctx) ) {
			case 1:
				{
				_localctx = new NOT_EXPRContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;

				setState(17);
				match(NOT);
				setState(19); 
				_errHandler.sync(this);
				_alt = 1+1;
				do {
					switch (_alt) {
					case 1+1:
						{
						{
						setState(18);
						match(WS);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(21); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,1,_ctx);
				} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				setState(23);
				expression(9);
				}
				break;
			case 2:
				{
				_localctx = new ATTR_PRContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(24);
				match(ATTRNAME);
				setState(26); 
				_errHandler.sync(this);
				_alt = 1+1;
				do {
					switch (_alt) {
					case 1+1:
						{
						{
						setState(25);
						match(WS);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(28); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,2,_ctx);
				} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				setState(30);
				match(PR);
				}
				break;
			case 3:
				{
				_localctx = new ATTR_OPER_EXPRContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(31);
				match(ATTRNAME);
				setState(33); 
				_errHandler.sync(this);
				_alt = 1+1;
				do {
					switch (_alt) {
					case 1+1:
						{
						{
						setState(32);
						match(WS);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(35); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,3,_ctx);
				} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				setState(37);
				operator();
				setState(39); 
				_errHandler.sync(this);
				_alt = 1+1;
				do {
					switch (_alt) {
					case 1+1:
						{
						{
						setState(38);
						match(WS);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(41); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,4,_ctx);
				} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				setState(43);
				expression(4);
				}
				break;
			case 4:
				{
				_localctx = new ATTR_OPER_CRITERIAContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(45);
				match(ATTRNAME);
				setState(47); 
				_errHandler.sync(this);
				_alt = 1+1;
				do {
					switch (_alt) {
					case 1+1:
						{
						{
						setState(46);
						match(WS);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(49); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,5,_ctx);
				} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				setState(51);
				operator();
				setState(53); 
				_errHandler.sync(this);
				_alt = 1+1;
				do {
					switch (_alt) {
					case 1+1:
						{
						{
						setState(52);
						match(WS);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(55); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,6,_ctx);
				} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				setState(57);
				criteria();
				}
				break;
			case 5:
				{
				_localctx = new LPAREN_EXPR_RPARENContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(59);
				match(LPAREN);
				setState(63);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,7,_ctx);
				while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
					if ( _alt==1+1 ) {
						{
						{
						setState(60);
						match(WS);
						}
						} 
					}
					setState(65);
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,7,_ctx);
				}
				setState(66);
				expression(0);
				setState(70);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,8,_ctx);
				while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
					if ( _alt==1+1 ) {
						{
						{
						setState(67);
						match(WS);
						}
						} 
					}
					setState(72);
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,8,_ctx);
				}
				setState(73);
				match(RPAREN);
				}
				break;
			case 6:
				{
				_localctx = new LBRAC_EXPR_RBRACContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(75);
				match(ATTRNAME);
				setState(76);
				match(LBRAC);
				setState(80);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,9,_ctx);
				while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
					if ( _alt==1+1 ) {
						{
						{
						setState(77);
						match(WS);
						}
						} 
					}
					setState(82);
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,9,_ctx);
				}
				setState(83);
				expression(0);
				setState(87);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,10,_ctx);
				while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
					if ( _alt==1+1 ) {
						{
						{
						setState(84);
						match(WS);
						}
						} 
					}
					setState(89);
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,10,_ctx);
				}
				setState(90);
				match(RBRAC);
				}
				break;
			}
			_ctx.stop = _input.LT(-1);
			setState(136);
			_errHandler.sync(this);
			_alt = getInterpreter().adaptivePredict(_input,19,_ctx);
			while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
				if ( _alt==1 ) {
					if ( _parseListeners!=null ) triggerExitRuleEvent();
					_prevctx = _localctx;
					{
					setState(134);
					_errHandler.sync(this);
					switch ( getInterpreter().adaptivePredict(_input,18,_ctx) ) {
					case 1:
						{
						_localctx = new EXPR_AND_EXPRContext(new ExpressionContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_expression);
						setState(94);
						if (!(precpred(_ctx, 8))) throw new FailedPredicateException(this, "precpred(_ctx, 8)");
						setState(96); 
						_errHandler.sync(this);
						_alt = 1+1;
						do {
							switch (_alt) {
							case 1+1:
								{
								{
								setState(95);
								match(WS);
								}
								}
								break;
							default:
								throw new NoViableAltException(this);
							}
							setState(98); 
							_errHandler.sync(this);
							_alt = getInterpreter().adaptivePredict(_input,12,_ctx);
						} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
						setState(100);
						match(AND);
						setState(102); 
						_errHandler.sync(this);
						_alt = 1+1;
						do {
							switch (_alt) {
							case 1+1:
								{
								{
								setState(101);
								match(WS);
								}
								}
								break;
							default:
								throw new NoViableAltException(this);
							}
							setState(104); 
							_errHandler.sync(this);
							_alt = getInterpreter().adaptivePredict(_input,13,_ctx);
						} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
						setState(106);
						expression(9);
						}
						break;
					case 2:
						{
						_localctx = new EXPR_OR_EXPRContext(new ExpressionContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_expression);
						setState(107);
						if (!(precpred(_ctx, 7))) throw new FailedPredicateException(this, "precpred(_ctx, 7)");
						setState(109); 
						_errHandler.sync(this);
						_alt = 1+1;
						do {
							switch (_alt) {
							case 1+1:
								{
								{
								setState(108);
								match(WS);
								}
								}
								break;
							default:
								throw new NoViableAltException(this);
							}
							setState(111); 
							_errHandler.sync(this);
							_alt = getInterpreter().adaptivePredict(_input,14,_ctx);
						} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
						setState(113);
						match(OR);
						setState(115); 
						_errHandler.sync(this);
						_la = _input.LA(1);
						do {
							{
							{
							setState(114);
							match(WS);
							}
							}
							setState(117); 
							_errHandler.sync(this);
							_la = _input.LA(1);
						} while ( _la==WS );
						setState(119);
						expression(8);
						}
						break;
					case 3:
						{
						_localctx = new EXPR_OPER_EXPRContext(new ExpressionContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_expression);
						setState(120);
						if (!(precpred(_ctx, 6))) throw new FailedPredicateException(this, "precpred(_ctx, 6)");
						setState(122); 
						_errHandler.sync(this);
						_alt = 1+1;
						do {
							switch (_alt) {
							case 1+1:
								{
								{
								setState(121);
								match(WS);
								}
								}
								break;
							default:
								throw new NoViableAltException(this);
							}
							setState(124); 
							_errHandler.sync(this);
							_alt = getInterpreter().adaptivePredict(_input,16,_ctx);
						} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
						setState(126);
						operator();
						setState(128); 
						_errHandler.sync(this);
						_alt = 1+1;
						do {
							switch (_alt) {
							case 1+1:
								{
								{
								setState(127);
								match(WS);
								}
								}
								break;
							default:
								throw new NoViableAltException(this);
							}
							setState(130); 
							_errHandler.sync(this);
							_alt = getInterpreter().adaptivePredict(_input,17,_ctx);
						} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
						setState(132);
						expression(7);
						}
						break;
					}
					} 
				}
				setState(138);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,19,_ctx);
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
			setState(139);
			match(T__0);
			setState(141); 
			_errHandler.sync(this);
			_alt = 1+1;
			do {
				switch (_alt) {
				case 1+1:
					{
					{
					setState(140);
					matchWildcard();
					}
					}
					break;
				default:
					throw new NoViableAltException(this);
				}
				setState(143); 
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,20,_ctx);
			} while ( _alt!=1 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
			setState(145);
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
		enterRule(_localctx, 6, RULE_operator);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(147);
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
			return precpred(_ctx, 8);
		case 1:
			return precpred(_ctx, 7);
		case 2:
			return precpred(_ctx, 6);
		}
		return true;
	}

	public static final String _serializedATN =
		"\3\u608b\ua72a\u8133\ub9ed\u417c\u3be7\u7786\u5964\3\30\u0098\4\2\t\2"+
		"\4\3\t\3\4\4\t\4\4\5\t\5\3\2\7\2\f\n\2\f\2\16\2\17\13\2\3\2\3\2\3\3\3"+
		"\3\3\3\6\3\26\n\3\r\3\16\3\27\3\3\3\3\3\3\6\3\35\n\3\r\3\16\3\36\3\3\3"+
		"\3\3\3\6\3$\n\3\r\3\16\3%\3\3\3\3\6\3*\n\3\r\3\16\3+\3\3\3\3\3\3\3\3\6"+
		"\3\62\n\3\r\3\16\3\63\3\3\3\3\6\38\n\3\r\3\16\39\3\3\3\3\3\3\3\3\7\3@"+
		"\n\3\f\3\16\3C\13\3\3\3\3\3\7\3G\n\3\f\3\16\3J\13\3\3\3\3\3\3\3\3\3\3"+
		"\3\7\3Q\n\3\f\3\16\3T\13\3\3\3\3\3\7\3X\n\3\f\3\16\3[\13\3\3\3\3\3\5\3"+
		"_\n\3\3\3\3\3\6\3c\n\3\r\3\16\3d\3\3\3\3\6\3i\n\3\r\3\16\3j\3\3\3\3\3"+
		"\3\6\3p\n\3\r\3\16\3q\3\3\3\3\6\3v\n\3\r\3\16\3w\3\3\3\3\3\3\6\3}\n\3"+
		"\r\3\16\3~\3\3\3\3\6\3\u0083\n\3\r\3\16\3\u0084\3\3\3\3\7\3\u0089\n\3"+
		"\f\3\16\3\u008c\13\3\3\4\3\4\6\4\u0090\n\4\r\4\16\4\u0091\3\4\3\4\3\5"+
		"\3\5\3\5\22\27\36%+\639AHRYdjq~\u0084\u0091\3\4\6\2\4\6\b\2\3\3\2\4\f"+
		"\2\u00ad\2\r\3\2\2\2\4^\3\2\2\2\6\u008d\3\2\2\2\b\u0095\3\2\2\2\n\f\5"+
		"\4\3\2\13\n\3\2\2\2\f\17\3\2\2\2\r\13\3\2\2\2\r\16\3\2\2\2\16\20\3\2\2"+
		"\2\17\r\3\2\2\2\20\21\7\2\2\3\21\3\3\2\2\2\22\23\b\3\1\2\23\25\7\r\2\2"+
		"\24\26\7\25\2\2\25\24\3\2\2\2\26\27\3\2\2\2\27\30\3\2\2\2\27\25\3\2\2"+
		"\2\30\31\3\2\2\2\31_\5\4\3\13\32\34\7\26\2\2\33\35\7\25\2\2\34\33\3\2"+
		"\2\2\35\36\3\2\2\2\36\37\3\2\2\2\36\34\3\2\2\2\37 \3\2\2\2 _\7\20\2\2"+
		"!#\7\26\2\2\"$\7\25\2\2#\"\3\2\2\2$%\3\2\2\2%&\3\2\2\2%#\3\2\2\2&\'\3"+
		"\2\2\2\')\5\b\5\2(*\7\25\2\2)(\3\2\2\2*+\3\2\2\2+,\3\2\2\2+)\3\2\2\2,"+
		"-\3\2\2\2-.\5\4\3\6._\3\2\2\2/\61\7\26\2\2\60\62\7\25\2\2\61\60\3\2\2"+
		"\2\62\63\3\2\2\2\63\64\3\2\2\2\63\61\3\2\2\2\64\65\3\2\2\2\65\67\5\b\5"+
		"\2\668\7\25\2\2\67\66\3\2\2\289\3\2\2\29:\3\2\2\29\67\3\2\2\2:;\3\2\2"+
		"\2;<\5\6\4\2<_\3\2\2\2=A\7\21\2\2>@\7\25\2\2?>\3\2\2\2@C\3\2\2\2AB\3\2"+
		"\2\2A?\3\2\2\2BD\3\2\2\2CA\3\2\2\2DH\5\4\3\2EG\7\25\2\2FE\3\2\2\2GJ\3"+
		"\2\2\2HI\3\2\2\2HF\3\2\2\2IK\3\2\2\2JH\3\2\2\2KL\7\22\2\2L_\3\2\2\2MN"+
		"\7\26\2\2NR\7\23\2\2OQ\7\25\2\2PO\3\2\2\2QT\3\2\2\2RS\3\2\2\2RP\3\2\2"+
		"\2SU\3\2\2\2TR\3\2\2\2UY\5\4\3\2VX\7\25\2\2WV\3\2\2\2X[\3\2\2\2YZ\3\2"+
		"\2\2YW\3\2\2\2Z\\\3\2\2\2[Y\3\2\2\2\\]\7\24\2\2]_\3\2\2\2^\22\3\2\2\2"+
		"^\32\3\2\2\2^!\3\2\2\2^/\3\2\2\2^=\3\2\2\2^M\3\2\2\2_\u008a\3\2\2\2`b"+
		"\f\n\2\2ac\7\25\2\2ba\3\2\2\2cd\3\2\2\2de\3\2\2\2db\3\2\2\2ef\3\2\2\2"+
		"fh\7\16\2\2gi\7\25\2\2hg\3\2\2\2ij\3\2\2\2jk\3\2\2\2jh\3\2\2\2kl\3\2\2"+
		"\2l\u0089\5\4\3\13mo\f\t\2\2np\7\25\2\2on\3\2\2\2pq\3\2\2\2qr\3\2\2\2"+
		"qo\3\2\2\2rs\3\2\2\2su\7\17\2\2tv\7\25\2\2ut\3\2\2\2vw\3\2\2\2wu\3\2\2"+
		"\2wx\3\2\2\2xy\3\2\2\2y\u0089\5\4\3\nz|\f\b\2\2{}\7\25\2\2|{\3\2\2\2}"+
		"~\3\2\2\2~\177\3\2\2\2~|\3\2\2\2\177\u0080\3\2\2\2\u0080\u0082\5\b\5\2"+
		"\u0081\u0083\7\25\2\2\u0082\u0081\3\2\2\2\u0083\u0084\3\2\2\2\u0084\u0085"+
		"\3\2\2\2\u0084\u0082\3\2\2\2\u0085\u0086\3\2\2\2\u0086\u0087\5\4\3\t\u0087"+
		"\u0089\3\2\2\2\u0088`\3\2\2\2\u0088m\3\2\2\2\u0088z\3\2\2\2\u0089\u008c"+
		"\3\2\2\2\u008a\u0088\3\2\2\2\u008a\u008b\3\2\2\2\u008b\5\3\2\2\2\u008c"+
		"\u008a\3\2\2\2\u008d\u008f\7\3\2\2\u008e\u0090\13\2\2\2\u008f\u008e\3"+
		"\2\2\2\u0090\u0091\3\2\2\2\u0091\u0092\3\2\2\2\u0091\u008f\3\2\2\2\u0092"+
		"\u0093\3\2\2\2\u0093\u0094\7\3\2\2\u0094\7\3\2\2\2\u0095\u0096\t\2\2\2"+
		"\u0096\t\3\2\2\2\27\r\27\36%+\639AHRY^djqw~\u0084\u0088\u008a\u0091";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}