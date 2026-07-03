package goeval

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func assertEvalAndCompile(t *testing.T, e *Evaluable, expr string, want any, args ...any) {
	t.Helper()

	evalVal, _, err := e.Eval(expr, args...)
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}

	compiled, err := e.Compile(expr)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	compiledVal, err := compiled.Eval(args...)
	if err != nil {
		t.Fatalf("Compiled Eval() error = %v", err)
	}

	if evalVal.vType != compiledVal.vType || !reflect.DeepEqual(evalVal.val, compiledVal.val) {
		t.Fatalf("Eval() = %#v, compiled = %#v", evalVal, compiledVal)
	}
	if !reflect.DeepEqual(evalVal.val, want) {
		t.Fatalf("both modes = %#v, want %#v", evalVal.val, want)
	}
}

type expressionCase struct {
	name string
	expr string
	args []any
	want any
}

func runExpressionCases(t *testing.T, e *Evaluable, tests []expressionCase) {
	t.Helper()

	for _, tt := range tests {
		name := tt.name
		if name == "" {
			name = tt.expr
		}
		t.Run(name, func(t *testing.T) {
			assertEvalAndCompile(t, e, tt.expr, tt.want, tt.args...)
		})
	}
}

func TestEvalAndCompileModes(t *testing.T) {
	yyDebug = 1
	yyErrorVerbose = true

	e := Full(
		WithFunc("max", func(args ...any) (any, error) {
			if len(args) == 0 {
				return 0, nil
			}
			var val float64
			val, ok := args[0].(float64)
			if !ok {
				return 0, fmt.Errorf("max() expects number arguments")
			}
			for i := 1; i < len(args); i++ {
				arg, ok := args[i].(float64)
				if !ok {
					return 0, fmt.Errorf("max() expects number arguments")
				}
				val = max(val, arg)
			}
			return val, nil
		}),
		WithFunc("randN", func(args ...any) (any, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("randN() expects exactly one argument")
			}
			count := NewValue("", args[0]).Int()
			return strings.Repeat("0", count), nil
		}),
	)
	params := map[string]any{
		"a":    7,
		"b":    2,
		"abc":  7,
		"flag": true,
		"body": map[string]any{"cardBin": "123456"},
		"tags": []string{"go", "eval"},
		"idx":  1,
		"name": "foo",
		"nested": map[string]bool{
			"ok": true,
		},
		"flags": []bool{true},
		"foo bar": map[string]any{
			"foo baz": map[string]any{"data": "baz"},
		},
	}
	builtinParams := map[string]any{
		"name":    "  GoEval  ",
		"email":   "dev@company.com",
		"tags":    []string{"go", "rules"},
		"missing": nil,
	}
	runExpressionCases(t, e, []expressionCase{
		{name: "true literal", expr: `true`, want: true},
		{name: "false literal", expr: `false`, want: false},
		{name: "basic arithmetic bool", expr: `1 + 2 >= 3`, want: true},
		{name: "grouped arithmetic bool", expr: `(1 + 2) > 5`, want: false},
		{name: "parameter comparison", expr: `abc > 6`, args: []any{params}, want: true},
		{name: "double quoted escape", expr: `str == "a\"bc"`, args: []any{map[string]any{"str": `a"bc`}}, want: true},
		{name: "single quoted escape", expr: `str == 'a\'bc'`, args: []any{map[string]any{"str": `a'bc`}}, want: true},
		{name: "backtick in string", expr: "str == 'a`bc'", args: []any{map[string]any{"str": "a`bc"}}, want: true},
		{name: "string literal comparison", expr: `"abc" == "abc"`, want: true},
		{name: "array membership miss", expr: `a in [3,4,5]`, args: []any{map[string]any{"a": 2}}, want: false},
		{name: "array membership variable", expr: `a in [3,4,5,b]`, args: []any{map[string]any{"a": 7, "b": 7}}, want: true},
		{name: "string array membership", expr: `a in b`, args: []any{map[string]any{"a": 7, "b": "[7]"}}, want: true},
		{name: "slice membership", expr: `a in b`, args: []any{map[string]any{"a": 7, "b": []int{7}}}, want: true},
		{name: "json path fast path", expr: `$["a.b.0"] > 5`, args: []any{map[string]any{"a": map[string]any{"b": []int{7, 8}}}}, want: true},
		{name: "json path gjson fallback", expr: `$["a.b.#"] == 2`, args: []any{map[string]any{"a": map[string]any{"b": []int{7, 8}}}}, want: true},
		{name: "underscore identifier", expr: `__hour > 6`, args: []any{map[string]any{"__hour": 7}}, want: true},
		{name: "special nested parameter", expr: `$.["foo bar"].$.["foo baz"].data == "baz"`, args: []any{params}, want: true},
		{name: "custom max function", expr: `max(1,2,3,a)`, args: []any{params}, want: float64(7)},
		{name: "subtraction without spaces", expr: `16-6`, want: float64(10)},
		{name: "subtraction space before right", expr: `16 -6`, want: float64(10)},
		{name: "subtraction space after operator", expr: `16- 6`, want: float64(10)},
		{name: "subtract negative number", expr: `16--6`, want: float64(22)},
		{name: "leading negative number", expr: `-1+2`, want: float64(1)},
		{name: "negative grouped expression", expr: `-(1+2)`, want: float64(-3)},
		{name: "operator precedence", expr: `a*b+3`, args: []any{params}, want: float64(17)},
		{name: "function argument expression", expr: `body.cardBin + randN(16 - strlen(body.cardBin))`, args: []any{params}, want: "1234560000000000"},
		{name: "dynamic index", expr: `tags[idx] == "eval"`, args: []any{params}, want: true},
		{name: "ternary", expr: `flag ? "yes" : "no"`, args: []any{params}, want: "yes"},
		{name: "array item expression", expr: `a in [3, 7, b+5]`, args: []any{params}, want: true},
		{name: "string concat number left", expr: `1 + "a"`, want: "1a"},
		{name: "string concat number right", expr: `"a" + 1`, want: "a1"},
		{name: "string concat strings", expr: `"a" + "b"`, want: "ab"},
		{name: "string concat numeric string", expr: `1 + "2"`, want: "12"},
		{name: "string concat chain", expr: `1 + "2" + 3`, want: "123"},
		{name: "string subtraction", expr: `"hello world" - " world"`, want: "hello"},
		{name: "compiled nested parameter", expr: `a.b[0] > 5`, args: []any{map[string]any{"a": map[string]any{"b": []int{7, 8}}}}, want: true},
		{name: "compiled string concat", expr: `1 + "2" + a`, args: []any{params}, want: "127"},
		{name: "method call", expr: `(date("2024-02-20") + duration("24h")).Format("2006-01-02")`, want: "2024-02-21"},
		{name: "compile with custom function", expr: `max(1, a, 3)`, args: []any{params}, want: float64(7)},
		{name: "eval map path", expr: `a.b[0] > 5`, args: []any{map[string]any{"a": map[string]any{"b": []int{7, 8}}}}, want: true},
		{name: "constant folded arithmetic", expr: `1000 / 2 > (80 * 100 % 2)`, want: true},
		{name: "constant folded string concat", expr: `"a" + "b" + 3`, want: "ab3"},
		{name: "constant folded method", expr: `(date("2024-02-20")).Format("2006-01-02")`, want: "2024-02-20"},
		{name: "duration subtraction", expr: `duration("48h") - duration("24h")`, want: time.Hour * 24},
		{name: "duration threshold", expr: `1000 / 2 > threshold`, args: []any{map[string]any{"threshold": 499}}, want: true},
		{name: "regex literal", expr: `name =~ "^fo+$"`, args: []any{params}, want: true},
		{name: "typed path fast path", expr: `flags[0] && nested.ok`, args: []any{params}, want: true},
		{name: "overridden builtin value", expr: `max(1, 2, 3)`, want: float64(3)},
		{name: "builtin contains", expr: `contains(email, "@company.com")`, args: []any{builtinParams}, want: true},
		{name: "builtin trim prefix suffix", expr: `startsWith(trim(name), "Go") && endsWith(trim(name), "Eval")`, args: []any{builtinParams}, want: true},
		{name: "builtin lower upper", expr: `lower("GO") == "go" && upper("go") == "GO"`, args: []any{builtinParams}, want: true},
		{name: "builtin replace", expr: `replace("go-eval-go", "go", "expr") == "expr-eval-expr"`, args: []any{builtinParams}, want: true},
		{name: "builtin len", expr: `len(tags) == 2`, args: []any{builtinParams}, want: true},
		{name: "builtin min max", expr: `min(3, 2, 5) == 2 && max(3, 2, 5) == 5`, args: []any{builtinParams}, want: true},
		{name: "builtin abs round", expr: `abs(-3) == 3 && round(1.6) == 2`, args: []any{builtinParams}, want: true},
		{name: "builtin conversions", expr: `int("3.9") == 3 && string(12) == "12" && bool("true")`, args: []any{builtinParams}, want: true},
		{name: "builtin presence checks", expr: `exists(email) && !exists(missing) && empty(missing) && notEmpty(tags)`, args: []any{builtinParams}, want: true},
		{name: "builtin coalesce default", expr: `coalesce("", 0, "ok") == "ok" && default("", "fallback") == "fallback"`, args: []any{builtinParams}, want: true},
		{name: "builtin regex", expr: `matches(email, "^[^@]+@company\\.com$") && regex(email, "company")`, args: []any{builtinParams}, want: true},
		{name: "builtin any all", expr: `any(false, "", true) && all(true, "true", true)`, args: []any{builtinParams}, want: true},
		{
			name: "nested method call",
			expr: "a.b.c.Value(d,5)",
			args: []any{
				map[string]any{
					"a": map[string]any{
						"b": map[string]any{
							"c": &S{},
						},
					},
					"d": "abc",
				},
			},
			want: []any{"abc", 5},
		},
		{
			name: "nested index",
			expr: "a.b.c[0]",
			args: []any{
				map[string]any{
					"a": map[string]any{
						"b": map[string]any{
							"c": []int{1, 2, 3},
						},
					},
				},
			},
			want: float64(1),
		},
		{
			name: "nested index comparison",
			expr: "a.b.c[0] == 1",
			args: []any{
				map[string]any{
					"a": map[string]any{
						"b": map[string]any{
							"c": []int{1, 2, 3},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "nested value",
			expr: "a.b.c",
			args: []any{
				map[string]any{
					"a": map[string]any{
						"b": map[string]any{
							"c": []int{1, 2, 3},
						},
					},
				},
			},
			want: []int{1, 2, 3},
		},
		{name: "date minus duration", expr: `(date("2024-02-20") - duration("24h")).Format("2006-01-02")`, want: "2024-02-19"},
		{name: "date plus duration", expr: `(date("2024-02-20") + duration("24h")).Format("2006-01-02")`, want: "2024-02-21"},
		{name: "duration multiply", expr: `duration("24h") * 2`, want: time.Hour * 48},
	})
}

func TestCompileShortCircuit(t *testing.T) {
	called := 0
	e := Full(
		WithFunc("boom", func(args ...any) (any, error) {
			called++
			return true, fmt.Errorf("boom should not be called")
		}),
	)

	tests := []struct {
		expr string
		args []any
	}{
		{expr: `false && boom()`},
		{expr: `true || boom()`},
		{expr: `flag && boom()`, args: []any{map[string]any{"flag": false}}},
		{expr: `flag || boom()`, args: []any{map[string]any{"flag": true}}},
		{expr: `value ?? boom()`, args: []any{map[string]any{"value": "ok"}}},
		{expr: `flag ? "ok" : boom()`, args: []any{map[string]any{"flag": true}}},
		{expr: `flag ? boom() : "ok"`, args: []any{map[string]any{"flag": false}}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			compiled, err := e.Compile(tt.expr)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			if _, err := compiled.Eval(tt.args...); err != nil {
				t.Fatalf("Compiled Eval() error = %v", err)
			}
		})
	}

	if called != 0 {
		t.Fatalf("short-circuited function called %d times", called)
	}
}

func TestCompileShortCircuitEvaluatesNeededBranch(t *testing.T) {
	called := 0
	e := Full(
		WithFunc("fallback", func(args ...any) (any, error) {
			called++
			return "fallback", nil
		}),
	)

	compiled, err := e.Compile(`value ?? fallback()`)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	got, err := compiled.EvalString(map[string]any{"value": ""})
	if err != nil {
		t.Fatalf("Compiled EvalString() error = %v", err)
	}
	if got != "fallback" || called != 1 {
		t.Fatalf("Compiled EvalString() = %q called=%d, want fallback called=1", got, called)
	}
}

func TestCompileInvalidRegexLiteral(t *testing.T) {
	if _, err := Compile(`name =~ "["`); err == nil {
		t.Fatal("Compile() error = nil, want invalid regexp error")
	}
}

func TestCompileCache(t *testing.T) {
	e := NewEvaluable()
	first, err := e.Compile(`a > 1`)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	second, err := e.Compile(`a > 1`)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	if first != second {
		t.Fatal("Compile() returned different pointers for cached expression")
	}

	uncached := NewEvaluable(WithCompileCache(0))
	first, err = uncached.Compile(`a > 1`)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	second, err = uncached.Compile(`a > 1`)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	if first == second {
		t.Fatal("Compile() returned cached pointer with cache disabled")
	}
}

func TestValidateAndMustCompile(t *testing.T) {
	if err := Validate(`a > 1 && contains(name, "go")`); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if err := Validate(`unknownFn(a)`); err == nil {
		t.Fatal("Validate() error = nil, want unknown function error")
	}

	compiled := MustCompile(`a > 1`)
	got, err := compiled.EvalBool(map[string]any{"a": 2})
	if err != nil {
		t.Fatalf("Compiled EvalBool() error = %v", err)
	}
	if !got {
		t.Fatal("Compiled EvalBool() = false, want true")
	}

	defer func() {
		if recover() == nil {
			t.Fatal("MustCompile() did not panic on invalid expression")
		}
	}()
	_ = MustCompile(`unknownFn(a)`)
}

func TestVariables(t *testing.T) {
	vars, err := Variables(`contains(email, "@company.com") && user.age >= minAge && $["items.0.price"] > 0 && tags[idx] == "go"`)
	if err != nil {
		t.Fatalf("Variables() error = %v", err)
	}
	want := []string{"email", "idx", "items.0.price", "minAge", "tags", "user.age"}
	if !reflect.DeepEqual(vars, want) {
		t.Fatalf("Variables() = %#v, want %#v", vars, want)
	}

	compiled := MustCompile(`a.b[0] > threshold && name =~ "^go"`)
	vars = compiled.Variables()
	want = []string{"a.b[0]", "name", "threshold"}
	if !reflect.DeepEqual(vars, want) {
		t.Fatalf("Compiled Variables() = %#v, want %#v", vars, want)
	}
}

func TestErrorPosition(t *testing.T) {
	_, err := EvalBool("a @ 1", map[string]any{"a": 1})
	if err == nil {
		t.Fatal("EvalBool() error = nil, want unexpected character error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "unexpected character") ||
		!strings.Contains(msg, "line 1") ||
		!strings.Contains(msg, "column 3") {
		t.Fatalf("EvalBool() error = %q, want line and column", msg)
	}
}

func TestExplain(t *testing.T) {
	steps, err := Explain(`age > 18 && contains(email, "@")`, map[string]any{
		"age":   20,
		"email": "dev@company.com",
	})
	if err != nil {
		t.Fatalf("Explain() error = %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("Explain() returned no steps")
	}
	last := steps[len(steps)-1]
	if last.Expression != `((age > 18) && contains(email, "@"))` {
		t.Fatalf("last expression = %q", last.Expression)
	}
	if last.Value != true || last.Type != "boolean" {
		t.Fatalf("last step = %#v, want true boolean", last)
	}

	called := 0
	e := Full(WithFunc("boom", func(args ...any) (any, error) {
		called++
		return true, nil
	}))
	steps, err = e.Explain(`false && boom()`)
	if err != nil {
		t.Fatalf("Explain() short-circuit error = %v", err)
	}
	for _, step := range steps {
		if step.Expression == "boom()" {
			t.Fatalf("Explain() evaluated short-circuited expression: %#v", step)
		}
	}
	if called != 0 {
		t.Fatalf("short-circuited function called %d times", called)
	}

	steps, err = Explain(`-(1+2)`)
	if err != nil {
		t.Fatalf("Explain() unary error = %v", err)
	}
	last = steps[len(steps)-1]
	if last.Value != float64(-3) || last.Type != "number" {
		t.Fatalf("last unary step = %#v, want -3 number", last)
	}
}

func TestDependencies(t *testing.T) {
	deps, err := Dependencies(`contains(email, "@") && $["items.0"] > limit && tags[idx] == "go"`)
	if err != nil {
		t.Fatalf("Dependencies() error = %v", err)
	}
	want := []Dependency{
		{Kind: DependencyFunction, Name: "contains"},
		{Kind: DependencyJSONPath, Name: "items.0"},
		{Kind: DependencyVariable, Name: "email"},
		{Kind: DependencyVariable, Name: "idx"},
		{Kind: DependencyVariable, Name: "limit"},
		{Kind: DependencyVariable, Name: "tags"},
		{Kind: DependencyVariable, Name: "tags", Dynamic: true},
	}
	if !reflect.DeepEqual(deps, want) {
		t.Fatalf("Dependencies() = %#v, want %#v", deps, want)
	}
}

func TestCustomFunctionOverridesDefault(t *testing.T) {
	assertEvalAndCompile(t, Full(
		WithFunc("max", func(args ...any) (any, error) {
			return 42, nil
		}),
	), `max(1, 2, 3)`, float64(42))
}

func TestOverriddenBuiltinIsNotConstantFolded(t *testing.T) {
	called := 0
	e := Full(
		WithFunc("max", func(args ...any) (any, error) {
			called++
			return called, nil
		}),
	)
	compiled, err := e.Compile(`max(1, 2)`)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	if called != 0 {
		t.Fatalf("custom max called during Compile(): %d", called)
	}

	first, err := compiled.EvalInt()
	if err != nil {
		t.Fatalf("first EvalInt() error = %v", err)
	}
	second, err := compiled.EvalInt()
	if err != nil {
		t.Fatalf("second EvalInt() error = %v", err)
	}
	if first != 1 || second != 2 {
		t.Fatalf("EvalInt() results = %d, %d; want 1, 2", first, second)
	}
}

type S struct {
}

func (s *S) Value(a string, b int) (string, int) {
	return a, b
}
