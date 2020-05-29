package grc

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/appootb/grc/backend"
)

func isValidSystemType(t reflect.Type, depth int) bool {
	if t.Kind() == reflect.Ptr {
		return isValidSystemType(t.Elem(), depth)
	}
	switch t.Kind() {
	case reflect.String,
		reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.Slice, reflect.Array,
		reflect.Map:
		if depth > 1 {
			return false
		}
		return isValidSystemType(t.Elem(), depth+1)
	default:
		return false
	}
}

func isBaseType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		return isBaseType(t.Elem())
	}
	switch t {
	case reflect.TypeOf(String{}),
		reflect.TypeOf(Bool{}),
		reflect.TypeOf(Int{}),
		reflect.TypeOf(Uint{}),
		reflect.TypeOf(Float{}),
		reflect.TypeOf(Array{}),
		reflect.TypeOf(Map{}):
		return true
	default:
		return false
	}
}

func isSliceOrMap(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Ptr:
		return isSliceOrMap(t.Elem())
	case reflect.Slice, reflect.Array,
		reflect.Map:
		return true
	default:
		return t == reflect.TypeOf(Array{}) || t == reflect.TypeOf(Map{})
	}
}

func formatDefaultValue(t reflect.Type, tag reflect.StructTag) string {
	val := tag.Get("default")
	if isSliceOrMap(t) && !strings.Contains(val, ";") {
		val = strings.ReplaceAll(val, ",", ";")
	}
	return val
}

func configElem(v reflect.Value) reflect.Value {
	if v.Type().Kind() == reflect.Ptr {
		return configElem(v.Elem())
	}
	return v
}

func parseConfig(t reflect.Type, baseName string) backend.ConfigItems {
	if t.Kind() == reflect.Ptr {
		return parseConfig(t.Elem(), baseName)
	}
	return parseConfigItems(t, baseName)
}

func parseConfigItems(t reflect.Type, baseName string) backend.ConfigItems {
	items := backend.ConfigItems{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if isBaseType(field.Type) || isValidSystemType(field.Type, 0) {
			items[baseName+field.Name] = &backend.ConfigItem{
				Type:    strings.ReplaceAll(field.Type.String(), "*", ""),
				Value:   formatDefaultValue(field.Type, field.Tag),
				Comment: field.Tag.Get("comment"),
			}
		} else if field.Type.Kind() == reflect.Ptr {
			items.Add(parseConfigItems(field.Type, baseName))
		} else if field.Type.Kind() == reflect.Struct {
			items.Add(parseConfigItems(field.Type, baseName+field.Name+"/"))
		} else {
			panic("grc: unsupported field type:" + field.Type.String())
		}
	}
	return items
}

func setSystemTypeValue(s string, v reflect.Value, recursion bool) bool {
	// Used for slice or map value.
	sep := ";"
	if recursion {
		sep = ","
	}

	switch v.Type().Kind() {
	case reflect.Ptr:
		e := reflect.New(v.Type().Elem())
		setSystemTypeValue(s, e.Elem(), false)
		v.Set(e)
	case reflect.String:
		v.SetString(s)
	case reflect.Bool:
		bv, _ := strconv.ParseBool(s)
		v.SetBool(bv)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		iv, _ := strconv.ParseInt(s, 10, 64)
		v.SetInt(iv)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uv, _ := strconv.ParseUint(s, 10, 64)
		v.SetUint(uv)
	case reflect.Float32, reflect.Float64:
		fv, _ := strconv.ParseFloat(s, 64)
		v.SetFloat(fv)
	case reflect.Slice, reflect.Array:
		fields := strings.Split(s, sep)
		sv := reflect.MakeSlice(v.Type(), len(fields), len(fields))
		for i, field := range fields {
			setSystemTypeValue(field, sv.Index(i), true)
		}
		v.Set(sv)
	case reflect.Map:
		vs := strings.Split(s, sep)
		mv := reflect.MakeMapWithSize(v.Type(), len(vs))
		for _, vv := range vs {
			kv := strings.SplitN(vv, ":", 2)
			k := reflect.New(v.Type().Key())
			v := reflect.New(v.Type().Elem())
			setSystemTypeValue(kv[0], k.Elem(), true)
			if len(kv) > 1 {
				setSystemTypeValue(kv[1], v.Elem(), true)
			}
			mv.SetMapIndex(k.Elem(), v.Elem())
		}
		v.Set(mv)
	default:
		return false
	}
	return true
}

func setBaseTypeValue(s string, v reflect.Value) bool {
	if v.CanInterface() {
		if v.Type().Kind() == reflect.Ptr && v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if u, ok := v.Interface().(AtomicUpdate); ok {
			u.Store(s)
			return true
		}
	}
	if v.CanAddr() && v.Addr().CanInterface() {
		if u, ok := v.Addr().Interface().(AtomicUpdate); ok {
			u.Store(s)
			return true
		}
	}
	return false
}
