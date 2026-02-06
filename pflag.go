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

type (
	FlagItem = pflag.Flag
	FlagSet  = pflag.FlagSet

	pSet = pflag.FlagSet
)

func pSetInit() *FlagSet {
	fSet := pflag.NewFlagSet(filepath.Base(os.Args[0]), pflag.ExitOnError)
	// 设置帮助信息错误
	pflag.ErrHelp = fmt.Errorf("\nstart with %s [...OPTIONS]", filepath.Base(os.Args[0]))
	// 不对标志进行排序
	fSet.SortFlags = false
	return fSet
}

// AddToSet 根据值类型添加相应的命令行标志
func AddToSet(fs *FlagSet, val any, name string, itemOptions ...ItemOption) (f *FlagItem, err error) {
	var item Item
	item.Name = name
	for _, apply := range itemOptions {
		apply(&item)
	}

	switch x := val.(type) {
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

// StructAddToSet 从结构体中定义命令行标志
func StructToSet(fs *FlagSet, pStruct any) (err error) {
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
		if _, err = AddToSet(
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
