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

%token <val> VALUE EQ NEQ GTE LTE RE NRE AND OR NC IN
%token <name> IDENTIFIER
%left EQ NEQ GTE LTE RE NRE AND OR NC IN IDENTIFIER
%left '+' '-' '*' '/' '%' '<' '>' '?' ':' '=' '(' ')' ',' '[' ']'
%right '!'

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
        $$ = NewValue("", $2)
    }
    | IDENTIFIER '(' params ')'
    {
	fn := yylex.(*lexer).fns[$1]
	if fn == nil {
	    panic(__yyfmt__.Errorf("unknown function %s", $1))
	}
	res, err := fn($3...)
	if err != nil {
	    panic(err)
	}
	$$ = NewValue("", res)
    }
    | quote
    | nested
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
	$$ = NewValue(name, val)
    }
    | quote '[' expr ']'
    {
	val := SelectValue($1.val, $3.String())
	name := __yyfmt__.Sprintf("%s[%#v]", $1.name, $3.val)
	$$ = NewValue(name, val)
    }
    | quote '(' params ')'
    {
        funcValue := toFunc($1.val)
	val, err := funcValue($3...)
	if err != nil {
	    panic(err)
	}
	name := $1.name + "("
	for _, v := range $3 {
	   name += __yyfmt__.Sprintf("%#v, ", v)
	}
	if name[len(name)-2] == ',' {
	    name = name[:len(name)-2]
	}
	name += ")"
	$$ = NewValue(name, val)
    }

nested:
    IDENTIFIER
    {
	val := yylex.(*lexer).kv[$1]
	$$ = NewValue($1, val)
    }
    | nested IDENTIFIER
    {
	val := SelectValue($1.val, $2)
	name := $1.name + "." + $2
	$$ = NewValue(name, val)
    }
    | nested '[' expr ']'
    {
	val := SelectValue($1.val, $3.String())
	name := __yyfmt__.Sprintf("%s[%#v]", $1.name, $3.val)
	$$ = NewValue(name, val)
    }
    | nested '(' params ')'
    {
        funcValue := toFunc($1.val)
	val, err := funcValue($3...)
	if err != nil {
	    panic(err)
	}
	name := $1.name + "("
	for _, v := range $3 {
	   name += __yyfmt__.Sprintf("%#v, ", v)
	}
	if name[len(name)-2] == ',' {
	    name = name[:len(name)-2]
	}
	name += ")"
	$$ = NewValue(name, val)
    }

params:
    {
	$$ = []any{}
    }
    | expr
    {
	$$ = []any{$1.val}
    }
    | params ',' expr
    {
	$$ = append($$, $3.val)
    }

rel:
  expr
| '!' rel {
        $$ = $2.Not()
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