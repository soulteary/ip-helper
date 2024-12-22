package response_test

import (
	"encoding/json"
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
			name:   "Normal IP with single info",
			ipaddr: "192.168.1.1",
			dbInfo: []string{"Location: New York"},
			expected: map[string]any{
				"ip":   "192.168.1.1",
				"info": []string{"Location: New York"},
			},
		},
		{
			name:   "IP with multiple info entries",
			ipaddr: "10.0.0.1",
			dbInfo: []string{"Country: US", "City: San Francisco", "ISP: Example"},
			expected: map[string]any{
				"ip":   "10.0.0.1",
				"info": []string{"Country: US", "City: San Francisco", "ISP: Example"},
			},
		},
		{
			name:   "Empty IP with empty info",
			ipaddr: "",
			dbInfo: []string{},
			expected: map[string]any{
				"ip":   "",
				"info": []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := response.RenderJSON(tt.ipaddr, tt.dbInfo)

			// Parse the JSON result back into a map for comparison
			var got map[string]any
			if err := json.Unmarshal(result, &got); err != nil {
				t.Fatalf("Failed to unmarshal result JSON: %v", err)
			}

			// Compare the IP field
			if got["ip"] != tt.expected["ip"] {
				t.Errorf("IP mismatch - got: %v, want: %v", got["ip"], tt.expected["ip"])
			}

			// Compare the info array
			gotInfo, ok := got["info"].([]any)
			if !ok {
				t.Fatal("Info field is not an array")
			}

			expectedInfo := tt.expected["info"].([]string)
			if len(gotInfo) != len(expectedInfo) {
				t.Errorf("Info array length mismatch - got: %d, want: %d", len(gotInfo), len(expectedInfo))
			}

			for i, v := range gotInfo {
				if v != expectedInfo[i] {
					t.Errorf("Info[%d] mismatch - got: %v, want: %v", i, v, expectedInfo[i])
				}
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
