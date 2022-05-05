package goeval

import (
	"encoding/json"
	"errors"
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
	Json
	Struct
	Map
	Interface
	Error
)

type Value struct {
	name  string
	val   interface{}
	vType uint8
}

func NewValue(name string, v interface{}) Value {
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
	case []interface{}:
		res.vType = Array
	case gjson.Result:
		res.vType = Json
	case time.Time:
		res.vType = Time
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
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

func (v Value) Float() (float64, error) {
	switch v.vType {
	case Number:
		if f, ok := v.val.(float64); ok {
			return f, nil
		}
		fallthrough
	case String:
		f, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			return 0, err
		}
		return f, nil
	case Time:
		return float64(v.val.(time.Time).Unix()), nil
	case Json:
		return v.val.(gjson.Result).Float(), nil
	case Nil:
		if v.name == "" {
			return 0, errors.New("can not convert nil value to number")
		}
		return 0, fmt.Errorf("variable '%s' is nil, can not convert to number", v.name)
	case Error:
		return 0, fmt.Errorf("%v", v.val)
	}
	if v.name == "" {
		return 0, fmt.Errorf("unknown value type of %#v", v.val)
	}
	return 0, fmt.Errorf("unknow value type of variable '%s'", v.name)
}

func (v Value) Int() (int, error) {
	val, err := v.Float()
	if err != nil {
		return 0, err
	}
	return int(val), nil
}

func (v Value) Array() ([]interface{}, error) {
	switch v.vType {
	case Array:
		if val, ok := v.val.([]interface{}); ok {
			return val, nil
		}
		rv := reflect.ValueOf(v.val)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		val := make([]interface{}, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			val = append(val, rv.Index(i).Interface())
		}
		return val, nil
	case Json:
		array := v.val.(gjson.Result).Array()
		results := make([]interface{}, len(array))
		for i, r := range array {
			results[i] = r.Value()
		}
		return results, nil
	case String:
		s := v.String()
		if s == "" {
			return nil, nil
		}
		if s[0] != '[' || s[len(s)-1] != ']' {
			return nil, fmt.Errorf("value '%s' is not array", s)
		}
		res := make([]interface{}, 0)
		err := json.Unmarshal([]byte(s), &res)
		if err != nil {
			return nil, err
		}
		return res, nil
	case Error:
		return nil, fmt.Errorf("%v", v.val)
	}
	if v.name == "" {
		return nil, fmt.Errorf("unknown value [%v] type for parse array", v.val)
	}
	return nil, fmt.Errorf("unknow value [%v] type of variable '%s' for parse array", v.val, v.name)
}

func (v Value) String() string {
	switch v.vType {
	case String:
		return v.val.(string)
	case Time:
		return v.val.(time.Time).Format(time.RFC3339)
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
	case Error, Nil:
		return false
	default:
		s := strings.ToLower(fmt.Sprintf("%v", v.val))
		return s != "" && s != "false"
	}
}

func (v Value) Error() error {
	if v.vType == Error {
		return fmt.Errorf("%v", v.val)
	}
	return nil
}

func (v Value) NOT() Value {
	if v.vType == Error {
		return v
	}
	return Value{
		val:   !v.Boolean(),
		vType: Boolean,
	}
}

func (v Value) AND(v2 Value) Value {
	if v.vType == Error {
		return v
	}
	if !v.Boolean() {
		return v
	}
	return v2
}

func (v Value) OR(v2 Value) Value {
	if v.vType == Error {
		return v
	}
	if v.Boolean() {
		return v
	}
	return v2
}

func (v Value) EQ(v2 Value) Value {
	return Value{
		val:   v.String() == v2.String(),
		vType: Boolean,
	}
}

func (v Value) RE(v2 Value) Value {
	exp, err := compileRegexp(v2.String())
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	return Value{
		val:   exp.MatchString(v.String()),
		vType: Boolean,
	}
}

func (v Value) NRE(v2 Value) Value {
	exp, err := compileRegexp(v2.String())
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}

	return Value{
		val:   !exp.MatchString(v.String()),
		vType: Boolean,
	}
}

func (v Value) NEQ(v2 Value) Value {
	return v.EQ(v2).NOT()
}

func (v Value) GT(v2 Value) Value {
	left, err := v.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	right, err := v2.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	return Value{
		val:   left > right,
		vType: Boolean,
	}
}

func (v Value) GTE(v2 Value) Value {
	left, err := v.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	right, err := v2.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	return Value{
		val:   left >= right,
		vType: Boolean,
	}
}

func (v Value) LT(v2 Value) Value {
	left, err := v.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	right, err := v2.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	return Value{
		val:   left < right,
		vType: Boolean,
	}
}

func (v Value) LTE(v2 Value) Value {
	left, err := v.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	right, err := v2.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	return Value{
		val:   left <= right,
		vType: Boolean,
	}
}

func (v Value) MATCH(v2 Value) Value {
	return Value{
		val:   simpleMatch(v2.String(), v.String()),
		vType: Boolean,
	}
}

func (v Value) ADD(v2 Value) Value {
	f, err := v.Float()
	if err != nil {
		return Value{
			val:   v.String() + v2.String(),
			vType: String,
		}
	}
	f2, err := v2.Float()
	if err != nil {
		return Value{
			val:   v.String() + v2.String(),
			vType: String,
		}
	}
	return Value{
		val:   f + f2,
		vType: Number,
	}
}

func (v Value) SUB(v2 Value) Value {
	f, err := v.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	f2, err := v2.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	return Value{
		val:   f - f2,
		vType: Number,
	}
}

func (v Value) MULTI(v2 Value) Value {
	f, err := v.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	f2, err := v2.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	return Value{
		val:   f * f2,
		vType: Number,
	}
}

func (v Value) DIV(v2 Value) Value {
	f, err := v.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	f2, err := v2.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	return Value{
		val:   f / f2,
		vType: Number,
	}
}

func (v Value) MOD(v2 Value) Value {
	f, err := v.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	f2, err := v2.Float()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	return Value{
		val:   float64(int(f) % int(f2)),
		vType: Number,
	}
}

func (v Value) NC(v2 Value) Value {
	if v.vType == Nil ||
		(v.vType == Boolean && !v.Boolean()) ||
		(v.vType == String && len(v.String()) == 0) ||
		(v.vType == Number && v.val.(float64) == 0) ||
		(v.vType == Array && len(v.val.([]interface{})) == 0) {
		return v2
	}
	return v
}

func (v Value) IN(v2 Value) Value {
	array, err := v2.Array()
	if err != nil {
		return Value{
			val:   err.Error(),
			vType: Error,
		}
	}
	for _, item := range array {
		if v.EQ(NewValue("", item)).Boolean() {
			return trueValue
		}
	}
	return falseValue
}

func (v Value) TERNARY(v2 Value, v3 Value) Value {
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
