// main.go
// 演示如何使用flags库的各种功能

package main

import (
	"fmt"
	"net"
	"time"

	"github.com/cnk3x/flags"
)

// 示例结构体，展示如何使用Struct方法
type Config struct {
	Host    string        `flag:"host,h" usage:"服务器主机地址**DEPRECATED** 使用 --listen 替代" env:"SERVER_HOST"`
	Port    int           `flag:"port,p" usage:"服务器端口号**DEPRECATED** 使用 --listen 替代" env:"SERVER_PORT"`
	Timeout time.Duration `flag:"t" usage:"请求超时时间"`
	Verbose bool          `flag:"v" usage:"启用详细日志"`
	Numbers []int         `flag:"numbers,n" usage:"数字列表"`
	IP      net.IP        `flag:"ip,i" usage:"IP地址"`
	Listen  string        `flag:"l" usage:"监听地址"`
}

func main() {
	// 方式一：使用结构体定义
	cfg := &Config{
		Host:    "localhost",
		Port:    8080,
		Timeout: 30 * time.Second,
		Numbers: []int{1, 2, 3},
		IP:      net.ParseIP("127.0.0.1"),
	}

	fs := flags.NewSet(
		flags.Version("1.0.0"),
		flags.Description("这是一个演示flags库功能的示例程序"),
		flags.BuildTime(time.Now()),
	)

	// 使用Struct方法从结构体标签生成命令行参数
	fs.Struct(cfg)

	// 方式二：手动添加参数
	var debug bool
	fs.Var(&debug, "debug", "d", "启用调试模式")

	// 解析命令行参数
	if fs.Parse() {
		fmt.Printf("配置信息:\n")
		fmt.Printf("- Host: %s\n", cfg.Host)
		fmt.Printf("- Port: %d\n", cfg.Port)
		fmt.Printf("- Timeout: %s\n", cfg.Timeout)
		fmt.Printf("- Verbose: %t\n", cfg.Verbose)
		fmt.Printf("- Numbers: %v\n", cfg.Numbers)
		fmt.Printf("- IP: %s\n", cfg.IP)
		fmt.Printf("- Debug: %t\n", debug)
	}
}
