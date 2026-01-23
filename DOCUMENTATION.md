# flags 库详细文档

## 概述

flags 是一个对 spf13/pflag 库的封装，提供了更简洁的命令行参数解析 API。它简化了命令行参数的定义和解析过程，并提供了一些高级功能。

## 核心概念

### FlagSet

`FlagSet` 是核心类型，封装了 `pflag.FlagSet` 并扩展了以下功能：

- `Name`: FlagSet 名称
- `Description`: 描述信息
- `Version`: 版本号
- `BuildTime`: 构建时间
- `envKeys`: 存储环境变量键名映射

### Option 函数

Option 是一种配置模式，用于在创建 FlagSet 时配置各种属性：

- `Version(version string)`: 设置版本号
- `Description(desc string)`: 设置描述信息
- `BuildTime(buildTime T)`: 设置构建时间（支持 string、time.Time、int、int64 类型）

## 使用方法

### 1. 基本用法

```go
var verbose bool
var count int

fs := flags.NewSet(
    flags.Version("1.0.0"),
    flags.Description("Example application"),
)

fs.Var(&verbose, "verbose", "v", "Enable verbose output")
fs.Var(&count, "count", "c", "Number of iterations")

if fs.Parse() {
    fmt.Printf("Verbose: %t, Count: %d\n", verbose, count)
}
```

### 2. 使用结构体标签

```go
type Config struct {
    Verbose bool   `flag:"verbose,v" usage:"Enable verbose output"`
    Count   int    `flag:"count,c" usage:"Number of iterations"`
    Output  string `flag:"output,o" usage:"Output file path"`
}

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
```

### 3. 环境变量支持

```go
var host string = "localhost"
var port int = 8080

fs := flags.NewSet()
fs.Var(&host, "host", "h", "Server host", "SERVER_HOST")  // 从环境变量 SERVER_HOST 获取值
fs.Var(&port, "port", "p", "Server port", "SERVER_PORT") // 从环境变量 SERVER_PORT 获取值

if fs.Parse() {
    fmt.Printf("Server: %s:%d\n", host, port)
}
```

## 支持的数据类型

flags 库支持以下数据类型：

### 基本类型
- `bool`
- `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `time.Duration`
- `net.IP`, `net.IPNet`

### 切片类型
- `[]bool`
- `[]string`
- `[]int`, `[]int32`, `[]int64`
- `[]uint`
- `[]float32`, `[]float64`
- `[]time.Duration`
- `[]net.IP`, `[]net.IPNet`

## 高级功能

### 1. 自定义名称转换

当使用 `Struct` 方法且未明确指定标志名称时，系统会自动将驼峰命名转换为小写下划线分隔的形式：

```go
type Config struct {
    MaxConnections int // -> --max_connections
    ServerPort     int // -> --server_port
}
```

### 2. 弃用标志支持

可以通过在使用说明中使用 `**DEPRECATED**` 来标记弃用的标志：

```go
fs.StringVar(&oldFlag, "old-flag", "", "Old flag **DEPRECATED** Use new-flag instead")
```

### 3. 自定义解析

`ParseFrom` 方法允许你从自定义的参数切片中解析：

```go
args := []string{"--verbose", "--count", "5"}
err := fs.ParseFrom(args)
```

## 最佳实践

1. **使用结构体标签**: 当有多个相关参数时，使用结构体标签可以使代码更清晰
2. **提供描述信息**: 使用 `Description` 选项为应用程序提供有意义的描述
3. **利用环境变量**: 在容器化环境中，环境变量是传递配置的重要途径
4. **设置版本信息**: 为应用程序提供版本信息便于管理和调试

## 错误处理

`Parse()` 方法返回布尔值表示解析是否成功，而 `ParseFrom()` 返回具体的错误信息，可根据需要选择使用。