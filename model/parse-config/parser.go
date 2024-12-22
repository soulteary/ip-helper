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

	flag.BoolVar(&config.Debug, "debug", false, "调试模式")
	flag.StringVar(&config.Port, "port", "", "服务器端口")
	flag.StringVar(&config.Domain, "domain", "", "服务器域名")
	flag.StringVar(&config.Token, "token", "", "API 访问令牌")
	flag.Parse()

	if !config.Debug {
		config.Debug = strings.ToLower(os.Getenv("DEBUG")) == "true"
	}
	if config.Port == "" {
		config.Port = os.Getenv("SERVER_PORT")
	}
	if config.Domain == "" {
		config.Domain = os.Getenv("SERVER_DOMAIN")
	}
	if config.Token == "" {
		config.Token = os.Getenv("TOKEN")
	}

	if config.Debug {
		log.Println("调试模式已开启")
	}
	if config.Port == "" {
		config.Port = "8080"
	}
	if config.Domain == "" {
		config.Domain = "http://localhost:8080"
	}
	if config.Token == "" {
		config.Token = ""
		log.Println("提醒：为了提高安全性，可以设置 `TOKEN` 环境变量或 `token` 命令行参数")
	}

	return config
}
