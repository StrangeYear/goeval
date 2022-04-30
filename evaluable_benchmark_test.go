package goeval

import (
	"testing"
)

/*
  Benchmarks the bare-minimum evaluation time
*/
func BenchmarkEvaluationSingle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		val, _, err := Eval("1")
		if err != nil {
			b.Fatal(err)
		}
		if val.val != float64(1) {
			b.Fatal("Expected 1, got", val)
		}
	}
}

/*
  Benchmarks evaluation times of literals (no variables, no modifiers)
*/
func BenchmarkEvaluationNumericLiteral(b *testing.B) {
	for i := 0; i < b.N; i++ {
		val, _, err := Eval("(2) > (1)")
		if err != nil {
			b.Fatal(err)
		}
		if !val.Boolean() {
			b.Fatal("Expected true, got", val)
		}
	}
}

/*
  Benchmarks evaluation times of literals with modifiers
*/
func BenchmarkEvaluationLiteralModifiers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		val, _, err := Eval("(2) + (2) == (4)")
		if err != nil {
			b.Fatal(err)
		}
		if !val.Boolean() {
			b.Fatal("Expected true, got", val)
		}
	}
}

func BenchmarkEvaluationParameter(b *testing.B) {
	parameters := map[string]interface{}{
		"requests_made": 99.0,
	}
	for i := 0; i < b.N; i++ {
		val, _, err := Eval("requests_made", parameters)
		if err != nil {
			b.Fatal(err)
		}
		if val.val != float64(99) {
			b.Fatal("Expected 99, got", val)
		}
	}
}

/*
  Benchmarks evaluation times of parameters
*/
func BenchmarkEvaluationParameters(b *testing.B) {
	parameters := map[string]interface{}{
		"requests_made":      99.0,
		"requests_succeeded": 90.0,
	}

	for i := 0; i < b.N; i++ {
		val, _, err := Eval("requests_made > requests_succeeded", parameters)
		if err != nil {
			b.Fatal(err)
		}
		if !val.Boolean() {
			b.Fatal("Expected true, got", val)
		}
	}
}

/*
  Benchmarks evaluation times of parameters + literals with modifiers
*/
func BenchmarkEvaluationParametersModifiers(b *testing.B) {
	parameters := map[string]interface{}{
		"requests_made":      99.0,
		"requests_succeeded": 90.0,
	}

	for i := 0; i < b.N; i++ {
		val, _, err := Eval("(requests_made * requests_succeeded / 100) >= 90", parameters)
		if err != nil {
			b.Fatal(err)
		}
		if val.Boolean() {
			b.Fatal("Expected false, got", val)
		}
	}
}

/*
  Benchmarks the ludicrously-unlikely worst-case expression,
  one which uses all features.
  This is largely a canary benchmark to make sure that any syntax additions don't
  unnecessarily bloat the evaluation time.
*/
func BenchmarkComplexExpression(b *testing.B) {
	expressionString := "2 > 1 &&" +
		"'something' != 'nothing' || " +
		"date('2014-01-20') < date('Wed Jul  8 23:07:35 MDT 2015') && " +
		"$.['escapedVariable name with spaces'] >= $.['unescaped-variableName'] &&" +
		"modifierTest + 1000 / 2 > (80 * 100 % 2)"
	parameters := map[string]interface{}{
		"escapedVariable name with spaces": 99.0,
		"unescaped-variableName":           90.0,
		"modifierTest":                     5.0,
	}

	for i := 0; i < b.N; i++ {
		val, _, err := Eval(expressionString, parameters)
		if err != nil {
			b.Fatal(err)
		}
		if !val.Boolean() {
			b.Fatal("Expected true, got", val)
		}
	}
}

func BenchmarkRegexExpression(b *testing.B) {
	expressionString := "(foo !~ bar) && ('foobar' =~ oba)"
	parameters := map[string]interface{}{
		"foo": "foo",
		"bar": "bar",
		"oba": ".*oba.*",
	}

	for i := 0; i < b.N; i++ {
		val, err := EvalBool(expressionString, parameters)
		if err != nil {
			b.Fatal(err)
		}
		if !val {
			b.Fatal("Expected true, got", val)
		}
	}
}

func BenchmarkFunc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := Full().Eval(`now() < date("2022-05-02 00:00:00") && strlen(abc) == 3`, map[string]interface{}{"abc": "abc"})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		val, err := EvalBool(`$["a.b.0"] > 5`, map[string]interface{}{"a": map[string]interface{}{"b": []int{7, 8}}})
		if err != nil {
			b.Fatal(err)
		}
		if !val {
			b.Fatal("not true")
		}
	}
}

func BenchmarkNested(b *testing.B) {
	for i := 0; i < b.N; i++ {
		val, err := EvalBool(`a.b[0] > 5`, map[string]interface{}{"a": map[string]interface{}{"b": []int{7, 8}}})
		if err != nil {
			b.Fatal(err)
		}
		if !val {
			b.Fatal("not true")
		}
	}
}

func BenchmarkSpecial(b *testing.B) {
	for i := 0; i < b.N; i++ {
		val, err := EvalString(`$.["foo bar"]`, map[string]interface{}{"foo bar": "baz"})
		if err != nil {
			b.Fatal(err)
		}
		if val != "baz" {
			b.Fatal("not true")
		}
	}
}
