package flags

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
)

func BindFile(fn string, value any) error {
	data, err := os.ReadFile(fn)
	if err != nil {
		return err
	}
	ext := strings.ToLower(filepath.Ext(fn))
	switch ext {
	case ".yaml", ".yml":
		if data, err = yaml.YAMLToJSON(data); err != nil {
			return err
		}
	case ".json":
	default:
		if _data, _ := yaml.YAMLToJSON(data); len(_data) > 0 {
			data = _data
		}
	}
	return json.Unmarshal(data, value)
}
