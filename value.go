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
	case []byte:
		res.val = string(r)
		res.vType = String
	case string:
		res.vType = String
	case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		res.vType = Number
		res.val = convertNumberToFloat(v)
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
			return NewValue(name, rv.Elem().Interface())
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

func (v Value) Float() float64 {
	switch v.vType {
	case Boolean:
		if v.Boolean() {
			return 1
		}
		return 0
	case Number:
		return v.val.(float64)
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

func (v Value) Int() int {
	return int(v.Float())
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
	return Value{
		val:   v.Float() > v2.Float(),
		vType: Boolean,
	}
}

func (v Value) Gte(v2 Value) Value {
	return Value{
		val:   v.Float() >= v2.Float(),
		vType: Boolean,
	}
}

func (v Value) Lt(v2 Value) Value {
	return Value{
		val:   v.Float() < v2.Float(),
		vType: Boolean,
	}
}

func (v Value) Lte(v2 Value) Value {
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
			vType: Number,
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
			vType: Number,
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
	return Value{
		val:   float64(v.Int() % v2.Int()),
		vType: Number,
	}
}

func (v Value) Nc(v2 Value) Value {
	if v.vType == Nil ||
		(v.vType == Boolean && !v.Boolean()) ||
		(v.vType == String && len(v.String()) == 0) ||
		(v.vType == Number && v.val.(float64) == 0) ||
		(v.vType == Array && len(v.val.([]any)) == 0) {
		return v2
	}
	return v
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
