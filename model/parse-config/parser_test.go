package configParser_test

import (
	"flag"
	"os"
	"testing"

	configParser "github.com/soulteary/ip-helper/model/parse-config"
)

// cleanEnv 清理测试用的环境变量
func cleanEnv() {
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("SERVER_DOMAIN")
	os.Unsetenv("TOKEN")
}

// cleanArgs 清理命令行参数并保存原始参数
func cleanArgs() []string {
	oldArgs := os.Args
	os.Args = []string{os.Args[0]}
	return oldArgs
}

// restoreArgs 恢复原始命令行参数
func restoreArgs(args []string) {
	os.Args = args
}

// resetFlags 重置 flag 包的状态
func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

// TestParseDefault 测试默认配置场景
func TestParseDefault(t *testing.T) {
	// 保存并清理环境
	cleanEnv()
	oldArgs := cleanArgs()
	defer restoreArgs(oldArgs)
	resetFlags()

	config := configParser.Parse()

	// 验证默认值
	if config.Port != "8080" {
		t.Errorf("Expected default port to be 8080, got %s", config.Port)
	}
	if config.Domain != "http://localhost:8080" {
		t.Errorf("Expected default domain to be http://localhost:8080, got %s", config.Domain)
	}
	if config.Token != "" {
		t.Errorf("Expected default token to be empty, got %s", config.Token)
	}
}

// TestParseCommandLine 测试命令行参数解析
func TestParseCommandLine(t *testing.T) {
	cleanEnv()
	oldArgs := cleanArgs()
	defer restoreArgs(oldArgs)
	resetFlags()

	// 设置命令行参数
	os.Args = []string{
		os.Args[0],
		"-port", "9090",
		"-domain", "https://example.com",
		"-token", "test-token",
	}

	config := configParser.Parse()

	// 验证命令行参数是否正确解析
	if config.Port != "9090" {
		t.Errorf("Expected port to be 9090, got %s", config.Port)
	}
	if config.Domain != "https://example.com" {
		t.Errorf("Expected domain to be https://example.com, got %s", config.Domain)
	}
	if config.Token != "test-token" {
		t.Errorf("Expected token to be test-token, got %s", config.Token)
	}
}

// TestParseEnvironment 测试环境变量解析
func TestParseEnvironment(t *testing.T) {
	cleanEnv()
	oldArgs := cleanArgs()
	defer restoreArgs(oldArgs)
	resetFlags()

	// 设置环境变量
	os.Setenv("SERVER_PORT", "7070")
	os.Setenv("SERVER_DOMAIN", "https://test.com")
	os.Setenv("TOKEN", "env-token")

	config := configParser.Parse()

	// 验证环境变量是否正确解析
	if config.Port != "7070" {
		t.Errorf("Expected port to be 7070, got %s", config.Port)
	}
	if config.Domain != "https://test.com" {
		t.Errorf("Expected domain to be https://test.com, got %s", config.Domain)
	}
	if config.Token != "env-token" {
		t.Errorf("Expected token to be env-token, got %s", config.Token)
	}
}

// TestParsePriority 测试参数优先级：命令行参数 > 环境变量 > 默认值
func TestParsePriority(t *testing.T) {
	cleanEnv()
	oldArgs := cleanArgs()
	defer restoreArgs(oldArgs)
	resetFlags()

	// 设置环境变量
	os.Setenv("SERVER_PORT", "7070")
	os.Setenv("SERVER_DOMAIN", "https://test.com")
	os.Setenv("TOKEN", "env-token")

	// 设置命令行参数
	os.Args = []string{
		os.Args[0],
		"-port", "9090",
		"-domain", "https://example.com",
		"-token", "test-token",
	}

	config := configParser.Parse()

	// 验证优先级：应该使用命令行参数的值
	if config.Port != "9090" {
		t.Errorf("Expected port to be 9090 (command line), got %s", config.Port)
	}
	if config.Domain != "https://example.com" {
		t.Errorf("Expected domain to be https://example.com (command line), got %s", config.Domain)
	}
	if config.Token != "test-token" {
		t.Errorf("Expected token to be test-token (command line), got %s", config.Token)
	}
}
