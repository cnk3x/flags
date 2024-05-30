package flags

import (
	"reflect"
	"strings"
)

func newValue(v reflect.Value, t reflect.Type) *value {
	vs := rGets(v)
	return &value{v: v, typ: t, defVal: vs, args: vs}
}

type value struct {
	v       reflect.Value
	typ     reflect.Type
	display string
	changed bool
	defVal  []string
	args    []string
}

func (v *value) String() string {
	if len(v.defVal) > 0 {
		if v.IsSlice() {
			return "[" + strings.Join(v.defVal, ",") + "]"
		} else {
			return v.defVal[0]
		}
	}
	return ""
}

func (v *value) Type() string {
	if v.display != "" {
		return "<" + v.display + ">"
	}
	return rType(v.typ)
}

func (v *value) Set(s string) (err error) {
	if err = rSets(v.v, s, !v.changed); err != nil {
		return
	}

	if !v.changed || !v.IsSlice() {
		v.args = v.args[:0]
	}

	v.args = append(v.args, s)
	v.changed = true
	return
}

func (v *value) SetDefault(args ...string) (err error) {
	for _, arg := range args {
		if err = v.Set(arg); err != nil {
			return
		}
	}

	v.changed = false
	v.defVal = v.args
	return
}

func (v *value) Args() []string { return v.args }

func (v *value) DirectType() reflect.Type {
	if v.typ.Kind() == reflect.Pointer {
		return v.typ.Elem()
	}
	return v.typ
}

func (v *value) IsBool() bool  { return v.DirectType().Kind() == reflect.Bool }
func (v *value) IsSlice() bool { return v.DirectType().Kind() == reflect.Slice }
