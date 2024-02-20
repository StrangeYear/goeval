## GoEval

Provides support for evaluating arbitrary C-like artithmetic/string expressions.

### How do I use it?
GoEval can evaluate expressions with parameters, arimethetic, logical, and string operations:

- basic expression: `10 > 0`
- parameterized expression: `foo > 0`
- special parameterized expression: `$.["foo bar"] > 0`、`$.["response-time"] > 0`
- nested parameterized expression: `foo.bar > 0`、`foo.bar[0] > 0`、`foo["bar"] > 0`、`$.["foo bar"].$.["foo baz"] > 0`
- gjson expression: `$["response-time"] > 0`、`$["data.items.0"] < 0`
- arithmetic expression: `(requests_made * requests_succeeded / 100) >= 90`
- string expression: `http_response_body == "service is ok"`
- float64 expression: `(mem_used / total_mem) * 100`
- date comparator: `date(`2022-05-02`) > date(`2022-05-01 23:59:59`)`
- date timestamp comparator: `date(1651467728) > date("2022-05-01 23:59:59")`
- `strlen("someReallyLongInputString") <= 16`

It can easily be extended with custom functions or operators:

custom comparator: max(1,2,3) > 1

### What operators and types does this support?
- Modifiers: + - / * %
- Comparators: > >= < <= == != =~ !~ in
- Logical ops: || &&
- Numeric constants, as 64-bit floating point (12345.678)
- String constants ("foo", 'bar', `baz`)
- Boolean constants: true false
- Parenthesis to control order of evaluation ( )
- Arrays (anything separated by , within parenthesis: [1, 2, 'foo', bar], bar is variable)
- Prefixes: ! - ~
- Ternary conditional: ? :
- Null coalescence: ??

### Functions
```go
func TestFunc(t *testing.T) {
	val, err := Full(
		WithFunc("max", func(args ...any) (any, error) {
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
	).EvalFloat(`max(1,2,3,a)`, map[string]any{"a": float64(6)})
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	t.Log(val == 6)
	// output: true
}
```

### Benchmarks
For a very rough idea of performance, here are the results output from a benchmark run on a Mac (i9-9900K).

```text
goos: darwin
goarch: amd64
pkg: github.com/StrangeYear/goeval
cpu: Intel(R) Core(TM) i9-9900K CPU @ 3.60GHz
BenchmarkEvaluationSingle-16                 	 2371440	       508.2 ns/op
BenchmarkEvaluationNumericLiteral-16         	  979606	      1178 ns/op
BenchmarkEvaluationLiteralModifiers-16       	  604010	      1903 ns/op
BenchmarkEvaluationParameter-16              	 1344694	       885.3 ns/op
BenchmarkEvaluationParameters-16             	  712880	      1594 ns/op
BenchmarkEvaluationParametersModifiers-16    	  478978	      2434 ns/op
BenchmarkComplexExpression-16                	  143646	      8187 ns/op
BenchmarkRegexExpression-16                  	  451497	      2595 ns/op
BenchmarkFunc-16                             	  268068	      4476 ns/op
BenchmarkJSON-16                             	  460448	      2557 ns/op
BenchmarkNested-16                           	  520017	      2287 ns/op
BenchmarkSpecial-16                          	 1323009	       908.7 ns/op
PASS
```
