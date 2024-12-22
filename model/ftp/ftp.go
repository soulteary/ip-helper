package ftp

import (
	"bytes"
	"log"
	"net"

	"github.com/soulteary/ip-helper/model/define"
	"github.com/soulteary/ip-helper/model/fn"
	ipInfo "github.com/soulteary/ip-helper/model/ip-info"
	"github.com/soulteary/ip-helper/model/response"
)

func Server(ipdb *ipInfo.IPDB) {
	listener, err := net.Listen("tcp", define.FTP_PORT)
	if err != nil {
		log.Fatalf("FTP 服务器启动失败: %v\n", err)
	}
	defer listener.Close()

	log.Println("FTP 服务器已启动，监听端口:", define.FTP_PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("FTP 服务器接受连接时发生错误: %v\n", err)
			continue
		}
		go handleConnection(ipdb, conn)
	}
}

func handleConnection(ipdb *ipInfo.IPDB, conn net.Conn) {
	defer conn.Close()

	clientIP := fn.GetBaseIP(conn.RemoteAddr().String())
	info := ipdb.FindByIPIP(clientIP)

	sendBuf := [][]byte{
		[]byte("220"),
		response.RenderJSON(clientIP, info),
		[]byte("\r\n"),
	}
	_, err := conn.Write(bytes.Join(sendBuf, []byte(" ")))
	if err != nil {
		log.Println("发送消息时发生错误: ", err)
		return
	}
	conn.Close()
}
