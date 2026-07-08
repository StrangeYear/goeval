package goeval

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
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
		"keyword_fields": map[string]any{
			"contains":    "AWS Marketplace",
			"starts_with": "GoEval",
			"not":         true,
		},
		"contains":    true,
		"starts_with": "Go",
		"ends_with":   "Eval",
		"between":     3,
		"within_last": "window",
		"in":          "membership",
		"not":         false,
		"flags":       []bool{true},
		"foo bar": map[string]any{
			"foo baz": map[string]any{"data": "baz"},
		},
	}
	builtinParams := map[string]any{
		"name":               "  GoEval  ",
		"email":              "dev@company.com",
		"tags":               []string{"go", "rules"},
		"amount":             100,
		"created_at":         "2024-02-20 12:00:00",
		"start_at":           "2024-02-20 00:00:00",
		"end_at":             "2024-02-21 00:00:00",
		"event_triggered_at": "2024-02-23 12:00:00",
		"missing":            nil,
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
		{name: "array not in", expr: `a not in [3,4,5]`, args: []any{map[string]any{"a": 2}}, want: true},
		{name: "array membership variable", expr: `a in [3,4,5,b]`, args: []any{map[string]any{"a": 7, "b": 7}}, want: true},
		{name: "string array membership", expr: `a in b`, args: []any{map[string]any{"a": 7, "b": "[7]"}}, want: true},
		{name: "slice membership", expr: `a in b`, args: []any{map[string]any{"a": 7, "b": []int{7}}}, want: true},
		{name: "json path fast path", expr: `$["a.b.0"] > 5`, args: []any{map[string]any{"a": map[string]any{"b": []int{7, 8}}}}, want: true},
		{name: "json path gjson fallback", expr: `$["a.b.#"] == 2`, args: []any{map[string]any{"a": map[string]any{"b": []int{7, 8}}}}, want: true},
		{name: "underscore identifier", expr: `__hour > 6`, args: []any{map[string]any{"__hour": 7}}, want: true},
		{name: "special nested parameter", expr: `$.["foo bar"].$.["foo baz"].data == "baz"`, args: []any{params}, want: true},
		{name: "keyword field contains", expr: `keyword_fields.contains contains "AWS"`, args: []any{params}, want: true},
		{name: "keyword field starts with", expr: `keyword_fields.starts_with starts_with "Go"`, args: []any{params}, want: true},
		{name: "keyword field not", expr: `keyword_fields.not`, args: []any{params}, want: true},
		{name: "keyword bracket field", expr: `keyword_fields["contains"] contains "AWS"`, args: []any{params}, want: true},
		{name: "keyword root variable contains", expr: `contains == true`, args: []any{params}, want: true},
		{name: "keyword root variable starts_with", expr: `starts_with == "Go"`, args: []any{params}, want: true},
		{name: "keyword root variable ends_with", expr: `ends_with == "Eval"`, args: []any{params}, want: true},
		{name: "keyword root variable between", expr: `between == 3`, args: []any{params}, want: true},
		{name: "keyword root variable within_last", expr: `within_last == "window"`, args: []any{params}, want: true},
		{name: "keyword root variable in", expr: `in == "membership"`, args: []any{params}, want: true},
		{name: "keyword root variable not", expr: `not == false`, args: []any{params}, want: true},
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
		{name: "operator contains", expr: `email contains "@company.com"`, args: []any{builtinParams}, want: true},
		{name: "operator not contains", expr: `email not contains "@example.com"`, args: []any{builtinParams}, want: true},
		{name: "builtin trim prefix suffix", expr: `startsWith(trim(name), "Go") && endsWith(trim(name), "Eval")`, args: []any{builtinParams}, want: true},
		{name: "builtin snake aliases", expr: `starts_with(trim(name), "Go") && ends_with(trim(name), "Eval") && not_empty(tags)`, args: []any{builtinParams}, want: true},
		{name: "operator starts ends with", expr: `trim(name) starts_with "Go" && trim(name) ends_with "Eval"`, args: []any{builtinParams}, want: true},
		{name: "operator not starts ends with", expr: `trim(name) not starts_with "No" && trim(name) not ends_with "No"`, args: []any{builtinParams}, want: true},
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
		{name: "builtin between numeric lower bound", expr: `between(amount, 100, 500)`, args: []any{builtinParams}, want: true},
		{name: "builtin between numeric upper bound", expr: `between(500, 100, 500)`, args: []any{builtinParams}, want: true},
		{name: "builtin between numeric outside", expr: `between(99, 100, 500)`, args: []any{builtinParams}, want: false},
		{name: "operator between numeric", expr: `amount between [100, 500]`, args: []any{builtinParams}, want: true},
		{name: "operator not between numeric", expr: `99 not between [100, 500]`, args: []any{builtinParams}, want: true},
		{name: "builtin between dates", expr: `between(date(created_at), date(start_at), date(end_at))`, args: []any{builtinParams}, want: true},
		{name: "operator between dates", expr: `date(created_at) between [date(start_at), date(end_at)]`, args: []any{builtinParams}, want: true},
		{name: "builtin withinLast inside", expr: `withinLast(date(created_at), duration("72h"), date(event_triggered_at))`, args: []any{builtinParams}, want: true},
		{name: "builtin within_last alias", expr: `within_last(date(created_at), duration("72h"), date(event_triggered_at))`, args: []any{builtinParams}, want: true},
		{name: "operator within_last inside", expr: `date(created_at) within_last [duration("72h"), date(event_triggered_at)]`, args: []any{builtinParams}, want: true},
		{name: "builtin withinLast lower bound", expr: `withinLast(date("2024-02-20 12:00:00"), duration("72h"), date(event_triggered_at))`, args: []any{builtinParams}, want: true},
		{name: "builtin withinLast outside", expr: `withinLast(date("2024-02-20 11:59:59"), duration("72h"), date(event_triggered_at))`, args: []any{builtinParams}, want: false},
		{name: "operator within_last outside", expr: `date("2024-02-20 11:59:59") within_last [duration("72h"), date(event_triggered_at)]`, args: []any{builtinParams}, want: false},
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
		{name: "date with location", expr: `(date("2024-02-20 10:30:00", "Asia/Shanghai")).Format("2006-01-02 15:04 MST")`, want: "2024-02-20 10:30 CST"},
		{name: "date with location and format", expr: `(date("2024/02/20 10:30", "Asia/Shanghai", "2006/01/02 15:04")).Format("2006-01-02 15:04 MST")`, want: "2024-02-20 10:30 CST"},
		{name: "date timestamp with utc", expr: `(date(0, "UTC")).Format("2006-01-02 15:04 MST")`, want: "1970-01-01 00:00 UTC"},
		{name: "date timestamp with location", expr: `(date(0, "Asia/Shanghai")).Format("2006-01-02 15:04 MST")`, want: "1970-01-01 08:00 CST"},
		{name: "duration multiply", expr: `duration("24h") * 2`, want: time.Hour * 48},
	})
}

func TestDateArgumentErrors(t *testing.T) {
	tests := []string{
		`date()`,
		`date("2024-02-20", 1)`,
		`date("2024-02-20", "Bad/Location")`,
		`date("2024-02-20", "UTC", 1)`,
		`date("2024/02/20", "UTC", "2006-01-02")`,
	}
	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			if _, _, err := Full().Eval(expr); err == nil {
				t.Fatal("Eval() error = nil, want error")
			}
			if _, err := Full().Compile(expr); err == nil {
				t.Fatal("Compile() error = nil, want error")
			}
		})
	}
}

func TestRiskBuiltinArgumentErrors(t *testing.T) {
	tests := []string{
		`between(1, 2)`,
		`between("abc", 1, 2)`,
		`1 between [1]`,
		`withinLast(date("2024-02-20"), duration("24h"))`,
		`withinLast(date("2024-02-20"), "24h", date("2024-02-21"))`,
		`date("2024-02-20") within_last [duration("24h")]`,
		`date("2024-02-20") within_last ["24h", date("2024-02-21")]`,
	}
	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			if _, _, err := Full().Eval(expr); err == nil {
				t.Fatal("Eval() error = nil, want error")
			}
			if _, err := Full().Compile(expr); err == nil {
				t.Fatal("Compile() error = nil, want error")
			}
		})
	}
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

func TestDecimalEvaluationAndCompile(t *testing.T) {
	e := Full(WithDecimal(true))
	tests := []struct {
		name string
		expr string
		args []any
		want bool
	}{
		{name: "addition", expr: `0.1 + 0.2 == 0.3`, want: true},
		{name: "subtraction", expr: `0.3 - 0.2 == 0.1`, want: true},
		{name: "multiplication", expr: `0.7 * 0.1 == 0.07`, want: true},
		{name: "division", expr: `1.00 / 4 == 0.25`, want: true},
		{name: "comparison", expr: `0.1 + 0.2 > 0.3`, want: false},
		{
			name: "params",
			expr: `a + b == 0.3`,
			args: []any{map[string]any{"a": 0.1, "b": 0.2}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := e.EvalBool(tt.expr, tt.args...)
			if err != nil {
				t.Fatalf("EvalBool() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("EvalBool() = %v, want %v", got, tt.want)
			}

			compiled, err := e.Compile(tt.expr)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			got, err = compiled.EvalBool(tt.args...)
			if err != nil {
				t.Fatalf("Compiled EvalBool() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("Compiled EvalBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimalValueAndBuiltins(t *testing.T) {
	e := Full(WithDecimal(true))
	val, _, err := e.Eval(`0.1 + 0.2`)
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	got, ok := val.val.(decimal.Decimal)
	if !ok {
		t.Fatalf("Eval() value type = %T, want decimal.Decimal", val.val)
	}
	if got.String() != "0.3" {
		t.Fatalf("Eval() value = %s, want 0.3", got)
	}

	assertEvalAndCompile(t, e, `date(1651467728) > date("2022-05-01 23:59:59")`, true)
	assertEvalAndCompile(t, e, `duration(24, "ns") == 24`, true)
	assertEvalAndCompile(t, e, `float("0.1") + 0.2 == 0.3`, true)

	defaultEval := Full()
	val, _, err = defaultEval.Eval(`decimal("0.1")`)
	if err != nil {
		t.Fatalf("Eval(decimal()) error = %v", err)
	}
	got, ok = val.val.(decimal.Decimal)
	if !ok {
		t.Fatalf("decimal() value type = %T, want decimal.Decimal", val.val)
	}
	if got.String() != "0.1" {
		t.Fatalf("decimal() value = %s, want 0.1", got)
	}
	assertEvalAndCompile(t, defaultEval, `decimal("0.1") + decimal("0.2") == decimal("0.3")`, true)
	assertEvalAndCompile(t, defaultEval, `decimal(0.1) + 0.2 == decimal("0.3")`, true)
}

func TestDecimalContrastsFloatPrecision(t *testing.T) {
	tests := []struct {
		name        string
		expr        string
		wantFloat   bool
		wantDecimal bool
	}{
		{
			name:        "decimal fraction is not exact in binary float",
			expr:        `0.1 + 0.2 == 0.3`,
			wantFloat:   false,
			wantDecimal: true,
		},
		{
			name:        "binary exact fractions stay exact in float",
			expr:        `0.5 + 0.25 == 0.75`,
			wantFloat:   true,
			wantDecimal: true,
		},
		{
			name:        "division can also be exactly representable",
			expr:        `1.00 / 4 == 0.25`,
			wantFloat:   true,
			wantDecimal: true,
		},
	}

	floatEval := Full()
	decimalEval := Full(WithDecimal(true))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFloat, err := floatEval.EvalBool(tt.expr)
			if err != nil {
				t.Fatalf("float EvalBool() error = %v", err)
			}
			if gotFloat != tt.wantFloat {
				t.Fatalf("float EvalBool() = %v, want %v", gotFloat, tt.wantFloat)
			}

			gotDecimal, err := decimalEval.EvalBool(tt.expr)
			if err != nil {
				t.Fatalf("decimal EvalBool() error = %v", err)
			}
			if gotDecimal != tt.wantDecimal {
				t.Fatalf("decimal EvalBool() = %v, want %v", gotDecimal, tt.wantDecimal)
			}
		})
	}
}

func TestCompileCache(t *testing.T) {
	e := New(WithCompileCache(2))
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

	uncached := New()
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

	_, err = e.Compile(`b > 1`)
	if err != nil {
		t.Fatalf("Compile() b error = %v", err)
	}
	_, err = e.Compile(`c > 1`)
	if err != nil {
		t.Fatalf("Compile() c error = %v", err)
	}
	third, err := e.Compile(`a > 1`)
	if err != nil {
		t.Fatalf("Compile() recached error = %v", err)
	}
	if third == first {
		t.Fatal("Compile() kept evicted LRU entry")
	}
}

func TestCompileCacheSeparatesDecimalMode(t *testing.T) {
	e := New(WithCompileCache(2))
	expr := `0.1 + 0.2 == 0.3`
	floatCompiled, err := e.Compile(expr)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	got, err := floatCompiled.EvalBool()
	if err != nil {
		t.Fatalf("Compiled EvalBool() error = %v", err)
	}
	if got {
		t.Fatal("float compiled expression = true, want false")
	}

	e.useDecimal = true
	decimalCompiled, err := e.Compile(expr)
	if err != nil {
		t.Fatalf("decimal Compile() error = %v", err)
	}
	if decimalCompiled == floatCompiled {
		t.Fatal("decimal Compile() reused float cache entry")
	}
	got, err = decimalCompiled.EvalBool()
	if err != nil {
		t.Fatalf("decimal Compiled EvalBool() error = %v", err)
	}
	if !got {
		t.Fatal("decimal compiled expression = false, want true")
	}
}

func TestNewAndFullFunctionSets(t *testing.T) {
	e := New()
	if _, err := e.EvalBool(`now() != nil`); err == nil {
		t.Fatal("New().EvalBool() error = nil, want unknown function error")
	}
	if _, err := e.Compile(`contains(name, "x")`); err == nil {
		t.Fatal("New().Compile() error = nil, want unknown function error")
	}

	e = New(WithFunc("ok", func(args ...any) (any, error) {
		return true, nil
	}))
	matched, err := e.EvalBool(`ok()`)
	if err != nil {
		t.Fatalf("EvalBool(ok) error = %v", err)
	}
	if !matched {
		t.Fatal("EvalBool(ok) = false, want true")
	}

	matched, err = Full().EvalBool(`contains("xyz", "x")`)
	if err != nil {
		t.Fatalf("Full().EvalBool(contains) error = %v", err)
	}
	if !matched {
		t.Fatal("Full().EvalBool(contains) = false, want true")
	}
}

func TestEvalMapAPIs(t *testing.T) {
	e := New(WithCompileCache(2))
	val, vars, err := e.EvalMap(`a + b`, map[string]any{"a": 1, "b": 2})
	if err != nil {
		t.Fatalf("EvalMap() error = %v", err)
	}
	if val.val != float64(3) {
		t.Fatalf("EvalMap() value = %#v, want 3", val.val)
	}
	if !reflect.DeepEqual(vars, []string{"a", "b"}) {
		t.Fatalf("EvalMap() vars = %#v, want a,b", vars)
	}
	if got, err := e.EvalMapBool(`a > 1 && b < 10`, map[string]any{"a": 2, "b": 3}); err != nil || !got {
		t.Fatalf("EvalMapBool() = %v, %v; want true, nil", got, err)
	}
	if got, err := e.EvalMapInt(`a + b`, map[string]any{"a": 1, "b": 2}); err != nil || got != 3 {
		t.Fatalf("EvalMapInt() = %v, %v; want 3, nil", got, err)
	}
	if got, err := e.EvalMapFloat(`a + b`, map[string]any{"a": 1, "b": 2}); err != nil || got != 3 {
		t.Fatalf("EvalMapFloat() = %v, %v; want 3, nil", got, err)
	}
	if got, err := e.EvalMapString(`name`, map[string]any{"name": "goeval"}); err != nil || got != "goeval" {
		t.Fatalf("EvalMapString() = %q, %v; want goeval, nil", got, err)
	}
}

func TestStrictVariables(t *testing.T) {
	e := New(WithStrictVariables(true))

	_, err := e.EvalMapBool(`a > 1`, map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "a") {
		t.Fatalf("missing top-level variable error = %v, want a", err)
	}

	_, err = e.EvalMapBool(`user.account.id == "1"`, map[string]any{
		"user": map[string]any{},
	})
	if err == nil || !strings.Contains(err.Error(), "user.account") {
		t.Fatalf("missing path error = %v, want user.account", err)
	}

	_, err = e.EvalMapBool(`items[2] == 1`, map[string]any{
		"items": []int{1},
	})
	if err == nil || !strings.Contains(err.Error(), "items[2]") {
		t.Fatalf("missing index error = %v, want items[2]", err)
	}

	_, err = e.EvalMapBool(`$["items.1.price"] > 0`, map[string]any{
		"items": []map[string]any{{"price": 10}},
	})
	if err == nil || !strings.Contains(err.Error(), "items.1.price") {
		t.Fatalf("missing JSON path error = %v, want items.1.price", err)
	}

	matched, err := New().EvalMapBool(`a > 1`, map[string]any{})
	if err != nil {
		t.Fatalf("non-strict missing variable error = %v", err)
	}
	if matched {
		t.Fatal("non-strict missing variable matched, want false")
	}
}

func TestUndefinedFunctionError(t *testing.T) {
	_, err := New().Compile(`missingFn(a)`)
	if err == nil || !strings.Contains(err.Error(), `undefined function "missingFn"`) {
		t.Fatalf("Compile() error = %v, want undefined function name", err)
	}

	_, err = New().EvalBool(`missingFn(a)`, map[string]any{"a": 1})
	if err == nil || !strings.Contains(err.Error(), `undefined function "missingFn"`) {
		t.Fatalf("EvalBool() error = %v, want undefined function name", err)
	}
}

func TestCompiledExpressionConcurrentEval(t *testing.T) {
	e := New(WithStrictVariables(true))
	compiled, err := e.Compile(`a >= 3 && b < 7`)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 100)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				matched, err := compiled.EvalMapBool(map[string]any{
					"a": 3,
					"b": 2,
				})
				if err != nil {
					errs <- err
					return
				}
				if !matched {
					errs <- fmt.Errorf("matched = false")
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
}

func TestWithFuncsAndCompiledFunctionSnapshot(t *testing.T) {
	e := New(WithFuncs(map[string]Func{
		"contains": func(args ...any) (any, error) {
			return strings.Contains(fmt.Sprint(args[0]), fmt.Sprint(args[1])), nil
		},
	}))
	compiled, err := e.Compile(`contains(name, "x")`)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	WithFunc("contains", func(args ...any) (any, error) {
		return false, nil
	})(e)

	matched, err := compiled.EvalMapBool(map[string]any{"name": "xyz"})
	if err != nil {
		t.Fatalf("Compiled EvalMapBool() error = %v", err)
	}
	if !matched {
		t.Fatal("Compiled EvalMapBool() = false, want snapshot to remain true")
	}
}

func TestCustomFunctionAliases(t *testing.T) {
	riskScore := func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("riskScore() expects exactly one argument")
		}
		return NewValue("", args[0]).Float() * 2, nil
	}
	e := New(
		WithFunc("riskScore", riskScore, "risk_score"),
		WithFuncAlias("riskScore", "score"),
	)

	for _, expr := range []string{
		`riskScore(amount) == 20`,
		`risk_score(amount) == 20`,
		`score(amount) == 20`,
	} {
		t.Run(expr, func(t *testing.T) {
			matched, err := e.EvalMapBool(expr, map[string]any{"amount": 10})
			if err != nil {
				t.Fatalf("EvalMapBool() error = %v", err)
			}
			if !matched {
				t.Fatal("EvalMapBool() = false, want true")
			}

			compiled, err := e.Compile(expr)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			matched, err = compiled.EvalMapBool(map[string]any{"amount": 10})
			if err != nil {
				t.Fatalf("Compiled EvalMapBool() error = %v", err)
			}
			if !matched {
				t.Fatal("Compiled EvalMapBool() = false, want true")
			}
		})
	}
}

func TestDecimalModeStringNumericEqualityAndIn(t *testing.T) {
	e := New(WithDecimal(true))

	matched, err := e.EvalMapBool(`amount == 0.3`, map[string]any{
		"amount": "0.30",
	})
	if err != nil {
		t.Fatalf("EvalMapBool(decimal equality) error = %v", err)
	}
	if !matched {
		t.Fatal("EvalMapBool(decimal equality) = false, want true")
	}

	matched, err = e.EvalMapBool(`count in [1, 2, 3]`, map[string]any{
		"count": int64(3),
	})
	if err != nil {
		t.Fatalf("EvalMapBool(decimal in) error = %v", err)
	}
	if !matched {
		t.Fatal("EvalMapBool(decimal in) = false, want true")
	}

	full := Full(WithDecimal(true))
	matched, err = full.EvalMapBool(`between(amount, 0.1, 0.3)`, map[string]any{
		"amount": "0.30",
	})
	if err != nil {
		t.Fatalf("EvalMapBool(decimal between) error = %v", err)
	}
	if !matched {
		t.Fatal("EvalMapBool(decimal between) = false, want true")
	}
}

func TestDefaultFloatEvaluationUnchanged(t *testing.T) {
	got, err := Full().EvalBool(`0.1 + 0.2 == 0.3`)
	if err != nil {
		t.Fatalf("EvalBool() error = %v", err)
	}
	if got {
		t.Fatal("EvalBool() = true, want existing float64 behavior false")
	}
}

func TestInvalidOperationsReturnErrors(t *testing.T) {
	assertReturnsErrorWithoutPanic := func(name string, fn func() error) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			t.Helper()
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("unexpected panic: %v", r)
				}
			}()
			if err := fn(); err == nil {
				t.Fatal("error = nil, want error")
			}
		})
	}

	assertReturnsErrorWithoutPanic("eval invalid multiply", func() error {
		_, _, err := Full().Eval(`"a" * "b"`)
		return err
	})
	assertReturnsErrorWithoutPanic("eval float invalid conversion", func() error {
		_, err := Full().EvalFloat(`"abc"`)
		return err
	})
	assertReturnsErrorWithoutPanic("eval int invalid conversion", func() error {
		_, err := Full().EvalInt(`"abc"`)
		return err
	})

	compiled := MustCompile(`"abc"`)
	assertReturnsErrorWithoutPanic("compiled eval float invalid conversion", func() error {
		_, err := compiled.EvalFloat()
		return err
	})
	assertReturnsErrorWithoutPanic("compiled eval map int invalid conversion", func() error {
		_, err := compiled.EvalMapInt(nil)
		return err
	})
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
