package telnet

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"github.com/soulteary/ip-helper/model/fn"
	ipInfo "github.com/soulteary/ip-helper/model/ip-info"
	"github.com/soulteary/ip-helper/model/response"
)

func Server(ipdb *ipInfo.IPDB, port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("TELNET 服务器启动失败: %v", err)
	}
	defer listener.Close()

	info := ipdb.FindByIPIP("127.0.0.1")
	if len(info) == 0 {
		return fmt.Errorf("IP 数据库加载失败")
	}

	log.Println("TELNET 服务器已启动，监听端口:", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("TELNET 服务器接受连接时发生错误: %v\n", err)
			continue
		}
		go HandleConnection(ipdb, conn)
	}
}

func HandleConnection(ipdb *ipInfo.IPDB, conn net.Conn) {
	defer conn.Close()

	clientIP := fn.GetBaseIP(conn.RemoteAddr().String())
	info := ipdb.FindByIPIP(clientIP)

	sendBuf := [][]byte{
		response.RenderJSON(clientIP, info),
		[]byte("\r\n"),
	}
	_, err := conn.Write(bytes.Join(sendBuf, []byte("")))
	if err != nil {
		fmt.Printf("TELNET 服务发送消息时发生错误: %v\n", err)
		return
	}
}
