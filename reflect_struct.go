package flags

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"
)

type FlagField struct {
	Field           reflect.StructField
	Name            string
	Shorthand       string
	Usage           string
	Value           *value
	Env             []string
	Deprecated      string
	ShortDeprecated string

	Struct  reflect.Value
	Referer reflect.Value
}

func (f *FlagField) UpdateFromEnv() {
	printDeprecatedEnvKey := func(keys []string, ck, ak string, deprecated bool, i int) {
		if deprecated {
			if ak == "" && i < len(keys)-1 {
				for _, ek := range keys {
					if ek != "" && !strings.HasPrefix(ek, "*") {
						ak = ek
						break
					}
				}
			}

			if ak != "" {
				fmt.Fprintf(os.Stderr, "[WARN] 环境变量参数[%s]已过期,请使用[%s]替代", ck, ak)
			} else {
				fmt.Fprintf(os.Stderr, "[WARN] 环境变量参数[%s]已过期", ck)
			}
		}
	}

	checkDeprecated := func(in string) (key string, deprecated bool) {
		if key, deprecated = in, strings.HasPrefix(in, "*"); deprecated {
			key = key[1:]
		}
		return
	}

	var ak string
	for i, k := range f.Env {
		if ck, deprecated := checkDeprecated(k); ck != "" {
			if !deprecated && ak == "" {
				ak = ck
			}
			if ev := os.Getenv(ck); ev != "" {
				if e := f.Value.SetDefault(ev); e == nil {
					printDeprecatedEnvKey(f.Env, ck, ak, deprecated, i)
					return
				}
			}
		}
	}
}

func ParseStruct(src any, checkSetable ...bool) (items []*FlagField, err error) {
	r := rVal(src)

	if len(checkSetable) > 0 && checkSetable[0] && !r.CanSet() {
		err = fmt.Errorf("can't set %T", src)
		return
	}

	for i, t := 0, r.Type(); i < t.NumField(); i++ {
		f := t.Field(i)

		if !f.IsExported() {
			continue
		}

		if f.Anonymous {
			children, e := ParseStruct(r.Field(i), checkSetable...)
			if e != nil {
				err = e
				return
			}
			items = append(items, children...)
			continue
		}

		if !isAllow(f.Type) {
			return
		}

		if item, ignored := parseField(r, f, i); !ignored {
			items = append(items, &item)
		}
	}

	return
}

func parseField(r reflect.Value, f reflect.StructField, i int) (item FlagField, ignored bool) {
	if flagTag := getTag(f.Tag, _TAG_FLAG); flagTag != "" {
		if ignored = flagTag == "-"; ignored {
			return
		}

		for _, s := range fieldSpilt(flagTag) {
			switch {
			case item.Name == "":
				item.Name = s
			case len(item.Name) == 1 && item.Shorthand == "":
				item.Shorthand = item.Name
				item.Name = s
			case len(s) == 1 && item.Shorthand == "":
				item.Shorthand = s
			default:
				item.Env = append(item.Env, s)
			}
		}
	}

	if deprecatedTag := getTag(f.Tag, _TAG_DEPRECATED); deprecatedTag != "" {
		nn := fieldSpilt(deprecatedTag)
		for _, n := range nn {
			if n != "" {
				if item.Deprecated == "" {
					item.Deprecated = n
				} else if item.ShortDeprecated == "" {
					item.ShortDeprecated = n
				}
			}
		}
		if len(item.Deprecated) <= 1 && len(item.ShortDeprecated) > 1 {
			item.Deprecated, item.ShortDeprecated = item.ShortDeprecated, item.Deprecated
		}
	}

	if envTag := getTag(f.Tag, _TAG_ENV); envTag != "" && envTag != "-" {
		item.Env = append(item.Env, fieldSpilt(envTag)...)
	}

	if item.Name == "" {
		item.Name = strings.ToLower(f.Name)
	}

	item.Field = f
	item.Struct = r
	item.Referer = r.Field(i)
	item.Usage = getTag(f.Tag, _TAG_USAGE)
	item.Value = newValue(item.Referer, f.Type)
	return
}

const (
	_TAG_FLAG       = "flag"
	_TAG_DEPRECATED = "deprecated"
	_TAG_ENV        = "env"
	_TAG_USAGE      = "usage"
)

var (
	fieldSpilt = func(s string) []string {
		fields := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ';' || r == '|' || unicode.IsSpace(r) })
		var x int
		for _, s := range fields {
			if s = strings.Trim(s, "-_"); s != "" {
				fields[x] = s
				x++
			}
		}
		return fields[:x]
	}

	getTag = func(tag reflect.StructTag, tagName string) string { return strings.TrimSpace(tag.Get(tagName)) }
)
