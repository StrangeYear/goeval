%{
package goeval

%}

%union {
  name string
  node exprNode
  nodes []exprNode
}

%type <node> start cond expr rel nested quote
%type <nodes> params

%token <node> C_VALUE
%token C_EQ C_NEQ C_GTE C_LTE C_RE C_NRE C_AND C_OR C_NC C_IN
%token <name> C_IDENTIFIER
%left C_EQ C_NEQ C_GTE C_LTE C_RE C_NRE C_AND C_OR C_NC C_IN C_IDENTIFIER
%left '+' '-' '*' '/' '%' '<' '>' '?' ':' '=' '(' ')' ',' '[' ']'
%right '!' C_UMINUS

%%

start: cond
    {
        compilelex.(*compileLexerAdapter).answer = $1
        $$ = $1
        return 0
    }

expr:
    C_VALUE
    {
        $$ = $1
    }
    | '[' params ']'
    {
        $$ = arrayNode{items: $2}
    }
    | C_IDENTIFIER '(' params ')'
    {
        $$ = functionNode{name: $1, args: $3}
    }
    | quote
    | nested
;

quote:
    '(' cond ')'
    {
        $$ = $2
    }
    | quote C_IDENTIFIER
    {
        $$ = selectNameNode{base: $1, key: $2}
    }
    | quote '[' expr ']'
    {
        $$ = selectIndexExpr($1, $3)
    }
    | quote '(' params ')'
    {
        $$ = callNode{base: $1, args: $3}
    }

nested:
    C_IDENTIFIER
    {
        $$ = variableNode{name: $1}
    }
    | nested C_IDENTIFIER
    {
        $$ = selectNameNode{base: $1, key: $2}
    }
    | nested '[' expr ']'
    {
        $$ = selectIndexExpr($1, $3)
    }
    | nested '(' params ')'
    {
        $$ = callNode{base: $1, args: $3}
    }

params:
    {
        $$ = nil
    }
    | cond
    {
        $$ = []exprNode{$1}
    }
    | params ',' cond
    {
        $$ = append($$, $3)
    }

rel:
  expr
| '!' rel {
        $$ = unaryNode{op: "!", x: $2}
}
| '-' rel %prec C_UMINUS {
        $$ = unaryNode{op: "-", x: $2}
}
| rel C_EQ rel {
        $$ = binaryNode{op: opEq, left: $1, right: $3}
}
| rel C_NEQ rel {
        $$ = binaryNode{op: opNeq, left: $1, right: $3}
}
| rel C_GTE rel {
        $$ = binaryNode{op: opGte, left: $1, right: $3}
}
| rel C_LTE rel {
        $$ = binaryNode{op: opLte, left: $1, right: $3}
}
| rel C_RE rel {
        $$ = binaryNode{op: opRe, left: $1, right: $3}
}
| rel C_NRE rel {
        $$ = binaryNode{op: opNre, left: $1, right: $3}
}
| rel C_NC rel {
        $$ = binaryNode{op: opNc, left: $1, right: $3}
}
| rel C_IN rel {
        $$ = binaryNode{op: opIn, left: $1, right: $3}
}
| rel '<' rel {
        $$ = binaryNode{op: opLt, left: $1, right: $3}
}
| rel '>' rel {
        $$ = binaryNode{op: opGt, left: $1, right: $3}
}
| rel '=' rel {
        $$ = binaryNode{op: opMatch, left: $1, right: $3}
}
| rel '+' rel {
        $$ = binaryNode{op: opAdd, left: $1, right: $3}
}
| rel '-' rel {
        $$ = binaryNode{op: opSub, left: $1, right: $3}
}
| rel '*' rel {
        $$ = binaryNode{op: opMulti, left: $1, right: $3}
}
| rel '/' rel {
        $$ = binaryNode{op: opDiv, left: $1, right: $3}
}
| rel '%' rel {
        $$ = binaryNode{op: opMod, left: $1, right: $3}
}
;

cond:
  rel
| cond C_AND cond {
        $$ = binaryNode{op: opAnd, left: $1, right: $3}
}
| cond C_OR cond {
        $$ = binaryNode{op: opOr, left: $1, right: $3}
}
| cond '?' cond ':' cond {
        $$ = ternaryNode{cond: $1, truthy: $3, falsy: $5}
  }
;
