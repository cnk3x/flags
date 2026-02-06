package flags

import (
	"os"
	"regexp"
	"strings"
	"unicode"
)

// 匹配弃用信息
var reDeprecated = regexp.MustCompile(`\s*\*\*DEPRECATED\*\*\s*(.*)$`)

// lower 将驼峰命名转换为下划线分隔的小写形式
//   - 例如：MaxConnection -> max_connection
func lower(s string) string {
	var b strings.Builder
	var prevUp bool
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i != 0 && !prevUp {
				b.WriteRune('_')
			}
			prevUp = true
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
			prevUp = false
		}
	}
	return b.String()
}

// getEnv 从环境变量中获取第一个非空值
//   - 按顺序检查提供的环境变量键，返回第一个存在的值
func getEnv(keys []string) (s string) {
	if len(keys) > 0 {
		for _, e := range keys {
			if s = os.Getenv(e); s != "" {
				return
			}
		}
	}
	return ""
}

// fSplit 定义分割字符串的规则
//   - 支持空格、逗号、分号和空白字符作为分隔符
func fSplit(r rune) bool { return r == ' ' || r == ',' || r == ';' || unicode.IsSpace(r) }

// f2ns 解析 flag 标签，提取完整名称和短名称
//   - flag 参数是标签内容，def 是默认名称，defParse 是默认名称解析函数
func f2ns(flag string, def string, defParse func(string) string) (name, short string) {
	for n := range strings.FieldsFuncSeq(flag, fSplit) {
		if name != "" && short != "" {
			break
		}
		if len(n) == 1 && short == "" {
			short = n
		} else if len(n) > 1 {
			name = n
		}
	}
	if name == "" {
		name = defParse(def)
	}
	return
}
