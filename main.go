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

func isValidIPAddress(ip string) bool {
	if parsedIP := net.ParseIP(ip); parsedIP != nil {
		return true
	}
	return false
}

// 判断请求发起方是否为“下载工具”
func IsDownloadTool(userAgent string) bool {
	// 转换为小写以便不区分大小写比较
	ua := strings.ToLower(userAgent)

	// 常见下载工具的特征字符串
	downloadTools := []string{
		"curl",
		"wget",
		"aria2",
		"python-requests",
		"axios",
		"got",
		"postman",
	}

	for _, tool := range downloadTools {
		if strings.Contains(ua, tool) {
			return true
		}
	}

	return false
}

//go:embed public
var EmbedFS embed.FS

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

	getClientIPInfo := func(c *gin.Context, ipaddr string) (resultIP string, resultDBInfo []string, err error) {
		if ipaddr == "" {
			ipInfo, exists := c.Get("ip_info")
			if !exists {
				return resultIP, resultDBInfo, fmt.Errorf("IP info not found")
			}
			ipaddr = ipInfo.(IPInfo).RealIP
		}

		dbInfo, err := db.Find(ipaddr, "CN")
		if err != nil {
			dbInfo = []string{"未找到 IP 地址信息"}
		}
		dbInfo = removeDuplicates(dbInfo)
		return ipaddr, dbInfo, nil
	}

	renderTemplate := func(globalTemplate []byte, ipaddr string, dbInfo []string) []byte {
		template := bytes.ReplaceAll(globalTemplate, []byte("%IP_ADDR%"), []byte(ipaddr))
		template = bytes.ReplaceAll(template, []byte("%DOMAIN%"), []byte(config.Domain))
		template = bytes.ReplaceAll(template, []byte("%DATA_1_INFO%"), []byte(strings.Join(removeDuplicates(dbInfo), " ")))
		return template
	}

	renderJSON := func(ipaddr string, dbInfo []string) map[string]any {
		return map[string]any{"ip": ipaddr, "info": dbInfo}
	}

	globalTemplate := []byte{}

	r.GET("/", func(c *gin.Context) {
		if len(globalTemplate) == 0 {
			globalTemplate, err = Get(fmt.Sprintf("http://localhost:%s/index.template.html", config.Port))
			if err != nil {
				log.Fatalf("读取模板文件失败: %v\n", err)
				return
			}
		}

		ipAddr, dbInfo, err := getClientIPInfo(c, "")
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// 获取请求头中的 User-Agent 信息
		userAgent := c.GetHeader("User-Agent")
		// 使用下载工具访问时返回 JSON 格式
		if IsDownloadTool(userAgent) {
			c.JSON(200, renderJSON(ipAddr, dbInfo))
		} else {
			c.Data(200, "text/html; charset=utf-8", renderTemplate(globalTemplate, ipAddr, dbInfo))
		}
	})

	r.POST("/", func(c *gin.Context) {
		ipInfo, exists := c.Get("ip_info")
		if !exists {
			c.JSON(500, gin.H{"error": "IP info not found"})
			return
		}
		ip := ""
		var form IPForm
		if err := c.ShouldBind(&form); err != nil {
			ip = ipInfo.(IPInfo).RealIP
		} else {
			ip = form.IP
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
		ip := c.Param("ip")
		ipAddr, dbInfo, err := getClientIPInfo(c, ip)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		// 获取请求头中的 User-Agent 信息
		userAgent := c.GetHeader("User-Agent")
		// 使用下载工具访问时返回 JSON 格式
		if IsDownloadTool(userAgent) {
			c.JSON(200, renderJSON(ipAddr, dbInfo))
		} else {
			c.Data(200, "text/html; charset=utf-8", renderTemplate(globalTemplate, ipAddr, dbInfo))
		}
	})

	serverAddr := fmt.Sprintf(":%s", config.Port)
	log.Printf("启动服务器于 %s:%s\n", "config.Domain", config.Port)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}

	r.Run(":8080")
}
