package ipInfo

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/soulteary/ip-helper/model/fn"
)

type Info struct {
	ClientIP     string `json:"client_ip"`
	ProxyIP      string `json:"proxy_ip,omitempty"`
	IsProxy      bool   `json:"is_proxy"`
	ForwardedFor string `json:"forwarded_for,omitempty"`
	RealIP       string `json:"real_ip"`
}

func AnalyzeRequestData(c *gin.Context) Info {
	var ipInfo Info

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

	if fn.IsPrivateIP(ipInfo.ClientIP) {
		ipInfo.IsProxy = true
	}
	return ipInfo
}
