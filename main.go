package main

import (
	"crypto/md5"
	"embed"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	static "github.com/soulteary/gin-static"
	"github.com/soulteary/ip-helper/model/define"
	fn "github.com/soulteary/ip-helper/model/fn"
	"github.com/soulteary/ip-helper/model/ftp"
	ipInfo "github.com/soulteary/ip-helper/model/ip-info"
	configParser "github.com/soulteary/ip-helper/model/parse-config"
	"github.com/soulteary/ip-helper/model/response"
	"github.com/soulteary/ip-helper/model/telnet"
)

func authMiddleware(config *define.Config) gin.HandlerFunc {
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

func IPAnalyzer() gin.HandlerFunc {
	return func(c *gin.Context) {
		ipInfo := ipInfo.AnalyzeRequestData(c)
		c.Set("ip_info", ipInfo)
		c.Next()
	}
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

//go:embed public
var EmbedFS embed.FS

type IPForm struct {
	IP string `form:"ip" binding:"required"`
}

func main() {
	config := configParser.Parse()

	ipdb, err := ipInfo.InitIPDB("./data/ipipfree.ipdb")
	if err != nil {
		log.Fatalf("初始化 IP 数据库失败: %v\n", err)
		return
	}

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
			info, exists := c.Get("ip_info")
			if !exists {
				return resultIP, resultDBInfo, fmt.Errorf("IP info not found")
			}
			ipaddr = info.(ipInfo.Info).RealIP
		}

		dbInfo := ipdb.FindByIPIP(ipaddr)
		return ipaddr, dbInfo, nil
	}

	globalTemplate := []byte{}

	r.GET("/", func(c *gin.Context) {
		if len(globalTemplate) == 0 {
			globalTemplate, err = fn.HTTPGet(fmt.Sprintf("http://localhost:%s/index.template.html", config.Port))
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
		if fn.IsDownloadTool(userAgent) {
			c.JSON(200, response.RenderJSON(ipAddr, dbInfo))
		} else {
			c.Data(200, "text/html; charset=utf-8", response.RenderHTML(config, c.Request.URL.Path, globalTemplate, ipAddr, dbInfo))
		}
	})

	r.POST("/", func(c *gin.Context) {
		info, exists := c.Get("ip_info")
		if !exists {
			c.JSON(500, gin.H{"error": "IP info not found"})
			return
		}
		ip := ""
		var form IPForm
		if err := c.ShouldBind(&form); err != nil {
			ip = info.(ipInfo.Info).RealIP
		} else {
			ip = form.IP
			if !fn.IsValidIPAddress(ip) {
				ip = info.(ipInfo.Info).RealIP
			}
		}
		c.Redirect(302, fmt.Sprintf("/ip/%s", ip))
	})

	r.GET("/ip", func(c *gin.Context) {
		info, exists := c.Get("ip_info")
		if !exists {
			c.JSON(500, gin.H{"error": "IP info not found"})
			return
		}
		c.String(200, info.(ipInfo.Info).ClientIP)
	})

	r.GET("/ip/:ip", func(c *gin.Context) {
		ip := c.Param("ip")
		ipAddr, dbInfo, err := getClientIPInfo(c, ip)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		userAgent := c.GetHeader("User-Agent")
		if fn.IsDownloadTool(userAgent) {
			c.JSON(200, response.RenderJSON(ipAddr, dbInfo))
		} else {
			c.Data(200, "text/html; charset=utf-8", response.RenderHTML(config, c.Request.URL.Path, globalTemplate, ipAddr, dbInfo))
		}
	})

	go telnet.Server(&ipdb)
	go ftp.Server(&ipdb)

	serverAddr := fmt.Sprintf(":%s", config.Port)
	log.Printf("启动服务器于 %s:%s\n", config.Domain, config.Port)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}

	r.Run(":8080")
}
