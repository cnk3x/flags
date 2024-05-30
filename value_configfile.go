package flags

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

var _ = isConfigFile

var typeConfigFile = reflect.TypeOf(ConfigFile(""))

func isConfigFile(t reflect.Type) bool { return t == typeConfigFile }

func BindFile(structPtr any, name, shorthand, defVal, usage string, flags ...*FlagSet) {
	v := &configFileValue{structPtr: structPtr, path: defVal}
	for _, set := range flags {
		set.VarP(v, shorthand, defVal, usage)
	}
}

type ConfigFile string

type configFileValue struct {
	path      string
	structPtr any
}

func (b *configFileValue) String() string { return b.path }
func (b *configFileValue) Type() string   { return "configfile" }
func (b *configFileValue) Set(s string) (err error) {
	if b.path = s; b.path != "" {
		ct, path := getCotentType(s)
		switch ct {
		case "json":
			_, err = readBytes(path, UnmarshalJSON(b.structPtr))
		case "yaml":
			_, err = readBytes(path, UnmarshalYAML(b.structPtr))
		case "toml":
			_, err = readBytes(path, UnmarshalToml(b.structPtr))
		case "ini":
			_, err = readBytes(path, UnmarshalIni(b.structPtr))
		default:
			err = fmt.Errorf("unsupported config file: %s", s)
		}
		if os.IsNotExist(err) {
			err = nil
		}
	}
	return
}

func getCotentType(s string) (ct, path string) {
	if s != "" {
		var ok bool
		if ct, path, ok = strings.Cut(s, ":"); ok {
			switch ct {
			case "json", "yaml", "toml", "ini":
			default:
				ok = false
			}
			return
		}

		if !ok {
			switch path, ct = s, strings.TrimPrefix(filepath.Ext(s), "."); ct {
			case "jsonc":
				ct = "json"
			case "yml":
				ct = "yaml"
			case "tml":
				ct = "toml"
			}
		}
	}

	return
}

type drFunc = func(data []byte) (err error)

func readBytes(filename string, read drFunc) (data []byte, err error) {
	if filename != "" {
		if data, err = os.ReadFile(filename); err == nil {
			err = read(data)
		}
	}
	return
}
