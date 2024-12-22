package ftp_test

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/soulteary/ip-helper/model/ftp"
	ipInfo "github.com/soulteary/ip-helper/model/ip-info"
)

// 创建一个辅助函数来获取可用的端口
func getFreePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", ":0")
	if err != nil {
		return "", err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	defer l.Close()
	return fmt.Sprintf(":%d", l.Addr().(*net.TCPAddr).Port), nil
}

func GetIPDB() (*ipInfo.IPDB, error) {
	workDir, _ := os.Getwd()
	dbPath := filepath.Join(workDir, "../../data/ipipfree.ipdb")

	ipdb, err := ipInfo.InitIPDB(dbPath)
	if err != nil {
		return nil, err
	}
	return &ipdb, nil
}

// TestServerStartupError 测试服务器启动失败的场景
// TestServerStartupError 测试服务器启动失败的场景
func TestServerStartupError(t *testing.T) {
	// 获取测试端口
	testPort, err := getFreePort()
	if err != nil {
		t.Fatalf("无法获取测试端口: %v", err)
	}

	// 先启动一个占用端口的服务器
	listener, err := net.Listen("tcp", testPort)
	if err != nil {
		t.Fatalf("无法启动测试服务器: %v", err)
	}
	defer listener.Close()

	ipdb, err := GetIPDB()
	if err != nil {
		t.Fatalf("初始化 IP 数据库失败: %v", err)
	}

	// 创建一个buffer来捕获日志输出
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr) // 测试结束后恢复标准输出

	// 启动服务器
	err = ftp.Server(ipdb, testPort)
	if err == nil {
		t.Error("期望服务器启动失败，但是成功了")
	}

	// 验证错误信息
	if !strings.Contains(err.Error(), "FTP 服务器启动失败") {
		t.Errorf("错误消息不符合预期，得到: %v", err)
	}
}

// TestSuccessfulServerStartup 测试服务器成功启动
func TestSuccessfulServerStartup(t *testing.T) {
	testPort, err := getFreePort()
	if err != nil {
		t.Fatalf("无法获取测试端口: %v", err)
	}

	ipdb, err := GetIPDB()
	if err != nil {
		t.Fatalf("初始化 IP 数据库失败: %v", err)
	}

	// 在 goroutine 中启动服务器
	go func() {
		ftp.Server(ipdb, testPort)
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 尝试连接服务器
	conn, err := net.Dial("tcp", "localhost"+testPort)
	if err != nil {
		t.Fatalf("无法连接到服务器: %v", err)
	}
	defer conn.Close()
}

// TestConnectionHandling 测试连接处理和响应
func TestConnectionHandling(t *testing.T) {
	testPort, err := getFreePort()
	if err != nil {
		t.Fatalf("无法获取测试端口: %v", err)
	}

	ipdb, err := GetIPDB()
	if err != nil {
		t.Fatalf("初始化 IP 数据库失败: %v", err)
	}

	// 启动服务器
	go func() {
		ftp.Server(ipdb, testPort)
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 连接服务器
	conn, err := net.Dial("tcp", "localhost"+testPort)
	if err != nil {
		t.Fatalf("无法连接到服务器: %v", err)
	}
	defer conn.Close()

	// 读取响应
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("读取响应失败: %v", err)
	}

	response := string(buf[:n])

	// 验证响应格式
	if !strings.HasPrefix(response, "220") {
		t.Errorf("响应应该以 220 开头，实际收到: %s", response)
	}

	// 验证响应中包含 JSON
	if !strings.Contains(response, "{") || !strings.Contains(response, "}") {
		t.Error("响应中应该包含 JSON 数据")
	}

	// 验证响应以 \r\n 结尾
	if !strings.HasSuffix(response, "\r\n") {
		t.Error("响应应该以 \\r\\n 结尾")
	}
}

// TestConcurrentConnections 测试并发连接
func TestConcurrentConnections(t *testing.T) {
	testPort, err := getFreePort()
	if err != nil {
		t.Fatalf("无法获取测试端口: %v", err)
	}

	ipdb, err := GetIPDB()
	if err != nil {
		t.Fatalf("初始化 IP 数据库失败: %v", err)
	}

	// 启动服务器
	go func() {
		ftp.Server(ipdb, testPort)
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 并发测试
	concurrentConnections := 5
	done := make(chan bool, concurrentConnections)

	for i := 0; i < concurrentConnections; i++ {
		go func() {
			conn, err := net.Dial("tcp", "localhost"+testPort)
			if err != nil {
				t.Errorf("并发连接失败: %v", err)
				done <- false
				return
			}
			defer conn.Close()

			buf := make([]byte, 1024)
			_, err = conn.Read(buf)
			if err != nil {
				t.Errorf("读取响应失败: %v", err)
				done <- false
				return
			}

			done <- true
		}()
	}

	// 等待所有连接完成
	for i := 0; i < concurrentConnections; i++ {
		success := <-done
		if !success {
			t.Error("一个或多个并发连接失败")
		}
	}
}

// TestInvalidIPDB 测试无效的 IPDB
func TestInvalidIPDB(t *testing.T) {
	testPort, err := getFreePort()
	if err != nil {
		t.Fatalf("无法获取测试端口: %v", err)
	}

	// 创建一个空的 IPDB
	var ipdb *ipInfo.IPDB

	// 使用 channel 来捕获 panic
	done := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- true
			}
		}()
		ftp.Server(ipdb, testPort)
	}()

	select {
	case <-done:
		// 服务器应该因为无效的 IPDB 而失败
	case <-time.After(time.Second):
		t.Error("服务器应该因为无效的 IPDB 而失败，但没有")
	}
}
