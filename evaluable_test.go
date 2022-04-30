package goeval

import (
	"fmt"
	"reflect"
	"testing"
)

func TestEvalBool(t *testing.T) {
	yyDebug = 1
	yyErrorVerbose = true
	type args struct {
		expr string
		args []interface{}
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
				args: []interface{}{map[string]interface{}{"abc": 7}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: `abc == 'abc"d'`,
				args: []interface{}{map[string]interface{}{"abc": "abc\"d"}},
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
				args: []interface{}{map[string]interface{}{"a": 2}},
			},
			want:    false,
			wantErr: false,
		},
		{
			args: args{
				expr: "a in [3,4,5,b]",
				args: []interface{}{map[string]interface{}{"a": 7, "b": 7}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: "a in b",
				args: []interface{}{map[string]interface{}{"a": 7, "b": "[7]"}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: "a in b",
				args: []interface{}{map[string]interface{}{"a": 7, "b": []int{7}}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: `$["a.b.0"] > 5`,
				args: []interface{}{map[string]interface{}{"a": map[string]interface{}{"b": []int{7, 8}}}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: "__hour > 6",
				args: []interface{}{map[string]interface{}{"__hour": 7}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: `a == "a\"bc"`,
				args: []interface{}{map[string]interface{}{"a": `a"bc`}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: `a == 'a\'bc'`,
				args: []interface{}{map[string]interface{}{"a": `a'bc`}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: "a == 'a`bc'",
				args: []interface{}{map[string]interface{}{"a": "a`bc"}},
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				expr: `$.["foo bar"].$.["foo baz"].data == "baz"`,
				args: []interface{}{map[string]interface{}{"foo bar": map[string]interface{}{"foo baz": map[string]interface{}{"data": "baz"}}}},
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
		WithFunc("max", func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return 0, nil
			}
			max := func(a, b float64) float64 {
				if a > b {
					return a
				}
				return b
			}
			var val float64
			val, ok := args[0].(float64)
			if !ok {
				return 0, fmt.Errorf("max() expects number arguments")
			}
			for i, arg := range args {
				if i == 0 {
					continue
				}
				arg, ok := arg.(float64)
				if !ok {
					return 0, fmt.Errorf("max() expects number arguments")
				}
				val = max(val, arg)
			}
			return val, nil
		}),
	).EvalFloat(`max(1,2,3,a)`, map[string]interface{}{"a": float64(6)})
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
		args []interface{}
	}
	tests := []struct {
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			args: args{
				expr: "a.b.c.Value(d,5)",
				args: []interface{}{
					map[string]interface{}{
						"a": map[string]interface{}{
							"b": map[string]interface{}{
								"c": &S{},
							},
						},
						"d": "abc",
					},
				},
			},
			want:    []interface{}{"abc", 5},
			wantErr: false,
		},
		{
			args: args{
				expr: "a.b.c[0]",
				args: []interface{}{
					map[string]interface{}{
						"a": map[string]interface{}{
							"b": map[string]interface{}{
								"c": []int{1, 2, 3},
							},
						},
					},
				},
			},
			want:    1,
			wantErr: false,
		},
		{
			args: args{
				expr: "a.b.c[0] == 1",
				args: []interface{}{
					map[string]interface{}{
						"a": map[string]interface{}{
							"b": map[string]interface{}{
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
				args: []interface{}{
					map[string]interface{}{
						"a": map[string]interface{}{
							"b": map[string]interface{}{
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
