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
	"strings"
	"time"

	"github.com/spf13/pflag"
)

// Set 包装了 pflag.Set，并扩展了一些功能
// 提供了设置版本号、描述、构建时间等额外功能
type Set struct {
	*pSet                 // 内部使用的pflag.FlagSet实例
	description string    // FlagSet描述
	version     string    // 版本号
	buildTime   time.Time // 构建时间
	errs        []error   // 错误信息
}

type (
	FlagItem = pflag.Flag
	FlagSet  = pflag.FlagSet

	pSet = pflag.FlagSet
)

// New 创建一个新的 App 实例
//   - 可以传入选项函数来配置实例属性
func New(options ...Option) *Set {
	fs := &Set{
		pSet: pflag.NewFlagSet(filepath.Base(os.Args[0]), pflag.ExitOnError),
	}
	for _, apply := range options {
		apply(fs)
	}
	return fs
}

// Version 返回 FlagSet 的版本号
//   - 返回值: 当前 FlagSet 实例的版本号字符串
func (fs *Set) Version() string { return fs.version }

// BuildTime 返回 FlagSet 的构建时间
//   - 返回值: 当前 FlagSet 实例的构建时间
func (fs *Set) BuildTime() time.Time { return fs.buildTime }

// Description 返回 FlagSet 的描述信息
//   - 返回值: 当前 FlagSet 实例的描述字符串
func (fs *Set) Description() string { return fs.description }

// ParseArgs 解析命令行参数
//   - 接收参数切片，返回错误信息
func (fs *Set) ParseArgs(args []string) (err error) {
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

	// 解析命令行参数
	return fs.pSet.Parse(os.Args[1:])
}

// Parse 解析命令行参数，返回布尔值表示是否成功
//   - 默认从 os.Args[1:] 解析
func (fs *Set) Parse() bool { return fs.ParseArgs(os.Args[1:]) == nil }

// Struct 从结构体标签中定义命令行标志
//   - 结构体字段必须有 "flag" 标签才能被识别为命令行标志
func (fs *Set) Struct(structPtr any) {
	if err := AddStruct(fs.pSet, structPtr); err != nil {
		fs.errs = append(fs.errs, err)
	}
}

// Var 为指定值添加命令行标志
//   - v 是值的指针，name 是完整名称，short 是短名称，usage 是使用说明，env 是环境变量名称
func (fs *Set) Var(v any, name string, itemOptions ...ItemOption) *FlagItem {
	f, err := AddVar(fs.pSet, v, name, itemOptions...)
	if err != nil {
		fs.errs = append(fs.errs, err)
	}
	return f
}

// AddVar 根据值类型添加相应的命令行标志
func AddVar(fs *FlagSet, v any, name string, itemOptions ...ItemOption) (f *FlagItem, err error) {
	var item Item
	item.Name = name
	for _, apply := range itemOptions {
		apply(&item)
	}

	switch x := v.(type) {
	case *time.Duration:
		fs.DurationVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *net.IP:
		fs.IPVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *net.IPNet:
		fs.IPNetVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *string:
		fs.StringVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *int:
		fs.IntVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *int8:
		fs.Int8VarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *int16:
		fs.Int16VarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *int32:
		fs.Int32VarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *int64:
		fs.Int64VarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *uint:
		fs.UintVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *uint8:
		fs.Uint8VarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *uint16:
		fs.Uint16VarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *uint32:
		fs.Uint32VarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *uint64:
		fs.Uint64VarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *float32:
		fs.Float32VarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *float64:
		fs.Float64VarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *bool:
		fs.BoolVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *[]time.Duration:
		fs.DurationSliceVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *[]net.IP:
		fs.IPSliceVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *[]net.IPNet:
		fs.IPNetSliceVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *[]string:
		fs.StringSliceVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *[]int:
		fs.IntSliceVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *[]int32:
		fs.Int32SliceVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *[]int64:
		fs.Int64SliceVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *[]uint:
		fs.UintSliceVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *[]float32:
		fs.Float32SliceVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *[]float64:
		fs.Float64SliceVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	case *[]bool:
		fs.BoolSliceVarP(x, item.Name, item.Shorthand, *x, item.Usage)
	default:
		err = fmt.Errorf("%s type %v(%T) not support", item.Name, x, x)
		return
	}

	f = fs.Lookup(item.Name)

	// 检查是否是被弃用的标记
	if matches := reDeprecated.FindStringSubmatch(f.Usage); len(matches) > 1 {
		f.Usage = f.Usage[:len(f.Usage)-len(matches[0])]
		f.Deprecated = matches[1]
		if f.Deprecated == "" {
			f.Deprecated = "deprecated"
		}
	}

	// 如果该标志有对应的环境变量，则处理环境变量值
	if len(item.Env) > 0 {
		if f.Usage != "" {
			f.Usage += " "
		}
		f.Usage = fmt.Sprintf("%s[%s]", f.Usage, strings.Join(item.Env, ", "))
		if s := getEnv(item.Env); s != "" {
			if e := f.Value.Set(s); e != nil {
				fmt.Fprintf(os.Stderr, "WARN: set flag `%s` value `%s` from environ: %s\n", f.Name, s, e)
			}
		}
	}

	return
}

// AddStruct 从结构体中定义命令行标志
func AddStruct(fs *FlagSet, pStruct any) (err error) {
	rv := reflect.Indirect(reflect.ValueOf(pStruct)) // 获取结构体值的反射对象
	rt := rv.Type()                                  // 获取结构体类型

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
		if _, err = AddVar(
			fs,
			rv.Field(i).Addr().Interface(),
			name,
			func(item *Item) { item.Shorthand, item.Usage, item.Env = short, usage, env },
		); err != nil {
			break
		}
	}

	return
}
