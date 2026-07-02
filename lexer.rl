package goeval

import (
	"encoding/json"
	"fmt"
	"strconv"
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
    build bool
    fns map[string]Func
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

func newLexer(data string, kv map[string]any, fns map[string]Func, collectTokens bool) *lexer {
	lex := lexerPool.Get().(*lexer)
	*lex = lexer{
			data: data,
			pe: len(data),
			kv: kv,
			fns: fns,
			collectTokens: collectTokens,
	}
	%% write init;
	return lex
}

func (lex *lexer) release() {
	*lex = lexer{}
	lexerPool.Put(lex)
}

func (lex *lexer) Lex(out *yySymType) int {
	eof := lex.pe
	tok := 0

	%%{
		action Bool {
			tok = VALUE
			out.val = lex.value(NewValue("", lex.token() == "true"))
			fbreak;
		}
		action JsonPath {
			tok = VALUE
			path := lex.data[lex.ts+3:lex.te-2]
			if lex.build {
				out.val = astValue(jsonPathNode{path: path})
			} else {
				if val, ok := SelectPath(lex.kv, path); ok {
					out.val = NewValue(path, val)
				} else {
					if !lex.jsonParsed {
						bs, err := json.Marshal(lex.kv)
						if err != nil {
							panic(fmt.Errorf("parameter json marshal failed, %s", err))
						}
						lex.json = gjson.ParseBytes(bs)
						lex.jsonParsed = true
					}
					res := lex.json.Get(path)
					out.val = NewValue(path, res)
				}
			}
			fbreak;
		}
		action Identifier {
            tok = IDENTIFIER
            out.name = lex.token()
            fbreak;
        }
        action Special {
            tok = IDENTIFIER
            out.name = lex.data[lex.ts+4:lex.te-2]
            fbreak;
        }
        action SubPath {
             tok = IDENTIFIER
             out.name = lex.data[lex.ts+1:lex.te]
             fbreak;
        }
        action SpecialSubPath {
             tok = IDENTIFIER
             out.name = lex.data[lex.ts+5:lex.te-2]
             fbreak;
        }
		action Float {
			tok = VALUE
			n, err := strconv.ParseFloat(lex.token(), 64)
			if err != nil {
				panic(err)
			}
			out.val = lex.value(NewValue("", n))
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
			out.val = lex.value(NewValue("", val[1:len(val)-1]))
			fbreak;
		}

		identifier = (alpha | '_')+ (alnum | '_')* ;
		string = '"' ((any - '"') | '\\\"')* '"' | "'" ((any - "'") | "\\\'")* "'" | '`' ((any - "`") | "\\\`")* "`" ;
		int = '-'? digit+ ;
		float = '-'? digit+ ('.' digit+)? ;
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
			"nil" => { tok = VALUE; out.val = lex.value(nilValue); fbreak; };

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
			"in" => { tok = IN; fbreak; };

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

	return tok
}

func (lex *lexer) Error(e string) {
    lex.err = lex.errorAt(lex.p, "%s", e)
}
