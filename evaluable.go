//go:generate ragel -Z lexer.rl
//go:generate goyacc -v /dev/null -o parser.go parser.y
//go:generate goyacc -p compile -v /dev/null -o compile_parser.go compile_parser.y
//go:generate gofmt -w parser.go compile_parser.go lexer.go
package goeval

import (
	"fmt"
	"reflect"
	"sync"
)

type Evaluable struct {
	// custom functions
	fns         map[string]Func
	foldableFns map[string]struct{}
	cache       *compileCache
}

type Func func(...any) (any, error)

type Operator func(any, any) (any, error)

type Option func(*Evaluable)

func WithFunc(name string, fn Func) Option {
	return func(e *Evaluable) {
		e.fns[name] = fn
		delete(e.foldableFns, name)
	}
}

func withFoldableFunc(name string, fn Func) Option {
	return func(e *Evaluable) {
		e.fns[name] = fn
		e.foldableFns[name] = struct{}{}
	}
}

func WithCompileCache(size int) Option {
	return func(e *Evaluable) {
		e.cache = newCompileCache(size)
	}
}

func NewEvaluable(opts ...Option) *Evaluable {
	e := &Evaluable{
		fns:         make(map[string]Func),
		foldableFns: make(map[string]struct{}),
		cache:       newCompileCache(defaultCompileCacheSize),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

var full = NewEvaluable(defaultOptions()...)

func Full(opts ...Option) *Evaluable {
	if len(opts) == 0 {
		return full
	}
	defaults := make([]Option, 0, len(full.fns)+len(opts))
	for name, fn := range full.fns {
		if _, ok := full.foldableFns[name]; ok {
			defaults = append(defaults, withFoldableFunc(name, fn))
		} else {
			defaults = append(defaults, WithFunc(name, fn))
		}
	}
	defaults = append(defaults, opts...)
	return NewEvaluable(defaults...)
}

func (e *Evaluable) Eval(expr string, args ...any) (Value, []string, error) {
	return e.eval(expr, true, args...)
}

func (e *Evaluable) eval(expr string, collectTokens bool, args ...any) (val Value, tokens []string, err error) {
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
	lex := newLexer(expr, pArgs, e.fns, collectTokens)
	defer lex.release()
	parseWithPool(lex)
	return lex.answer, lex.tokens, lex.err
}

func (e *Evaluable) EvalBool(expr string, args ...any) (bool, error) {
	val, _, err := e.eval(expr, false, args...)
	if err != nil {
		return false, err
	}
	return val.Boolean(), nil
}

func (e *Evaluable) EvalInt(expr string, args ...any) (int, error) {
	val, _, err := e.eval(expr, false, args...)
	if err != nil {
		return 0, err
	}
	return val.Int(), nil
}

func (e *Evaluable) EvalFloat(expr string, args ...any) (float64, error) {
	val, _, err := e.eval(expr, false, args...)
	if err != nil {
		return 0, err
	}
	return val.Float(), nil
}

func (e *Evaluable) EvalString(expr string, args ...any) (string, error) {
	val, _, err := e.eval(expr, false, args...)
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
	if len(args) == 1 {
		switch v := args[0].(type) {
		case nil:
			return nil, nil
		case map[string]any:
			return v, nil
		}
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

const defaultCompileCacheSize = 256

type compileCache struct {
	mu     sync.Mutex
	max    int
	order  []string
	values map[string]*CompiledExpression
}

func newCompileCache(max int) *compileCache {
	if max <= 0 {
		return nil
	}
	return &compileCache{
		max:    max,
		order:  make([]string, 0, max),
		values: make(map[string]*CompiledExpression, max),
	}
}

func (c *compileCache) get(expr string) (*CompiledExpression, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	compiled, ok := c.values[expr]
	return compiled, ok
}

func (c *compileCache) set(expr string, compiled *CompiledExpression) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.values[expr]; ok {
		c.values[expr] = compiled
		return
	}
	if len(c.order) >= c.max {
		oldest := c.order[0]
		copy(c.order, c.order[1:])
		c.order[len(c.order)-1] = expr
		delete(c.values, oldest)
	} else {
		c.order = append(c.order, expr)
	}
	c.values[expr] = compiled
}
