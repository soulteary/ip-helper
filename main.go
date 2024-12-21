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
	"github.com/soulteary/ipdb-go"
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

type IPInfo struct {
	ClientIP     string `json:"client_ip"`
	ProxyIP      string `json:"proxy_ip,omitempty"`
	IsProxy      bool   `json:"is_proxy"`
	ForwardedFor string `json:"forwarded_for,omitempty"`
	RealIP       string `json:"real_ip"`
}

func IPAnalyzer() gin.HandlerFunc {
	return func(c *gin.Context) {
		ipInfo := analyzeIP(c)
		c.Set("ip_info", ipInfo)
		c.Next()
	}
}

func analyzeIP(c *gin.Context) IPInfo {
	var ipInfo IPInfo

	ipInfo.ClientIP = c.ClientIP()

	forwardedFor := c.GetHeader("X-Forwarded-For")
	if forwardedFor != "" {
		ipInfo.ForwardedFor = forwardedFor
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

	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" && xRealIP != ipInfo.RealIP {
		ipInfo.IsProxy = true
		ipInfo.ProxyIP = ipInfo.ClientIP
		ipInfo.RealIP = xRealIP
	}

	if isPrivateIP(ipInfo.ClientIP) {
		ipInfo.IsProxy = true
	}

	return ipInfo
}

func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

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

func removeDuplicates(strSlice []string) []string {
	encountered := make(map[string]bool)
	result := []string{}

	for _, str := range strSlice {
		if !encountered[str] {
			encountered[str] = true
			result = append(result, str)
		}
	}
	return result
}

// 使用 net 包验证 IP 地址
func isValidIPAddress(ip string) bool {
	if parsedIP := net.ParseIP(ip); parsedIP != nil {
		return true
	}
	return false
}

//go:embed public
var EmbedFS embed.FS

// IPForm 定义表单结构
type IPForm struct {
	IP string `form:"ip" binding:"required"`
}

func main() {
	config := parseConfig()

	db, err := ipdb.NewCity("./data/ipipfree.ipdb")
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"domain": config.Domain,
		})
	})

	r.Use(static.Serve("/", static.LocalFile("./public", false)))

	r.Use(authMiddleware(config))
	r.Use(IPAnalyzer())
	r.GET("/", func(c *gin.Context) {
		ipInfo, exists := c.Get("ip_info")
		if !exists {
			c.JSON(500, gin.H{"error": "IP info not found"})
			return
		}

		// 查询 IP 地址具体信息
		dbInfo, err := db.Find(ipInfo.(IPInfo).RealIP, "CN")
		if err != nil {
			dbInfo = []string{"未找到 IP 地址信息"}
		}
		// 读取默认模版
		template, err := Get(fmt.Sprintf("http://localhost:%s/index.template.html", config.Port))
		if err != nil {
			c.String(500, "读取模板文件失败: %v", err)
			return
		}
		// 更新模版中的 IP 地址
		template = bytes.ReplaceAll(template, []byte("%IP_ADDR%"), []byte(ipInfo.(IPInfo).ClientIP))
		// 更新模版中的域名
		template = bytes.ReplaceAll(template, []byte("%DOMAIN%"), []byte(config.Domain))
		// 更新模版中的 IP 地址信息
		template = bytes.ReplaceAll(template, []byte("%DATA_1_INFO%"), []byte(strings.Join(removeDuplicates(dbInfo), " ")))

		c.Data(200, "text/html; charset=utf-8", template)
	})
	// 处理 POST 请求，解析表单数据
	r.POST("/", func(c *gin.Context) {
		// 获取请求中的 IP 地址信息
		ipInfo, exists := c.Get("ip_info")
		if !exists {
			c.JSON(500, gin.H{"error": "IP info not found"})
			return
		}
		// 默认 IP 地址为空
		ip := ""
		var form IPForm
		// 使用 ShouldBind 绑定表单数据
		if err := c.ShouldBind(&form); err != nil {
			// 如果绑定失败，使用请求中的 IP 地址
			ip = ipInfo.(IPInfo).RealIP
		} else {
			// 获取到 IP 地址后的处理逻辑
			ip = form.IP
			// 如果 IP 地址不合法，使用请求中的 IP 地址
			if !isValidIPAddress(ip) {
				ip = ipInfo.(IPInfo).RealIP
			}
		}
		c.Redirect(302, fmt.Sprintf("/ip/%s", ip))
	})

	r.GET("/ip", func(c *gin.Context) {
		ipInfo, exists := c.Get("ip_info")
		if !exists {
			c.JSON(500, gin.H{"error": "IP info not found"})
			return
		}
		c.JSON(200, ipInfo)
	})
	r.GET("/ip/:ip", func(c *gin.Context) {
		ipaddr := c.Param("ip")
		if ipaddr == "" {
			ipInfo, exists := c.Get("ip_info")
			if !exists {
				c.JSON(500, gin.H{"error": "IP info not found"})
				return
			}
			ipaddr = ipInfo.(IPInfo).RealIP
		}

		dbInfo, err := db.Find(ipaddr, "CN")
		if err != nil {
			dbInfo = []string{"未找到 IP 地址信息"}
		}
		dbInfo = removeDuplicates(dbInfo)
		c.JSON(200, map[string]any{"ip": ipaddr, "info": dbInfo})
	})

	serverAddr := fmt.Sprintf(":%s", config.Port)
	log.Printf("启动服务器于 %s:%s\n", "config.Domain", config.Port)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}

	r.Run(":8080")
}
