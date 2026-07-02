package goeval

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestEvalBool(t *testing.T) {
	yyDebug = 1
	yyErrorVerbose = true
	type args struct {
		expr string
		args []any
	}
	tests := []struct {
		args    args
		want    bool
		wantErr bool
	}{
		{
			args: args{
				expr: "true",
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: "false",
			},
			want:    false,
			wantErr: false,
		},
		{
			args: args{
				expr: "1 + 2 >= 3",
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: "(1 + 2) > 5",
			},
			want:    false,
			wantErr: false,
		},
		{
			args: args{
				expr: "abc > 6",
				args: []any{map[string]any{"abc": 7}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: `abc == 'abc"d'`,
				args: []any{map[string]any{"abc": "abc\"d"}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: `"abc" == "abc"`,
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: "a in [3,4,5]",
				args: []any{map[string]any{"a": 2}},
			},
			want:    false,
			wantErr: false,
		},
		{
			args: args{
				expr: "a in [3,4,5,b]",
				args: []any{map[string]any{"a": 7, "b": 7}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: "a in b",
				args: []any{map[string]any{"a": 7, "b": "[7]"}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: "a in b",
				args: []any{map[string]any{"a": 7, "b": []int{7}}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: `$["a.b.0"] > 5`,
				args: []any{map[string]any{"a": map[string]any{"b": []int{7, 8}}}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: `$["a.b.#"] == 2`,
				args: []any{map[string]any{"a": map[string]any{"b": []int{7, 8}}}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: "__hour > 6",
				args: []any{map[string]any{"__hour": 7}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: `a == "a\"bc"`,
				args: []any{map[string]any{"a": `a"bc`}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: `a == 'a\'bc'`,
				args: []any{map[string]any{"a": `a'bc`}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: "a == 'a`bc'",
				args: []any{map[string]any{"a": "a`bc"}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: `$.["foo bar"].$.["foo baz"].data == "baz"`,
				args: []any{map[string]any{"foo bar": map[string]any{"foo baz": map[string]any{"data": "baz"}}}},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.args.expr, func(t *testing.T) {
			got, tokens, err := Eval(tt.args.expr, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvalBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Boolean() != tt.want {
				t.Errorf("EvalBool() got = %v, want %v", got, tt.want)
			}
			t.Logf("%#v", tokens)
		})
	}
}

func TestFunc(t *testing.T) {
	val, err := Full(
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
	).EvalFloat(`max(1,2,3,a)`, map[string]any{"a": float64(6)})
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	t.Log(val == 6)
	// output: true
}

func TestStringConcatenation(t *testing.T) {
	tests := []struct {
		expr string
		want string
	}{
		{expr: `1 + "a"`, want: "1a"},
		{expr: `"a" + 1`, want: "a1"},
		{expr: `"a" + "b"`, want: "ab"},
		{expr: `1 + "2"`, want: "12"},
		{expr: `1 + "2" + 3`, want: "123"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, _, err := Eval(tt.expr)
			if err != nil {
				t.Fatalf("Eval() error = %v", err)
			}
			if got.vType != String {
				t.Fatalf("Eval() type = %v, want String", got.vType)
			}
			if got.String() != tt.want {
				t.Fatalf("Eval() = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestStringSubtraction(t *testing.T) {
	got, _, err := Eval(`"hello world" - " world"`)
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	if got.vType != String {
		t.Fatalf("Eval() type = %v, want String", got.vType)
	}
	if got.String() != "hello" {
		t.Fatalf("Eval() = %q, want hello", got.String())
	}
}

func TestCompile(t *testing.T) {
	tests := []struct {
		name string
		expr string
		args []any
		want any
	}{
		{
			name: "nested parameter",
			expr: `a.b[0] > 5`,
			args: []any{map[string]any{"a": map[string]any{"b": []int{7, 8}}}},
			want: true,
		},
		{
			name: "json fast path",
			expr: `$["a.b.0"] > 5`,
			args: []any{map[string]any{"a": map[string]any{"b": []int{7, 8}}}},
			want: true,
		},
		{
			name: "json gjson fallback",
			expr: `$["a.b.#"] == 2`,
			args: []any{map[string]any{"a": map[string]any{"b": []int{7, 8}}}},
			want: true,
		},
		{
			name: "string concat",
			expr: `1 + "2" + a`,
			args: []any{map[string]any{"a": 3}},
			want: "123",
		},
		{
			name: "method call",
			expr: `(date("2024-02-20") + duration("24h")).Format("2006-01-02")`,
			want: "2024-02-21",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiled, err := Compile(tt.expr)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			got, err := compiled.Eval(tt.args...)
			if err != nil {
				t.Fatalf("Compiled Eval() error = %v", err)
			}
			if !reflect.DeepEqual(got.val, tt.want) {
				t.Fatalf("Compiled Eval() = %#v, want %#v", got.val, tt.want)
			}
		})
	}
}

func TestCompileWithFunc(t *testing.T) {
	e := Full(
		WithFunc("max", func(args ...any) (any, error) {
			if len(args) == 0 {
				return 0, nil
			}
			val := args[0].(float64)
			for i := 1; i < len(args); i++ {
				val = max(val, args[i].(float64))
			}
			return val, nil
		}),
	)
	compiled, err := e.Compile(`max(1, a, 3)`)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	got, err := compiled.EvalFloat(map[string]any{"a": 6})
	if err != nil {
		t.Fatalf("Compiled EvalFloat() error = %v", err)
	}
	if got != 6 {
		t.Fatalf("Compiled EvalFloat() = %v, want 6", got)
	}
}

func TestCompileEvalMap(t *testing.T) {
	compiled, err := Compile(`a.b[0] > 5`)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	got, err := compiled.EvalMapBool(map[string]any{"a": map[string]any{"b": []int{7, 8}}})
	if err != nil {
		t.Fatalf("EvalMapBool() error = %v", err)
	}
	if !got {
		t.Fatal("EvalMapBool() = false, want true")
	}
}

func TestCompileConstantFolding(t *testing.T) {
	tests := []struct {
		expr string
		args []any
		want any
	}{
		{expr: `1000 / 2 > (80 * 100 % 2)`, want: true},
		{expr: `"a" + "b" + 3`, want: "ab3"},
		{expr: `(date("2024-02-20")).Format("2006-01-02")`, want: "2024-02-20"},
		{
			expr: `duration("48h") - duration("24h")`,
			want: time.Hour * 24,
		},
		{
			expr: `1000 / 2 > threshold`,
			args: []any{map[string]any{"threshold": 499}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			compiled, err := Compile(tt.expr)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			got, err := compiled.Eval(tt.args...)
			if err != nil {
				t.Fatalf("Compiled Eval() error = %v", err)
			}
			if !reflect.DeepEqual(got.val, tt.want) {
				t.Fatalf("Compiled Eval() = %#v, want %#v", got.val, tt.want)
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

func TestCompileRegexLiteral(t *testing.T) {
	compiled, err := Compile(`name =~ "^fo+$"`)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	got, err := compiled.EvalBool(map[string]any{"name": "foo"})
	if err != nil {
		t.Fatalf("Compiled EvalBool() error = %v", err)
	}
	if !got {
		t.Fatal("Compiled EvalBool() = false, want true")
	}

	if _, err := Compile(`name =~ "["`); err == nil {
		t.Fatal("Compile() error = nil, want invalid regexp error")
	}
}

func TestCompileTypedPathFastPaths(t *testing.T) {
	compiled, err := Compile(`flags[0] && nested.ok`)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	got, err := compiled.EvalMapBool(map[string]any{
		"flags":  []bool{true},
		"nested": map[string]bool{"ok": true},
	})
	if err != nil {
		t.Fatalf("Compiled EvalMapBool() error = %v", err)
	}
	if !got {
		t.Fatal("Compiled EvalMapBool() = false, want true")
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

func TestBuiltins(t *testing.T) {
	params := map[string]any{
		"name":    "  GoEval  ",
		"email":   "dev@company.com",
		"tags":    []string{"go", "rules"},
		"missing": nil,
	}
	tests := []struct {
		expr string
		want bool
	}{
		{expr: `contains(email, "@company.com")`, want: true},
		{expr: `startsWith(trim(name), "Go") && endsWith(trim(name), "Eval")`, want: true},
		{expr: `lower("GO") == "go" && upper("go") == "GO"`, want: true},
		{expr: `replace("go-eval-go", "go", "expr") == "expr-eval-expr"`, want: true},
		{expr: `len(tags) == 2`, want: true},
		{expr: `min(3, 2, 5) == 2 && max(3, 2, 5) == 5`, want: true},
		{expr: `abs(-3) == 3 && round(1.6) == 2`, want: true},
		{expr: `int("3.9") == 3 && string(12) == "12" && bool("true")`, want: true},
		{expr: `exists(email) && !exists(missing) && empty(missing) && notEmpty(tags)`, want: true},
		{expr: `coalesce("", 0, "ok") == "ok" && default("", "fallback") == "fallback"`, want: true},
		{expr: `matches(email, "^[^@]+@company\\.com$") && regex(email, "company")`, want: true},
		{expr: `any(false, "", true) && all(true, "true", true)`, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, err := EvalBool(tt.expr, params)
			if err != nil {
				t.Fatalf("EvalBool() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("EvalBool() = %v, want %v", got, tt.want)
			}
		})
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
	got, err := Full(
		WithFunc("max", func(args ...any) (any, error) {
			return 42, nil
		}),
	).EvalInt(`max(1, 2, 3)`)
	if err != nil {
		t.Fatalf("EvalInt() error = %v", err)
	}
	if got != 42 {
		t.Fatalf("EvalInt() = %v, want 42", got)
	}
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

func TestNested(t *testing.T) {
	type args struct {
		expr string
		args []any
	}
	tests := []struct {
		args    args
		want    any
		wantErr bool
	}{
		{
			args: args{
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
			},
			want:    []any{"abc", 5},
			wantErr: false,
		},
		{
			args: args{
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
			},
			want:    float64(1),
			wantErr: false,
		},
		{
			args: args{
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
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
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
			},
			want:    []int{1, 2, 3},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.args.expr, func(t *testing.T) {
			v, _, err := Eval(tt.args.expr, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Eval() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(v.val, tt.want) {
				t.Errorf("Eval() got = %#v, want %#v", v.val, tt.want)
			}
		})
	}
}

func TestTimeAndDuration(t *testing.T) {
	type args struct {
		expr string
		args []any
	}
	tests := []struct {
		args    args
		want    any
		wantErr bool
	}{
		{
			args: args{
				expr: `(date("2024-02-20") - duration("24h")).Format("2006-01-02")`,
				args: []any{},
			},
			want:    "2024-02-19",
			wantErr: false,
		},
		{
			args: args{
				expr: `(date("2024-02-20") + duration("24h")).Format("2006-01-02")`,
				args: []any{},
			},
			want:    "2024-02-21",
			wantErr: false,
		},
		{
			args: args{
				expr: `duration("48h") - duration("24h")`,
				args: []any{},
			},
			want:    time.Hour * 24,
			wantErr: false,
		},
		{
			args: args{
				expr: `duration("24h") * 2`,
				args: []any{},
			},
			want:    time.Hour * 48,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.args.expr, func(t *testing.T) {
			v, _, err := Eval(tt.args.expr, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Eval() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(v.val, tt.want) {
				t.Errorf("Eval() got = %#v, want %#v", v.val, tt.want)
			}
		})
	}
}
