package ipInfo_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	ipInfo "github.com/soulteary/ip-helper/model/ip-info"
)

func TestIPDB_FindByIPIP(t *testing.T) {
	workDir, _ := os.Getwd()
	dbPath := filepath.Join(workDir, "../../data/ipipfree.ipdb")

	db, err := ipInfo.InitIPDB(dbPath)
	if err != nil {
		t.Errorf("Failed to initialize IPDB: %v", err)
		return
	}

	tests := []struct {
		name     string
		ip       string
		mockData []string
		mockErr  error
		want     []string
	}{
		{
			name: "Test 1.1.1.1",
			ip:   "1.1.1.1",
			want: []string{
				"CLOUDFLARE.COM",
				"",
			},
		},
		{
			name: "Test 123.123.123.123",
			ip:   "123.123.123.123",
			want: []string{
				"中国",
				"北京",
			},
		},
		{
			name: "Test 0.0.0.0",
			ip:   "0.0.0.0",
			want: []string{"保留地址", ""},
		},
		{
			name: "Test 2.2.2.2",
			ip:   "2.2.2.2",
			want: []string{"法国", ""},
		},
		{
			name: "Test ::1",
			ip:   "::1",
			want: []string{"未找到 IP 地址信息"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := db.FindByIPIP(tt.ip)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindByIPIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
