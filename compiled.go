package goeval

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/tidwall/gjson"
)

type CompiledExpression struct {
	root exprNode
	fns  map[string]Func
}

type evalContext struct {
	kv  map[string]any
	fns map[string]Func
}

type foldContext struct {
	fns         map[string]Func
	foldableFns map[string]struct{}
}

type exprNode interface {
	Eval(*evalContext) Value
}

func (e *Evaluable) Compile(expr string) (compiled *CompiledExpression, err error) {
	if compiled, ok := e.cache.get(expr); ok {
		return compiled, nil
	}
	defer func() {
		if r := recover(); r != nil {
			compiled = nil
			err = fmt.Errorf("%v", r)
		}
	}()
	lex := newLexer(expr, nil, e.fns, false)
	defer lex.release()
	compileLex := &compileLexerAdapter{lexer: lex}
	compileParseWithPool(compileLex)
	if lex.err != nil {
		return nil, lex.err
	}
	compiled = &CompiledExpression{
		root: foldNode(compileLex.answer, foldContext{
			fns:         e.fns,
			foldableFns: e.foldableFns,
		}),
		fns: e.fns,
	}
	e.cache.set(expr, compiled)
	return compiled, nil
}

func Compile(expr string) (*CompiledExpression, error) {
	return Full().Compile(expr)
}

func (e *Evaluable) Validate(expr string) error {
	_, err := e.Compile(expr)
	return err
}

func Validate(expr string) error {
	return Full().Validate(expr)
}

func (e *Evaluable) MustCompile(expr string) *CompiledExpression {
	compiled, err := e.Compile(expr)
	if err != nil {
		panic(err)
	}
	return compiled
}

func MustCompile(expr string) *CompiledExpression {
	return Full().MustCompile(expr)
}

func (e *Evaluable) Variables(expr string) ([]string, error) {
	compiled, err := e.Compile(expr)
	if err != nil {
		return nil, err
	}
	return compiled.Variables(), nil
}

func Variables(expr string) ([]string, error) {
	return Full().Variables(expr)
}

func (c *CompiledExpression) Variables() []string {
	collector := variableCollector{
		seen: make(map[string]struct{}),
	}
	collector.collect(c.root)
	sort.Strings(collector.names)
	return collector.names
}

type literalNode struct {
	value Value
}

func (n literalNode) Eval(*evalContext) Value {
	return n.value
}

type arrayNode struct {
	items []exprNode
}

func (n arrayNode) Eval(ctx *evalContext) Value {
	vals := make([]any, len(n.items))
	for i, item := range n.items {
		vals[i] = item.Eval(ctx).val
	}
	return NewValue("", vals)
}

type functionNode struct {
	name string
	args []exprNode
}

func (n functionNode) Eval(ctx *evalContext) Value {
	fn := ctx.fns[n.name]
	if fn == nil {
		panic(fmt.Errorf("unknown function %s", n.name))
	}
	res, err := evalNodeCall(ctx, fn, n.args)
	if err != nil {
		panic(err)
	}
	return NewValue("", res)
}

type variableNode struct {
	name string
}

func (n variableNode) Eval(ctx *evalContext) Value {
	return NewValue(n.name, ctx.kv[n.name])
}

type pathStep struct {
	key     string
	display string
	index   int
	isIndex bool
}

type pathNode struct {
	root  string
	steps []pathStep
	name  string
}

func (n pathNode) Eval(ctx *evalContext) Value {
	val := any(ctx.kv[n.root])
	for _, step := range n.steps {
		if step.isIndex {
			val = selectStaticIndex(val, step.index, step.key)
		} else {
			val = SelectValue(val, step.key)
		}
	}
	return NewValue(n.name, val)
}

type selectNameNode struct {
	base exprNode
	key  string
}

func (n selectNameNode) Eval(ctx *evalContext) Value {
	base := n.base.Eval(ctx)
	name := base.name + "." + n.key
	return NewValue(name, SelectValue(base.val, n.key))
}

type selectIndexNode struct {
	base         exprNode
	key          exprNode
	staticKey    string
	hasStaticKey bool
}

func (n selectIndexNode) Eval(ctx *evalContext) Value {
	base := n.base.Eval(ctx)
	key := n.staticKey
	keyValue := Value{val: key, vType: String}
	if !n.hasStaticKey {
		keyValue = n.key.Eval(ctx)
		key = keyValue.String()
	}
	name := indexName(base.name, keyValue)
	return NewValue(name, SelectValue(base.val, key))
}

type callNode struct {
	base exprNode
	args []exprNode
}

func (n callNode) Eval(ctx *evalContext) Value {
	base := n.base.Eval(ctx)
	val, err := evalNodeCall(ctx, toFunc(base.val), n.args)
	if err != nil {
		panic(err)
	}
	return NewValue(callName(base.name), val)
}

type jsonPathNode struct {
	path string
}

func (n jsonPathNode) Eval(ctx *evalContext) Value {
	return evalJSONPath(ctx.kv, n.path)
}

type unaryNode struct {
	op string
	x  exprNode
}

func (n unaryNode) Eval(ctx *evalContext) Value {
	return evalUnaryValue(n.op, n.x.Eval(ctx))
}

type binaryOp uint8

const (
	opEq binaryOp = iota + 1
	opNeq
	opGte
	opLte
	opRe
	opNre
	opNc
	opIn
	opLt
	opGt
	opMatch
	opAdd
	opSub
	opMulti
	opDiv
	opMod
	opAnd
	opOr
)

func (op binaryOp) String() string {
	switch op {
	case opEq:
		return "=="
	case opNeq:
		return "!="
	case opGte:
		return ">="
	case opLte:
		return "<="
	case opRe:
		return "=~"
	case opNre:
		return "!~"
	case opNc:
		return "??"
	case opIn:
		return "in"
	case opLt:
		return "<"
	case opGt:
		return ">"
	case opMatch:
		return "="
	case opAdd:
		return "+"
	case opSub:
		return "-"
	case opMulti:
		return "*"
	case opDiv:
		return "/"
	case opMod:
		return "%"
	case opAnd:
		return "&&"
	case opOr:
		return "||"
	default:
		return "unknown"
	}
}

type binaryNode struct {
	op          binaryOp
	left, right exprNode
}

func (n binaryNode) Eval(ctx *evalContext) Value {
	left := n.left.Eval(ctx)
	if val, ok := evalBinaryShortCircuit(n.op, left); ok {
		return val
	}
	return evalBinaryValue(n.op, left, n.right.Eval(ctx))
}

func evalUnaryValue(op string, x Value) Value {
	switch op {
	case "!":
		return x.Not()
	case "-":
		return NewValue("", float64(0)).Sub(x)
	default:
		panic(fmt.Errorf("unsupported unary operator %s", op))
	}
}

func evalBinaryShortCircuit(op binaryOp, left Value) (Value, bool) {
	switch op {
	case opAnd:
		if !left.Boolean() {
			return left, true
		}
	case opOr:
		if left.Boolean() {
			return left, true
		}
	case opNc:
		if !isCoalesceEmpty(left) {
			return left, true
		}
	default:
		return nilValue, false
	}
	return nilValue, false
}

func evalBinaryValue(op binaryOp, left, right Value) Value {
	switch op {
	case opEq:
		return left.Eq(right)
	case opNeq:
		return left.Neq(right)
	case opGte:
		return left.Gte(right)
	case opLte:
		return left.Lte(right)
	case opRe:
		return left.Re(right)
	case opNre:
		return left.Nre(right)
	case opIn:
		return left.In(right)
	case opLt:
		return left.Lt(right)
	case opGt:
		return left.Gt(right)
	case opMatch:
		return left.Match(right)
	case opAdd:
		return left.Add(right)
	case opSub:
		return left.Sub(right)
	case opMulti:
		return left.Multi(right)
	case opDiv:
		return left.Div(right)
	case opMod:
		return left.Mod(right)
	case opAnd, opOr, opNc:
		return right
	default:
		panic(fmt.Errorf("unsupported binary operator %s", op))
	}
}

type regexNode struct {
	left    exprNode
	exp     *regexp.Regexp
	pattern string
	negate  bool
}

func (n regexNode) Eval(ctx *evalContext) Value {
	matched := n.exp.MatchString(n.left.Eval(ctx).String())
	if n.negate {
		matched = !matched
	}
	return Value{
		val:   matched,
		vType: Boolean,
	}
}

type ternaryNode struct {
	cond, truthy, falsy exprNode
}

func (n ternaryNode) Eval(ctx *evalContext) Value {
	if n.cond.Eval(ctx).Boolean() {
		return n.truthy.Eval(ctx)
	}
	return n.falsy.Eval(ctx)
}

func (c *CompiledExpression) Eval(args ...any) (val Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			val = nilValue
			err = fmt.Errorf("%v", r)
		}
	}()
	pArgs, err := parseArgs(args...)
	if err != nil {
		return nilValue, err
	}
	return c.root.Eval(&evalContext{kv: pArgs, fns: c.fns}), nil
}

func (c *CompiledExpression) EvalMap(args map[string]any) (val Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			val = nilValue
			err = fmt.Errorf("%v", r)
		}
	}()
	return c.evalMapFast(args), nil
}

func (c *CompiledExpression) EvalBool(args ...any) (bool, error) {
	val, err := c.Eval(args...)
	if err != nil {
		return false, err
	}
	return val.Boolean(), nil
}

func (c *CompiledExpression) EvalMapBool(args map[string]any) (bool, error) {
	val, err := c.EvalMap(args)
	if err != nil {
		return false, err
	}
	return val.Boolean(), nil
}

func (c *CompiledExpression) EvalInt(args ...any) (int, error) {
	val, err := c.Eval(args...)
	if err != nil {
		return 0, err
	}
	return val.Int(), nil
}

func (c *CompiledExpression) EvalMapInt(args map[string]any) (int, error) {
	val, err := c.EvalMap(args)
	if err != nil {
		return 0, err
	}
	return val.Int(), nil
}

func (c *CompiledExpression) EvalFloat(args ...any) (float64, error) {
	val, err := c.Eval(args...)
	if err != nil {
		return 0, err
	}
	return val.Float(), nil
}

func (c *CompiledExpression) EvalMapFloat(args map[string]any) (float64, error) {
	val, err := c.EvalMap(args)
	if err != nil {
		return 0, err
	}
	return val.Float(), nil
}

func (c *CompiledExpression) EvalString(args ...any) (string, error) {
	val, err := c.Eval(args...)
	if err != nil {
		return "", err
	}
	return val.String(), nil
}

func (c *CompiledExpression) EvalMapString(args map[string]any) (string, error) {
	val, err := c.EvalMap(args)
	if err != nil {
		return "", err
	}
	return val.String(), nil
}

func (c *CompiledExpression) evalMapFast(args map[string]any) Value {
	return c.root.Eval(&evalContext{kv: args, fns: c.fns})
}

func foldNode(node exprNode, ctx foldContext) exprNode {
	switch n := node.(type) {
	case literalNode, variableNode, jsonPathNode:
		return node
	case pathNode:
		return node
	case arrayNode:
		items := foldNodes(n.items, ctx)
		if allLiterals(items) {
			return foldConstant(arrayNode{items: items}, ctx.fns)
		}
		n.items = items
		return n
	case functionNode:
		if ctx.fns[n.name] == nil {
			panic(fmt.Errorf("unknown function %s", n.name))
		}
		n.args = foldNodes(n.args, ctx)
		if isFoldableFunction(n.name, ctx) && allLiterals(n.args) {
			return foldConstant(n, ctx.fns)
		}
		return n
	case selectNameNode:
		n.base = foldNode(n.base, ctx)
		if path, ok := appendPathName(n.base, n.key, "."+n.key); ok {
			return path
		}
		if _, ok := n.base.(literalNode); ok {
			return foldConstant(n, ctx.fns)
		}
		return n
	case selectIndexNode:
		n.base = foldNode(n.base, ctx)
		n.key = foldNode(n.key, ctx)
		if literal, ok := n.key.(literalNode); ok {
			n.staticKey = literal.value.String()
			n.hasStaticKey = true
		}
		if n.hasStaticKey {
			if path, ok := appendPathName(n.base, n.staticKey, "["+n.staticKey+"]"); ok {
				return path
			}
		}
		if _, ok := n.base.(literalNode); ok && n.hasStaticKey {
			return foldConstant(n, ctx.fns)
		}
		return n
	case callNode:
		n.base = foldNode(n.base, ctx)
		n.args = foldNodes(n.args, ctx)
		if _, ok := n.base.(literalNode); ok && allLiterals(n.args) {
			return foldConstant(n, ctx.fns)
		}
		return n
	case unaryNode:
		n.x = foldNode(n.x, ctx)
		if _, ok := n.x.(literalNode); ok {
			return foldConstant(n, ctx.fns)
		}
		return n
	case binaryNode:
		n.left = foldNode(n.left, ctx)
		n.right = foldNode(n.right, ctx)
		if _, ok := n.left.(literalNode); ok {
			if _, ok := n.right.(literalNode); ok {
				return foldConstant(n, ctx.fns)
			}
		}
		if regex, ok := foldRegexNode(n); ok {
			return regex
		}
		return n
	case ternaryNode:
		n.cond = foldNode(n.cond, ctx)
		n.truthy = foldNode(n.truthy, ctx)
		n.falsy = foldNode(n.falsy, ctx)
		if _, ok := n.cond.(literalNode); ok {
			if _, ok := n.truthy.(literalNode); ok {
				if _, ok := n.falsy.(literalNode); ok {
					return foldConstant(n, ctx.fns)
				}
			}
		}
		return n
	default:
		return node
	}
}

func foldRegexNode(n binaryNode) (exprNode, bool) {
	if n.op != opRe && n.op != opNre {
		return nil, false
	}
	right, ok := n.right.(literalNode)
	if !ok {
		return nil, false
	}
	exp, err := compileRegexp(right.value.String())
	if err != nil {
		panic(fmt.Errorf("compile regexp error: %s, name: %s", err, right.value.name))
	}
	return regexNode{
		left:    n.left,
		exp:     exp,
		pattern: right.value.String(),
		negate:  n.op == opNre,
	}, true
}

func foldNodes(nodes []exprNode, ctx foldContext) []exprNode {
	if len(nodes) == 0 {
		return nil
	}
	folded := make([]exprNode, len(nodes))
	for i, node := range nodes {
		folded[i] = foldNode(node, ctx)
	}
	return folded
}

func allLiterals(nodes []exprNode) bool {
	for _, node := range nodes {
		if _, ok := node.(literalNode); !ok {
			return false
		}
	}
	return true
}

func foldConstant(node exprNode, fns map[string]Func) exprNode {
	return literalNode{value: node.Eval(&evalContext{fns: fns})}
}

func isFoldableFunction(name string, ctx foldContext) bool {
	_, ok := ctx.foldableFns[name]
	return ok && ctx.fns[name] != nil
}

func appendPathName(base exprNode, key, display string) (pathNode, bool) {
	step := pathStep{key: key, display: display}
	if strings.HasPrefix(display, "[") {
		if index, ok := parsePathIndex(key); ok {
			step.index = index
			step.isIndex = true
		}
	}
	switch n := base.(type) {
	case variableNode:
		return pathNode{
			root:  n.name,
			steps: []pathStep{step},
			name:  n.name + display,
		}, true
	case pathNode:
		steps := make([]pathStep, len(n.steps), len(n.steps)+1)
		copy(steps, n.steps)
		steps = append(steps, step)
		n.steps = steps
		n.name += display
		return n, true
	default:
		return pathNode{}, false
	}
}

func selectStaticIndex(value any, index int, fallbackKey string) any {
	switch v := value.(type) {
	case []any:
		if index < len(v) {
			return v[index]
		}
		return nil
	case []int:
		if index < len(v) {
			return v[index]
		}
		return nil
	case []float64:
		if index < len(v) {
			return v[index]
		}
		return nil
	case []string:
		if index < len(v) {
			return v[index]
		}
		return nil
	case []bool:
		if index < len(v) {
			return v[index]
		}
		return nil
	default:
		return SelectValue(value, fallbackKey)
	}
}

func evalNodeArgs(ctx *evalContext, nodes []exprNode) []any {
	if len(nodes) == 0 {
		return nil
	}
	args := make([]any, len(nodes))
	for i, node := range nodes {
		args[i] = node.Eval(ctx).val
	}
	return args
}

func evalNodeCall(ctx *evalContext, fn func(...any) (any, error), nodes []exprNode) (any, error) {
	switch len(nodes) {
	case 0:
		return fn()
	case 1:
		return fn(nodes[0].Eval(ctx).val)
	case 2:
		return fn(nodes[0].Eval(ctx).val, nodes[1].Eval(ctx).val)
	case 3:
		return fn(nodes[0].Eval(ctx).val, nodes[1].Eval(ctx).val, nodes[2].Eval(ctx).val)
	case 4:
		return fn(nodes[0].Eval(ctx).val, nodes[1].Eval(ctx).val, nodes[2].Eval(ctx).val, nodes[3].Eval(ctx).val)
	default:
		return fn(evalNodeArgs(ctx, nodes)...)
	}
}

type variableCollector struct {
	seen  map[string]struct{}
	names []string
}

func (c *variableCollector) add(name string) {
	if name == "" {
		return
	}
	if _, ok := c.seen[name]; ok {
		return
	}
	c.seen[name] = struct{}{}
	c.names = append(c.names, name)
}

func (c *variableCollector) collect(node exprNode) {
	switch n := node.(type) {
	case nil, literalNode:
		return
	case variableNode:
		c.add(n.name)
	case pathNode:
		c.add(n.name)
	case jsonPathNode:
		c.add(n.path)
	case arrayNode:
		c.collectAll(n.items)
	case functionNode:
		c.collectAll(n.args)
	case selectNameNode:
		c.collect(n.base)
	case selectIndexNode:
		c.collect(n.base)
		if !n.hasStaticKey {
			c.collect(n.key)
		}
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

func (c *variableCollector) collectAll(nodes []exprNode) {
	for _, node := range nodes {
		c.collect(node)
	}
}

func (lex *lexer) param(val Value) any {
	return val.val
}

func selectIndexExpr(base exprNode, key exprNode) exprNode {
	node := selectIndexNode{
		base: base,
		key:  key,
	}
	if literal, ok := node.key.(literalNode); ok {
		node.staticKey = literal.value.String()
		node.hasStaticKey = true
	}
	return node
}

func evalJSONPath(kv map[string]any, path string) Value {
	if val, ok := SelectPath(kv, path); ok {
		return NewValue(path, val)
	}
	bs, err := json.Marshal(kv)
	if err != nil {
		panic(fmt.Errorf("parameter json marshal failed, %s", err))
	}
	return NewValue(path, gjson.ParseBytes(bs).Get(path))
}
