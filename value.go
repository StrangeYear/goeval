package goeval

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

var (
	regexpCache sync.Map
	nilValue    Value
	trueValue   Value
	falseValue  Value
)

func init() {
	nilValue = NewValue("", nil)
	trueValue = NewValue("", true)
	falseValue = NewValue("", false)
}

type ValueType uint8

const (
	Nil = iota
	Boolean
	Number
	String
	Array
	Time
	Duration
	Json
	Struct
	Map
	Interface
)

type Value struct {
	name  string
	val   any
	vType uint8
}

func (v Value) Raw() any {
	return v.val
}

func (v Value) Any() any {
	switch v.vType {
	case Nil:
		return nil
	case Boolean:
		return v.Boolean()
	case Number:
		return json.Number(v.Decimal().String())
	case String:
		return v.String()
	case Array:
		return v.Array()
	default:
		return v.val
	}
}

func (v Value) Type() ValueType {
	return ValueType(v.vType)
}

func convertNumberToFloat(num any) float64 {
	switch n := num.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int8:
		return float64(n)
	case int16:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case uint:
		return float64(n)
	case uint8:
		return float64(n)
	case uint16:
		return float64(n)
	case uint32:
		return float64(n)
	case uint64:
		return float64(n)
	}
	return 0
}

func NewValue(name string, v any) Value {
	return newValue(name, v, false)
}

func NewDecimalValue(name string, v any) Value {
	return newValue(name, v, true)
}

func newValue(name string, v any, useDecimal bool) Value {
	res := Value{
		name:  name,
		val:   v,
		vType: Nil,
	}
	if v == nil {
		return res
	}
	switch r := v.(type) {
	case bool:
		res.vType = Boolean
	case decimal.Decimal:
		res.vType = Number
	case []byte:
		res.val = string(r)
		res.vType = String
	case string:
		res.vType = String
	case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		res.vType = Number
		if useDecimal {
			res.val = convertNumberToDecimal(v)
		} else {
			res.val = convertNumberToFloat(v)
		}
	case []any:
		res.vType = Array
	case gjson.Result:
		res.vType = Json
	case time.Time:
		res.vType = Time
	case time.Duration:
		res.vType = Duration
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return res
			}
			elem := rv.Elem()
			switch elem.Kind() {
			case reflect.Struct:
				res.vType = Struct
			case reflect.Map:
				res.vType = Map
			case reflect.Array, reflect.Slice:
				res.vType = Array
			default:
				return newValue(name, elem.Interface(), useDecimal)
			}
			return res
		}

		switch rv.Kind() {
		case reflect.Struct:
			res.vType = Struct
		case reflect.Map:
			res.vType = Map
		case reflect.Array, reflect.Slice:
			res.vType = Array
		default:
			res.vType = Interface
		}
	}
	return res
}

func convertNumberToDecimal(num any) decimal.Decimal {
	switch n := num.(type) {
	case decimal.Decimal:
		return n
	case float64:
		return decimal.NewFromFloat(n)
	case float32:
		return decimal.NewFromFloat32(n)
	case int:
		return decimal.NewFromInt(int64(n))
	case int8:
		return decimal.NewFromInt(int64(n))
	case int16:
		return decimal.NewFromInt(int64(n))
	case int32:
		return decimal.NewFromInt(int64(n))
	case int64:
		return decimal.NewFromInt(n)
	case uint:
		return decimal.NewFromUint64(uint64(n))
	case uint8:
		return decimal.NewFromUint64(uint64(n))
	case uint16:
		return decimal.NewFromUint64(uint64(n))
	case uint32:
		return decimal.NewFromUint64(uint64(n))
	case uint64:
		return decimal.NewFromUint64(n)
	}
	return decimal.Zero
}

func newNumberLiteral(name, token string, useDecimal bool) Value {
	if useDecimal {
		d, err := decimal.NewFromString(token)
		if err != nil {
			panic(err)
		}
		return NewDecimalValue(name, d)
	}
	f, err := strconv.ParseFloat(token, 64)
	if err != nil {
		panic(err)
	}
	return NewValue(name, f)
}

func (v Value) Float() float64 {
	switch v.vType {
	case Boolean:
		if v.Boolean() {
			return 1
		}
		return 0
	case Number:
		switch n := v.val.(type) {
		case float64:
			return n
		case decimal.Decimal:
			f, _ := n.Float64()
			return f
		default:
			return convertNumberToFloat(n)
		}
	case String:
		f, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			panic(fmt.Errorf("convert value to float error: %s, name: %s", err, v.name))
		}
		return f
	case Time:
		return float64(v.val.(time.Time).Unix())
	case Duration:
		return float64(v.val.(time.Duration))
	case Json:
		return v.val.(gjson.Result).Float()
	case Nil:
		return 0
	default:
		panic(fmt.Errorf("invalid value type of variable '%s' for parse float", v.name))
	}
}

func safeValueFloat(v Value) (val float64, err error) {
	defer func() {
		if r := recover(); r != nil {
			val = 0
			err = fmt.Errorf("%v", r)
		}
	}()
	return v.Float(), nil
}

func (v Value) Decimal() decimal.Decimal {
	switch v.vType {
	case Boolean:
		if v.Boolean() {
			return decimal.NewFromInt(1)
		}
		return decimal.Zero
	case Number:
		switch n := v.val.(type) {
		case decimal.Decimal:
			return n
		default:
			return convertNumberToDecimal(n)
		}
	case String:
		d, err := decimal.NewFromString(v.String())
		if err != nil {
			panic(fmt.Errorf("convert value to decimal error: %s, name: %s", err, v.name))
		}
		return d
	case Time:
		return decimal.NewFromInt(v.val.(time.Time).Unix())
	case Duration:
		return decimal.NewFromInt(int64(v.val.(time.Duration)))
	case Json:
		r := v.val.(gjson.Result)
		if r.Raw != "" {
			d, err := decimal.NewFromString(r.Raw)
			if err == nil {
				return d
			}
		}
		d, err := decimal.NewFromString(r.String())
		if err != nil {
			panic(fmt.Errorf("convert value to decimal error: %s, name: %s", err, v.name))
		}
		return d
	case Nil:
		return decimal.Zero
	default:
		panic(fmt.Errorf("invalid value type of variable '%s' for parse decimal", v.name))
	}
}

func (v Value) Int() int {
	return int(v.Float())
}

func safeValueInt(v Value) (val int, err error) {
	defer func() {
		if r := recover(); r != nil {
			val = 0
			err = fmt.Errorf("%v", r)
		}
	}()
	return v.Int(), nil
}

func (v Value) Array() []any {
	switch v.vType {
	case Array:
		if val, ok := v.val.([]any); ok {
			return val
		}
		rv := reflect.ValueOf(v.val)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		val := make([]any, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			val = append(val, rv.Index(i).Interface())
		}
		return val
	case Json:
		array := v.val.(gjson.Result).Array()
		results := make([]any, len(array))
		for i, r := range array {
			results[i] = r.Value()
		}
		return results
	case String:
		s := v.String()
		if s == "" {
			return nil
		}
		if s[0] != '[' || s[len(s)-1] != ']' {
			panic(fmt.Errorf("value '%s' is not array", s))
		}
		res := make([]any, 0)
		err := json.Unmarshal([]byte(s), &res)
		if err != nil {
			panic(fmt.Errorf("parse array error: %s, name: %s", err, v.name))
		}
		return res
	default:
		panic(fmt.Errorf("invalid value [%v] type of variable '%s' for parse array", v.val, v.name))
	}
}

func (v Value) String() string {
	switch v.vType {
	case String:
		return v.val.(string)
	case Time:
		return v.val.(time.Time).Format(time.RFC3339)
	case Duration:
		return v.val.(time.Duration).String()
	case Json:
		return v.val.(gjson.Result).String()
	default:
		return fmt.Sprintf("%v", v.val)
	}
}

func (v Value) Boolean() bool {
	switch v.vType {
	case Boolean:
		return v.val.(bool)
	case Json:
		return v.val.(gjson.Result).Bool()
	case Nil:
		return false
	default:
		s := strings.ToLower(fmt.Sprintf("%v", v.val))
		return s == "true"
	}
}

func (v Value) Not() Value {
	return Value{
		val:   !v.Boolean(),
		vType: Boolean,
	}
}

func (v Value) And(v2 Value) Value {
	if !v.Boolean() {
		return v
	}
	return v2
}

func (v Value) Or(v2 Value) Value {
	if v.Boolean() {
		return v
	}
	return v2
}

func (v Value) Eq(v2 Value) Value {
	if useDecimalMath(v, v2) {
		left, leftOK := decimalValue(v)
		right, rightOK := decimalValue(v2)
		if leftOK && rightOK {
			return Value{
				val:   left.Equal(right),
				vType: Boolean,
			}
		}
	}
	return Value{
		val:   v.String() == v2.String(),
		vType: Boolean,
	}
}

func (v Value) Re(v2 Value) Value {
	exp, err := compileRegexp(v2.String())
	if err != nil {
		panic(fmt.Errorf("compile regexp error: %s, name: %s", err, v.name))
	}
	return Value{
		val:   exp.MatchString(v.String()),
		vType: Boolean,
	}
}

func (v Value) Nre(v2 Value) Value {
	exp, err := compileRegexp(v2.String())
	if err != nil {
		panic(fmt.Errorf("compile regexp error: %s, name: %s", err, v.name))
	}

	return Value{
		val:   !exp.MatchString(v.String()),
		vType: Boolean,
	}
}

func (v Value) Neq(v2 Value) Value {
	return v.Eq(v2).Not()
}

func (v Value) Gt(v2 Value) Value {
	if useDecimalMath(v, v2) {
		return Value{
			val:   v.Decimal().GreaterThan(v2.Decimal()),
			vType: Boolean,
		}
	}
	return Value{
		val:   v.Float() > v2.Float(),
		vType: Boolean,
	}
}

func (v Value) Gte(v2 Value) Value {
	if useDecimalMath(v, v2) {
		return Value{
			val:   v.Decimal().GreaterThanOrEqual(v2.Decimal()),
			vType: Boolean,
		}
	}
	return Value{
		val:   v.Float() >= v2.Float(),
		vType: Boolean,
	}
}

func (v Value) Lt(v2 Value) Value {
	if useDecimalMath(v, v2) {
		return Value{
			val:   v.Decimal().LessThan(v2.Decimal()),
			vType: Boolean,
		}
	}
	return Value{
		val:   v.Float() < v2.Float(),
		vType: Boolean,
	}
}

func (v Value) Lte(v2 Value) Value {
	if useDecimalMath(v, v2) {
		return Value{
			val:   v.Decimal().LessThanOrEqual(v2.Decimal()),
			vType: Boolean,
		}
	}
	return Value{
		val:   v.Float() <= v2.Float(),
		vType: Boolean,
	}
}

func (v Value) Match(v2 Value) Value {
	return Value{
		val:   simpleMatch(v2.String(), v.String()),
		vType: Boolean,
	}
}

func (v Value) Add(v2 Value) Value {
	switch {
	case v.vType == Time && v2.vType == Duration:
		return Value{
			val:   v.val.(time.Time).Add(v2.val.(time.Duration)),
			vType: Time,
		}
	case v.vType == Duration && v2.vType == Time:
		return Value{
			val:   v2.val.(time.Time).Add(v.val.(time.Duration)),
			vType: Time,
		}
	case v.vType == Duration && v2.vType == Duration:
		return Value{
			val:   v.val.(time.Duration) + v2.val.(time.Duration),
			vType: Duration,
		}
	case (v.vType == Duration && v2.vType == Number) || (v.vType == Number && v2.vType == Duration):
		return Value{
			val:   time.Duration(v.Float()) + time.Duration(v2.Float()),
			vType: Duration,
		}
	case v.vType == Number && v2.vType == Number:
		if useDecimalMath(v, v2) {
			return Value{
				val:   v.Decimal().Add(v2.Decimal()),
				vType: Number,
			}
		}
		f := v.val.(float64)
		f2 := v2.val.(float64)
		return Value{
			val:   f + f2,
			vType: Number,
		}
	case v.vType == Array && v2.vType == Array:
		return Value{
			val:   append(v.Array(), v2.Array()...),
			vType: Array,
		}
	case v.vType == Array:
		return Value{
			val:   append(v.Array(), v2.val),
			vType: Array,
		}
	case v2.vType == Array:
		return Value{
			val:   append(v2.Array(), v.val),
			vType: Array,
		}
	default:
		return Value{
			val:   v.String() + v2.String(),
			vType: String,
		}
	}
}

func (v Value) Sub(v2 Value) Value {
	switch {
	case v.vType == Time && v2.vType == Duration:
		return Value{
			val:   v.val.(time.Time).Add(-v2.val.(time.Duration)),
			vType: Time,
		}
	case v.vType == Duration && v2.vType == Time:
		panic("time - duration is not supported")
	case v.vType == Duration && v2.vType == Duration:
		return Value{
			val:   v.val.(time.Duration) - v2.val.(time.Duration),
			vType: Duration,
		}
	case (v.vType == Duration && v2.vType == Number) || (v.vType == Number && v2.vType == Duration):
		return Value{
			val:   time.Duration(v.Float()) - time.Duration(v2.Float()),
			vType: Duration,
		}
	case v.vType == Time && v2.vType == Time:
		return Value{
			val:   v.val.(time.Time).Sub(v2.val.(time.Time)),
			vType: Duration,
		}
	case v.vType == Number && v2.vType == Number:
		if useDecimalMath(v, v2) {
			return Value{
				val:   v.Decimal().Sub(v2.Decimal()),
				vType: Number,
			}
		}
		f := v.val.(float64)
		f2 := v2.val.(float64)
		return Value{
			val:   f - f2,
			vType: Number,
		}
	case v.vType == Array && v2.vType == Array:
		res := make([]any, 0, len(v.Array()))
		for _, item := range v.Array() {
			if !v2.In(NewValue("", item)).Boolean() {
				res = append(res, item)
			}
		}
		return Value{
			val:   res,
			vType: Array,
		}
	case v.vType == Array:
		res := make([]any, 0, len(v.Array()))
		for _, item := range v.Array() {
			if !v2.Eq(NewValue("", item)).Boolean() {
				res = append(res, item)
			}
		}
		return Value{
			val:   res,
			vType: Array,
		}
	case v2.vType == Array:
		res := make([]any, 0, len(v2.Array()))
		for _, item := range v2.Array() {
			if !v.Eq(NewValue("", item)).Boolean() {
				res = append(res, item)
			}
		}
		return Value{
			val:   res,
			vType: Array,
		}
	default:
		return Value{
			val:   strings.ReplaceAll(v.String(), v2.String(), ""),
			vType: String,
		}
	}
}

func (v Value) Multi(v2 Value) Value {
	switch {
	case v.vType == Duration && v2.vType == Duration:
		return Value{
			val:   v.val.(time.Duration) * v2.val.(time.Duration),
			vType: Duration,
		}
	case (v.vType == Duration && v2.vType == Number) || (v.vType == Number && v2.vType == Duration):
		return Value{
			val:   time.Duration(v.Float()) * time.Duration(v2.Float()),
			vType: Duration,
		}
	case v.vType == Number && v2.vType == Number:
		if useDecimalMath(v, v2) {
			return Value{
				val:   v.Decimal().Mul(v2.Decimal()),
				vType: Number,
			}
		}
		f := v.val.(float64)
		f2 := v2.val.(float64)
		return Value{
			val:   f * f2,
			vType: Number,
		}
	default:
		panic(fmt.Errorf("invalid value [%v] type of variable '%s' for multiply", v.val, v.name))
	}
}

func (v Value) Div(v2 Value) Value {
	switch {
	case v.vType == Duration && v2.vType == Duration:
		return Value{
			val:   v.val.(time.Duration) / v2.val.(time.Duration),
			vType: Duration,
		}
	case (v.vType == Duration && v2.vType == Number) || (v.vType == Number && v2.vType == Duration):
		return Value{
			val:   time.Duration(v.Float()) / time.Duration(v2.Float()),
			vType: Duration,
		}
	case v.vType == Number && v2.vType == Number:
		if useDecimalMath(v, v2) {
			return Value{
				val:   v.Decimal().Div(v2.Decimal()),
				vType: Number,
			}
		}
		f := v.val.(float64)
		f2 := v2.val.(float64)
		return Value{
			val:   f / f2,
			vType: Number,
		}
	default:
		panic(fmt.Errorf("invalid value [%v] type of variable '%s' for divide", v.val, v.name))
	}
}

func (v Value) Mod(v2 Value) Value {
	if useDecimalMath(v, v2) {
		return Value{
			val:   v.Decimal().Mod(v2.Decimal()),
			vType: Number,
		}
	}
	return Value{
		val:   float64(v.Int() % v2.Int()),
		vType: Number,
	}
}

func (v Value) Nc(v2 Value) Value {
	if isCoalesceEmpty(v) {
		return v2
	}
	return v
}

func isCoalesceEmpty(v Value) bool {
	switch v.vType {
	case Nil:
		return true
	case Boolean:
		return !v.Boolean()
	case String:
		return v.String() == ""
	case Number:
		return isNumberZero(v)
	case Array:
		length, err := valueLen(v.val)
		return err == nil && length == 0
	default:
		return false
	}
}

func useDecimalMath(v, v2 Value) bool {
	return isDecimalNumber(v) || isDecimalNumber(v2)
}

func isDecimalNumber(v Value) bool {
	if v.vType != Number {
		return false
	}
	_, ok := v.val.(decimal.Decimal)
	return ok
}

func decimalValue(v Value) (val decimal.Decimal, ok bool) {
	switch v.vType {
	case Number, String, Json:
	default:
		return decimal.Decimal{}, false
	}
	defer func() {
		if recover() != nil {
			val = decimal.Decimal{}
			ok = false
		}
	}()
	return v.Decimal(), true
}

func isNumberZero(v Value) bool {
	if v.vType != Number {
		return false
	}
	if d, ok := v.val.(decimal.Decimal); ok {
		return d.IsZero()
	}
	return v.Float() == 0
}

func (v Value) In(v2 Value) Value {
	switch v2.vType {
	case String:
		return Value{
			val:   strings.Contains(v2.String(), v.String()),
			vType: Boolean,
		}
	case Array:
		for _, item := range v2.Array() {
			if v.Eq(NewValue("", item)).Boolean() {
				return trueValue
			}
		}
		return falseValue
	default:
		panic(fmt.Errorf("invalid value [%v] type of variable '%s' for in", v2.val, v2.name))
	}
}

func (v Value) Ternary(v2 Value, v3 Value) Value {
	if v.Boolean() {
		return v2
	}
	return v3
}

func compileRegexp(s string) (*regexp.Regexp, error) {
	v, ok := regexpCache.Load(s)
	if ok {
		return v.(*regexp.Regexp), nil
	}
	exp, err := regexp.Compile(s)
	if err != nil {
		return nil, err
	}
	regexpCache.Store(s, exp)
	return exp, nil
}

func simpleMatch(pattern, s string) bool {
	i, j, star, match := 0, 0, -1, 0
	for i < len(s) {
		if j < len(pattern) && (s[i] == pattern[j] || pattern[j] == '?') {
			i++
			j++
		} else if j < len(pattern) && pattern[j] == '*' {
			match, star = i, j
			j++
		} else if star != -1 {
			j = star + 1
			match++
			i = match
		} else {
			return false
		}
	}
	for ; j < len(pattern); j++ {
		if pattern[j] != '*' {
			return false
		}
	}
	return true
}
