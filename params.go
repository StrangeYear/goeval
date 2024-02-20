// Copyright (c) 2017, Paessler AG <support@paessler.com>
// All rights reserved.

package goeval

import (
	"reflect"
)

func SelectValue(value any, key string) any {
	if value == nil {
		return nil
	}

	vv := reflect.ValueOf(value)
	vvElem := resolvePotentialPointer(vv)

	switch vvElem.Kind() {
	case reflect.Map:
		mapKey, ok := reflectConvertTo(vv.Type().Key().Kind(), key)
		if !ok {
			return nil
		}

		vvElem = vv.MapIndex(reflect.ValueOf(mapKey))
		vvElem = resolvePotentialPointer(vvElem)

		if vvElem.IsValid() {
			return vvElem.Interface()
		}
	case reflect.Slice, reflect.Array, reflect.String:
		if i, err := NewValue("", key).Int(); err == nil && i >= 0 && vv.Len() > i {
			vvElem = resolvePotentialPointer(vv.Index(i))
			return vvElem.Interface()
		}
	case reflect.Struct:
		field := vvElem.FieldByName(key)
		if field.IsValid() {
			return field.Interface()
		}
		method := vv.MethodByName(key)
		if method.IsValid() {
			return method.Interface()
		}
	default:
		return nil
	}

	return nil
}

func resolvePotentialPointer(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Ptr {
		return value.Elem()
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
		if i, err := NewValue("", value).Int(); err == nil {
			return i, true
		}
	case reflect.Float64:
		if f, err := NewValue("", value).Float(); err == nil {
			return f, true
		}
	case reflect.Float32:
		if f, err := NewValue("", value).Float(); err == nil {
			return float32(f), true
		}
	case reflect.Bool:
		return NewValue("", value).Boolean(), true
	case reflect.Int8:
		if i, err := NewValue("", value).Int(); err == nil {
			return int8(i), true
		}
	case reflect.Int16:
		if i, err := NewValue("", value).Int(); err == nil {
			return int16(i), true
		}
	case reflect.Int32:
		if i, err := NewValue("", value).Int(); err == nil {
			return int32(i), true
		}
	case reflect.Int64:
		if i, err := NewValue("", value).Int(); err == nil {
			return int64(i), true
		}
	case reflect.Uint:
		if i, err := NewValue("", value).Int(); err == nil && i > 0 {
			return uint(i), true
		}
	case reflect.Uint8:
		if i, err := NewValue("", value).Int(); err == nil && i > 0 {
			return uint8(i), true
		}
	case reflect.Uint16:
		if i, err := NewValue("", value).Int(); err == nil && i > 0 {
			return uint16(i), true
		}
	case reflect.Uint32:
		if i, err := NewValue("", value).Int(); err == nil && i > 0 {
			return uint32(i), true
		}
	case reflect.Uint64:
		if i, err := NewValue("", value).Int(); err == nil && i > 0 {
			return uint64(i), true
		}
	default:
		return nil, false
	}
	return nil, false
}
