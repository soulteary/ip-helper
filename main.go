package main

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	static "github.com/soulteary/gin-static"
)

type Config struct {
	Domain string
	Port   string
	Token  string
}

func parseConfig() *Config {
	config := &Config{}

	flag.StringVar(&config.Port, "port", "", "服务器端口")
	flag.StringVar(&config.Domain, "domain", "", "服务器域名")
	flag.StringVar(&config.Token, "token", "", "API 访问令牌")
	flag.Parse()

	if config.Port == "" {
		config.Port = os.Getenv("SERVER_PORT")
	}
	if config.Domain == "" {
		config.Domain = os.Getenv("SERVER_DOMAIN")
	}
	if config.Token == "" {
		config.Token = os.Getenv("TOKEN")
	}

	if config.Port == "" {
		config.Port = "8080"
	}
	if config.Domain == "" {
		config.Domain = "localhost"
	}
	if config.Token == "" {
		config.Token = ""
		log.Println("提醒：为了提高安全性，可以设置 `TOKEN` 环境变量或 `token` 命令行参数")
	}

	return config
}

func authMiddleware(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.Token != "" {
			token := c.Query("token")
			if token == "" {
				token = c.GetHeader("X-Token")
			}
			if token != config.Token {
				c.JSON(401, gin.H{"error": "无效的认证令牌"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func Get(link string) ([]byte, error) {
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务器返回非200状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应内容失败: %v", err)
	}
	return body, nil
}

// IPInfo 存储 IP 相关信息
type IPInfo struct {
	ClientIP     string `json:"client_ip"`
	ProxyIP      string `json:"proxy_ip,omitempty"`
	IsProxy      bool   `json:"is_proxy"`
	ForwardedFor string `json:"forwarded_for,omitempty"`
	RealIP       string `json:"real_ip"`
}

// 获取并分析 IP 信息的中间件
func IPAnalyzer() gin.HandlerFunc {
	return func(c *gin.Context) {
		ipInfo := analyzeIP(c)
		// 将 IP 信息存储到上下文中
		c.Set("ip_info", ipInfo)
		c.Next()
	}
}

// 分析 IP 信息
func analyzeIP(c *gin.Context) IPInfo {
	var ipInfo IPInfo

	// 获取客户端 IP
	ipInfo.ClientIP = c.ClientIP()

	// 获取 X-Forwarded-For 头信息
	forwardedFor := c.GetHeader("X-Forwarded-For")
	if forwardedFor != "" {
		ipInfo.ForwardedFor = forwardedFor
		// X-Forwarded-For 可能包含多个 IP，第一个是原始客户端 IP
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			ipInfo.RealIP = strings.TrimSpace(ips[0])
			if len(ips) > 1 {
				ipInfo.IsProxy = true
				ipInfo.ProxyIP = strings.TrimSpace(ips[len(ips)-1])
			}
		}
	} else {
		ipInfo.RealIP = ipInfo.ClientIP
	}

	// 获取 X-Real-IP 头信息
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" && xRealIP != ipInfo.RealIP {
		ipInfo.IsProxy = true
		ipInfo.ProxyIP = ipInfo.ClientIP
		ipInfo.RealIP = xRealIP
	}

	// 检查是否为私有 IP
	if isPrivateIP(ipInfo.ClientIP) {
		ipInfo.IsProxy = true
	}

	return ipInfo
}

// 检查是否为私有 IP 地址
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// 检查是否为私有 IP 范围
	privateIPRanges := []struct {
		start net.IP
		end   net.IP
	}{
		{net.ParseIP("10.0.0.0"), net.ParseIP("10.255.255.255")},
		{net.ParseIP("172.16.0.0"), net.ParseIP("172.31.255.255")},
		{net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255")},
	}

	for _, r := range privateIPRanges {
		if bytes.Compare(ip, r.start) >= 0 && bytes.Compare(ip, r.end) <= 0 {
			return true
		}
	}
	return false
}

//go:embed public
var EmbedFS embed.FS

func main() {
	config := parseConfig()

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"domain": config.Domain,
		})
	})

	r.Use(static.Serve("/", static.LocalFile("./public", false)))

	r.Use(authMiddleware(config))
	// 使用IP分析中间件
	r.Use(IPAnalyzer())
	r.GET("/", func(c *gin.Context) {
		// 先获取 IP 信息
		ipInfo, exists := c.Get("ip_info")
		if !exists {
			c.JSON(500, gin.H{"error": "IP info not found"})
			return
		}

		buf, err := Get(fmt.Sprintf("http://localhost:%s/index.template.html", config.Port))
		if err != nil {
			c.String(500, "读取模板文件失败: %v", err)
			return
		}

		// TODO 将 IP 信息传递给模板
		fmt.Println(ipInfo)

		c.Data(200, "text/html; charset=utf-8", buf)
	})
	// 单独提供一个接口，来获取 IP 信息
	r.GET("/ip", func(c *gin.Context) {
		ipInfo, exists := c.Get("ip_info")
		if !exists {
			c.JSON(500, gin.H{"error": "IP info not found"})
			return
		}
		c.JSON(200, ipInfo)
	})

	serverAddr := fmt.Sprintf(":%s", config.Port)
	log.Printf("启动服务器于 %s:%s\n", "config.Domain", config.Port)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}

	r.Run(":8080")
}
