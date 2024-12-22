package configParser_test

import (
	"bytes"
	"flag"
	"log"
	"os"
	"strings"
	"testing"

	configParser "github.com/soulteary/ip-helper/model/parse-config"
)

var originalArgs []string
var originalFlagCommandLine *flag.FlagSet

func init() {
	originalArgs = os.Args
	originalFlagCommandLine = flag.CommandLine
}

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	// 清除所有已定义的标志
	flag.CommandLine.Init("", flag.ExitOnError)
}

func clearEnv() {
	os.Unsetenv("DEBUG")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("SERVER_DOMAIN")
	os.Unsetenv("TOKEN")
}

func captureLog(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr)
	return buf.String()
}

func TestParseDefaultValues(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"cmd"}

	defer func() {
		os.Args = oldArgs
		resetFlags()
		clearEnv()
	}()

	config := configParser.Parse()

	if config.Debug {
		t.Error("Debug 默认值应该为 false")
	}

	if config.Port != "8080" {
		t.Errorf("Port 默认值应该为 8080，实际为 %s", config.Port)
	}

	if config.Domain != "http://localhost:8080" {
		t.Errorf("Domain 默认值应该为 http://localhost:8080，实际为 %s", config.Domain)
	}

	if config.Token != "" {
		t.Errorf("Token 默认值应该为空，实际为 %s", config.Token)
	}
}

func TestParseEmptyValues(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{
		"cmd",
		"-port=",
		"-domain=",
	}

	defer func() {
		os.Args = oldArgs
		resetFlags()
		clearEnv()
	}()

	config := configParser.Parse()

	if config.Port != "8080" {
		t.Errorf("空 Port 应该使用默认值 8080，实际为 %s", config.Port)
	}

	if config.Domain != "http://localhost:8080" {
		t.Errorf("空 Domain 应该使用默认值 http://localhost:8080，实际为 %s", config.Domain)
	}
}

func TestParseCommandLineArgs(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{
		"cmd",
		"-debug=true",
		"-port=9090",
		"-domain=https://example.com",
		"-token=test-token",
	}

	defer func() {
		os.Args = oldArgs
		resetFlags()
		clearEnv()
	}()

	config := configParser.Parse()

	if !config.Debug {
		t.Error("Debug 应该为 true")
	}

	if config.Port != "9090" {
		t.Errorf("Port 应该为 9090，实际为 %s", config.Port)
	}

	if config.Domain != "https://example.com" {
		t.Errorf("Domain 应该为 https://example.com，实际为 %s", config.Domain)
	}

	if config.Token != "test-token" {
		t.Errorf("Token 应该为 test-token，实际为 %s", config.Token)
	}
}

func TestParseEnvironmentVars(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"cmd"}

	defer func() {
		os.Args = oldArgs
		resetFlags()
		clearEnv()
	}()

	os.Setenv("DEBUG", "true")
	os.Setenv("SERVER_PORT", "7070")
	os.Setenv("SERVER_DOMAIN", "https://env-example.com")
	os.Setenv("TOKEN", "env-token")

	config := configParser.Parse()

	if !config.Debug {
		t.Error("Debug 应该为 true")
	}

	if config.Port != "7070" {
		t.Errorf("Port 应该为 7070，实际为 %s", config.Port)
	}

	if config.Domain != "https://env-example.com" {
		t.Errorf("Domain 应该为 https://env-example.com，实际为 %s", config.Domain)
	}

	if config.Token != "env-token" {
		t.Errorf("Token 应该为 env-token，实际为 %s", config.Token)
	}
}

func TestEnvironmentDebugCaseInsensitive(t *testing.T) {
	t.Run("DEBUG=TRUE", func(t *testing.T) {
		oldArgs := os.Args
		os.Args = []string{"cmd"}

		defer func() {
			os.Args = oldArgs
			resetFlags()
			clearEnv()
		}()

		os.Setenv("DEBUG", "TRUE")
		config := configParser.Parse()
		if !config.Debug {
			t.Error("DEBUG=TRUE 应该设置 Debug 为 true")
		}
	})

	t.Run("DEBUG=True", func(t *testing.T) {
		oldArgs := os.Args
		os.Args = []string{"cmd"}

		defer func() {
			os.Args = oldArgs
			resetFlags()
			clearEnv()
		}()

		os.Setenv("DEBUG", "True")
		config := configParser.Parse()
		if !config.Debug {
			t.Error("DEBUG=True 应该设置 Debug 为 true")
		}
	})
}

func TestLogOutput(t *testing.T) {
	oldArgs := os.Args

	defer func() {
		os.Args = oldArgs
		resetFlags()
		clearEnv()
	}()

	// 测试 Debug 日志
	os.Args = []string{"cmd", "-debug=true"}
	output := captureLog(func() {
		configParser.Parse()
	})
	if !strings.Contains(output, "调试模式已开启") {
		t.Error("应该输出调试模式开启的日志")
	}

	// 测试 Token 提醒日志
	resetFlags()
	clearEnv()
	os.Args = []string{"cmd"}
	output = captureLog(func() {
		configParser.Parse()
	})
	if !strings.Contains(output, "TOKEN") {
		t.Error("应该输出 Token 提醒的日志")
	}
}

func TestPriorityOrder(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{
		"cmd",
		"-debug=true",
		"-port=9090",
	}

	defer func() {
		os.Args = oldArgs
		resetFlags()
		clearEnv()
	}()

	os.Setenv("DEBUG", "false")
	os.Setenv("SERVER_PORT", "7070")

	config := configParser.Parse()

	if !config.Debug {
		t.Error("命令行参数 Debug=true 应该覆盖环境变量")
	}

	if config.Port != "9090" {
		t.Errorf("命令行参数 Port=9090 应该覆盖环境变量，实际为 %s", config.Port)
	}
}

func TestEmptyEnvironmentVars(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"cmd"}

	defer func() {
		os.Args = oldArgs
		resetFlags()
		clearEnv()
	}()

	os.Setenv("SERVER_PORT", "")
	os.Setenv("SERVER_DOMAIN", "")
	os.Setenv("TOKEN", "")

	config := configParser.Parse()

	if config.Port != "8080" {
		t.Errorf("空环境变量 SERVER_PORT 应该使用默认值 8080，实际为 %s", config.Port)
	}

	if config.Domain != "http://localhost:8080" {
		t.Errorf("空环境变量 SERVER_DOMAIN 应该使用默认值 http://localhost:8080，实际为 %s", config.Domain)
	}

	if config.Token != "" {
		t.Errorf("空环境变量 TOKEN 应该为空字符串，实际为 %s", config.Token)
	}
}
