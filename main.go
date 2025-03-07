/**
 * Copyright (c) 2024-2025 Su Yang (soulteary)
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"embed"
	"fmt"
	"os"

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
		fmt.Println("初始化 IP 数据库失败: ", err)
		fmt.Println("请自行准备数据库文件，并确保 data/ipipfree.ipdb 路径文件存在")
		os.Exit(1)
		return
	}

	go telnet.Server(&ipdb, define.TELNET_PORT)
	go ftp.Server(&ipdb, define.FTP_PORT)
	web.Server(config, &ipdb)
}
