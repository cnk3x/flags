# flags

flags 是一个对 [spf13/pflag](https://github.com/spf13/pflag) 库的简单封装，旨在简化命令行参数解析的使用方式。

## 功能特点

- 封装 pflag 提供更简洁的 API
- 支持全局命令行标志（flags）的定义与解析
- 提供默认的 CommandLine 实例管理
- 支持从结构体标签自动生成命令行标志
- 支持环境变量自动填充
- 支持版本、描述和构建时间等元数据

## 安装

```bash
go get github.com/cnk3x/flags
```

## 使用示例

### 基本用法

```go
package main

import (
    "fmt"
    "github.com/cnk3x/flags"
)

func main() {
    var verbose bool
    var count int
    
    fs := flags.NewSet(
        flags.Version("1.0.0"),
        flags.Description("Example application demonstrating flags usage"),
    )
    
    fs.Var(&verbose, "verbose", "v", "Enable verbose output")
    fs.Var(&count, "count", "c", "Number of iterations")
    
    if fs.Parse() {
        fmt.Printf("Verbose: %t, Count: %d\n", verbose, count)
    }
}
```

### 使用结构体标签

```go
package main

import (
    "fmt"
    "github.com/cnk3x/flags"
)

type Config struct {
    Verbose bool   `flag:"verbose,v" usage:"Enable verbose output"`
    Count   int    `flag:"count,c" usage:"Number of iterations"`
    Output  string `flag:"output,o" usage:"Output file path"`
}

func main() {
    cfg := &Config{
        Verbose: false,
        Count:   1,
        Output:  "output.txt",
    }
    
    fs := flags.NewSet(flags.Description("Struct example"))
    fs.Struct(cfg)
    
    if fs.Parse() {
        fmt.Printf("Config: %+v\n", cfg)
    }
}
```

### 环境变量支持

```go
package main

import (
    "fmt"
    "github.com/cnk3x/flags"
)

func main() {
    var host string = "localhost"
    var port int = 8080
    
    fs := flags.NewSet()
    fs.Var(&host, "host", "h", "Server host", "SERVER_HOST")  // 从环境变量 SERVER_HOST 获取值
    fs.Var(&port, "port", "p", "Server port", "SERVER_PORT") // 从环境变量 SERVER_PORT 获取值
    
    if fs.Parse() {
        fmt.Printf("Server: %s:%d\n", host, port)
    }
}
```

## API 文档

### `NewSet(options ...Option) *FlagSet`

创建一个新的 FlagSet 实例，可选地应用一些配置选项。

### `Version(version string) Option`

设置应用程序版本号。

### `Description(desc string) Option`

设置应用程序描述信息。

### `BuildTime[T string | time.Time | int | int64](buildTime T) Option`

设置构建时间，支持多种时间格式。

### `(*FlagSet) Parse() bool`

解析命令行参数，如果解析成功返回 true。

### `(*FlagSet) ParseFrom(args []string) (err error)`

从指定参数切片解析命令行参数。

### `(*FlagSet) Struct(structPtr any)`

从结构体标签生成并注册命令行标志。

### `(*FlagSet) Var(v any, name, short, usage string, env ...string)`

添加一个命令行标志，支持完整名称、短名称、使用说明和环境变量。

## 支持的数据类型

- `time.Duration`
- `net.IP`, `net.IPNet`
- `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `bool`
- 以上类型的切片形式

## 许可证

MIT License