// Generated from /home/r2t2/git/arturoeanton/goscim/ScimFilter.g4 by ANTLR 4.8
import org.antlr.v4.runtime.Lexer;
import org.antlr.v4.runtime.CharStream;
import org.antlr.v4.runtime.Token;
import org.antlr.v4.runtime.TokenStream;
import org.antlr.v4.runtime.*;
import org.antlr.v4.runtime.atn.*;
import org.antlr.v4.runtime.dfa.DFA;
import org.antlr.v4.runtime.misc.*;

@SuppressWarnings({"all", "warnings", "unchecked", "unused", "cast"})
public class ScimFilterLexer extends Lexer {
	static { RuntimeMetaData.checkVersion("4.8", RuntimeMetaData.VERSION); }

	protected static final DFA[] _decisionToDFA;
	protected static final PredictionContextCache _sharedContextCache =
		new PredictionContextCache();
	public static final int
		T__0=1, EQ=2, NE=3, CO=4, SW=5, EW=6, GT=7, LT=8, GE=9, LE=10, NOT=11, 
		AND=12, OR=13, PR=14, LPAREN=15, RPAREN=16, LBRAC=17, RBRAC=18, WS=19, 
		ATTRNAME=20, ANY=21, EOL=22;
	public static String[] channelNames = {
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN"
	};

	public static String[] modeNames = {
		"DEFAULT_MODE"
	};

	private static String[] makeRuleNames() {
		return new String[] {
			"T__0", "EQ", "NE", "CO", "SW", "EW", "GT", "LT", "GE", "LE", "NOT", 
			"AND", "OR", "PR", "LPAREN", "RPAREN", "LBRAC", "RBRAC", "WS", "ATTRNAME", 
			"ANY", "EOL"
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


	public ScimFilterLexer(CharStream input) {
		super(input);
		_interp = new LexerATNSimulator(this,_ATN,_decisionToDFA,_sharedContextCache);
	}

	@Override
	public String getGrammarFileName() { return "ScimFilter.g4"; }

	@Override
	public String[] getRuleNames() { return ruleNames; }

	@Override
	public String getSerializedATN() { return _serializedATN; }

	@Override
	public String[] getChannelNames() { return channelNames; }

	@Override
	public String[] getModeNames() { return modeNames; }

	@Override
	public ATN getATN() { return _ATN; }

	public static final String _serializedATN =
		"\3\u608b\ua72a\u8133\ub9ed\u417c\u3be7\u7786\u5964\2\30r\b\1\4\2\t\2\4"+
		"\3\t\3\4\4\t\4\4\5\t\5\4\6\t\6\4\7\t\7\4\b\t\b\4\t\t\t\4\n\t\n\4\13\t"+
		"\13\4\f\t\f\4\r\t\r\4\16\t\16\4\17\t\17\4\20\t\20\4\21\t\21\4\22\t\22"+
		"\4\23\t\23\4\24\t\24\4\25\t\25\4\26\t\26\4\27\t\27\3\2\3\2\3\3\3\3\3\3"+
		"\3\4\3\4\3\4\3\5\3\5\3\5\3\6\3\6\3\6\3\7\3\7\3\7\3\b\3\b\3\b\3\t\3\t\3"+
		"\t\3\n\3\n\3\n\3\13\3\13\3\13\3\f\3\f\3\f\3\f\3\r\3\r\3\r\3\r\3\16\3\16"+
		"\3\16\3\17\3\17\3\17\3\20\3\20\3\21\3\21\3\22\3\22\3\23\3\23\3\24\3\24"+
		"\3\25\6\25f\n\25\r\25\16\25g\3\26\3\26\3\27\6\27m\n\27\r\27\16\27n\3\27"+
		"\3\27\2\2\30\3\3\5\4\7\5\t\6\13\7\r\b\17\t\21\n\23\13\25\f\27\r\31\16"+
		"\33\17\35\20\37\21!\22#\23%\24\'\25)\26+\27-\30\3\2\23\4\2GGgg\4\2SSs"+
		"s\4\2PPpp\4\2EEee\4\2QQqq\4\2UUuu\4\2YYyy\4\2IIii\4\2VVvv\4\2NNnn\4\2"+
		"CCcc\4\2FFff\4\2TTtt\4\2RRrr\7\2/\60\62<C\\aac|\6\2$$*+]]__\4\2\13\f\16"+
		"\17\2s\2\3\3\2\2\2\2\5\3\2\2\2\2\7\3\2\2\2\2\t\3\2\2\2\2\13\3\2\2\2\2"+
		"\r\3\2\2\2\2\17\3\2\2\2\2\21\3\2\2\2\2\23\3\2\2\2\2\25\3\2\2\2\2\27\3"+
		"\2\2\2\2\31\3\2\2\2\2\33\3\2\2\2\2\35\3\2\2\2\2\37\3\2\2\2\2!\3\2\2\2"+
		"\2#\3\2\2\2\2%\3\2\2\2\2\'\3\2\2\2\2)\3\2\2\2\2+\3\2\2\2\2-\3\2\2\2\3"+
		"/\3\2\2\2\5\61\3\2\2\2\7\64\3\2\2\2\t\67\3\2\2\2\13:\3\2\2\2\r=\3\2\2"+
		"\2\17@\3\2\2\2\21C\3\2\2\2\23F\3\2\2\2\25I\3\2\2\2\27L\3\2\2\2\31P\3\2"+
		"\2\2\33T\3\2\2\2\35W\3\2\2\2\37Z\3\2\2\2!\\\3\2\2\2#^\3\2\2\2%`\3\2\2"+
		"\2\'b\3\2\2\2)e\3\2\2\2+i\3\2\2\2-l\3\2\2\2/\60\7$\2\2\60\4\3\2\2\2\61"+
		"\62\t\2\2\2\62\63\t\3\2\2\63\6\3\2\2\2\64\65\t\4\2\2\65\66\t\2\2\2\66"+
		"\b\3\2\2\2\678\t\5\2\289\t\6\2\29\n\3\2\2\2:;\t\7\2\2;<\t\b\2\2<\f\3\2"+
		"\2\2=>\t\2\2\2>?\t\b\2\2?\16\3\2\2\2@A\t\t\2\2AB\t\n\2\2B\20\3\2\2\2C"+
		"D\t\13\2\2DE\t\n\2\2E\22\3\2\2\2FG\t\t\2\2GH\t\2\2\2H\24\3\2\2\2IJ\t\13"+
		"\2\2JK\t\2\2\2K\26\3\2\2\2LM\t\4\2\2MN\t\6\2\2NO\t\n\2\2O\30\3\2\2\2P"+
		"Q\t\f\2\2QR\t\4\2\2RS\t\r\2\2S\32\3\2\2\2TU\t\6\2\2UV\t\16\2\2V\34\3\2"+
		"\2\2WX\t\17\2\2XY\t\16\2\2Y\36\3\2\2\2Z[\7*\2\2[ \3\2\2\2\\]\7+\2\2]\""+
		"\3\2\2\2^_\7]\2\2_$\3\2\2\2`a\7_\2\2a&\3\2\2\2bc\7\"\2\2c(\3\2\2\2df\t"+
		"\20\2\2ed\3\2\2\2fg\3\2\2\2ge\3\2\2\2gh\3\2\2\2h*\3\2\2\2ij\n\21\2\2j"+
		",\3\2\2\2km\t\22\2\2lk\3\2\2\2mn\3\2\2\2nl\3\2\2\2no\3\2\2\2op\3\2\2\2"+
		"pq\b\27\2\2q.\3\2\2\2\5\2gn\3\b\2\2";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}