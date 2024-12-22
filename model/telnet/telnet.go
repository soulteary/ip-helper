package telnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"

	"github.com/soulteary/ip-helper/model/fn"
	ipInfo "github.com/soulteary/ip-helper/model/ip-info"
	"github.com/soulteary/ip-helper/model/response"
)

func Server(ipdb *ipInfo.IPDB) {
	listener, err := net.Listen("tcp", ":23")
	if err != nil {
		fmt.Printf("无法启动telnet服务器: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Println("Telnet服务器已启动，监听端口 23...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("接受连接时发生错误: %v\n", err)
			continue
		}
		go handleConnection(ipdb, conn)
	}
}

func handleConnection(ipdb *ipInfo.IPDB, conn net.Conn) {
	defer conn.Close()

	clientIP := fn.GetBaseIP(conn.RemoteAddr().String())
	info := ipdb.FindByIPIP(clientIP)

	sendBuf := [][]byte{}
	message, err := json.Marshal(response.RenderJSON(clientIP, info))
	if err != nil {
		fmt.Println("序列化 JSON 数据时发生错误: ", err)
		return
	}

	sendBuf = append(sendBuf, message)
	sendBuf = append(sendBuf, []byte("\r\n"))
	_, err = conn.Write(bytes.Join(sendBuf, []byte("")))
	if err != nil {
		fmt.Printf("发送消息时发生错误: %v\n", err)
		return
	}
}
