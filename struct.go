package flags

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slog"
)

type Options struct {
	Prefix      string
	EnvPrefix   string
	Version     string
	Description string
}

func ParseStruct(set *FlagSet, value any, fOpts *Options) (err error) {
	if fOpts == nil {
		fOpts = &Options{}
	}

	rv := reflect.ValueOf(value).Elem()
	rt := rv.Type()

	var cfgFile *FieldItem
	var flagItems []*FieldItem
	var nl, sl, tl, el int
	var pl = len(fOpts.EnvPrefix)

	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)

		name := fieldName(ft, fOpts.Prefix)
		if name == "-" {
			continue
		}

		fit := &FieldItem{Name: name, Value: rv.Field(i), Field: ft}

		if strings.HasSuffix(fit.Name, ",file") && ft.Type.Kind() == reflect.String {
			fit.Name = strings.TrimSuffix(fit.Name, ",file")
			cfgFile = fit
		}

		if fit.EnvKey = ft.Tag.Get("env"); fit.EnvKey != "-" {
			if fit.EnvKey == "" {
				fit.EnvKey = jEnvKey(strings.ToUpper(fit.Name), fOpts.EnvPrefix)
			} else {
				fit.EnvKey = jEnvKey(fit.EnvKey, fOpts.EnvPrefix)
			}

			if v := os.Getenv(fit.EnvKey); v != "" {
				if err = fit.Set(v); err != nil {
					return
				}
			}
		} else {
			fit.EnvKey = ""
		}

		fit.Short = fieldTag(ft, "short")
		fit.Usage = fieldTag(ft, "usage")
		fit.Alias = fieldsSplit(fieldTag(ft, "alias"))
		flagItems = append(flagItems, fit)

		//apply
		set.Var(fit, fit.Name, fit.Usage)
		for _, alias := range fit.Alias {
			set.Var(fit, alias, "alias of "+fit.Name)
		}
		if fit.Short != "" {
			set.Var(fit, fit.Short, "short of "+fit.Name)
		}

		if l := len(fit.Name) + 2; l > nl {
			nl = l
		}

		for _, n := range fit.Alias {
			if l := len(n) + 2; l > nl {
				nl = l
			}
		}

		if fit.Short != "" {
			if l := len(fit.Short) + 2; l > sl {
				sl = l
			}
		}

		if fit.EnvKey != "" {
			if l := len(fit.EnvKey) + pl + 2; l > el {
				el = l
			}
		}

		if l := len(ft.Type.Kind().String()); l > tl {
			tl = l
		}
	}

	set.Usage = func() {
		fmt.Fprint(os.Stderr, filepath.Base(set.Name()))
		if fOpts.Version != "" {
			fmt.Fprint(os.Stderr, " - version ", fOpts.Version)
		}
		if fOpts.Description != "" {
			fmt.Fprint(os.Stderr, " - ", fOpts.Description)
		}
		fmt.Fprintln(os.Stderr)

		if len(flagItems) == 0 {
			return
		}

		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "命令格式:")
		fmt.Fprintf(os.Stderr, "    %s [...参数选项]\n", filepath.Base(set.Name()))
		fmt.Fprintln(os.Stderr)

		fmt.Fprintln(os.Stderr, "参数选项:")
		sort.Slice(flagItems, func(i, j int) bool { return flagItems[i].Name < flagItems[j].Name })
		for _, it := range flagItems {
			fmt.Fprint(os.Stderr, "    ")

			if sl > 0 {
				if it.Short != "" {
					fmt.Fprintf(os.Stderr, "-%*s,", -sl+2, it.Short)
				} else {
					fmt.Fprintf(os.Stderr, "%*s", sl, "")
				}
				fmt.Fprint(os.Stderr, " ")
			}

			fmt.Fprintf(os.Stderr, "--%*s  ", -nl+2, it.Name)

			if it.Field.Type.Kind() != reflect.Bool {
				fmt.Fprintf(os.Stderr, "%*s  ", -tl, it.Value.Kind().String())
			} else {
				fmt.Fprintf(os.Stderr, "%*s  ", tl, "")
			}

			if el > 0 {
				if it.EnvKey != "" {
					fmt.Fprintf(os.Stderr, "%*s", -el, "["+jEnvKey(it.EnvKey, fOpts.EnvPrefix)+"]")
				} else {
					fmt.Fprintf(os.Stderr, "%*s", el, "")
				}
				fmt.Fprint(os.Stderr, " ")
			}

			fmt.Fprint(os.Stderr, it.Usage)

			if vs := it.String(); vs != "" {
				if it.Usage != "" {
					fmt.Fprint(os.Stderr, " ")
				}
				fmt.Fprintf(os.Stderr, `(默认: "%s")`, vs)
			}

			fmt.Fprintln(os.Stderr)
		}
	}

	if err = set.Parse(os.Args); err != nil {
		return
	}

	if cfgFile != nil {
		if fn := cfgFile.Value.String(); fn != "" {
			if err = BindFile(fn, value); err != nil {
				return
			}
		}
	}

	return
}

func PrintFields(value any) {
	rv := reflect.Indirect(reflect.ValueOf(value))
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		slog.Debug(fmt.Sprintf("%s=%v", Snakecase(ft.Name), rv.Field(i).Interface()))
	}
}

type FieldItem struct {
	Name   string
	Alias  []string
	Short  string
	Usage  string
	EnvKey string
	Value  reflect.Value
	Field  reflect.StructField
}

func (fv *FieldItem) String() string {
	if !fv.Value.IsValid() || fv.Value.IsZero() {
		return ""
	}
	switch v := fv.Value.Interface().(type) {
	case time.Duration:
		return v.String()
	case time.Time:
		return v.Format(time.RFC3339)
	case net.IP:
		if len(v) == 0 {
			return ""
		}
		return v.String()
	case net.IPMask:
		if len(v) == 0 {
			return ""
		}
		return v.String()
	case *net.IPNet:
		return v.String()
	}

	switch fv.Value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(fv.Value.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(fv.Value.Uint(), 10)
	case reflect.String:
		return fv.Value.String()
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(fv.Value.Float(), 'f', 19, 64)
	case reflect.Bool:
		return strconv.FormatBool(fv.Value.Bool())
	case reflect.Struct:
		return ""
	}
	return ""
}

func (fv *FieldItem) Set(s string) (err error) {
	defer func() {
		if re := recover(); re != nil {
			if er, ok := re.(error); ok {
				err = er
			} else {
				err = fmt.Errorf("%v", er)
			}
		}
	}()

	if s == "" && !fv.IsBoolFlag() {
		return
	}
	switch fv.Value.Interface().(type) {
	case time.Duration:
		var d time.Duration
		if d, err = time.ParseDuration(s); err != nil {
			return
		}
		fv.Value.SetInt(int64(d))
	case time.Time:
		var t time.Time
		if t, err = time.Parse(time.RFC3339, s); err != nil {
			return
		}
		fv.Value.Set(reflect.ValueOf(t))
	case net.IP, net.IPMask:
		if ip := net.ParseIP(s); ip != nil {
			fv.Value.SetBytes(ip)
		}
	case *net.IPNet:
		if _, _, ni := net.ParseCIDR(s); ni != nil {
			fv.Value.Set(reflect.ValueOf(ni))
		}
	case []byte:
		fv.Value.SetBytes([]byte(s))
	default:
		switch fv.Value.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var v int64
			if v, err = strconv.ParseInt(s, 10, 64); err != nil {
				return
			}
			fv.Value.SetInt(v)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			var v uint64
			if v, err = strconv.ParseUint(s, 10, 64); err != nil {
				return
			}
			fv.Value.SetUint(v)
		case reflect.String:
			fv.Value.SetString(s)
		case reflect.Float32, reflect.Float64:
			var v float64
			if v, err = strconv.ParseFloat(s, 64); err != nil {
				return
			}
			fv.Value.SetFloat(v)
		case reflect.Bool:
			var v = s == ""
			if !v {
				if v, err = strconv.ParseBool(s); err != nil {
					return
				}
			}
			fv.Value.SetBool(v)
		}
	}

	return
}

func (fv *FieldItem) IsBoolFlag() bool { return fv.Field.Type.Kind() == reflect.Bool }

func fieldName(ft reflect.StructField, prefix string) (name string) {
	if !ft.IsExported() {
		name = "-"
	}

	if name != "-" {
		if name = fieldTag(ft, "flag"); name == "" {
			name = Snakecase(ft.Name)
		}

		if prefix != "" && name != "-" {
			name = prefix + name
		}

		name = strings.ReplaceAll(name, " ", "")
	}

	return
}

func fieldTag(ft reflect.StructField, name string) string {
	return strings.TrimSpace(ft.Tag.Get(name))
}

func jEnvKey(key string, prefix string) string {
	if key == "" {
		prefix = ""
	}
	return strings.ReplaceAll(prefix+key, "-", "_")
}

var fieldsSplitRe = regexp.MustCompile(`[\s,;|]+`)

func fieldsSplit(s string) (arr []string) {
	var x int
	arr = fieldsSplitRe.Split(s, -1)
	for _, it := range arr {
		if it != "" {
			arr[x] = it
			x++
		}
	}
	arr = arr[:x]
	return
}
