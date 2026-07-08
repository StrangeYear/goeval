// Copyright (c) 2017, Paessler AG <support@paessler.com>
// All rights reserved.

package goeval

import (
	"reflect"
	"strconv"
	"strings"
)

func SelectPath(value any, path string) (any, bool) {
	if value == nil {
		return nil, false
	}
	cur := value
	for {
		dot := strings.IndexByte(path, '.')
		part := path
		if dot >= 0 {
			part = path[:dot]
		}
		if part == "" {
			return nil, false
		}
		next, ok := selectPathPart(cur, part)
		if !ok {
			return nil, false
		}
		if dot < 0 {
			return next, true
		}
		cur = next
		path = path[dot+1:]
	}
}

func selectPathPart(value any, key string) (any, bool) {
	switch v := value.(type) {
	case map[string]any:
		return selectStringMapPath(v, key)
	case map[string]string:
		return selectStringMapPath(v, key)
	case map[string]int:
		return selectStringMapPath(v, key)
	case map[string]float64:
		return selectStringMapPath(v, key)
	case map[string]bool:
		return selectStringMapPath(v, key)
	case []any:
		return selectSlicePath(v, key)
	case []int:
		return selectSlicePath(v, key)
	case []float64:
		return selectSlicePath(v, key)
	case []string:
		return selectSlicePath(v, key)
	case []bool:
		return selectSlicePath(v, key)
	}
	return nil, false
}

func parsePathIndex(key string) (int, bool) {
	i, err := strconv.Atoi(key)
	return i, err == nil && i >= 0
}

func SelectValue(value any, key string) any {
	val, _ := SelectValueOK(value, key)
	return val
}

func SelectValueOK(value any, key string) (any, bool) {
	if value == nil {
		return nil, false
	}

	switch v := value.(type) {
	case map[string]any:
		return selectStringMap(v, key)
	case map[string]string:
		return selectStringMap(v, key)
	case map[string]int:
		return selectStringMap(v, key)
	case map[string]float64:
		return selectStringMap(v, key)
	case map[string]bool:
		return selectStringMap(v, key)
	case []any:
		return selectSlice(v, key)
	case []int:
		return selectSlice(v, key)
	case []float64:
		return selectSlice(v, key)
	case []string:
		return selectSlice(v, key)
	case []bool:
		return selectSlice(v, key)
	}

	vv := reflect.ValueOf(value)
	vvElem := resolvePotentialPointer(vv)
	if !vvElem.IsValid() {
		return nil, false
	}

	switch vvElem.Kind() {
	case reflect.Map:
		mapKey, ok := reflectConvertTo(vvElem.Type().Key().Kind(), key)
		if !ok {
			return nil, false
		}

		vvElem = vvElem.MapIndex(reflect.ValueOf(mapKey))
		vvElem = resolvePotentialPointer(vvElem)

		if vvElem.IsValid() && vvElem.CanInterface() {
			return vvElem.Interface(), true
		}
	case reflect.Slice, reflect.Array, reflect.String:
		if i := NewValue("", key).Int(); i >= 0 && vvElem.Len() > i {
			vvElem = resolvePotentialPointer(vvElem.Index(i))
			if vvElem.IsValid() && vvElem.CanInterface() {
				return vvElem.Interface(), true
			}
		}
	case reflect.Struct:
		field := vvElem.FieldByName(key)
		if field.IsValid() && field.CanInterface() {
			return field.Interface(), true
		}
		method := vv.MethodByName(key)
		if !method.IsValid() {
			method = vvElem.MethodByName(key)
		}
		if method.IsValid() {
			return method.Interface(), true
		}
	default:
		return nil, false
	}

	return nil, false
}

func selectStringMap[V any](values map[string]V, key string) (any, bool) {
	val, ok := values[key]
	return val, ok
}

func selectStringMapPath[V any](values map[string]V, key string) (any, bool) {
	val, ok := values[key]
	return val, ok
}

func selectSlice[V any](values []V, key string) (any, bool) {
	if i := parseIndex(key); i >= 0 && len(values) > i {
		return values[i], true
	}
	return nil, false
}

func selectSlicePath[V any](values []V, key string) (any, bool) {
	i, ok := parsePathIndex(key)
	if ok && len(values) > i {
		return values[i], true
	}
	return nil, false
}

func parseIndex(key string) int {
	i, err := strconv.Atoi(key)
	if err == nil {
		return i
	}
	return NewValue("", key).Int()
}

func resolvePotentialPointer(value reflect.Value) reflect.Value {
	for value.IsValid() && value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

func reflectConvertTo(k reflect.Kind, value any) (any, bool) {
	switch k {
	case reflect.String:
		return NewValue("", value).String(), true
	case reflect.Interface:
		return value, true
	case reflect.Int:
		return NewValue("", value).Int(), true
	case reflect.Float64:
		return NewValue("", value).Float(), true
	case reflect.Float32:
		return float32(NewValue("", value).Float()), true
	case reflect.Bool:
		return NewValue("", value).Boolean(), true
	case reflect.Int8:
		return int8(NewValue("", value).Int()), true
	case reflect.Int16:
		return int16(NewValue("", value).Int()), true
	case reflect.Int32:
		return int32(NewValue("", value).Int()), true
	case reflect.Int64:
		return int64(NewValue("", value).Int()), true
	case reflect.Uint:
		return uint(NewValue("", value).Int()), true
	case reflect.Uint8:
		return uint8(NewValue("", value).Int()), true
	case reflect.Uint16:
		return uint16(NewValue("", value).Int()), true
	case reflect.Uint32:
		return uint32(NewValue("", value).Int()), true
	case reflect.Uint64:
		return uint64(NewValue("", value).Int()), true
	default:
		return nil, false
	}
}
