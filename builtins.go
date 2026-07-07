package goeval

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

func defaultOptions() []Option {
	opts := []Option{
		withFoldableFunc("date", builtinDate),
		withFoldableFunc("strlen", builtinStrlen),
		withFoldableFunc("duration", builtinDuration),
		withFoldableFunc("float", builtinFloat),
		withFoldableFunc("decimal", builtinDecimal),
		WithFunc("now", builtinNow),
	}
	return append(opts, builtinOptions()...)
}

func builtinOptions() []Option {
	return []Option{
		withFoldableFunc("contains", builtinContains),
		withFoldableFunc("startsWith", builtinStartsWith),
		withFoldableFunc("endsWith", builtinEndsWith),
		withFoldableFunc("lower", builtinLower),
		withFoldableFunc("upper", builtinUpper),
		withFoldableFunc("trim", builtinTrim),
		withFoldableFunc("replace", builtinReplace),
		withFoldableFunc("len", builtinLen),
		withFoldableFunc("min", builtinMin),
		withFoldableFunc("max", builtinMax),
		withFoldableFunc("abs", builtinAbs),
		withFoldableFunc("round", builtinRound),
		withFoldableFunc("int", builtinInt),
		withFoldableFunc("string", builtinString),
		withFoldableFunc("bool", builtinBool),
		withFoldableFunc("exists", builtinExists),
		withFoldableFunc("empty", builtinEmpty),
		withFoldableFunc("notEmpty", builtinNotEmpty),
		withFoldableFunc("coalesce", builtinCoalesce),
		withFoldableFunc("default", builtinDefault),
		withFoldableFunc("matches", builtinMatches),
		withFoldableFunc("regex", builtinMatches),
		withFoldableFunc("any", builtinAny),
		withFoldableFunc("all", builtinAll),
	}
}

func builtinNow(args ...any) (any, error) {
	return time.Now(), nil
}

func builtinDate(args ...any) (any, error) {
	if len(args) < 1 || len(args) > 3 {
		return nil, fmt.Errorf("date() expects 1 to 3 arguments")
	}
	loc := time.Local
	if len(args) >= 2 {
		parsedLoc, err := parseDateLocation(args[1])
		if err != nil {
			return nil, err
		}
		loc = parsedLoc
	}
	if len(args) == 3 {
		format, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("date() expects a string format argument")
		}
		return parseDateValue(args[0], loc, []string{format})
	}
	return parseDateValue(args[0], loc, defaultDateFormats())
}

func parseDateLocation(arg any) (*time.Location, error) {
	name, ok := arg.(string)
	if !ok {
		return nil, fmt.Errorf("date() expects a string location argument")
	}
	switch name {
	case "Local", "":
		return time.Local, nil
	case "UTC":
		return time.UTC, nil
	default:
		loc, err := time.LoadLocation(name)
		if err != nil {
			return nil, fmt.Errorf("date() invalid location %q: %w", name, err)
		}
		return loc, nil
	}
}

func parseDateValue(arg any, loc *time.Location, formats []string) (any, error) {
	if s, ok := arg.(string); ok {
		for _, format := range formats {
			ret, err := time.ParseInLocation(format, s, loc)
			if err == nil {
				return ret, nil
			}
		}
	} else {
		value := NewValue("", arg)
		if value.vType == Number {
			return time.Unix(int64(value.Float()), 0).In(loc), nil
		}
	}
	return nil, fmt.Errorf("date() expects a string or a number argument")
}

func defaultDateFormats() []string {
	return []string{
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
	}
}

func builtinStrlen(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("strlen() expects exactly one string argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("strlen() expects exactly one string argument")
	}
	return len(s), nil
}

func builtinDuration(args ...any) (any, error) {
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
	default:
		value := NewValue("", arg)
		if value.vType != Number {
			err = fmt.Errorf("duration() expects a string or a number argument")
			break
		}
		d = time.Duration(value.Float())
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
}

func builtinFloat(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("float() expects exactly one argument")
	}
	switch arg := args[0].(type) {
	case string:
		return strconv.ParseFloat(arg, 64)
	case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return arg, nil
	default:
		value := NewValue("", arg)
		if value.vType == Number {
			return value.Float(), nil
		}
		return nil, fmt.Errorf("float() expects a string or a number argument")
	}
}

func builtinDecimal(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("decimal() expects exactly one argument")
	}
	switch arg := args[0].(type) {
	case decimal.Decimal:
		return arg, nil
	case string:
		d, err := decimal.NewFromString(arg)
		if err != nil {
			return nil, fmt.Errorf("decimal() expects a valid decimal string: %w", err)
		}
		return d, nil
	case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return convertNumberToDecimal(arg), nil
	default:
		value := NewValue("", arg)
		if value.vType == Number {
			return value.Decimal(), nil
		}
		return nil, fmt.Errorf("decimal() expects a string or a number argument")
	}
}

func expectArgCount(name string, args []any, count int) error {
	if len(args) != count {
		return fmt.Errorf("%s() expects exactly %d argument(s)", name, count)
	}
	return nil
}

func expectMinArgCount(name string, args []any, count int) error {
	if len(args) < count {
		return fmt.Errorf("%s() expects at least %d argument(s)", name, count)
	}
	return nil
}

func builtinContains(args ...any) (any, error) {
	if err := expectArgCount("contains", args, 2); err != nil {
		return nil, err
	}
	return strings.Contains(NewValue("", args[0]).String(), NewValue("", args[1]).String()), nil
}

func builtinStartsWith(args ...any) (any, error) {
	if err := expectArgCount("startsWith", args, 2); err != nil {
		return nil, err
	}
	return strings.HasPrefix(NewValue("", args[0]).String(), NewValue("", args[1]).String()), nil
}

func builtinEndsWith(args ...any) (any, error) {
	if err := expectArgCount("endsWith", args, 2); err != nil {
		return nil, err
	}
	return strings.HasSuffix(NewValue("", args[0]).String(), NewValue("", args[1]).String()), nil
}

func builtinLower(args ...any) (any, error) {
	if err := expectArgCount("lower", args, 1); err != nil {
		return nil, err
	}
	return strings.ToLower(NewValue("", args[0]).String()), nil
}

func builtinUpper(args ...any) (any, error) {
	if err := expectArgCount("upper", args, 1); err != nil {
		return nil, err
	}
	return strings.ToUpper(NewValue("", args[0]).String()), nil
}

func builtinTrim(args ...any) (any, error) {
	if err := expectArgCount("trim", args, 1); err != nil {
		return nil, err
	}
	return strings.TrimSpace(NewValue("", args[0]).String()), nil
}

func builtinReplace(args ...any) (any, error) {
	if err := expectArgCount("replace", args, 3); err != nil {
		return nil, err
	}
	return strings.ReplaceAll(
		NewValue("", args[0]).String(),
		NewValue("", args[1]).String(),
		NewValue("", args[2]).String(),
	), nil
}

func builtinLen(args ...any) (any, error) {
	if err := expectArgCount("len", args, 1); err != nil {
		return nil, err
	}
	return valueLen(args[0])
}

func builtinMin(args ...any) (any, error) {
	if err := expectMinArgCount("min", args, 1); err != nil {
		return nil, err
	}
	minVal := NewValue("", args[0]).Float()
	for i := 1; i < len(args); i++ {
		minVal = math.Min(minVal, NewValue("", args[i]).Float())
	}
	return minVal, nil
}

func builtinMax(args ...any) (any, error) {
	if err := expectMinArgCount("max", args, 1); err != nil {
		return nil, err
	}
	maxVal := NewValue("", args[0]).Float()
	for i := 1; i < len(args); i++ {
		maxVal = math.Max(maxVal, NewValue("", args[i]).Float())
	}
	return maxVal, nil
}

func builtinAbs(args ...any) (any, error) {
	if err := expectArgCount("abs", args, 1); err != nil {
		return nil, err
	}
	return math.Abs(NewValue("", args[0]).Float()), nil
}

func builtinRound(args ...any) (any, error) {
	if err := expectArgCount("round", args, 1); err != nil {
		return nil, err
	}
	return math.Round(NewValue("", args[0]).Float()), nil
}

func builtinInt(args ...any) (any, error) {
	if err := expectArgCount("int", args, 1); err != nil {
		return nil, err
	}
	return NewValue("", args[0]).Int(), nil
}

func builtinString(args ...any) (any, error) {
	if err := expectArgCount("string", args, 1); err != nil {
		return nil, err
	}
	return NewValue("", args[0]).String(), nil
}

func builtinBool(args ...any) (any, error) {
	if err := expectArgCount("bool", args, 1); err != nil {
		return nil, err
	}
	return NewValue("", args[0]).Boolean(), nil
}

func builtinExists(args ...any) (any, error) {
	if err := expectArgCount("exists", args, 1); err != nil {
		return nil, err
	}
	return args[0] != nil, nil
}

func builtinEmpty(args ...any) (any, error) {
	if err := expectArgCount("empty", args, 1); err != nil {
		return nil, err
	}
	return isEmptyValue(args[0]), nil
}

func builtinNotEmpty(args ...any) (any, error) {
	if err := expectArgCount("notEmpty", args, 1); err != nil {
		return nil, err
	}
	return !isEmptyValue(args[0]), nil
}

func builtinCoalesce(args ...any) (any, error) {
	if err := expectMinArgCount("coalesce", args, 1); err != nil {
		return nil, err
	}
	for _, arg := range args {
		value := NewValue("", arg)
		if !isCoalesceEmpty(value) {
			return arg, nil
		}
	}
	return nil, nil
}

func builtinDefault(args ...any) (any, error) {
	if err := expectArgCount("default", args, 2); err != nil {
		return nil, err
	}
	if isCoalesceEmpty(NewValue("", args[0])) {
		return args[1], nil
	}
	return args[0], nil
}

func builtinMatches(args ...any) (any, error) {
	if err := expectArgCount("matches", args, 2); err != nil {
		return nil, err
	}
	exp, err := compileRegexp(NewValue("", args[1]).String())
	if err != nil {
		return nil, err
	}
	return exp.MatchString(NewValue("", args[0]).String()), nil
}

func builtinAny(args ...any) (any, error) {
	if err := expectMinArgCount("any", args, 1); err != nil {
		return nil, err
	}
	for _, arg := range args {
		if NewValue("", arg).Boolean() {
			return true, nil
		}
	}
	return false, nil
}

func builtinAll(args ...any) (any, error) {
	if err := expectMinArgCount("all", args, 1); err != nil {
		return nil, err
	}
	for _, arg := range args {
		if !NewValue("", arg).Boolean() {
			return false, nil
		}
	}
	return true, nil
}

func valueLen(value any) (int, error) {
	if value == nil {
		return 0, nil
	}
	switch v := value.(type) {
	case string:
		return len(v), nil
	case []any:
		return len(v), nil
	}
	rv := reflect.ValueOf(value)
	for rv.IsValid() && rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return 0, nil
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return 0, nil
	}
	switch rv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len(), nil
	default:
		return 0, fmt.Errorf("len() expects a string, array, slice, map, or channel argument")
	}
}

func isEmptyValue(value any) bool {
	if value == nil {
		return true
	}
	v := NewValue("", value)
	switch v.vType {
	case Nil:
		return true
	case Boolean:
		return !v.Boolean()
	case String:
		return v.String() == ""
	case Number:
		return v.Float() == 0
	case Array:
		length, err := valueLen(value)
		return err == nil && length == 0
	}
	rv := reflect.ValueOf(value)
	for rv.IsValid() && rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return true
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return true
	}
	switch rv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len() == 0
	default:
		return rv.IsZero()
	}
}
