package goeval

import (
	"encoding/json"
	"fmt"
	"sync"
    "strings"

	"github.com/tidwall/gjson"
)

%%{
	machine expression;
	write data;
	access lex.;
	variable p lex.p;
	variable pe lex.pe;
}%%

type lexer struct {
    data string
    p, pe, cs int
    ts, te, act int
    answer Value
    kv map[string]any
    err error
    tokens []string
    collectTokens bool
    fns map[string]Func
    useDecimal bool
    jsonParsed bool
    json gjson.Result
}

var lexerPool = sync.Pool{
	New: func() any {
		return new(lexer)
	},
}

func (lex *lexer) token() string {
	return lex.data[lex.ts:lex.te]
}

func (lex *lexer) position(pos int) (int, int) {
	if pos < 0 {
		pos = 0
	}
	if pos > len(lex.data) {
		pos = len(lex.data)
	}
	line := 1
	column := 1
	for i := 0; i < pos; i++ {
		if lex.data[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}

func (lex *lexer) errorAt(pos int, format string, args ...any) error {
	line, column := lex.position(pos)
	return fmt.Errorf("%s at line %d, column %d", fmt.Sprintf(format, args...), line, column)
}

func newLexer(data string, kv map[string]any, fns map[string]Func, collectTokens bool, useDecimal bool) *lexer {
	lex := lexerPool.Get().(*lexer)
	*lex = lexer{
			data: data,
			pe: len(data),
			kv: kv,
			fns: fns,
			collectTokens: collectTokens,
			useDecimal: useDecimal,
	}
	%% write init;
	return lex
}

func (lex *lexer) newValue(name string, val any) Value {
	return newValue(name, val, lex.useDecimal)
}

func (lex *lexer) release() {
	*lex = lexer{}
	lexerPool.Put(lex)
}

type lexeme struct {
	val        Value
	name       string
	jsonPath   string
	isJSONPath bool
}

func (lex *lexer) jsonPathValue(path string) Value {
	if val, ok := SelectPath(lex.kv, path); ok {
		return lex.newValue(path, val)
	}
	if !lex.jsonParsed {
		bs, err := json.Marshal(lex.kv)
		if err != nil {
			panic(fmt.Errorf("parameter json marshal failed, %s", err))
		}
		lex.json = gjson.ParseBytes(bs)
		lex.jsonParsed = true
	}
	return lex.newValue(path, lex.json.Get(path))
}

func (lex *lexer) next() (int, lexeme) {
	eof := lex.pe
	tok := 0
	var item lexeme

	%%{
		action Bool {
			tok = VALUE
			item.val = lex.newValue("", lex.token() == "true")
			fbreak;
		}
		action JsonPath {
			tok = VALUE
			item.jsonPath = lex.data[lex.ts+3:lex.te-2]
			item.isJSONPath = true
			fbreak;
		}
		action Identifier {
            tok = IDENTIFIER
            item.name = lex.token()
            fbreak;
        }
		action Keyword {
            item.name = lex.token()
            switch item.name {
            case "contains":
                tok = CONTAINS
            case "starts_with":
                tok = STARTS_WITH
            case "ends_with":
                tok = ENDS_WITH
            case "between":
                tok = BETWEEN
            case "within_last":
                tok = WITHIN_LAST
            case "in":
                tok = IN
            case "not":
                tok = NOT
            }
            fbreak;
        }
        action Special {
            tok = IDENTIFIER
            item.name = lex.data[lex.ts+4:lex.te-2]
            fbreak;
        }
        action SubPath {
             tok = IDENTIFIER
             item.name = lex.data[lex.ts+1:lex.te]
             fbreak;
        }
        action SpecialSubPath {
             tok = IDENTIFIER
             item.name = lex.data[lex.ts+5:lex.te-2]
             fbreak;
        }
		action Float {
			tok = VALUE
			item.val = newNumberLiteral("", lex.token(), lex.useDecimal)
			fbreak;
		}
		action String {
			tok = VALUE
			val := lex.token()
			if strings.IndexByte(val, '\\') >= 0 {
				val = strings.ReplaceAll(val, "\\\\", "\\")
				if val[0] == '"' {
					val = strings.ReplaceAll(val, "\\\"", "\"")
				}else if val[0] == '\'' {
					val = strings.ReplaceAll(val, "\\'", "'")
				}else if val[0] == '`' {
					val = strings.ReplaceAll(val, "\\`", "`")
				}
			}
			item.val = lex.newValue("", val[1:len(val)-1])
			fbreak;
		}

		identifier = (alpha | '_')+ (alnum | '_')* ;
		string = '"' ((any - '"') | '\\\"')* '"' | "'" ((any - "'") | "\\\'")* "'" | '`' ((any - "`") | "\\\`")* "`" ;
		int = digit+ ;
		float = digit+ ('.' digit+)? ;
        jsonpath = "$[" string ']' ;
        special = "$.[" string ']' ;
        letter = '+' | '-' | '*' | '/' | '%' | '<' | '>' | '?' | ':' | '=' | '!' | '(' | ')' | ',' | '[' | ']';
        sub_path = '.' (alnum | '_')+ ;
        special_sub_path = '.' special ;

		main := |*
			# values
			string => String;
			(int | float) => Float;
			("true" | "false") => Bool;
			"nil" => { tok = VALUE; item.val = nilValue; fbreak; };

			# relations
			"==" => { tok = EQ; fbreak; };
			"!=" => { tok = NEQ; fbreak; };
			">=" => { tok = GTE; fbreak; };
			"<=" => { tok = LTE; fbreak; };
			"=~" => { tok = RE; fbreak; };
			"!~" => { tok = NRE; fbreak; };
			"&&" => { tok = AND; fbreak; };
			"||" => { tok = OR; fbreak; };
			"??" => { tok = NC; fbreak; };

			# keywords
			("contains" | "starts_with" | "ends_with" | "between" | "within_last" | "in" | "not") => Keyword;

			identifier => Identifier;
			sub_path => SubPath;
            special_sub_path => SpecialSubPath;
            jsonpath => JsonPath;
            special => Special;
			space+;
			letter => { tok = int(lex.data[lex.ts]); fbreak; };
			any => { panic(lex.errorAt(lex.ts, "unexpected character %q", lex.data[lex.ts])) };
		*|;

		write exec;
	}%%

	if tok != 0 && lex.collectTokens {
       lex.tokens = append(lex.tokens, lex.token())
    }

	return tok, item
}

func (lex *lexer) Lex(out *yySymType) int {
	tok, item := lex.next()
	switch tok {
	case VALUE:
		if item.isJSONPath {
			out.val = lex.jsonPathValue(item.jsonPath)
		} else {
			out.val = item.val
		}
	case IDENTIFIER:
		out.name = item.name
	case CONTAINS, STARTS_WITH, ENDS_WITH, BETWEEN, WITHIN_LAST, IN, NOT:
		out.name = item.name
	}
	return tok
}

type compileLexerAdapter struct {
	*lexer
	answer exprNode
}

func (lex *compileLexerAdapter) Lex(out *compileSymType) int {
	tok, item := lex.next()
	switch tok {
	case VALUE:
		if item.isJSONPath {
			out.node = jsonPathNode{path: item.jsonPath}
		} else {
			out.node = literalNode{value: item.val}
		}
		return C_VALUE
	case IDENTIFIER:
		out.name = item.name
		return C_IDENTIFIER
	case CONTAINS:
		out.name = item.name
		return C_CONTAINS
	case STARTS_WITH:
		out.name = item.name
		return C_STARTS_WITH
	case ENDS_WITH:
		out.name = item.name
		return C_ENDS_WITH
	case BETWEEN:
		out.name = item.name
		return C_BETWEEN
	case WITHIN_LAST:
		out.name = item.name
		return C_WITHIN_LAST
	case NOT:
		out.name = item.name
		return C_NOT
	case IN:
		out.name = item.name
		return C_IN
	case EQ:
		return C_EQ
	case NEQ:
		return C_NEQ
	case GTE:
		return C_GTE
	case LTE:
		return C_LTE
	case RE:
		return C_RE
	case NRE:
		return C_NRE
	case AND:
		return C_AND
	case OR:
		return C_OR
	case NC:
		return C_NC
	default:
		return tok
	}
}

func (lex *lexer) Error(e string) {
    lex.err = lex.errorAt(lex.p, "%s", e)
}
