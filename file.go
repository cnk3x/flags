package flags

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"golang.org/x/exp/errors"
)

func BindFile(fn string, value any) (err error) {
	var data []byte

	ext := strings.ToLower(filepath.Ext(fn))
	switch ext {
	case ".json", ".yaml", ".yml":
		if data, err = os.ReadFile(fn); err != nil {
			return
		}
	default:
		err = errors.New(fmt.Sprintf("config type %s not support", ext))
		return
	}

	switch ext {
	case ".json":
		err = json.Unmarshal(data, value)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, value)
	}

	return
}
