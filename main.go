package main

import (
	"bytes"
	"crypto/md5"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/gzip"
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

func IsDownloadTool(userAgent string) bool {
	ua := strings.ToLower(userAgent)

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

func cacheMiddleware() gin.HandlerFunc {
	data := []byte(time.Now().String())
	etag := fmt.Sprintf("W/%x", md5.Sum(data))

	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/") {
			c.Header("Cache-Control", "private, max-age=86400")

			if match := c.GetHeader("If-None-Match"); match != "" {
				if strings.Contains(match, etag) {
					c.Status(http.StatusNotModified)
					return
				}
			}
		}

		c.Next()
	}
}

// 从包含端口的地址中，获取客户端 IP 地址
func getBaseIP(addrWithPort string) string {
	host, _, err := net.SplitHostPort(addrWithPort)
	if err != nil {
		return ""
	}
	return host
}

// 生成 JSON 数据
func renderJSON(ipaddr string, dbInfo []string) map[string]any {
	return map[string]any{"ip": ipaddr, "info": dbInfo}
}

// 添加 IPDB 参数
func TelnetServer(ipdb *IPDB) {
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
		// 将 IPDB 参数传入处理函数
		go handleConnection(ipdb, conn)
	}
}

// 添加 IPDB 参数
func handleConnection(ipdb *IPDB, conn net.Conn) {
	defer conn.Close()

	// 去除端口号的 IP 地址
	clientIP := getBaseIP(conn.RemoteAddr().String())
	// 去除端口号，查询详细信息
	info := ipdb.FindByIPIP(clientIP)
	// 发送消息给客户端
	sendBuf := [][]byte{}
	message, err := json.Marshal(renderJSON(clientIP, info))
	// 发生错误时，打印错误信息
	if err != nil {
		fmt.Println("序列化 JSON 数据时发生错误: ", err)
		return
	}
	// 添加消息到发送缓冲区，确保消息以 CRLF 结尾
	sendBuf = append(sendBuf, message)
	sendBuf = append(sendBuf, []byte("\r\n"))
	_, err = conn.Write(bytes.Join(sendBuf, []byte("")))
	if err != nil {
		fmt.Printf("发送消息时发生错误: %v\n", err)
		return
	}
}

type IPDB struct {
	IPIP *ipdb.City
}

// 初始化 IPDB 数据库实例
func initIPDB() IPDB {
	db, err := ipdb.NewCity("./data/ipipfree.ipdb")
	if err != nil {
		log.Fatal(err)
	}
	return IPDB{IPIP: db}
}

// 根据 IP 地址查询信息（IPIP 数据库）
func (db IPDB) FindByIPIP(ip string) []string {
	info, err := db.IPIP.Find(ip, "CN")
	if err != nil {
		info = []string{"未找到 IP 地址信息"}
	}
	return removeDuplicates(info)
}

//go:embed public
var EmbedFS embed.FS

type IPForm struct {
	IP string `form:"ip" binding:"required"`
}

func main() {
	config := parseConfig()

	// 初始化 IPDB 数据库
	ipdb := initIPDB()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(gin.Recovery())
	r.Use(gzip.Gzip(gzip.BestCompression))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"domain": config.Domain,
		})
	})

	r.Use(cacheMiddleware())

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

		// 简化 IP 地址查询
		dbInfo := ipdb.FindByIPIP(ipaddr)
		return ipaddr, dbInfo, nil
	}

	renderTemplate := func(c *gin.Context, globalTemplate []byte, ipaddr string, dbInfo []string) []byte {
		template := bytes.ReplaceAll(globalTemplate, []byte("%IP_ADDR%"), []byte(ipaddr))
		template = bytes.ReplaceAll(template, []byte("%DOMAIN%"), []byte(config.Domain))
		template = bytes.ReplaceAll(template, []byte("%DATA_1_INFO%"), []byte(strings.Join(removeDuplicates(dbInfo), " ")))
		template = bytes.ReplaceAll(template, []byte("%DOCUMENT_PATH%"), []byte(c.Request.URL.Path))
		return template
	}

	globalTemplate := []byte{}
	err := error(nil)

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

		userAgent := c.GetHeader("User-Agent")
		if IsDownloadTool(userAgent) {
			c.JSON(200, renderJSON(ipAddr, dbInfo))
		} else {
			c.Data(200, "text/html; charset=utf-8", renderTemplate(c, globalTemplate, ipAddr, dbInfo))
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

		userAgent := c.GetHeader("User-Agent")
		if IsDownloadTool(userAgent) {
			c.JSON(200, renderJSON(ipAddr, dbInfo))
		} else {
			c.Data(200, "text/html; charset=utf-8", renderTemplate(c, globalTemplate, ipAddr, dbInfo))
		}
	})

	// 将 IPDB 参数传入 TelnetServer
	go TelnetServer(&ipdb)

	serverAddr := fmt.Sprintf(":%s", config.Port)
	log.Printf("启动服务器于 %s:%s\n", config.Domain, config.Port)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}

	r.Run(":8080")
}
