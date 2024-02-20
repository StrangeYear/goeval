package goeval

import (
	"fmt"
	"reflect"
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
