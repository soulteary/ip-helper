package web

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	static "github.com/soulteary/gin-static"
	"github.com/soulteary/ip-helper/model/define"
	"github.com/soulteary/ip-helper/model/fn"
	ipInfo "github.com/soulteary/ip-helper/model/ip-info"
	"github.com/soulteary/ip-helper/model/page"
	"github.com/soulteary/ip-helper/model/response"
)

func GetClientIP(c *gin.Context, ip string, ipdb *ipInfo.IPDB) (resultIP string, resultDBInfo []string, err error) {
	if ip == "" {
		info, exists := c.Get("ip_info")
		if !exists {
			return resultIP, resultDBInfo, fmt.Errorf("IP info not found")
		}
		ip = info.(ipInfo.Info).RealIP
	}
	return ip, ipdb.FindByIPIP(ip), nil
}

func Response(c *gin.Context, config *define.Config, ipdb *ipInfo.IPDB, ip string, template []byte) {
	err := error(nil)
	if config.Debug {
		template, err = fn.HTTPGet(fmt.Sprintf("http://localhost:%s/index.template.html", config.Port))
		if err != nil {
			log.Fatalf("读取模板文件失败: %v\n", err)
			return
		}
	}

	ipAddr, dbInfo, err := GetClientIP(c, ip, ipdb)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	userAgent := c.GetHeader("User-Agent")
	if fn.IsDownloadTool(userAgent) {
		c.Data(200, "application/json; charset=utf-8", response.RenderJSON(ipAddr, dbInfo))
	} else {
		c.Data(200, "text/html; charset=utf-8", response.RenderHTML(config, c.Request.URL.Path, template, ipAddr, dbInfo))
	}
}

type IPForm struct {
	IP string `form:"ip" binding:"required"`
}

func Server(config *define.Config, ipdb *ipInfo.IPDB) {
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

	r.Use(CacheMiddleware())
	r.Use(static.Serve("/", static.LocalFile("./public", false)))
	r.Use(AuthMiddleware(config))
	r.Use(IPAnalyzerMiddleware())

	globalTemplate := []byte(page.Template)
	if config.Debug {
		os.WriteFile("./public/index.template.html", globalTemplate, 0644)
	}

	r.GET("/", func(c *gin.Context) {
		Response(c, config, ipdb, "", globalTemplate)
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
		Response(c, config, ipdb, c.Param("ip"), globalTemplate)
	})

	serverAddr := fmt.Sprintf(":%s", config.Port)
	log.Printf("WEB 启动服务器于 %s\n", config.Port)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("WEB 服务器启动失败: %v", err)
	}
}
