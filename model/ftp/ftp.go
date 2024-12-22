package ftp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/soulteary/ip-helper/model/fn"
	ipInfo "github.com/soulteary/ip-helper/model/ip-info"
	"github.com/soulteary/ip-helper/model/response"
)

func Server(ipdb *ipInfo.IPDB) {
	listener, err := net.Listen("tcp", ":21")
	if err != nil {
		log.Fatalf("Error creating server: %v", err)
	}
	defer listener.Close()

	log.Println("FTP Server listening on :21")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
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
	sendBuf = append(sendBuf, []byte("220"))
	sendBuf = append(sendBuf, message)
	sendBuf = append(sendBuf, []byte("\r\n"))
	_, err = conn.Write(bytes.Join(sendBuf, []byte(" ")))
	if err != nil {
		log.Println("发送消息时发生错误: ", err)
		return
	}
	conn.Close()
}
