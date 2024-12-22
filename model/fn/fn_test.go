package fn_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	fn "github.com/soulteary/ip-helper/model/fn"
)

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"Private IP Class A", "10.0.0.1", true},
		{"Private IP Class B", "172.16.0.1", true},
		{"Private IP Class C", "192.168.1.1", true},
		{"Public IP", "8.8.8.8", false},
		{"Invalid IP", "256.256.256.256", false},
		{"Empty IP", "", false},
		{"Malformed IP", "192.168.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fn.IsPrivateIP(tt.ip); got != tt.expected {
				t.Errorf("IsPrivateIP(%v) = %v, want %v", tt.ip, got, tt.expected)
			}
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "No duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "With duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "All duplicates",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fn.RemoveDuplicates(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("RemoveDuplicates() length = %v, want %v", len(got), len(tt.expected))
			}
			// Convert slices to maps for comparison
			gotMap := make(map[string]bool)
			for _, v := range got {
				gotMap[v] = true
			}
			expectedMap := make(map[string]bool)
			for _, v := range tt.expected {
				expectedMap[v] = true
			}
			for k := range expectedMap {
				if !gotMap[k] {
					t.Errorf("RemoveDuplicates() missing expected value %v", k)
				}
			}
		})
	}
}

func TestIsValidIPAddress(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"Valid IPv4", "192.168.1.1", true},
		{"Valid IPv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"Invalid IP", "256.256.256.256", false},
		{"Empty string", "", false},
		{"Malformed IP", "192.168.1", false},
		{"Not an IP", "example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fn.IsValidIPAddress(tt.ip); got != tt.expected {
				t.Errorf("IsValidIPAddress(%v) = %v, want %v", tt.ip, got, tt.expected)
			}
		})
	}
}

func TestIsDownloadTool(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  bool
	}{
		{"Curl user agent", "curl/7.64.1", true},
		{"Wget user agent", "Wget/1.20.3 (linux-gnu)", true},
		{"Python requests", "python-requests/2.25.1", true},
		{"Regular browser", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", false},
		{"Empty string", "", false},
		{"Postman agent", "PostmanRuntime/7.26.8", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fn.IsDownloadTool(tt.userAgent); got != tt.expected {
				t.Errorf("IsDownloadTool(%v) = %v, want %v", tt.userAgent, got, tt.expected)
			}
		})
	}
}

func TestGetDomainOnly(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"Simple domain", "example.com", "example.com"},
		{"Domain with protocol", "https://example.com", "example.com"},
		{"Domain with path", "https://example.com/path", "example.com"},
		{"Domain with port", "example.com:8080", "example.com"},
		{"Full URL", "https://example.com:8080/path?query=1", "example.com"},
		{"Invalid URL", "not a url", "not a url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fn.GetDomainOnly(tt.url); got != tt.expected {
				t.Errorf("GetDomainOnly(%v) = %v, want %v", tt.url, got, tt.expected)
			}
		})
	}
}

func TestGetDomainWithPort(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"Domain with port", "example.com:8080", "example.com:8080"},
		{"Domain with protocol and port", "https://example.com:8080", "example.com:8080"},
		{"Domain without port", "example.com", "example.com"},
		{"Full URL with port", "https://example.com:8080/path?query=1", "example.com:8080"},
		{"Invalid URL", "not a url", "not a url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fn.GetDomainWithPort(tt.url); got != tt.expected {
				t.Errorf("GetDomainWithPort(%v) = %v, want %v", tt.url, got, tt.expected)
			}
		})
	}
}

func TestGetBaseIP(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected string
	}{
		{"Valid IP and port", "127.0.0.1:8080", "127.0.0.1"},
		{"IPv6 and port", "[::1]:8080", "::1"},
		{"Invalid format", "127.0.0.1", ""},
		{"Empty string", "", ""},
		{"Only port", ":8080", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fn.GetBaseIP(tt.addr); got != tt.expected {
				t.Errorf("GetBaseIP(%v) = %v, want %v", tt.addr, got, tt.expected)
			}
		})
	}
}

func TestHTTPGet(t *testing.T) {
	tests := []struct {
		name          string
		setupServer   func() (*httptest.Server, error)
		expectedError string
		expectedBody  []byte
	}{
		{
			name: "Success response",
			setupServer: func() (*httptest.Server, error) {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("test content"))
				})), nil
			},
			expectedBody: []byte("test content"),
		},
		{
			name: "Non-200 status code",
			setupServer: func() (*httptest.Server, error) {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				})), nil
			},
			expectedError: "服务器返回非200状态码: 404",
		},
		{
			name: "Network error - connection refused",
			setupServer: func() (*httptest.Server, error) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				ts.Close() // Close immediately to simulate connection error
				return ts, nil
			},
			expectedError: "connection refused",
		},
		{
			name: "Invalid URL",
			setupServer: func() (*httptest.Server, error) {
				return nil, nil // No server needed
			},
			expectedError: "unsupported protocol scheme",
		},
		{
			name: "Read body error",
			setupServer: func() (*httptest.Server, error) {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Length", "100") // 声明100字节的内容
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("incomplete")) // 但只写入一部分数据
					// 刷新数据，确保部分数据被发送
					if f, ok := w.(http.Flusher); ok {
						f.Flush()
					}
					// 等待一小段时间确保数据被发送
					time.Sleep(10 * time.Millisecond)
					// 然后关闭连接，导致读取错误
					if cn, ok := w.(http.CloseNotifier); ok {
						cn.(interface{ CloseConnection() }).CloseConnection()
					}
				})), nil
			},
			expectedError: "unexpected EOF",
		},
		{
			name: "Empty response body",
			setupServer: func() (*httptest.Server, error) {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})), nil
			},
			expectedBody: []byte{},
		},
		{
			name: "Large response body",
			setupServer: func() (*httptest.Server, error) {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					// Generate 1MB of data
					data := make([]byte, 1024*1024)
					for i := range data {
						data[i] = byte(i % 256)
					}
					w.Write(data)
				})), nil
			},
			expectedBody: nil, // We'll check the length instead
		},
		{
			name: "Slow response",
			setupServer: func() (*httptest.Server, error) {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(100 * time.Millisecond) // Simulate slow response
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("delayed content"))
				})), nil
			},
			expectedBody: []byte("delayed content"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts *httptest.Server
			var testURL string

			// Setup test server if provided
			if tt.setupServer != nil {
				var err error
				ts, err = tt.setupServer()
				if err != nil {
					t.Fatalf("Failed to setup test server: %v", err)
				}
				if ts != nil {
					defer ts.Close()
					testURL = ts.URL
				}
			}

			if testURL == "" && tt.name == "Invalid URL" {
				testURL = "invalid-url"
			}

			// Test the HTTPGet function
			body, err := fn.HTTPGet(testURL)

			// Verify error cases
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s' but got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s' but got '%s'", tt.expectedError, err.Error())
				}
				return
			}

			// Verify success cases
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Special handling for large response test
			if tt.name == "Large response body" {
				if len(body) != 1024*1024 {
					t.Errorf("Expected body length %d, got %d", 1024*1024, len(body))
				}
				return
			}

			// Verify response body
			if tt.expectedBody != nil && !bytes.Equal(body, tt.expectedBody) {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}
