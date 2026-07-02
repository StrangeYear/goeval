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
	if yylex.(*lexer).build {
	    $$ = astValue(arrayNode{items: exprNodes($2)})
	} else {
            $$ = NewValue("", $2)
	}
    }
    | IDENTIFIER '(' params ')'
    {
	if yylex.(*lexer).build {
	    $$ = astValue(functionNode{name: $1, args: exprNodes($3)})
	} else {
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
	if yylex.(*lexer).build {
	    $$ = astValue(selectNameNode{base: asExprNode($1), key: $2})
	} else {
	    val := SelectValue($1.val, $2)
	    name := $1.name + "." + $2
	    $$ = NewValue(name, val)
	}
    }
    | quote '[' expr ']'
    {
	if yylex.(*lexer).build {
	    $$ = astValue(selectIndexExpr($1, $3))
	} else {
	    val := SelectValue($1.val, $3.String())
	    name := indexName($1.name, $3)
	    $$ = NewValue(name, val)
	}
    }
    | quote '(' params ')'
    {
	if yylex.(*lexer).build {
	    $$ = astValue(callNode{base: asExprNode($1), args: exprNodes($3)})
	} else {
            funcValue := toFunc($1.val)
	    val, err := funcValue($3...)
	    if err != nil {
	        panic(err)
	    }
	    name := callName($1.name)
	    $$ = NewValue(name, val)
	}
    }

nested:
    IDENTIFIER
    {
	if yylex.(*lexer).build {
	    $$ = astValue(variableNode{name: $1})
	} else {
	    val := yylex.(*lexer).kv[$1]
	    $$ = NewValue($1, val)
	}
    }
    | nested IDENTIFIER
    {
	if yylex.(*lexer).build {
	    $$ = astValue(selectNameNode{base: asExprNode($1), key: $2})
	} else {
	    val := SelectValue($1.val, $2)
	    name := $1.name + "." + $2
	    $$ = NewValue(name, val)
	}
    }
    | nested '[' expr ']'
    {
	if yylex.(*lexer).build {
	    $$ = astValue(selectIndexExpr($1, $3))
	} else {
	    val := SelectValue($1.val, $3.String())
	    name := indexName($1.name, $3)
	    $$ = NewValue(name, val)
	}
    }
    | nested '(' params ')'
    {
	if yylex.(*lexer).build {
	    $$ = astValue(callNode{base: asExprNode($1), args: exprNodes($3)})
	} else {
            funcValue := toFunc($1.val)
	    val, err := funcValue($3...)
	    if err != nil {
	        panic(err)
	    }
	    name := callName($1.name)
	    $$ = NewValue(name, val)
	}
    }

params:
    {
	$$ = nil
    }
    | expr
    {
	$$ = []any{yylex.(*lexer).param($1)}
    }
    | params ',' expr
    {
	$$ = append($$, yylex.(*lexer).param($3))
    }

rel:
  expr
| '!' rel {
	if yylex.(*lexer).build {
	    $$ = astValue(unaryNode{op: "!", x: asExprNode($2)})
	} else {
            $$ = $2.Not()
	}
}
| rel EQ rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opEq, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Eq($3) }
}
| rel NEQ rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opNeq, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Neq($3) }
}
| rel GTE rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opGte, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Gte($3) }
}
| rel LTE rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opLte, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Lte($3) }
}
| rel RE rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opRe, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Re($3) }
}
| rel NRE rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opNre, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Nre($3) }
}
| rel NC rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opNc, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Nc($3) }
}
| rel IN rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opIn, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.In($3) }
}
| rel '<' rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opLt, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Lt($3) }
}
| rel '>' rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opGt, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Gt($3) }
}
| rel '=' rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opMatch, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Match($3) }
}
| rel '+' rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opAdd, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Add($3) }
}
| rel '-' rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opSub, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Sub($3) }
}
| rel '*' rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opMulti, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Multi($3) }
}
| rel '/' rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opDiv, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Div($3) }
}
| rel '%' rel {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opMod, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Mod($3) }
}
;

cond:
  rel
| cond AND cond {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opAnd, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.And($3) }
}
| cond OR cond {
	if yylex.(*lexer).build { $$ = astValue(binaryNode{op: opOr, left: asExprNode($1), right: asExprNode($3)}) } else { $$ = $1.Or($3) }
}
| cond '?' cond ':' cond {
	if yylex.(*lexer).build { $$ = astValue(ternaryNode{cond: asExprNode($1), truthy: asExprNode($3), falsy: asExprNode($5)}) } else { $$ = $1.Ternary($3, $5) }
  }
;
