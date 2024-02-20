package goeval

import (
	"encoding/json"
	"errors"
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
    data []byte
    p, pe, cs int
    ts, te, act int
    answer Value
    kv map[string]any
    err error
    tokens []string
    fns map[string]Func
    once sync.Once
    json gjson.Result
}

func (lex *lexer) token() string {
	return string(lex.data[lex.ts:lex.te])
}

func newLexer(data []byte, kv map[string]any, fns map[string]Func) *lexer {
	lex := &lexer{
			data: data,
			pe: len(data),
			kv: kv,
			fns: fns,
	}
	%% write init;
	return lex
}

func (lex *lexer) Lex(out *yySymType) int {
	eof := lex.pe
	tok := 0

	%%{
		action Bool {
			tok = VALUE
			out.val = NewValue("", lex.token() == "true")
			fbreak;
		}
		action JsonPath {
			tok = VALUE
			path := string(lex.data[lex.ts+3:lex.te-2])
			lex.once.Do(func() {
            	bs, err := json.Marshal(lex.kv)
            	if err != nil {
            		panic(fmt.Errorf("parameter json marshal failed, %s", err))
            	}
            	lex.json = gjson.ParseBytes(bs)
            })
            res := lex.json.Get(path)
			out.val = NewValue(path, res)
			fbreak;
		}
		action Identifier {
            tok = IDENTIFIER
            out.name = lex.token()
            fbreak;
        }
        action Special {
            tok = IDENTIFIER
            out.name = string(lex.data[lex.ts+4:lex.te-2])
            fbreak;
        }
        action SubPath {
             tok = IDENTIFIER
             out.name = string(lex.data[lex.ts+1:lex.te])
             fbreak;
        }
        action SpecialSubPath {
             tok = IDENTIFIER
             out.name = string(lex.data[lex.ts+5:lex.te-2])
             fbreak;
        }
		action Float {
			tok = VALUE
			n, err := strconv.ParseFloat(lex.token(), 64)
			if err != nil {
				panic(err)
			}
			out.val = NewValue("", n)
			fbreak;
		}
		action String {
			tok = VALUE
			val := string(lex.token())
			val = strings.ReplaceAll(val, "\\\\", "\\")
			if val[0] == '"' {
			    val = strings.ReplaceAll(val, "\\\"", "\"")
			}else if val[0] == '\'' {
                val = strings.ReplaceAll(val, "\\'", "'")
            }else if val[0] == '`' {
                val = strings.ReplaceAll(val, "\\`", "`")
            }
			out.val = NewValue("", val[1:len(val)-1])
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
			"nil" => { tok = VALUE; out.val = nilValue; fbreak; };

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
			any => { panic(errors.New("unexpected character: " + string(lex.data[lex.ts]))) };
		*|;

		write exec;
	}%%

	if tok != 0 {
       lex.tokens = append(lex.tokens, lex.token())
    }

	return tok
}

func (lex *lexer) Error(e string) {
    lex.err = errors.New(e)
}
