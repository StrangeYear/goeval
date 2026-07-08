%{
package goeval

%}

%union {
  name string
  val Value
  vals []any
}

%type <val> start cond expr rel nested quote
%type <vals> params
%type <name> keyword_name

%token <val> VALUE EQ NEQ GTE LTE RE NRE AND OR NC
%token <name> IDENTIFIER IN CONTAINS STARTS_WITH ENDS_WITH BETWEEN WITHIN_LAST NOT
%left EQ NEQ GTE LTE RE NRE AND OR NC IN CONTAINS STARTS_WITH ENDS_WITH BETWEEN WITHIN_LAST NOT IDENTIFIER
%left '+' '-' '*' '/' '%' '<' '>' '?' ':' '=' '(' ')' ',' '[' ']'
%right '!' UMINUS

%%

start: cond
    {
        yylex.(*lexer).answer = $1
        $$ = $1
        return 0
    }

expr:
    VALUE
    {
        $$ = $1
    }
    | '[' params ']'
    {
        $$ = yylex.(*lexer).newValue("", $2)
    }
    | IDENTIFIER '(' params ')'
    {
        fn := yylex.(*lexer).fns[$1]
        if fn == nil {
            panic(__yyfmt__.Errorf("undefined function %q", $1))
        }
        res, err := fn($3...)
        if err != nil {
            panic(err)
        }
        $$ = yylex.(*lexer).newValue("", res)
    }
    | keyword_name '(' params ')'
    {
        fn := yylex.(*lexer).fns[$1]
        if fn == nil {
            panic(__yyfmt__.Errorf("undefined function %q", $1))
        }
        res, err := fn($3...)
        if err != nil {
            panic(err)
        }
        $$ = yylex.(*lexer).newValue("", res)
    }
    | quote
    | nested
;

keyword_name:
    CONTAINS
    {
        $$ = $1
    }
    | STARTS_WITH
    {
        $$ = $1
    }
    | ENDS_WITH
    {
        $$ = $1
    }
    | BETWEEN
    {
        $$ = $1
    }
    | WITHIN_LAST
    {
        $$ = $1
    }
    | IN
    {
        $$ = $1
    }
    | NOT
    {
        $$ = $1
    }
;

quote:
    '(' cond ')'
    {
        $$ = $2
    }
    | quote IDENTIFIER
    {
        val := SelectValue($1.val, $2)
        name := $1.name + "." + $2
        $$ = yylex.(*lexer).newValue(name, val)
    }
    | quote '[' expr ']'
    {
        val := SelectValue($1.val, $3.String())
        name := indexName($1.name, $3)
        $$ = yylex.(*lexer).newValue(name, val)
    }
    | quote '(' params ')'
    {
        funcValue := toFunc($1.val)
        val, err := funcValue($3...)
        if err != nil {
            panic(err)
        }
        name := callName($1.name)
        $$ = yylex.(*lexer).newValue(name, val)
    }
;

nested:
    IDENTIFIER
    {
        val := yylex.(*lexer).kv[$1]
        $$ = yylex.(*lexer).newValue($1, val)
    }
    | keyword_name %prec IDENTIFIER
    {
        val := yylex.(*lexer).kv[$1]
        $$ = yylex.(*lexer).newValue($1, val)
    }
    | nested IDENTIFIER
    {
        val := SelectValue($1.val, $2)
        name := $1.name + "." + $2
        $$ = yylex.(*lexer).newValue(name, val)
    }
    | nested '[' expr ']'
    {
        val := SelectValue($1.val, $3.String())
        name := indexName($1.name, $3)
        $$ = yylex.(*lexer).newValue(name, val)
    }
    | nested '(' params ')'
    {
        funcValue := toFunc($1.val)
        val, err := funcValue($3...)
        if err != nil {
            panic(err)
        }
        name := callName($1.name)
        $$ = yylex.(*lexer).newValue(name, val)
    }
;

params:
    {
	$$ = nil
    }
    | cond
    {
	$$ = []any{yylex.(*lexer).param($1)}
    }
    | params ',' cond
    {
	$$ = append($$, yylex.(*lexer).param($3))
    }

rel:
  expr
| '!' rel {
        $$ = $2.Not()
}
| '-' rel %prec UMINUS {
        $$ = yylex.(*lexer).newValue("", 0).Sub($2)
}
| rel EQ rel {
	$$ = $1.Eq($3)
}
| rel NEQ rel {
	$$ = $1.Neq($3)
}
| rel GTE rel {
	$$ = $1.Gte($3)
}
| rel LTE rel {
	$$ = $1.Lte($3)
}
| rel RE rel {
	$$ = $1.Re($3)
}
| rel NRE rel {
	$$ = $1.Nre($3)
}
| rel NC rel {
	$$ = $1.Nc($3)
}
| rel IN rel {
	$$ = $1.In($3)
}
| rel NOT IN rel {
	$$ = $1.In($4).Not()
}
| rel CONTAINS rel {
	$$ = evalStringOperatorValue("contains", $1, $3)
}
| rel NOT CONTAINS rel {
	$$ = evalStringOperatorValue("contains", $1, $4).Not()
}
| rel STARTS_WITH rel {
	$$ = evalStringOperatorValue("starts_with", $1, $3)
}
| rel NOT STARTS_WITH rel {
	$$ = evalStringOperatorValue("starts_with", $1, $4).Not()
}
| rel ENDS_WITH rel {
	$$ = evalStringOperatorValue("ends_with", $1, $3)
}
| rel NOT ENDS_WITH rel {
	$$ = evalStringOperatorValue("ends_with", $1, $4).Not()
}
| rel BETWEEN rel {
	$$ = evalBetweenOperatorValue($1, $3)
}
| rel NOT BETWEEN rel {
	$$ = evalBetweenOperatorValue($1, $4).Not()
}
| rel WITHIN_LAST rel {
	$$ = evalWithinLastOperatorValue($1, $3)
}
| rel '<' rel {
	$$ = $1.Lt($3)
}
| rel '>' rel {
	$$ = $1.Gt($3)
}
| rel '=' rel {
	$$ = $1.Match($3)
}
| rel '+' rel {
	$$ = $1.Add($3)
}
| rel '-' rel {
	$$ = $1.Sub($3)
}
| rel '*' rel {
	$$ = $1.Multi($3)
}
| rel '/' rel {
	$$ = $1.Div($3)
}
| rel '%' rel {
	$$ = $1.Mod($3)
}
;

cond:
  rel
| cond AND cond {
	$$ = $1.And($3)
}
| cond OR cond {
	$$ = $1.Or($3)
}
| cond '?' cond ':' cond {
	$$ = $1.Ternary($3, $5)
  }
;
