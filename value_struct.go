package flags

import (
	"fmt"
	"reflect"
	"strings"
)

func StructBind(structPtr any, flags ...*FlagSet) {
	if len(flags) == 0 {
		flags = append(flags, Default())
	}

	v := reflect.Indirect(reflect.ValueOf(structPtr))

	if !v.CanSet() {
		panic(fmt.Errorf("cannot set %T", structPtr))
	}

	fields, err := ParseStruct(v, true)
	if err != nil {
		panic(err)
	}

	for _, field := range fields {
		field.UpdateFromEnv()

		usage := field.Usage
		if usage == "" {
			usage = field.Field.Name
		}

		if len(field.Env) > 0 {
			usage += fmt.Sprintf(" (env: %s)", strings.Join(field.Env, ", "))
		}

		item := flagSet(flags).VarPF(field.Value, field.Name, field.Shorthand, usage)
		item.Deprecated = field.Deprecated
		item.ShorthandDeprecated = field.ShortDeprecated

		fv := reflect.Indirect(field.Value.v)
		if fv.IsValid() && fv.Kind() == reflect.Bool {
			item.NoOptDefVal = "true"
		}
	}
}

func StructPrint(structPtr any, print func(s string)) {
	fields, err := ParseStruct(reflect.Indirect(reflect.ValueOf(structPtr)))
	if err != nil {
		print(err.Error())
		return
	}

	max := 0
	for _, f := range fields {
		if l := len(f.Field.Name); l > max {
			max = l
		}
	}

	for _, f := range fields {
		print(fmt.Sprintf("%-*s | %s", max, f.Name, strings.Join(f.Value.Args(), ", ")))
	}
}

func StructToArgs(structPtr any) (args []string) {
	fields, _ := ParseStruct(reflect.Indirect(reflect.ValueOf(structPtr)))
	for _, f := range fields {
		for _, s := range f.Value.Args() {
			args = append(args, "--"+f.Name)
			args = append(args, s)
		}
	}
	return
}
