package ipInfo_test

import (
	"os"
	"path/filepath"
	"testing"

	ipInfo "github.com/soulteary/ip-helper/model/ip-info"
)

func TestInitIPDB(t *testing.T) {
	// 创建临时测试目录
	tempDir := t.TempDir()

	// 测试用例1: 使用有效的数据库文件
	validDBPath := filepath.Join(tempDir, "valid.ipdb")
	// 创建一个模拟的 IPIP 数据库文件
	err := os.WriteFile(validDBPath, []byte("mock ipip db content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock DB file: %v", err)
	}

	workDir, _ := os.Getwd()

	tests := []struct {
		name    string
		dbPath  string
		wantErr bool
	}{
		{
			name:    "Valid DB path",
			dbPath:  filepath.Join(workDir, "../../data/ipipfree.ipdb"),
			wantErr: false, // 实际环境中这里会失败，因为我们使用了模拟数据
		},
		{
			name:    "Invalid DB path",
			dbPath:  filepath.Join(tempDir, "nonexistent.ipdb"),
			wantErr: true,
		},
		{
			name:    "Empty DB path",
			dbPath:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ipInfo.InitIPDB(tt.dbPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitIPDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// TestInitIPDB_RealDB 测试真实的IPIP数据库文件
// 注意：这个测试需要真实的IPIP数据库文件才能运行
func TestInitIPDB_RealDB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// 这里需要替换为实际的IPIP数据库文件路径
	realDBPath := "/path/to/real/ipip.ipdb"

	// 检查文件是否存在
	if _, err := os.Stat(realDBPath); os.IsNotExist(err) {
		t.Skip("Skipping test: real IPIP database file not found")
	}

	db, err := ipInfo.InitIPDB(realDBPath)
	if err != nil {
		t.Fatalf("InitIPDB() failed with real database: %v", err)
	}

	if db.IPIP == nil {
		t.Error("InitIPDB() returned nil IPIP database handler")
	}
}
