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
		if i := NewValue("", key).Int(); i >= 0 && vv.Len() > i {
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
