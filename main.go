package main

import (
	"embed"
	"log"

	"github.com/soulteary/ip-helper/model/define"
	"github.com/soulteary/ip-helper/model/ftp"
	ipInfo "github.com/soulteary/ip-helper/model/ip-info"
	configParser "github.com/soulteary/ip-helper/model/parse-config"
	"github.com/soulteary/ip-helper/model/telnet"
	"github.com/soulteary/ip-helper/model/web"
)

//go:embed public
var EmbedFS embed.FS

func main() {
	config := configParser.Parse()

	ipdb, err := ipInfo.InitIPDB("./data/ipipfree.ipdb")
	if err != nil {
		log.Fatalf("初始化 IP 数据库失败: %v\n", err)
		return
	}

	go telnet.Server(&ipdb, define.TELNET_PORT)
	go ftp.Server(&ipdb, define.FTP_PORT)
	web.Server(config, &ipdb)
}
