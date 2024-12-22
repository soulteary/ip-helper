package web_test

import (
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/soulteary/ip-helper/model/define"
	ipInfo "github.com/soulteary/ip-helper/model/ip-info"
	"github.com/soulteary/ip-helper/model/web"
)

func GetIPDB() (*ipInfo.IPDB, error) {
	workDir, _ := os.Getwd()
	dbPath := filepath.Join(workDir, "../../data/ipipfree.ipdb")

	ipdb, err := ipInfo.InitIPDB(dbPath)
	if err != nil {
		return nil, err
	}
	return &ipdb, nil
}

// 测试 GetClientIP 函数
func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		inputIP        string
		expectedIP     string
		expectedDBInfo []string
		expectError    bool
	}{
		{
			name: "With Direct IP Input",
			setupContext: func(c *gin.Context) {
				// 空设置,因为直接使用输入的 IP
			},
			inputIP:        "1.2.3.4",
			expectedIP:     "1.2.3.4",
			expectedDBInfo: []string{"APNIC.NET", ""},
			expectError:    false,
		},
		{
			name: "With Direct IP Input",
			setupContext: func(c *gin.Context) {
				// 空设置,因为直接使用输入的 IP
			},
			inputIP:        "123.123.123.123",
			expectedIP:     "123.123.123.123",
			expectedDBInfo: []string{"中国", "北京"},
			expectError:    false,
		},
		{
			name: "With Context IP Info",
			setupContext: func(c *gin.Context) {
				c.Set("ip_info", ipInfo.Info{RealIP: "5.6.7.8"})
			},
			inputIP:        "",
			expectedIP:     "5.6.7.8",
			expectedDBInfo: []string{"德国", ""},
			expectError:    false,
		},
		{
			name: "Without IP Info",
			setupContext: func(c *gin.Context) {
				// 不设置 IP 信息
			},
			inputIP:        "",
			expectedIP:     "",
			expectedDBInfo: nil,
			expectError:    true,
		},
	}

	db, err := GetIPDB()
	if err != nil {
		t.Fatalf("Failed to get IPDB: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试环境
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setupContext(c)

			// 执行测试
			ip, dbInfo, err := web.GetClientIP(c, tt.inputIP, db)

			// 验证结果
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if ip != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, ip)
			}
			if !tt.expectError {
				if len(dbInfo) != len(tt.expectedDBInfo) {
					fmt.Println(dbInfo)
					t.Errorf("Expected DB info length %d, got %d", len(tt.expectedDBInfo), len(dbInfo))
				}
				for i := range tt.expectedDBInfo {
					if dbInfo[i] != tt.expectedDBInfo[i] {
						t.Errorf("Expected DB info[%d]=%s, got %s", i, tt.expectedDBInfo[i], dbInfo[i])
					}
				}
			}
		})
	}
}

// 测试 Response 函数
func TestResponse(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func(*gin.Context)
		config       *define.Config
		ip           string
		template     []byte
		userAgent    string
		expectedCode int
	}{
		{
			name: "Normal HTML Response",
			setupContext: func(c *gin.Context) {
				c.Set("ip_info", ipInfo.Info{RealIP: "1.2.3.4"})
			},
			config: &define.Config{
				Debug:  false,
				Domain: "example.com",
			},
			ip:           "",
			template:     []byte("<html>{{.IP}}</html>"),
			userAgent:    "Mozilla/5.0",
			expectedCode: 200,
		},
		{
			name: "Download Tool JSON Response",
			setupContext: func(c *gin.Context) {
				c.Set("ip_info", ipInfo.Info{RealIP: "1.2.3.4"})
			},
			config: &define.Config{
				Debug:  false,
				Domain: "example.com",
			},
			ip:           "",
			template:     []byte("<html>{{.IP}}</html>"),
			userAgent:    "curl/7.64.1",
			expectedCode: 200,
		},
		{
			name: "Missing IP Info Error",
			setupContext: func(c *gin.Context) {
				// 不设置 IP 信息
			},
			config: &define.Config{
				Debug:  false,
				Domain: "example.com",
			},
			ip:           "",
			template:     []byte("<html>{{.IP}}</html>"),
			userAgent:    "Mozilla/5.0",
			expectedCode: 500,
		},
	}

	db, err := GetIPDB()
	if err != nil {
		t.Fatalf("Failed to get IPDB: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试环境
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// 设置请求头
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Request.Header.Set("User-Agent", tt.userAgent)

			tt.setupContext(c)

			// 执行测试
			web.Response(c, tt.config, db, tt.ip, tt.template)

			// 验证响应状态码
			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			// 验证响应头
			contentType := w.Header().Get("Content-Type")
			if tt.userAgent == "curl/7.64.1" {
				if contentType != "application/json; charset=utf-8" {
					t.Errorf("Expected JSON content type, got %s", contentType)
				}
			} else if tt.expectedCode == 200 {
				if contentType != "text/html; charset=utf-8" {
					t.Errorf("Expected HTML content type, got %s", contentType)
				}
			}
		})
	}
}
