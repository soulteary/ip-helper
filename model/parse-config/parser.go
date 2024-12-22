package configParser

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/soulteary/ip-helper/model/define"
)

func Parse() *define.Config {
	config := &define.Config{}

	// 先读取环境变量
	debug := strings.ToLower(os.Getenv("DEBUG"))
	port := os.Getenv("SERVER_PORT")
	domain := os.Getenv("SERVER_DOMAIN")
	token := os.Getenv("TOKEN")

	// 设置命令行参数默认值，如果环境变量存在则使用环境变量值
	defaultDebug := debug == "true"
	defaultPort := "8080"
	if port != "" {
		defaultPort = port
	}
	defaultDomain := "http://localhost:8080"
	if domain != "" {
		defaultDomain = domain
	}
	defaultToken := token

	// 解析命令行参数，会覆盖环境变量的值
	flag.BoolVar(&config.Debug, "debug", defaultDebug, "调试模式")
	flag.StringVar(&config.Port, "port", defaultPort, "服务器端口")
	flag.StringVar(&config.Domain, "domain", defaultDomain, "服务器域名")
	flag.StringVar(&config.Token, "token", defaultToken, "API 访问令牌")
	flag.Parse()

	// 处理特殊的空值情况
	if config.Port == "" {
		config.Port = "8080"
	}
	if config.Domain == "" {
		config.Domain = "http://localhost:8080"
	}

	// 输出相关日志
	if config.Debug {
		log.Println("调试模式已开启")
	}
	if config.Token == "" {
		log.Println("提醒：为了提高安全性，可以设置 `TOKEN` 环境变量或 `token` 命令行参数")
	}

	return config
}
