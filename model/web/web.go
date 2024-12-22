package web

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/gzip"
	static "github.com/soulteary/gin-static"

	"github.com/gin-gonic/gin"
	"github.com/soulteary/ip-helper/model/define"
	"github.com/soulteary/ip-helper/model/fn"
	ipInfo "github.com/soulteary/ip-helper/model/ip-info"
	"github.com/soulteary/ip-helper/model/page"
	"github.com/soulteary/ip-helper/model/response"
)

type IPForm struct {
	IP string `form:"ip" binding:"required"`
}

func GetClientIPInfo(c *gin.Context, ipaddr string, ipdb *ipInfo.IPDB) (resultIP string, resultDBInfo []string, err error) {
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

	err := error(nil)

	r.GET("/", func(c *gin.Context) {
		if config.Debug {
			globalTemplate, err = fn.HTTPGet(fmt.Sprintf("http://localhost:%s/index.template.html", config.Port))
			if err != nil {
				log.Fatalf("读取模板文件失败: %v\n", err)
				return
			}
		}

		ipAddr, dbInfo, err := GetClientIPInfo(c, "", ipdb)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		userAgent := c.GetHeader("User-Agent")
		if fn.IsDownloadTool(userAgent) {
			c.Data(200, "application/json; charset=utf-8", response.RenderJSON(ipAddr, dbInfo))
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
		ipAddr, dbInfo, err := GetClientIPInfo(c, ip, ipdb)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		userAgent := c.GetHeader("User-Agent")
		if fn.IsDownloadTool(userAgent) {
			c.Data(200, "application/json; charset=utf-8", response.RenderJSON(ipAddr, dbInfo))
		} else {
			c.Data(200, "text/html; charset=utf-8", response.RenderHTML(config, c.Request.URL.Path, globalTemplate, ipAddr, dbInfo))
		}
	})

	serverAddr := fmt.Sprintf(":%s", config.Port)
	log.Printf("启动服务器于 %s:%s\n", config.Domain, config.Port)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}

	r.Run(":8080")
}
