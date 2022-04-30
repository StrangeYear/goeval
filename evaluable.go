//go:generate ragel -Z lexer.rl
//go:generate goyacc -o parser.go parser.y
//go:generate gofmt -w parser.go lexer.go
package goeval

import (
	"fmt"
	"reflect"
	"time"
)

type Evaluable struct {
	// custom functions
	fns map[string]Func
	// custom operators
	ops map[string]Operator
}

type Func func(...interface{}) (interface{}, error)

type Operator func(interface{}, interface{}) (interface{}, error)

type Option func(*Evaluable)

func WithFunc(name string, fn Func) Option {
	return func(e *Evaluable) {
		e.fns[name] = fn
	}
}

func WithOperator(op string, fn Operator) Option {
	return func(e *Evaluable) {
		e.ops[op] = fn
	}
}

func NewEvaluable(opts ...Option) *Evaluable {
	e := &Evaluable{
		fns: make(map[string]Func),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

var full = NewEvaluable(
	WithFunc("now", func(args ...interface{}) (interface{}, error) {
		return time.Now(), nil
	}),
	WithFunc("date", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("date() expects exactly one argument")
		}
		arg := args[0]
		if s, ok := arg.(string); ok {
			for _, format := range [...]string{
				time.ANSIC,
				time.UnixDate,
				time.RubyDate,
				time.Kitchen,
				time.RFC3339,
				time.RFC3339Nano,
				"2006-01-02",                         // RFC 3339
				"2006-01-02 15:04",                   // RFC 3339 with minutes
				"2006-01-02 15:04:05",                // RFC 3339 with seconds
				"2006-01-02 15:04:05-07:00",          // RFC 3339 with seconds and timezone
				"2006-01-02T15Z0700",                 // ISO8601 with hour
				"2006-01-02T15:04Z0700",              // ISO8601 with minutes
				"2006-01-02T15:04:05Z0700",           // ISO8601 with seconds
				"2006-01-02T15:04:05.999999999Z0700", // ISO8601 with nanoseconds
			} {
				ret, err := time.ParseInLocation(format, s, time.Local)
				if err == nil {
					return ret, nil
				}
			}
		} else if f, ok := arg.(float64); ok {
			return time.Unix(int64(f), 0), nil
		}
		return nil, fmt.Errorf("date() expects a string or a number argument")
	}),
	WithFunc("strlen", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("strlen() expects exactly one string argument")
		}
		s, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("strlen() expects exactly one string argument")
		}
		return len(s), nil
	}),
)

func Full(opts ...Option) *Evaluable {
	if len(opts) == 0 {
		return full
	}
	for name, fn := range full.fns {
		opts = append(opts, WithFunc(name, fn))
	}
	return NewEvaluable(opts...)
}

func (e *Evaluable) Eval(expr string, args ...interface{}) (val Value, tokens []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			val = nilValue
			err = fmt.Errorf("%v", r)
		}
	}()
	pArgs, err := parseArgs(args...)
	if err != nil {
		return nilValue, nil, err
	}
	lex := newLexer([]byte(expr), pArgs, e.fns)
	yyParse(lex)
	return lex.answer, lex.tokens, lex.err
}

func (e *Evaluable) EvalBool(expr string, args ...interface{}) (bool, error) {
	val, _, err := e.Eval(expr, args...)
	if err != nil {
		return false, err
	}
	return val.Boolean(), nil
}

func (e *Evaluable) EvalInt(expr string, args ...interface{}) (int, error) {
	val, _, err := e.Eval(expr, args...)
	if err != nil {
		return 0, err
	}
	return val.Int()
}

func (e *Evaluable) EvalFloat(expr string, args ...interface{}) (float64, error) {
	val, _, err := e.Eval(expr, args...)
	if err != nil {
		return 0, err
	}
	return val.Float()
}

func (e *Evaluable) EvalString(expr string, args ...interface{}) (string, error) {
	val, _, err := e.Eval(expr, args...)
	if err != nil {
		return "", err
	}
	return val.String(), nil
}

func Eval(expr string, args ...interface{}) (Value, []string, error) {
	return Full().Eval(expr, args...)
}

func EvalBool(expr string, args ...interface{}) (bool, error) {
	return Full().EvalBool(expr, args...)
}

func EvalInt(expr string, args ...interface{}) (int, error) {
	return Full().EvalInt(expr, args...)
}

func EvalFloat(expr string, args ...interface{}) (float64, error) {
	return Full().EvalFloat(expr, args...)
}

func EvalString(expr string, args ...interface{}) (string, error) {
	return Full().EvalString(expr, args...)
}

func parseArgs(args ...interface{}) (map[string]interface{}, error) {
	if len(args) == 0 {
		return nil, nil
	}
	pArgs := make(map[string]interface{})
	for _, arg := range args {
		switch v := arg.(type) {
		case map[string]interface{}:
			for k, v := range v {
				pArgs[k] = v
			}
		case map[interface{}]interface{}:
			for k, v := range v {
				pArgs[fmt.Sprintf("%d", k)] = v
			}
		default:
			rv := reflect.ValueOf(arg)
			if rv.Kind() == reflect.Ptr {
				rv = rv.Elem()
			}
			if rv.Kind() == reflect.Struct {
				for i := 0; i < rv.NumField(); i++ {
					f := rv.Field(i)
					if f.CanInterface() {
						pArgs[rv.Type().Field(i).Name] = f.Interface()
					}
				}
			} else {
				return nil, fmt.Errorf("invalid argument type: %T", arg)
			}
		}
	}
	return pArgs, nil
}
