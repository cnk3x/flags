package flags

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
)

var defaultSet = New()

func Default() *Set { return defaultSet }

func SetDefault(options ...Option) {
	for _, option := range options {
		option(defaultSet)
	}
}

// ParseStruct 从结构体中定义命令行标志并解析命令行参数
//   - 默认从 os.Args[1:] 解析
//   - 从结构体中获取名称，版本信息, 构建时间, 描述信息
func ParseStruct(pStruct any, options ...Option) bool {
	type IName interface{ GetName() string }
	if iName, ok := pStruct.(IName); ok {
		defaultSet.Init(iName.GetName(), pflag.ExitOnError)
	}

	type IVersion interface{ GetVersion() string }
	if iVersion, ok := pStruct.(IVersion); ok {
		defaultSet.version = iVersion.GetVersion()
	}

	type IBuildTime interface{ GetBuildTime() time.Time }
	if iBuild, ok := pStruct.(IBuildTime); ok {
		defaultSet.buildTime = iBuild.GetBuildTime()
	}

	type IDescription interface{ GetDescription() string }
	if iDescription, ok := pStruct.(IDescription); ok {
		defaultSet.description = iDescription.GetDescription()
	}

	for _, option := range options {
		option(defaultSet)
	}

	if err := StructToSet(defaultSet.pSet, pStruct); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: apply struct to flag: %s", err.Error())
		os.Exit(1)
	}
	return defaultSet.Parse()
}

func Var(v any, name string, itemOptions ...ItemOption) *FlagItem {
	return defaultSet.Var(v, name, itemOptions...)
}
