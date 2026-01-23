// Package flags 是对 github.com/spf13/pflag 的轻量封装，目的是更方便地从结构体定义
// 生成命令行标志（flags），并支持通过结构体 tag 指定环境变量、帮助信息、短参名等。
//
// 设计要点：
// - 从结构体导出字段自动注册命令行标志；
// - 支持常见基础类型、切片类型、网络类型（net.IP/net.IPNet）和 time.Duration；
// - 支持在 tag 中指定 `flag`、`usage`、`env`，其中 `flag` 可同时包含长名与单字符短名；
// - 在 Usage 中自动显示绑定的环境变量，并在启动时从环境变量读取值作为覆盖默认值；
// - 支持在 NewSet 时通过 Option 注入版本与构建时间等元信息，Usage 会显示这些信息。
//
// 简要示例：
//
//	type Config struct {
//	    Host string `flag:"host h" usage:"server host" env:"HOST,SERVER_HOST"`
//	    Port int    `flag:"port p" usage:"server port" env:"PORT"`
//	}
//
//	func main() {
//	    var cfg Config
//	    fs := NewSet(Description("示例程序"), Version("v1.0"))
//	    fs.Struct(&cfg)
//	    fs.Parse()
//	    // 使用 cfg.Host, cfg.Port
//	}
//
// 该包试图保留 pflag 的所有能力，同时提供更适合在结构体驱动配置场景下使用的便利函数。
package flags

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/pflag"
)

// FlagSet 包装了 pflag.FlagSet，并扩展了一些功能
// 提供了设置版本号、描述、构建时间等额外功能
type FlagSet struct {
	*pSet                           // 内部使用的pflag.FlagSet实例
	description string              // FlagSet描述
	version     string              // 版本号
	buildTime   time.Time           // 构建时间
	envKeys     map[string][]string // 存储环境变量键名映射
}

// pSet 和 pFlag 是 pflag 包中对应类型的别名，用于内部隐藏实现细节
type (
	pSet  = pflag.FlagSet // 隐藏 FlagSet.FlagSet 的具体实现
	pFlag = pflag.Flag
)

// Option 函数类型，用于配置 FlagSet 的选项
type Option func(*FlagSet)

// Name 返回一个 Option，用于设置 FlagSet 的名称
func Name(name string) Option { return func(fs *FlagSet) { fs.Init(name, pflag.ExitOnError) } }

// Version 返回一个 Option，用于设置 FlagSet 的版本号
func Version(version string) Option { return func(fs *FlagSet) { fs.version = version } }

// Description 返回一个 Option，用于设置 FlagSet 的描述信息
func Description(desc string) Option { return func(fs *FlagSet) { fs.description = desc } }

// BuildTime 返回一个 Option，用于设置 FlagSet 的构建时间
//   - 支持 string、time.Time、int、int64 类型的时间表示
func BuildTime[T string | time.Time | int | int64](buildTime T) Option {
	return func(fs *FlagSet) {
		if buildTimeString, ok := any(buildTime).(string); ok {
			if buildTimeString != "" {
				fs.buildTime, _ = time.Parse(time.RFC3339, buildTimeString)
			}
			return
		}

		if buildTimeInt, ok := any(buildTime).(int); ok {
			if buildTimeInt > 0 {
				fs.buildTime = time.Unix(int64(buildTimeInt), 0)
			}
			return
		}

		if buildTimeMs, ok := any(buildTime).(int64); ok {
			if buildTimeMs > 0 {
				fs.buildTime = time.Unix(0, buildTimeMs*int64(time.Millisecond))
			}
			return
		}

		fs.buildTime, _ = any(buildTime).(time.Time)
	}
}

// NewSet 创建一个新的 FlagSet 实例
//   - 可以传入选项函数来配置实例属性
func NewSet(options ...Option) *FlagSet {
	fs := &FlagSet{
		pSet:    pflag.NewFlagSet(filepath.Base(os.Args[0]), pflag.ExitOnError),
		envKeys: map[string][]string{},
	}
	for _, apply := range options {
		apply(fs)
	}
	return fs
}

// Version 返回 FlagSet 的版本号
//   - 返回值: 当前 FlagSet 实例的版本号字符串
func (fs *FlagSet) Version() string { return fs.version }

// BuildTime 返回 FlagSet 的构建时间
//   - 返回值: 当前 FlagSet 实例的构建时间
func (fs *FlagSet) BuildTime() time.Time { return fs.buildTime }

// Description 返回 FlagSet 的描述信息
//   - 返回值: 当前 FlagSet 实例的描述字符串
func (fs *FlagSet) Description() string { return fs.description }

// ParseFrom 解析命令行参数
//   - 接收参数切片，返回错误信息
func (fs *FlagSet) ParseFrom(args []string) (err error) {
	// 设置帮助信息错误
	pflag.ErrHelp = fmt.Errorf("\nstart with %s [...OPTIONS]", filepath.Base(os.Args[0]))
	// 不对标志进行排序
	fs.SortFlags = false
	// 自定义使用说明输出格式
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, filepath.Base(os.Args[0]))
		if fs.version != "" {
			fmt.Fprintf(os.Stderr, " - version %s", fs.version)
		}

		if !fs.buildTime.IsZero() {
			fmt.Fprintf(os.Stderr, " - build %s", fs.buildTime.In(time.Local).Format(time.RFC3339))
		}

		fmt.Fprintln(os.Stderr)

		if fs.description != "" {
			fmt.Fprintln(os.Stderr, fs.description)
			fmt.Fprintln(os.Stderr)
		}

		fmt.Fprintln(os.Stderr, "OPTIONS:")
		fmt.Fprintln(os.Stderr, fs.FlagUsagesWrapped(0))
	}

	// 编译正则表达式，用于匹配被弃用的标记
	// reDeprecated := regexp.MustCompile(`\s*\*\*(.+)\*\*\s*`)
	reDeprecated := regexp.MustCompile(`\s*\*\*DEPRECATED\*\*\s*(.*)$`)
	fs.VisitAll(func(f *pFlag) {
		// 检查是否是被弃用的标记
		if matches := reDeprecated.FindStringSubmatch(f.Usage); len(matches) > 1 {
			f.Usage = f.Usage[:len(f.Usage)-len(matches[0])]
			f.Deprecated = matches[1]
			if f.Deprecated == "" {
				f.Deprecated = "deprecated"
			}
		}

		// 如果该标志有对应的环境变量，则处理环境变量值
		if keys, find := fs.envKeys[f.Name]; find && len(keys) > 0 {
			if f.Usage != "" {
				f.Usage += " "
			}
			f.Usage = fmt.Sprintf("%s[%s]", f.Usage, strings.Join(keys, ", "))

			if s := getEnv(keys); s != "" {
				if err := f.Value.Set(s); err != nil {
					fmt.Fprintf(os.Stderr, "WARN: set flag `%s` value `%s` from environ: %s\n", f.Name, s, err)
				}
			}
		}
	})
	// 解析命令行参数
	return fs.pSet.Parse(os.Args[1:])
}

// Parse 解析命令行参数，返回布尔值表示是否成功
//   - 默认从 os.Args[1:] 解析
func (fs *FlagSet) Parse() bool { return fs.ParseFrom(os.Args[1:]) == nil }

// Struct 从结构体标签中定义命令行标志
//   - 结构体字段必须有 "flag" 标签才能被识别为命令行标志
func (fs *FlagSet) Struct(structPtr any) {
	rv := reflect.Indirect(reflect.ValueOf(structPtr)) // 获取结构体值的反射对象
	rt := rv.Type()                                    // 获取结构体类型

	// 遍历结构体的所有字段
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)

		// 跳过非导出字段
		if !sf.IsExported() {
			continue
		}

		// 获取 flag 和 usage 标签
		flag, usage := sf.Tag.Get("flag"), sf.Tag.Get("usage")
		if flag == "-" || (flag == "" && usage == "") {
			continue
		}

		// 解析 flag 标签，获取完整名称和短名称
		name, short := f2ns(flag, sf.Name, lower)
		// 解析 env 标签，获取环境变量名称列表
		env := strings.FieldsFunc(sf.Tag.Get("env"), fSplit)
		// 添加标志到 FlagSet
		fs.add(rv.Field(i).Addr().Interface(), name, short, usage, env...)
	}
}

// Var 为指定值添加命令行标志
//   - v 是值的指针，name 是完整名称，short 是短名称，usage 是使用说明，env 是环境变量名称
func (fs *FlagSet) Var(v any, name, short, usage string, env ...string) {
	if err := fs.add(v, name, short, usage, env...); err != nil {
		panic(err)
	}
}

// add 根据值类型添加相应的命令行标志
//   - 内部方法，根据传入的值类型调用相应的方法添加标志
func (fs *FlagSet) add(v any, name, short, usage string, env ...string) error {
	switch x := v.(type) {
	case *time.Duration:
		fs.DurationVarP(x, name, short, *x, usage)
	case *net.IP:
		fs.IPVarP(x, name, short, *x, usage)
	case *net.IPNet:
		fs.IPNetVarP(x, name, short, *x, usage)
	case *string:
		fs.StringVarP(x, name, short, *x, usage)
	case *int:
		fs.IntVarP(x, name, short, *x, usage)
	case *int8:
		fs.Int8VarP(x, name, short, *x, usage)
	case *int16:
		fs.Int16VarP(x, name, short, *x, usage)
	case *int32:
		fs.Int32VarP(x, name, short, *x, usage)
	case *int64:
		fs.Int64VarP(x, name, short, *x, usage)
	case *uint:
		fs.UintVarP(x, name, short, *x, usage)
	case *uint8:
		fs.Uint8VarP(x, name, short, *x, usage)
	case *uint16:
		fs.Uint16VarP(x, name, short, *x, usage)
	case *uint32:
		fs.Uint32VarP(x, name, short, *x, usage)
	case *uint64:
		fs.Uint64VarP(x, name, short, *x, usage)
	case *float32:
		fs.Float32VarP(x, name, short, *x, usage)
	case *float64:
		fs.Float64VarP(x, name, short, *x, usage)
	case *bool:
		fs.BoolVarP(x, name, short, *x, usage)
	case *[]time.Duration:
		fs.DurationSliceVarP(x, name, short, *x, usage)
	case *[]net.IP:
		fs.IPSliceVarP(x, name, short, *x, usage)
	case *[]net.IPNet:
		fs.IPNetSliceVarP(x, name, short, *x, usage)
	case *[]string:
		fs.StringSliceVarP(x, name, short, *x, usage)
	case *[]int:
		fs.IntSliceVarP(x, name, short, *x, usage)
	case *[]int32:
		fs.Int32SliceVarP(x, name, short, *x, usage)
	case *[]int64:
		fs.Int64SliceVarP(x, name, short, *x, usage)
	case *[]uint:
		fs.UintSliceVarP(x, name, short, *x, usage)
	case *[]float32:
		fs.Float32SliceVarP(x, name, short, *x, usage)
	case *[]float64:
		fs.Float64SliceVarP(x, name, short, *x, usage)
	case *[]bool:
		fs.BoolSliceVarP(x, name, short, *x, usage)
	default:
		return fmt.Errorf("%s type %v(%T) not support", name, x, x)
	}
	// 如果提供了环境变量名，则记录到映射表中
	if len(env) > 0 {
		fs.envKeys[name] = env
	}
	return nil
}

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
