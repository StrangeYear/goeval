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
- date location and format: `date("2024/02/20 10:30", "Asia/Shanghai", "2006/01/02 15:04")`
- `strlen("someReallyLongInputString") <= 16`
- compiled expression: `compiled, err := goeval.Compile("foo > 0")`
- expression validation: `err := goeval.Validate("foo > 0")`
- variable dependency extraction: `vars, err := goeval.Variables("foo.bar > min")`
- structured dependency extraction: `deps, err := goeval.Dependencies("foo[bar] > 0")`
- evaluation trace: `steps, err := goeval.Explain("foo > 0", params)`
- decimal arithmetic: `goeval.Full(goeval.WithDecimal(true)).EvalBool("0.1 + 0.2 == 0.3")`
- strict variables: `goeval.New(goeval.WithStrictVariables(true)).EvalMapBool("foo > 0", params)`
- compile cache: `goeval.Full(goeval.WithCompileCache(1024))`

It can easily be extended with custom functions or operators:

custom comparator: max(1,2,3) > 1

### Decimal arithmetic

By default numbers use the existing `float64` behavior. Enable decimal mode when arithmetic expressions need exact decimal addition, subtraction, multiplication, and division:

```go
ok, err := goeval.Full(goeval.WithDecimal(true)).EvalBool(`0.1 + 0.2 == 0.3`)
// ok == true
```

Compiled expressions also use decimal arithmetic during constant folding when compiled from an evaluable configured with `WithDecimal(true)`.

### Constructors and function sets

Use `New` when you want a locked-down evaluator with no built-in functions. Register only the functions you allow:

```go
e := goeval.New(goeval.WithFunc("contains", func(args ...any) (any, error) {
    return strings.Contains(fmt.Sprint(args[0]), fmt.Sprint(args[1])), nil
}))

ok, err := e.EvalMapBool(`contains(name, "AWS")`, map[string]any{
    "name": "AWS Marketplace",
})
```

Use `Full` or the package-level helpers (`EvalBool`, `Compile`, `Variables`, and so on) when you want the built-in function set.

Custom functions can also be registered with aliases:

```go
e := goeval.New(
    goeval.WithFunc("riskScore", riskScore, "risk_score"),
    goeval.WithFuncAlias("riskScore", "score"),
)

ok, err := e.EvalMapBool(`risk_score(amount) > 10`, map[string]any{
    "amount": 320,
})
```

### Strict variables and compile cache

Strict variables return errors for missing top-level variables, missing map keys, missing struct fields, nil path traversal, and out-of-range indexes:

```go
e := goeval.New(goeval.WithStrictVariables(true))
_, err := e.EvalMapBool(`user.account.id == "1"`, map[string]any{
    "user": map[string]any{},
})
// err contains: undefined path "user.account"
```

Repeated expressions can be compiled once and reused directly, or cached at evaluator level:

```go
e := goeval.New(
    goeval.WithStrictVariables(true),
    goeval.WithDecimal(true),
    goeval.WithCompileCache(1024),
    goeval.WithFunc("contains", func(args ...any) (any, error) {
        return strings.Contains(fmt.Sprint(args[0]), fmt.Sprint(args[1])), nil
    }),
)

compiled, err := e.Compile(`count >= 3 && contains(name, "AWS")`)
if err != nil {
    return err
}

ok, err := compiled.EvalMapBool(map[string]any{
    "count": "3.0",
    "name":  "AWS Marketplace",
})
```

### What operators and types does this support?
- Modifiers: + - / * %
- Comparators: > >= < <= == != =~ !~ in
- String, range, window, and membership operators: `contains`, `not contains`, `starts_with`, `not starts_with`, `ends_with`, `not ends_with`, `between`, `not between`, `within_last`, `in`, `not in`
- Logical ops: || &&
- Numeric constants, as 64-bit floating point (12345.678)
- String constants ("foo", 'bar', `baz`)
- Boolean constants: true false
- Parenthesis to control order of evaluation ( )
- Arrays (anything separated by , within parenthesis: [1, 2, 'foo', bar], bar is variable)
- Prefixes: ! - ~
- Ternary conditional: ? :
- Null coalescence: ??

### Built-in functions
- Strings: `contains`, `startsWith` / `starts_with`, `endsWith` / `ends_with`, `lower`, `upper`, `trim`, `replace`, `strlen`
- Collections: `len`
- Numbers: `min`, `max`, `abs`, `round`
- Conversion: `int`, `float`, `decimal`, `string`, `bool`
- Presence checks: `exists`, `empty`, `notEmpty` / `not_empty`, `coalesce`, `default`
- Matching and booleans: `matches`, `regex`, `any`, `all`
- Ranges and windows: `between`, `withinLast` / `within_last`
- Time: `now`, `date`, `duration`

Examples:

```go
ok, err := goeval.Full().EvalMapBool(`between(amount, 100, 500)`, map[string]any{
    "amount": 320,
})

ok, err = goeval.Full().EvalMapBool(
    `withinLast(date(created_at), duration("72h"), date(event_triggered_at))`,
    map[string]any{
        "created_at":         "2024-02-20 12:00:00",
        "event_triggered_at": "2024-02-23 12:00:00",
    },
)
```

The same range and window checks can be written as operators:

```go
ok, err := goeval.Full().EvalMapBool(`amount between [100, 500]`, map[string]any{
    "amount": 320,
})

ok, err = goeval.Full().EvalMapBool(
    `date(created_at) within_last [duration("72h"), date(event_triggered_at)]`,
    map[string]any{
        "created_at":         "2024-02-20 12:00:00",
        "event_triggered_at": "2024-02-23 12:00:00",
    },
)
```

### Compile and validate
```go
compiled, err := goeval.Compile(`user.age >= 18 && contains(email, "@company.com")`)
if err != nil {
    return err
}

ok, err := compiled.EvalMapBool(map[string]any{
    "user":  map[string]any{"age": 20},
    "email": "dev@company.com",
})
```

Lexer and parser errors include line and column information, for example:
`unexpected character "@" at line 1, column 3`.

```go
vars, err := goeval.Variables(`user.age >= minAge && tags[idx] == "go"`)
// vars == []string{"idx", "minAge", "tags", "user.age"}
```

```go
deps, err := goeval.Dependencies(`contains(email, "@") && tags[idx] == "go"`)
// deps includes function, variable, and dynamic-index dependency metadata.
```

```go
steps, err := goeval.Explain(`age > 18 && contains(email, "@")`, map[string]any{
    "age":   20,
    "email": "dev@company.com",
})
// steps contains the evaluated subexpressions and their values.
```

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
