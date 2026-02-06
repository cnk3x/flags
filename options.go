package flags

import (
	"time"

	"github.com/spf13/pflag"
)

// Option 函数类型，用于配置 FlagSet 的选项
type Option func(*Set)

// Name 返回一个 Option，用于设置 FlagSet 的名称
func Name(name string) Option { return func(fs *Set) { fs.Init(name, pflag.ExitOnError) } }

// Version 返回一个 Option，用于设置 FlagSet 的版本号
func Version(version string) Option { return func(fs *Set) { fs.version = version } }

// Description 返回一个 Option，用于设置 FlagSet 的描述信息
func Description(desc string) Option { return func(fs *Set) { fs.description = desc } }

// BuildTime 返回一个 Option，用于设置 FlagSet 的构建时间
//   - 支持 string、time.Time、int、int64 类型的时间表示
func BuildTime[T string | time.Time | int | int64](buildTime T) Option {
	return func(fs *Set) {
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
