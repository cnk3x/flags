package flags

import (
	"encoding/json"

	"github.com/BurntSushi/toml"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
)

func UnmarshalYAML(v any) drFunc { return func(data []byte) error { return yaml.Unmarshal(data, v) } }
func UnmarshalToml(v any) drFunc { return func(data []byte) error { return toml.Unmarshal(data, v) } }
func UnmarshalIni(v any) drFunc  { return func(data []byte) error { return ini.MapTo(v, data) } }

func UnmarshalJSON(v any) drFunc {
	return func(data []byte) error { return json.Unmarshal(jcTranslate(data), v) }
}
