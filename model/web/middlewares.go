package web

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/soulteary/ip-helper/model/define"
	ipInfo "github.com/soulteary/ip-helper/model/ip-info"
)

func AuthMiddleware(config *define.Config) gin.HandlerFunc {
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

func IPAnalyzerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ipInfo := ipInfo.AnalyzeRequestData(c)
		c.Set("ip_info", ipInfo)
		c.Next()
	}
}

func CacheMiddleware() gin.HandlerFunc {
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
