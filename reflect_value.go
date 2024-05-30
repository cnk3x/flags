package flags

import (
	"fmt"
	"reflect"
	"strconv"
)

func rVal(src any, indirect ...bool) reflect.Value {
	r, ok := src.(reflect.Value)
	if !ok {
		r = reflect.ValueOf(src)
	}
	if len(indirect) > 0 && indirect[0] {
		r = reflect.Indirect(r)
	}
	return r
}

func rGets(v reflect.Value) (out []string) {
	if !v.IsValid() || v.IsZero() {
		return
	}

	if te := GetExtend(v.Type()); te != nil {
		return newSlice(te.Get(v))
	}

	switch kind := v.Kind(); kind {
	case reflect.Pointer:
		out = rGets(v.Elem())
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			out = append(out, rGets(v.Index(i))...)
		}
	case reflect.String:
		out = newSlice(v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		out = newSlice(strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		out = newSlice(strconv.FormatUint(v.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		out = newSlice(strconv.FormatFloat(v.Float(), 'f', -1, 64))
	case reflect.Bool:
		out = newSlice(strconv.FormatBool(v.Bool()))
	}
	return
}

func rSets(v reflect.Value, s string, reset ...bool) (err error) {
	if !v.CanSet() && v.Kind() != reflect.Pointer {
		return fmt.Errorf("value can not set, kind=%s", v.Kind())
	}

	if te := GetExtend(v.Type()); te != nil {
		return te.Set(v, s)
	}

	switch kind := v.Kind(); kind {
	case reflect.String:
		v.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		err = rSetInt(v, s)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = rSetUint(v, s)
	case reflect.Float32, reflect.Float64:
		err = rSetFloat(v, s)
	case reflect.Bool:
		err = rSetBool(v, s)
	case reflect.Pointer:
		err = rSetPtr(v, s, len(reset) > 0 && reset[0])
	case reflect.Slice:
		err = rSetSs(v, s, len(reset) > 0 && reset[0])
	default:
		err = fmt.Errorf("unknown kind: %s", kind.String())
	}

	return
}

func rSetSs(v reflect.Value, s string, reset bool) (err error) {
	if !v.IsValid() || v.Kind() != reflect.Slice {
		return invalid("rAppend")
	}

	if v.IsNil() {
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	} else if reset {
		v.Set(v.Slice(0, 0))
	}

	var el reflect.Value
	if et := v.Type().Elem(); et.Kind() == reflect.Pointer {
		el = reflect.New(et.Elem())
	} else {
		el = reflect.New(et).Elem()
	}

	if err = rSets(el, s, false); err != nil {
		return
	}

	v.Set(reflect.Append(v, el))
	return
}

func rSetPtr(v reflect.Value, s string, reset bool) (err error) {
	if !v.IsValid() || v.Kind() != reflect.Pointer {
		return invalid("rSetPtr")
	}

	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	return rSets(v.Elem(), s, reset)
}

func rSetInt(v reflect.Value, s string) (err error) {
	if !v.IsValid() || !isIntKind(v.Kind()) {
		return invalid("rSetInt")
	}
	r, e := strconv.ParseInt(s, 0, bits(v.Type()))
	if e == nil {
		v.SetInt(r)
	}
	return e
}

func rSetUint(v reflect.Value, s string) (err error) {
	if !v.IsValid() || !isUintKind(v.Kind()) {
		return invalid("rSetUint")
	}
	r, e := strconv.ParseUint(s, 0, bits(v.Type()))
	if e == nil {
		v.SetUint(r)
	}
	return e
}

func rSetFloat(v reflect.Value, s string) (err error) {
	if !v.IsValid() || !isFloatKind(v.Kind()) {
		return invalid("rSetFloat")
	}
	r, e := strconv.ParseFloat(s, bits(v.Type()))
	if e == nil {
		v.SetFloat(r)
	}
	return e
}

func rSetBool(v reflect.Value, s string) (err error) {
	if !v.IsValid() || v.Kind() != reflect.Bool {
		return invalid("rSetBool")
	}
	r, e := strconv.ParseBool(s)
	if e == nil {
		v.SetBool(r)
	}
	return e
}

func invalid(method string) error {
	return &reflect.ValueError{Method: method, Kind: reflect.Invalid}
}

func newSlice[T any](n ...T) []T { return n }
