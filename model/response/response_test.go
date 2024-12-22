package response_test

import (
	"reflect"
	"testing"

	"github.com/soulteary/ip-helper/model/define"
	"github.com/soulteary/ip-helper/model/response"
)

func TestRenderJSON(t *testing.T) {
	tests := []struct {
		name     string
		ipaddr   string
		dbInfo   []string
		expected map[string]any
	}{
		{
			name:   "basic test",
			ipaddr: "127.0.0.1",
			dbInfo: []string{"info1", "info2"},
			expected: map[string]any{
				"ip":   "127.0.0.1",
				"info": []string{"info1", "info2"},
			},
		},
		{
			name:   "empty dbInfo",
			ipaddr: "192.168.1.1",
			dbInfo: []string{},
			expected: map[string]any{
				"ip":   "192.168.1.1",
				"info": []string{},
			},
		},
		{
			name:   "empty ipaddr",
			ipaddr: "",
			dbInfo: []string{"test"},
			expected: map[string]any{
				"ip":   "",
				"info": []string{"test"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := response.RenderJSON(tt.ipaddr, tt.dbInfo)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("RenderJSON() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRenderHTML(t *testing.T) {
	tests := []struct {
		name            string
		config          *define.Config
		urlPath         string
		globalTemplate  []byte
		ipaddr          string
		dbInfo          []string
		expectedContent string
	}{
		{
			name: "all replacements",
			config: &define.Config{
				Domain: "example.com:8080",
			},
			urlPath:         "/test/path",
			globalTemplate:  []byte("IP: %IP_ADDR% Domain: %DOMAIN% Info: %DATA_1_INFO% Path: %DOCUMENT_PATH% Domain Only: %ONLY_DOMAIN% Domain With Port: %ONLY_DOMAIN_WITH_PORT%"),
			ipaddr:          "127.0.0.1",
			dbInfo:          []string{"info1", "info2"},
			expectedContent: "IP: 127.0.0.1 Domain: example.com:8080 Info: info1 info2 Path: /test/path Domain Only: example.com Domain With Port: example.com:8080",
		},
		{
			name: "empty values",
			config: &define.Config{
				Domain: "",
			},
			urlPath:         "",
			globalTemplate:  []byte("%IP_ADDR%%DOMAIN%%DATA_1_INFO%%DOCUMENT_PATH%%ONLY_DOMAIN%%ONLY_DOMAIN_WITH_PORT%"),
			ipaddr:          "",
			dbInfo:          []string{},
			expectedContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := response.RenderHTML(tt.config, tt.urlPath, tt.globalTemplate, tt.ipaddr, tt.dbInfo)
			if string(result) != tt.expectedContent {
				t.Errorf("RenderHTML() = %s, want %s", string(result), tt.expectedContent)
			}
		})
	}
}
