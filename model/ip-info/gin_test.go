package ipInfo_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	ipInfo "github.com/soulteary/ip-helper/model/ip-info"
)

func TestAnalyzeRequestData(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       ipInfo.Info
	}{
		{
			name:       "Simple direct connection",
			remoteAddr: "1.2.3.4:1234",
			headers:    map[string]string{},
			want: ipInfo.Info{
				ClientIP: "1.2.3.4",
				RealIP:   "1.2.3.4",
				IsProxy:  false,
			},
		},
		{
			name:       "Connection with X-Forwarded-For",
			remoteAddr: "1.2.3.4:1234",
			headers: map[string]string{
				"X-Forwarded-For": "5.6.7.8, 1.2.3.4",
			},
			want: ipInfo.Info{
				ClientIP:     "5.6.7.8",
				ProxyIP:      "1.2.3.4",
				RealIP:       "5.6.7.8",
				IsProxy:      true,
				ForwardedFor: "5.6.7.8, 1.2.3.4",
			},
		},
		{
			name:       "Connection with X-Real-IP",
			remoteAddr: "1.2.3.4:1234",
			headers: map[string]string{
				"X-Real-IP": "5.6.7.8",
			},
			want: ipInfo.Info{
				ClientIP: "5.6.7.8",
				ProxyIP:  "",
				RealIP:   "5.6.7.8",
				IsProxy:  false,
			},
		},
		{
			name:       "Connection with private IP",
			remoteAddr: "192.168.1.1:1234",
			headers:    map[string]string{},
			want: ipInfo.Info{
				ClientIP: "192.168.1.1",
				RealIP:   "192.168.1.1",
				IsProxy:  true,
			},
		},
		{
			name:       "Connection with multiple X-Forwarded-For IPs",
			remoteAddr: "1.2.3.4:1234",
			headers: map[string]string{
				"X-Forwarded-For": "5.6.7.8, 9.10.11.12, 1.2.3.4",
			},
			want: ipInfo.Info{
				ClientIP:     "5.6.7.8",
				ProxyIP:      "1.2.3.4",
				RealIP:       "5.6.7.8",
				IsProxy:      true,
				ForwardedFor: "5.6.7.8, 9.10.11.12, 1.2.3.4",
			},
		},
		{
			name:       "Connection with both X-Forwarded-For and different X-Real-IP",
			remoteAddr: "1.2.3.4:1234",
			headers: map[string]string{
				"X-Forwarded-For": "5.6.7.8, 1.2.3.4",
				"X-Real-IP":       "9.10.11.12",
			},
			want: ipInfo.Info{
				ClientIP:     "5.6.7.8",
				ProxyIP:      "5.6.7.8",    // ClientIP 被设置为代理IP
				RealIP:       "9.10.11.12", // X-Real-IP 的值覆盖了之前的 RealIP
				IsProxy:      true,
				ForwardedFor: "5.6.7.8, 1.2.3.4",
			},
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建一个带有信任代理设置的路由器
			router := gin.New()
			router.SetTrustedProxies([]string{"0.0.0.0/0", "::/0"})

			// 添加测试路由
			router.GET("/test", func(c *gin.Context) {
				result := ipInfo.AnalyzeRequestData(c)
				// 将结果存储在context中以便后续检查
				c.Set("result", result)
			})

			// 创建请求
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			// 设置请求头
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// 创建响应记录器
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 创建新的context来获取结果
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// 直接调用被测试的函数
			got := ipInfo.AnalyzeRequestData(c)

			// 验证结果
			if got.ClientIP != tt.want.ClientIP {
				t.Errorf("ClientIP = %v, want %v", got.ClientIP, tt.want.ClientIP)
			}
			if got.ProxyIP != tt.want.ProxyIP {
				t.Errorf("ProxyIP = %v, want %v", got.ProxyIP, tt.want.ProxyIP)
			}
			if got.RealIP != tt.want.RealIP {
				t.Errorf("RealIP = %v, want %v", got.RealIP, tt.want.RealIP)
			}
			if got.IsProxy != tt.want.IsProxy {
				t.Errorf("IsProxy = %v, want %v", got.IsProxy, tt.want.IsProxy)
			}
			if got.ForwardedFor != tt.want.ForwardedFor {
				t.Errorf("ForwardedFor = %v, want %v", got.ForwardedFor, tt.want.ForwardedFor)
			}
		})
	}
}
