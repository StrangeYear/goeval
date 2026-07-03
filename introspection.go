package goeval

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	DependencyVariable = "variable"
	DependencyJSONPath = "json_path"
	DependencyFunction = "function"
)

type TraceStep struct {
	Expression string
	Value      any
	Type       string
}

type Dependency struct {
	Name    string
	Kind    string
	Dynamic bool
}

func (e *Evaluable) Explain(expr string, args ...any) (steps []TraceStep, err error) {
	compiled, err := e.Compile(expr)
	if err != nil {
		return nil, err
	}
	return compiled.Explain(args...)
}

func Explain(expr string, args ...any) ([]TraceStep, error) {
	return Full().Explain(expr, args...)
}

func (c *CompiledExpression) Explain(args ...any) (steps []TraceStep, err error) {
	pArgs, err := parseArgs(args...)
	if err != nil {
		return nil, err
	}
	return c.ExplainMap(pArgs)
}

func (c *CompiledExpression) ExplainMap(args map[string]any) (steps []TraceStep, err error) {
	defer func() {
		if r := recover(); r != nil {
			steps = nil
			err = fmt.Errorf("%v", r)
		}
	}()
	ctx := &evalContext{kv: args, fns: c.fns}
	trace := explainTrace{}
	trace.eval(ctx, c.root)
	return trace.steps, nil
}

func (e *Evaluable) Dependencies(expr string) ([]Dependency, error) {
	compiled, err := e.Compile(expr)
	if err != nil {
		return nil, err
	}
	return compiled.Dependencies(), nil
}

func Dependencies(expr string) ([]Dependency, error) {
	return Full().Dependencies(expr)
}

func (c *CompiledExpression) Dependencies() []Dependency {
	collector := dependencyCollector{
		seen: make(map[string]struct{}),
	}
	collector.collect(c.root)
	sort.Slice(collector.items, func(i, j int) bool {
		if collector.items[i].Kind != collector.items[j].Kind {
			return collector.items[i].Kind < collector.items[j].Kind
		}
		if collector.items[i].Name != collector.items[j].Name {
			return collector.items[i].Name < collector.items[j].Name
		}
		return !collector.items[i].Dynamic && collector.items[j].Dynamic
	})
	return collector.items
}

type explainTrace struct {
	steps []TraceStep
}

func (t *explainTrace) add(expr string, value Value) Value {
	t.steps = append(t.steps, TraceStep{
		Expression: expr,
		Value:      value.val,
		Type:       valueTypeName(value.vType),
	})
	return value
}

func (t *explainTrace) eval(ctx *evalContext, node exprNode) Value {
	switch n := node.(type) {
	case literalNode:
		return t.add(nodeExpr(node), n.value)
	case arrayNode:
		vals := make([]any, len(n.items))
		for i, item := range n.items {
			vals[i] = t.eval(ctx, item).val
		}
		return t.add(nodeExpr(node), NewValue("", vals))
	case functionNode:
		fn := ctx.fns[n.name]
		if fn == nil {
			panic(fmt.Errorf("unknown function %s", n.name))
		}
		res, err := fn(t.evalArgs(ctx, n.args)...)
		if err != nil {
			panic(err)
		}
		return t.add(nodeExpr(node), NewValue("", res))
	case variableNode:
		return t.add(nodeExpr(node), NewValue(n.name, ctx.kv[n.name]))
	case pathNode:
		val := any(ctx.kv[n.root])
		for _, step := range n.steps {
			if step.isIndex {
				val = selectStaticIndex(val, step.index, step.key)
			} else {
				val = SelectValue(val, step.key)
			}
		}
		return t.add(nodeExpr(node), NewValue(n.name, val))
	case selectNameNode:
		base := t.eval(ctx, n.base)
		name := base.name + "." + n.key
		return t.add(nodeExpr(node), NewValue(name, SelectValue(base.val, n.key)))
	case selectIndexNode:
		base := t.eval(ctx, n.base)
		key := n.staticKey
		keyValue := Value{val: key, vType: String}
		if !n.hasStaticKey {
			keyValue = t.eval(ctx, n.key)
			key = keyValue.String()
		}
		name := indexName(base.name, keyValue)
		return t.add(nodeExpr(node), NewValue(name, SelectValue(base.val, key)))
	case callNode:
		base := t.eval(ctx, n.base)
		val, err := toFunc(base.val)(t.evalArgs(ctx, n.args)...)
		if err != nil {
			panic(err)
		}
		return t.add(nodeExpr(node), NewValue(callName(base.name), val))
	case jsonPathNode:
		return t.add(nodeExpr(node), evalJSONPath(ctx.kv, n.path))
	case unaryNode:
		return t.add(nodeExpr(node), evalUnaryValue(n.op, t.eval(ctx, n.x)))
	case binaryNode:
		return t.evalBinary(ctx, n)
	case regexNode:
		matched := n.exp.MatchString(t.eval(ctx, n.left).String())
		if n.negate {
			matched = !matched
		}
		return t.add(nodeExpr(node), Value{val: matched, vType: Boolean})
	case ternaryNode:
		cond := t.eval(ctx, n.cond)
		if cond.Boolean() {
			return t.add(nodeExpr(node), t.eval(ctx, n.truthy))
		}
		return t.add(nodeExpr(node), t.eval(ctx, n.falsy))
	default:
		panic(fmt.Errorf("unsupported explain node %T", node))
	}
}

func (t *explainTrace) evalArgs(ctx *evalContext, nodes []exprNode) []any {
	if len(nodes) == 0 {
		return nil
	}
	args := make([]any, len(nodes))
	for i, node := range nodes {
		args[i] = t.eval(ctx, node).val
	}
	return args
}

func (t *explainTrace) evalBinary(ctx *evalContext, n binaryNode) Value {
	left := t.eval(ctx, n.left)
	if val, ok := evalBinaryShortCircuit(n.op, left); ok {
		return t.add(nodeExpr(n), val)
	}
	return t.add(nodeExpr(n), evalBinaryValue(n.op, left, t.eval(ctx, n.right)))
}

type dependencyCollector struct {
	seen  map[string]struct{}
	items []Dependency
}

func (c *dependencyCollector) add(kind, name string, dynamic bool) {
	if name == "" {
		return
	}
	key := kind + "\x00" + name + "\x00" + strconv.FormatBool(dynamic)
	if _, ok := c.seen[key]; ok {
		return
	}
	c.seen[key] = struct{}{}
	c.items = append(c.items, Dependency{Kind: kind, Name: name, Dynamic: dynamic})
}

func (c *dependencyCollector) collect(node exprNode) {
	switch n := node.(type) {
	case nil, literalNode:
		return
	case variableNode:
		c.add(DependencyVariable, n.name, false)
	case pathNode:
		c.add(DependencyVariable, n.name, false)
	case jsonPathNode:
		c.add(DependencyJSONPath, n.path, false)
	case arrayNode:
		c.collectAll(n.items)
	case functionNode:
		c.add(DependencyFunction, n.name, false)
		c.collectAll(n.args)
	case selectNameNode:
		c.collect(n.base)
	case selectIndexNode:
		if !n.hasStaticKey {
			c.add(DependencyVariable, nodeExpr(n.base), true)
			c.collect(n.key)
		}
		c.collect(n.base)
	case callNode:
		c.collect(n.base)
		c.collectAll(n.args)
	case unaryNode:
		c.collect(n.x)
	case binaryNode:
		c.collect(n.left)
		c.collect(n.right)
	case regexNode:
		c.collect(n.left)
	case ternaryNode:
		c.collect(n.cond)
		c.collect(n.truthy)
		c.collect(n.falsy)
	}
}

func (c *dependencyCollector) collectAll(nodes []exprNode) {
	for _, node := range nodes {
		c.collect(node)
	}
}

func nodeExpr(node exprNode) string {
	switch n := node.(type) {
	case literalNode:
		return valueExpr(n.value)
	case arrayNode:
		parts := make([]string, len(n.items))
		for i, item := range n.items {
			parts[i] = nodeExpr(item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case functionNode:
		return n.name + "(" + nodeExprList(n.args) + ")"
	case variableNode:
		return n.name
	case pathNode:
		return n.name
	case selectNameNode:
		return nodeExpr(n.base) + "." + n.key
	case selectIndexNode:
		if n.hasStaticKey {
			return nodeExpr(n.base) + "[" + n.staticKey + "]"
		}
		return nodeExpr(n.base) + "[" + nodeExpr(n.key) + "]"
	case callNode:
		return nodeExpr(n.base) + "(" + nodeExprList(n.args) + ")"
	case jsonPathNode:
		return "$[" + strconv.Quote(n.path) + "]"
	case unaryNode:
		return n.op + nodeExpr(n.x)
	case binaryNode:
		return "(" + nodeExpr(n.left) + " " + n.op.String() + " " + nodeExpr(n.right) + ")"
	case regexNode:
		op := "=~"
		if n.negate {
			op = "!~"
		}
		return "(" + nodeExpr(n.left) + " " + op + " " + strconv.Quote(n.pattern) + ")"
	case ternaryNode:
		return "(" + nodeExpr(n.cond) + " ? " + nodeExpr(n.truthy) + " : " + nodeExpr(n.falsy) + ")"
	default:
		return fmt.Sprintf("%T", node)
	}
}

func nodeExprList(nodes []exprNode) string {
	if len(nodes) == 0 {
		return ""
	}
	parts := make([]string, len(nodes))
	for i, node := range nodes {
		parts[i] = nodeExpr(node)
	}
	return strings.Join(parts, ", ")
}

func valueExpr(v Value) string {
	switch v.vType {
	case String:
		return strconv.Quote(v.String())
	default:
		return v.String()
	}
}

func valueTypeName(t uint8) string {
	switch t {
	case Nil:
		return "nil"
	case Boolean:
		return "boolean"
	case Number:
		return "number"
	case String:
		return "string"
	case Array:
		return "array"
	case Time:
		return "time"
	case Duration:
		return "duration"
	case Json:
		return "json"
	case Struct:
		return "struct"
	case Map:
		return "map"
	case Interface:
		return "interface"
	default:
		return "unknown"
	}
}
