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
}

type Func func(...any) (any, error)

type Operator func(any, any) (any, error)

type Option func(*Evaluable)

func WithFunc(name string, fn Func) Option {
	return func(e *Evaluable) {
		e.fns[name] = fn
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
	WithFunc("now", func(args ...any) (any, error) {
		return time.Now(), nil
	}),
	WithFunc("date", func(args ...any) (any, error) {
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
	WithFunc("strlen", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("strlen() expects exactly one string argument")
		}
		s, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("strlen() expects exactly one string argument")
		}
		return len(s), nil
	}),
	WithFunc("duration", func(args ...any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("date() expects at least one argument")
		}
		var (
			d   time.Duration
			err error
		)
		switch arg := args[0].(type) {
		case string:
			d, err = time.ParseDuration(arg)
		case float64:
			d = time.Duration(arg)
		default:
			err = fmt.Errorf("duration() expects a string or a number argument")
		}
		if err != nil {
			return nil, err
		}
		if len(args) == 1 {
			return d, nil
		}

		unit, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("duration() expects a string unit argument")
		}
		switch unit {
		case "ns":
			return d.Nanoseconds(), nil
		case "us":
			return d.Microseconds(), nil
		case "ms":
			return d.Milliseconds(), nil
		case "s":
			return d.Seconds(), nil
		case "m":
			return d.Minutes(), nil
		case "h":
			return d.Hours(), nil
		case "d":
			return d.Hours() / 24, nil
		default:
			return nil, fmt.Errorf("duration() expects a valid unit argument")
		}
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

func (e *Evaluable) Eval(expr string, args ...any) (val Value, tokens []string, err error) {
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

func (e *Evaluable) EvalBool(expr string, args ...any) (bool, error) {
	val, _, err := e.Eval(expr, args...)
	if err != nil {
		return false, err
	}
	return val.Boolean(), nil
}

func (e *Evaluable) EvalInt(expr string, args ...any) (int, error) {
	val, _, err := e.Eval(expr, args...)
	if err != nil {
		return 0, err
	}
	return val.Int()
}

func (e *Evaluable) EvalFloat(expr string, args ...any) (float64, error) {
	val, _, err := e.Eval(expr, args...)
	if err != nil {
		return 0, err
	}
	return val.Float()
}

func (e *Evaluable) EvalString(expr string, args ...any) (string, error) {
	val, _, err := e.Eval(expr, args...)
	if err != nil {
		return "", err
	}
	return val.String(), nil
}

func Eval(expr string, args ...any) (Value, []string, error) {
	return Full().Eval(expr, args...)
}

func EvalBool(expr string, args ...any) (bool, error) {
	return Full().EvalBool(expr, args...)
}

func EvalInt(expr string, args ...any) (int, error) {
	return Full().EvalInt(expr, args...)
}

func EvalFloat(expr string, args ...any) (float64, error) {
	return Full().EvalFloat(expr, args...)
}

func EvalString(expr string, args ...any) (string, error) {
	return Full().EvalString(expr, args...)
}

// parseArgs parses the arguments to the evaluable function. supports map[string]any or map[string]otherType
func parseArgs(args ...any) (map[string]any, error) {
	if len(args) == 0 {
		return nil, nil
	}
	pArgs := make(map[string]any)
	for _, arg := range args {
		if arg == nil {
			continue
		}
		switch v := arg.(type) {
		case map[string]any:
			for k, v := range v {
				pArgs[k] = v
			}
		default:
			rv := reflect.ValueOf(arg)
			if rv.Kind() == reflect.Ptr {
				rv = rv.Elem()
			}
			if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
				for _, k := range rv.MapKeys() {
					pArgs[k.String()] = rv.MapIndex(k).Interface()
				}
			} else {
				return nil, fmt.Errorf("unsupported argument type: %T", arg)
			}
		}
	}
	return pArgs, nil
}
