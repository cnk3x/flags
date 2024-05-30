package flags

import (
	"reflect"
)

var extends = ExtendMap{}

type ExtendMap map[reflect.Type]*ExtendType

type ExtendType struct {
	typ     reflect.Type
	setFunc func(reflect.Value, string) error
	getFunc func(reflect.Value) string
	newFunc func() reflect.Value
}

func (te *ExtendType) New() (v reflect.Value) {
	if te != nil && te.newFunc != nil {
		v = te.newFunc()
	}
	return
}

func (te *ExtendType) Get(v reflect.Value) (s string) {
	if te != nil && te.getFunc != nil {
		s = te.getFunc(v)
	}
	return
}

func (te *ExtendType) Set(v reflect.Value, s string) (err error) {
	if te != nil && te.setFunc != nil {
		if err = te.setFunc(v, s); err != nil {
			return
		}
	}
	return
}

func (te *ExtendType) Type() string { return rType(te.typ, true) }

func Extend[T any](parse func(string) (T, error), format func(T) string) {
	setFunc := func(v reflect.Value, s string) (err error) {
		if parse != nil {
			if r, e := parse(s); e == nil {
				v.Set(reflect.ValueOf(r))
			} else {
				err = e
			}
		}
		return
	}

	getFunc := func(v reflect.Value) (s string) {
		if format != nil {
			r, _ := v.Interface().(T)
			s = format(r)
		}
		return
	}

	var x T
	it := &ExtendType{typ: reflect.TypeOf(x), setFunc: setFunc, getFunc: getFunc}
	if extends == nil {
		extends = ExtendMap{it.typ: it}
	} else {
		extends[it.typ] = it
	}
}

func GetExtend(t reflect.Type) *ExtendType {
	if extends != nil {
		return extends[t]
	}
	return nil
}

func HasExtend(t reflect.Type) bool {
	if extends != nil {
		_, found := extends[t]
		return found
	}
	return false
}
