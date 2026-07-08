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
	root            exprNode
	fns             map[string]Func
	useDecimal      bool
	strictVariables bool
}

type evalContext struct {
	kv              map[string]any
	fns             map[string]Func
	useDecimal      bool
	strictVariables bool
}

type foldContext struct {
	fns         map[string]Func
	foldableFns map[string]struct{}
	useDecimal  bool
}

type exprNode interface {
	Eval(*evalContext) Value
}

func (ctx *evalContext) newValue(name string, val any) Value {
	return newValue(name, val, ctx.useDecimal)
}

func (e *Evaluable) Compile(expr string) (compiled *CompiledExpression, err error) {
	cacheKey := compileCacheKey(expr, e.useDecimal, e.strictVariables)
	if compiled, ok := e.cache.get(cacheKey); ok {
		return compiled, nil
	}
	defer func() {
		if r := recover(); r != nil {
			compiled = nil
			err = fmt.Errorf("%v", r)
		}
	}()
	lex := newLexer(expr, nil, e.fns, false, e.useDecimal)
	defer lex.release()
	compileLex := &compileLexerAdapter{lexer: lex}
	compileParseWithPool(compileLex)
	if lex.err != nil {
		return nil, lex.err
	}
	fns := cloneFuncs(e.fns)
	compiled = &CompiledExpression{
		root: foldNode(compileLex.answer, foldContext{
			fns:         fns,
			foldableFns: cloneFoldableFuncs(e.foldableFns),
			useDecimal:  e.useDecimal,
		}),
		fns:             fns,
		useDecimal:      e.useDecimal,
		strictVariables: e.strictVariables,
	}
	e.cache.set(cacheKey, compiled)
	return compiled, nil
}

func cloneFuncs(funcs map[string]Func) map[string]Func {
	if len(funcs) == 0 {
		return nil
	}
	clone := make(map[string]Func, len(funcs))
	for name, fn := range funcs {
		clone[name] = fn
	}
	return clone
}

func cloneFoldableFuncs(funcs map[string]struct{}) map[string]struct{} {
	if len(funcs) == 0 {
		return nil
	}
	clone := make(map[string]struct{}, len(funcs))
	for name := range funcs {
		clone[name] = struct{}{}
	}
	return clone
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
	return ctx.newValue("", vals)
}

type functionNode struct {
	name string
	args []exprNode
}

func (n functionNode) Eval(ctx *evalContext) Value {
	fn := ctx.fns[n.name]
	if fn == nil {
		panic(fmt.Errorf("undefined function %q", n.name))
	}
	res, err := evalNodeCall(ctx, fn, n.args)
	if err != nil {
		panic(err)
	}
	return ctx.newValue("", res)
}

type variableNode struct {
	name string
}

func (n variableNode) Eval(ctx *evalContext) Value {
	val, ok := ctx.kv[n.name]
	if ctx.strictVariables && !ok {
		panic(fmt.Errorf("undefined variable %q", n.name))
	}
	return ctx.newValue(n.name, val)
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
	val, ok := ctx.kv[n.root]
	if ctx.strictVariables && !ok {
		panic(fmt.Errorf("undefined variable %q", n.root))
	}
	path := n.root
	for _, step := range n.steps {
		path += step.display
		if step.isIndex {
			val = selectStaticIndex(ctx, val, step.index, step.key, path)
		} else {
			val = ctx.selectValue(val, step.key, path)
		}
	}
	return ctx.newValue(n.name, val)
}

type selectNameNode struct {
	base exprNode
	key  string
}

func (n selectNameNode) Eval(ctx *evalContext) Value {
	base := n.base.Eval(ctx)
	name := base.name + "." + n.key
	return ctx.newValue(name, ctx.selectValue(base.val, n.key, name))
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
	return ctx.newValue(name, ctx.selectValue(base.val, key, name))
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
	return ctx.newValue(callName(base.name), val)
}

type jsonPathNode struct {
	path string
}

func (n jsonPathNode) Eval(ctx *evalContext) Value {
	return evalJSONPath(ctx, n.path)
}

type unaryNode struct {
	op string
	x  exprNode
}

func (n unaryNode) Eval(ctx *evalContext) Value {
	return evalUnaryValue(n.op, n.x.Eval(ctx), ctx.useDecimal)
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
	opNotIn
	opContains
	opNotContains
	opStartsWith
	opNotStartsWith
	opEndsWith
	opNotEndsWith
	opBetween
	opNotBetween
	opWithinLast
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
	case opNotIn:
		return "not in"
	case opContains:
		return "contains"
	case opNotContains:
		return "not contains"
	case opStartsWith:
		return "starts_with"
	case opNotStartsWith:
		return "not starts_with"
	case opEndsWith:
		return "ends_with"
	case opNotEndsWith:
		return "not ends_with"
	case opBetween:
		return "between"
	case opNotBetween:
		return "not between"
	case opWithinLast:
		return "within_last"
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

func evalUnaryValue(op string, x Value, useDecimal bool) Value {
	switch op {
	case "!":
		return x.Not()
	case "-":
		return newValue("", 0, useDecimal).Sub(x)
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
	case opNotIn:
		return left.In(right).Not()
	case opContains:
		return evalStringOperatorValue("contains", left, right)
	case opNotContains:
		return evalStringOperatorValue("contains", left, right).Not()
	case opStartsWith:
		return evalStringOperatorValue("starts_with", left, right)
	case opNotStartsWith:
		return evalStringOperatorValue("starts_with", left, right).Not()
	case opEndsWith:
		return evalStringOperatorValue("ends_with", left, right)
	case opNotEndsWith:
		return evalStringOperatorValue("ends_with", left, right).Not()
	case opBetween:
		return evalBetweenOperatorValue(left, right)
	case opNotBetween:
		return evalBetweenOperatorValue(left, right).Not()
	case opWithinLast:
		return evalWithinLastOperatorValue(left, right)
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

func evalStringOperatorValue(op string, left, right Value) Value {
	var matched bool
	switch op {
	case "contains":
		matched = strings.Contains(left.String(), right.String())
	case "starts_with":
		matched = strings.HasPrefix(left.String(), right.String())
	case "ends_with":
		matched = strings.HasSuffix(left.String(), right.String())
	default:
		panic(fmt.Errorf("unsupported string operator %s", op))
	}
	return Value{val: matched, vType: Boolean}
}

func evalBetweenOperatorValue(left, right Value) Value {
	args := operatorArrayArgs("between", right, 2)
	minVal := NewValue("", args[0])
	maxVal := NewValue("", args[1])
	return Value{
		val:   left.Gte(minVal).Boolean() && left.Lte(maxVal).Boolean(),
		vType: Boolean,
	}
}

func evalWithinLastOperatorValue(left, right Value) Value {
	args := operatorArrayArgs("within_last", right, 2)
	window := NewValue("", args[0])
	anchor := NewValue("", args[1])
	if left.vType != Time || window.vType != Duration || anchor.vType != Time {
		panic(fmt.Errorf("within_last operator expects a time value and [duration window, time anchor]"))
	}
	lower := anchor.Sub(window)
	return Value{
		val:   lower.Lte(left).Boolean() && left.Lte(anchor).Boolean(),
		vType: Boolean,
	}
}

func operatorArrayArgs(name string, right Value, count int) []any {
	args := right.Array()
	if len(args) != count {
		panic(fmt.Errorf("%s operator expects an array with exactly %d item(s)", name, count))
	}
	return args
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
	return c.root.Eval(c.newEvalContext(pArgs)), nil
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
	return safeValueInt(val)
}

func (c *CompiledExpression) EvalMapInt(args map[string]any) (int, error) {
	val, err := c.EvalMap(args)
	if err != nil {
		return 0, err
	}
	return safeValueInt(val)
}

func (c *CompiledExpression) EvalFloat(args ...any) (float64, error) {
	val, err := c.Eval(args...)
	if err != nil {
		return 0, err
	}
	return safeValueFloat(val)
}

func (c *CompiledExpression) EvalMapFloat(args map[string]any) (float64, error) {
	val, err := c.EvalMap(args)
	if err != nil {
		return 0, err
	}
	return safeValueFloat(val)
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
	return c.root.Eval(c.newEvalContext(args))
}

func (c *CompiledExpression) newEvalContext(args map[string]any) *evalContext {
	return &evalContext{
		kv:              args,
		fns:             c.fns,
		useDecimal:      c.useDecimal,
		strictVariables: c.strictVariables,
	}
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
			return foldConstant(arrayNode{items: items}, ctx)
		}
		n.items = items
		return n
	case functionNode:
		if ctx.fns[n.name] == nil {
			panic(fmt.Errorf("undefined function %q", n.name))
		}
		n.args = foldNodes(n.args, ctx)
		if isFoldableFunction(n.name, ctx) && allLiterals(n.args) {
			return foldConstant(n, ctx)
		}
		return n
	case selectNameNode:
		n.base = foldNode(n.base, ctx)
		if path, ok := appendPathName(n.base, n.key, "."+n.key); ok {
			return path
		}
		if _, ok := n.base.(literalNode); ok {
			return foldConstant(n, ctx)
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
			return foldConstant(n, ctx)
		}
		return n
	case callNode:
		n.base = foldNode(n.base, ctx)
		n.args = foldNodes(n.args, ctx)
		if _, ok := n.base.(literalNode); ok && allLiterals(n.args) {
			return foldConstant(n, ctx)
		}
		return n
	case unaryNode:
		n.x = foldNode(n.x, ctx)
		if _, ok := n.x.(literalNode); ok {
			return foldConstant(n, ctx)
		}
		return n
	case binaryNode:
		n.left = foldNode(n.left, ctx)
		n.right = foldNode(n.right, ctx)
		if _, ok := n.left.(literalNode); ok {
			if _, ok := n.right.(literalNode); ok {
				return foldConstant(n, ctx)
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
					return foldConstant(n, ctx)
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

func foldConstant(node exprNode, ctx foldContext) exprNode {
	return literalNode{value: node.Eval(&evalContext{fns: ctx.fns, useDecimal: ctx.useDecimal})}
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

func selectStaticIndex(ctx *evalContext, value any, index int, fallbackKey, path string) any {
	switch v := value.(type) {
	case []any:
		if index < len(v) {
			return v[index]
		}
		if ctx.strictVariables {
			panic(fmt.Errorf("undefined path %q", path))
		}
		return nil
	case []int:
		if index < len(v) {
			return v[index]
		}
		if ctx.strictVariables {
			panic(fmt.Errorf("undefined path %q", path))
		}
		return nil
	case []float64:
		if index < len(v) {
			return v[index]
		}
		if ctx.strictVariables {
			panic(fmt.Errorf("undefined path %q", path))
		}
		return nil
	case []string:
		if index < len(v) {
			return v[index]
		}
		if ctx.strictVariables {
			panic(fmt.Errorf("undefined path %q", path))
		}
		return nil
	case []bool:
		if index < len(v) {
			return v[index]
		}
		if ctx.strictVariables {
			panic(fmt.Errorf("undefined path %q", path))
		}
		return nil
	default:
		return ctx.selectValue(value, fallbackKey, path)
	}
}

func (ctx *evalContext) selectValue(value any, key, path string) any {
	val, ok := SelectValueOK(value, key)
	if ctx.strictVariables && !ok {
		panic(fmt.Errorf("undefined path %q", path))
	}
	return val
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

func evalJSONPath(ctx *evalContext, path string) Value {
	if val, ok := SelectPath(ctx.kv, path); ok {
		return ctx.newValue(path, val)
	}
	bs, err := json.Marshal(ctx.kv)
	if err != nil {
		panic(fmt.Errorf("parameter json marshal failed, %s", err))
	}
	result := gjson.ParseBytes(bs).Get(path)
	if ctx.strictVariables && !result.Exists() {
		panic(fmt.Errorf("undefined path %q", path))
	}
	return ctx.newValue(path, result)
}
