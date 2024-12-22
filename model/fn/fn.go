package fn

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func IsPrivateIP(ipStr string) bool {
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

func RemoveDuplicates(strSlice []string) []string {
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

func IsValidIPAddress(ip string) bool {
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

func GetDomainOnly(urlStr string) string {
	if !strings.Contains(urlStr, "://") {
		urlStr = "http://" + urlStr
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	host := parsedURL.Hostname()
	return host
}

func GetDomainWithPort(urlStr string) string {
	if !strings.Contains(urlStr, "://") {
		urlStr = "http://" + urlStr
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	return parsedURL.Host
}

func GetBaseIP(addrWithPort string) string {
	host, _, err := net.SplitHostPort(addrWithPort)
	if err != nil {
		return ""
	}
	return host
}

func HTTPGet(link string) ([]byte, error) {
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
