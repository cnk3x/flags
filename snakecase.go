package flags

import (
	"unicode"
)

func Snakecase(s string) string {
	in := []rune(s)
	isLower := func(idx int) bool {
		return idx >= 0 && idx < len(in) && unicode.IsLower(in[idx])
	}

	out := make([]rune, 0, len(in)+len(in)/2)
	for i, r := range in {
		if unicode.IsUpper(r) {
			r = unicode.ToLower(r)
			if i > 0 && in[i-1] != '-' && (isLower(i-1) || isLower(i+1)) {
				out = append(out, '-')
			}
		}
		out = append(out, r)
	}

	return string(out)
}
