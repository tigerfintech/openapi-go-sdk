package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pgregory.net/rapid"
)

// TestNewClientConfig_WithOptions tests setting fields via Option pattern
func TestNewClientConfig_WithOptions(t *testing.T) {
	cfg, err := NewClientConfig(
		WithTigerID("test_tiger_id"),
		WithPrivateKey("test_private_key"),
		WithAccount("DU123456"),
		WithLanguage("en_US"),
		WithTimezone("America/New_York"),
		WithTimeout(30*time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	if cfg.TigerID != "test_tiger_id" {
		t.Fatalf("TigerID: expected %q, got %q", "test_tiger_id", cfg.TigerID)
	}
	if cfg.PrivateKey != "test_private_key" {
		t.Fatalf("PrivateKey: expected %q, got %q", "test_private_key", cfg.PrivateKey)
	}
	if cfg.Account != "DU123456" {
		t.Fatalf("Account: expected %q, got %q", "DU123456", cfg.Account)
	}
	if cfg.Language != "en_US" {
		t.Fatalf("Language: expected %q, got %q", "en_US", cfg.Language)
	}
	if cfg.Timezone != "America/New_York" {
		t.Fatalf("Timezone: expected %q, got %q", "America/New_York", cfg.Timezone)
	}
	if cfg.Timeout != 30*time.Second {
		t.Fatalf("Timeout: expected %v, got %v", 30*time.Second, cfg.Timeout)
	}
}

// TestNewClientConfig_Defaults tests default values
func TestNewClientConfig_Defaults(t *testing.T) {
	// Use a temp HOME so auto-discovery doesn't find a real config file
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	cfg, err := NewClientConfig(
		WithTigerID("tid"),
		WithPrivateKey("pk"),
	)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	if cfg.Language != "zh_CN" {
		t.Fatalf("default Language: expected %q, got %q", "zh_CN", cfg.Language)
	}
	if cfg.Timeout != 15*time.Second {
		t.Fatalf("default Timeout: expected %v, got %v", 15*time.Second, cfg.Timeout)
	}
	if cfg.ServerURL != "https://openapi.tigerfintech.com/gateway" {
		t.Fatalf("default ServerURL: expected production URL, got %q", cfg.ServerURL)
	}
	if cfg.TigerPublicKey == "" {
		t.Fatal("default TigerPublicKey should not be empty")
	}
}

// TestNewClientConfig_MissingTigerID tests error when tiger_id is missing
func TestNewClientConfig_MissingTigerID(t *testing.T) {
	// Use a temp HOME so auto-discovery doesn't find a real config file
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	_, err := NewClientConfig(
		WithPrivateKey("pk"),
	)
	if err == nil {
		t.Fatal("期望返回错误（缺少 tiger_id），但没有")
	}
}

// TestNewClientConfig_MissingPrivateKey 测试缺少 private_key 时返回错误
func TestNewClientConfig_MissingPrivateKey(t *testing.T) {
	// Use a temp HOME so auto-discovery doesn't find a real config file
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	_, err := NewClientConfig(
		WithTigerID("tid"),
	)
	if err == nil {
		t.Fatal("期望返回错误（缺少 private_key），但没有")
	}
}

// TestNewClientConfig_EnvOverride 测试环境变量覆盖代码设置
func TestNewClientConfig_EnvOverride(t *testing.T) {
	// 设置环境变量
	os.Setenv("TIGEROPEN_TIGER_ID", "env_tiger_id")
	os.Setenv("TIGEROPEN_PRIVATE_KEY", "env_private_key")
	os.Setenv("TIGEROPEN_ACCOUNT", "env_account")
	defer func() {
		os.Unsetenv("TIGEROPEN_TIGER_ID")
		os.Unsetenv("TIGEROPEN_PRIVATE_KEY")
		os.Unsetenv("TIGEROPEN_ACCOUNT")
	}()

	cfg, err := NewClientConfig(
		WithTigerID("code_tiger_id"),
		WithPrivateKey("code_private_key"),
		WithAccount("code_account"),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}

	// 环境变量应覆盖代码设置
	if cfg.TigerID != "env_tiger_id" {
		t.Fatalf("TigerID: 期望环境变量值 %q，实际 %q", "env_tiger_id", cfg.TigerID)
	}
	if cfg.PrivateKey != "env_private_key" {
		t.Fatalf("PrivateKey: 期望环境变量值 %q，实际 %q", "env_private_key", cfg.PrivateKey)
	}
	if cfg.Account != "env_account" {
		t.Fatalf("Account: 期望环境变量值 %q，实际 %q", "env_account", cfg.Account)
	}
}

// TestNewClientConfig_EnvOverridePartial 测试环境变量部分覆盖
func TestNewClientConfig_EnvOverridePartial(t *testing.T) {
	os.Setenv("TIGEROPEN_TIGER_ID", "env_tid")
	defer os.Unsetenv("TIGEROPEN_TIGER_ID")

	cfg, err := NewClientConfig(
		WithTigerID("code_tid"),
		WithPrivateKey("code_pk"),
		WithAccount("code_acc"),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}

	// 只有 tiger_id 被环境变量覆盖
	if cfg.TigerID != "env_tid" {
		t.Fatalf("TigerID: 期望 %q，实际 %q", "env_tid", cfg.TigerID)
	}
	if cfg.PrivateKey != "code_pk" {
		t.Fatalf("PrivateKey: 期望 %q，实际 %q", "code_pk", cfg.PrivateKey)
	}
	if cfg.Account != "code_acc" {
		t.Fatalf("Account: 期望 %q，实际 %q", "code_acc", cfg.Account)
	}
}

// TestNewClientConfig_FromPropertiesFile 测试从配置文件加载
func TestNewClientConfig_FromPropertiesFile(t *testing.T) {
	content := "tiger_id=file_tid\nprivate_key=file_pk\naccount=file_acc\nlicense=TBNZ\n"
	path := writeTempFile(t, content)

	cfg, err := NewClientConfig(
		WithPropertiesFile(path),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}

	if cfg.TigerID != "file_tid" {
		t.Fatalf("TigerID: 期望 %q，实际 %q", "file_tid", cfg.TigerID)
	}
	if cfg.PrivateKey != "file_pk" {
		t.Fatalf("PrivateKey: 期望 %q，实际 %q", "file_pk", cfg.PrivateKey)
	}
	if cfg.Account != "file_acc" {
		t.Fatalf("Account: 期望 %q，实际 %q", "file_acc", cfg.Account)
	}
	if cfg.License != "TBNZ" {
		t.Fatalf("License: 期望 %q，实际 %q", "TBNZ", cfg.License)
	}
}

// TestNewClientConfig_EnvOverridesFile 测试环境变量优先级高于配置文件
func TestNewClientConfig_EnvOverridesFile(t *testing.T) {
	content := "tiger_id=file_tid\nprivate_key=file_pk\naccount=file_acc\n"
	path := writeTempFile(t, content)

	os.Setenv("TIGEROPEN_TIGER_ID", "env_tid")
	defer os.Unsetenv("TIGEROPEN_TIGER_ID")

	cfg, err := NewClientConfig(
		WithPropertiesFile(path),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}

	// 环境变量覆盖配置文件
	if cfg.TigerID != "env_tid" {
		t.Fatalf("TigerID: 期望环境变量值 %q，实际 %q", "env_tid", cfg.TigerID)
	}
	// 配置文件值保留
	if cfg.PrivateKey != "file_pk" {
		t.Fatalf("PrivateKey: 期望配置文件值 %q，实际 %q", "file_pk", cfg.PrivateKey)
	}
}

// TestNewClientConfig_WithToken 测试 Token 相关字段
func TestNewClientConfig_WithToken(t *testing.T) {
	cfg, err := NewClientConfig(
		WithTigerID("tid"),
		WithPrivateKey("pk"),
		WithToken("my_token"),
		WithTokenRefreshDuration(3600*time.Second),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}

	if cfg.Token != "my_token" {
		t.Fatalf("Token: 期望 %q，实际 %q", "my_token", cfg.Token)
	}
	if cfg.TokenRefreshDuration != 3600*time.Second {
		t.Fatalf("TokenRefreshDuration: 期望 %v，实际 %v", 3600*time.Second, cfg.TokenRefreshDuration)
	}
}

// TestNewClientConfig_WithLicense 测试 License 字段
func TestNewClientConfig_WithLicense(t *testing.T) {
	cfg, err := NewClientConfig(
		WithTigerID("tid"),
		WithPrivateKey("pk"),
		WithLicense("TBHK"),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}

	if cfg.License != "TBHK" {
		t.Fatalf("License: 期望 %q，实际 %q", "TBHK", cfg.License)
	}
}

// TestNewClientConfig_PKCS1FromFile 测试从配置文件加载 PKCS#1 私钥
func TestNewClientConfig_PKCS1FromFile(t *testing.T) {
	content := "tiger_id=tid\nprivate_key_pk1=pk1_key_content\naccount=acc\n"
	path := writeTempFile(t, content)

	cfg, err := NewClientConfig(
		WithPropertiesFile(path),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}

	if cfg.PrivateKey != "pk1_key_content" {
		t.Fatalf("PrivateKey: 期望 %q，实际 %q", "pk1_key_content", cfg.PrivateKey)
	}
}

// TestNewClientConfig_PKCS8FromFile 测试从配置文件加载 PKCS#8 私钥
func TestNewClientConfig_PKCS8FromFile(t *testing.T) {
	content := "tiger_id=tid\nprivate_key_pk8=pk8_key_content\naccount=acc\n"
	path := writeTempFile(t, content)

	cfg, err := NewClientConfig(
		WithPropertiesFile(path),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}

	if cfg.PrivateKey != "pk8_key_content" {
		t.Fatalf("PrivateKey: 期望 %q，实际 %q", "pk8_key_content", cfg.PrivateKey)
	}
}

// TestNewClientConfig_WithTokenWriter 测试 WithTokenWriter 选项
func TestNewClientConfig_WithTokenWriter(t *testing.T) {
	var got string
	cfg, err := NewClientConfig(
		WithTigerID("tid"),
		WithPrivateKey("pk"),
		WithTokenWriter(func(token string) { got = token }),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}
	if cfg.TokenWriter == nil {
		t.Fatal("TokenWriter 应已设置")
	}
	cfg.TokenWriter("test_token")
	if got != "test_token" {
		t.Fatalf("TokenWriter 未触发，got=%q", got)
	}
}

// TestNewClientConfig_WithTokenLoader 测试 WithTokenLoader 初始化时自动加载 token
func TestNewClientConfig_WithTokenLoader(t *testing.T) {
	cfg, err := NewClientConfig(
		WithTigerID("tid"),
		WithPrivateKey("pk"),
		WithTokenLoader(func() (string, error) { return "loader_token", nil }),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}
	if cfg.TokenLoader == nil {
		t.Fatal("TokenLoader 应已设置")
	}
	// NewClientConfig 应在初始化时调用 TokenLoader 填充 cfg.Token
	if cfg.Token != "loader_token" {
		t.Fatalf("Token: 期望 TokenLoader 返回值 %q，实际 %q", "loader_token", cfg.Token)
	}
}

// TestNewClientConfig_WithTokenLoader_EnvTokenTakesPriority 测试 TIGEROPEN_TOKEN 优先于 TokenLoader
func TestNewClientConfig_WithTokenLoader_EnvTokenTakesPriority(t *testing.T) {
	os.Setenv("TIGEROPEN_TOKEN", "env_token")
	defer os.Unsetenv("TIGEROPEN_TOKEN")

	called := false
	cfg, err := NewClientConfig(
		WithTigerID("tid"),
		WithPrivateKey("pk"),
		WithTokenLoader(func() (string, error) {
			called = true
			return "loader_token", nil
		}),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}
	if cfg.Token != "env_token" {
		t.Fatalf("Token: 环境变量应优先，期望 %q，实际 %q", "env_token", cfg.Token)
	}
	if called {
		t.Fatal("TIGEROPEN_TOKEN 已设置时不应调用 TokenLoader")
	}
}

// TestNewClientConfig_WithTokenCheckInterval 测试 WithTokenCheckInterval 选项
func TestNewClientConfig_WithTokenCheckInterval(t *testing.T) {
	cfg, err := NewClientConfig(
		WithTigerID("tid"),
		WithPrivateKey("pk"),
		WithTokenCheckInterval(10*time.Minute),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}
	if cfg.TokenCheckInterval != 10*time.Minute {
		t.Fatalf("TokenCheckInterval: 期望 %v，实际 %v", 10*time.Minute, cfg.TokenCheckInterval)
	}
}

// TestNewClientConfig_WithDeviceID 测试 WithDeviceID 选项
func TestNewClientConfig_WithDeviceID(t *testing.T) {
	cfg, err := NewClientConfig(
		WithTigerID("tid"),
		WithPrivateKey("pk"),
		WithDeviceID("custom-device-id"),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}
	if cfg.DeviceID != "custom-device-id" {
		t.Fatalf("DeviceID: 期望 %q，实际 %q", "custom-device-id", cfg.DeviceID)
	}
}

// TestNewClientConfig_WithQuoteServerURL 测试 WithQuoteServerURL 选项
func TestNewClientConfig_WithQuoteServerURL(t *testing.T) {
	cfg, err := NewClientConfig(
		WithTigerID("tid"),
		WithPrivateKey("pk"),
		WithQuoteServerURL("https://quote.example.com"),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}
	if cfg.QuoteServerURL != "https://quote.example.com" {
		t.Fatalf("QuoteServerURL: 期望 %q，实际 %q", "https://quote.example.com", cfg.QuoteServerURL)
	}
}

// TestNewClientConfig_DisableDynamicDomain 测试禁用动态域名
func TestNewClientConfig_DisableDynamicDomain(t *testing.T) {
	cfg, err := NewClientConfig(
		WithTigerID("tid"),
		WithPrivateKey("pk"),
		WithEnableDynamicDomain(false),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}
	if cfg.EnableDynamicDomain {
		t.Fatal("EnableDynamicDomain 应为 false")
	}
}

// TestNewClientConfig_EnvToken 测试 TIGEROPEN_TOKEN 环境变量
func TestNewClientConfig_EnvToken(t *testing.T) {
	os.Setenv("TIGEROPEN_TOKEN", "env_token_value")
	defer os.Unsetenv("TIGEROPEN_TOKEN")

	cfg, err := NewClientConfig(
		WithTigerID("tid"),
		WithPrivateKey("pk"),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}
	if cfg.Token != "env_token_value" {
		t.Fatalf("Token: 期望环境变量值 %q，实际 %q", "env_token_value", cfg.Token)
	}
}

// TestNewClientConfig_EnvTokenFile 测试 TIGEROPEN_TOKEN_FILE 环境变量
func TestNewClientConfig_EnvTokenFile(t *testing.T) {
	tokenFile := writeTempFile(t, "token=file_token_value\n")
	os.Setenv("TIGEROPEN_TOKEN_FILE", tokenFile)
	defer os.Unsetenv("TIGEROPEN_TOKEN_FILE")

	cfg, err := NewClientConfig(
		WithTigerID("tid"),
		WithPrivateKey("pk"),
	)
	if err != nil {
		t.Fatalf("创建配置失败: %v", err)
	}
	if cfg.Token != "file_token_value" {
		t.Fatalf("Token: 期望 token 文件值 %q，实际 %q", "file_token_value", cfg.Token)
	}
}

// TestNewClientConfig_WithPropertiesFile_InvalidPath 测试无效配置文件路径时静默忽略
func TestNewClientConfig_WithPropertiesFile_InvalidPath(t *testing.T) {
	// 文件不存在时 WithPropertiesFile 静默跳过；再补上必填字段仍可正常创建
	cfg, err := NewClientConfig(
		WithPropertiesFile("/nonexistent/config.properties"),
		WithTigerID("fallback_tid"),
		WithPrivateKey("fallback_pk"),
	)
	if err != nil {
		t.Fatalf("无效路径应静默忽略，但返回了错误: %v", err)
	}
	if cfg.TigerID != "fallback_tid" {
		t.Fatalf("TigerID: 期望 %q，实际 %q", "fallback_tid", cfg.TigerID)
	}
}

// TestApplyProperties_PrivateKeyPriority 测试私钥优先级：private_key > pk8 > pk1
func TestApplyProperties_PrivateKeyPriority(t *testing.T) {
	t.Run("private_key_takes_priority_over_pk8", func(t *testing.T) {
		content := "tiger_id=tid\nprivate_key=pk_main\nprivate_key_pk8=pk8_key\n"
		cfg, err := NewClientConfig(WithPropertiesFile(writeTempFile(t, content)))
		if err != nil {
			t.Fatalf("创建配置失败: %v", err)
		}
		if cfg.PrivateKey != "pk_main" {
			t.Fatalf("期望 pk_main，实际 %q", cfg.PrivateKey)
		}
	})

	t.Run("pk8_takes_priority_over_pk1", func(t *testing.T) {
		content := "tiger_id=tid\nprivate_key_pk8=pk8_key\nprivate_key_pk1=pk1_key\n"
		cfg, err := NewClientConfig(WithPropertiesFile(writeTempFile(t, content)))
		if err != nil {
			t.Fatalf("创建配置失败: %v", err)
		}
		if cfg.PrivateKey != "pk8_key" {
			t.Fatalf("期望 pk8_key，实际 %q", cfg.PrivateKey)
		}
	})
}

// TestClientConfigFieldsRoundTrip 属性测试：任意有效配置参数 round-trip
func TestClientConfigFieldsRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tigerID := rapid.StringMatching(`[a-zA-Z0-9]{1,20}`).Draw(t, "tigerID")
		privateKey := rapid.StringMatching(`[a-zA-Z0-9]{1,100}`).Draw(t, "privateKey")
		account := rapid.StringMatching(`[A-Z]{2}[0-9]{4,10}`).Draw(t, "account")
		language := rapid.SampledFrom([]string{"zh_CN", "zh_TW", "en_US"}).Draw(t, "language")
		timezone := rapid.StringMatching(`[A-Za-z/_]{3,30}`).Draw(t, "timezone")
		timeoutSec := rapid.IntRange(1, 120).Draw(t, "timeoutSec")

		cfg, err := NewClientConfig(
			WithTigerID(tigerID),
			WithPrivateKey(privateKey),
			WithAccount(account),
			WithLanguage(language),
			WithTimezone(timezone),
			WithTimeout(time.Duration(timeoutSec)*time.Second),
		)
		if err != nil {
			t.Fatalf("failed to create config: %v", err)
		}

		if cfg.TigerID != tigerID {
			t.Fatalf("TigerID: expected %q, got %q", tigerID, cfg.TigerID)
		}
		if cfg.PrivateKey != privateKey {
			t.Fatalf("PrivateKey: expected %q, got %q", privateKey, cfg.PrivateKey)
		}
		if cfg.Account != account {
			t.Fatalf("Account: expected %q, got %q", account, cfg.Account)
		}
		if cfg.Language != language {
			t.Fatalf("Language: expected %q, got %q", language, cfg.Language)
		}
		if cfg.Timezone != timezone {
			t.Fatalf("Timezone: expected %q, got %q", timezone, cfg.Timezone)
		}
		if cfg.Timeout != time.Duration(timeoutSec)*time.Second {
			t.Fatalf("Timeout: expected %v, got %v", time.Duration(timeoutSec)*time.Second, cfg.Timeout)
		}
	})
}

// TestEnvOverridesConfigFile 属性测试：环境变量优先级高于配置文件
func TestEnvOverridesConfigFile(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		fileTigerID := rapid.StringMatching(`file_[a-zA-Z0-9]{1,15}`).Draw(t, "fileTigerID")
		filePrivateKey := rapid.StringMatching(`file_[a-zA-Z0-9]{1,50}`).Draw(t, "filePrivateKey")
		fileAccount := rapid.StringMatching(`file_[A-Z]{2}[0-9]{4,8}`).Draw(t, "fileAccount")

		envTigerID := rapid.StringMatching(`env_[a-zA-Z0-9]{1,15}`).Draw(t, "envTigerID")
		envPrivateKey := rapid.StringMatching(`env_[a-zA-Z0-9]{1,50}`).Draw(t, "envPrivateKey")
		envAccount := rapid.StringMatching(`env_[A-Z]{2}[0-9]{4,8}`).Draw(t, "envAccount")

		content := fmt.Sprintf("tiger_id=%s\nprivate_key=%s\naccount=%s\n",
			fileTigerID, filePrivateKey, fileAccount)
		dir := os.TempDir()
		path := filepath.Join(dir, fmt.Sprintf("env_test_%d.properties",
			rapid.IntRange(0, 999999).Draw(t, "fileId")))
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("写入临时文件失败: %v", err)
		}
		defer os.Remove(path)

		os.Setenv("TIGEROPEN_TIGER_ID", envTigerID)
		os.Setenv("TIGEROPEN_PRIVATE_KEY", envPrivateKey)
		os.Setenv("TIGEROPEN_ACCOUNT", envAccount)
		defer func() {
			os.Unsetenv("TIGEROPEN_TIGER_ID")
			os.Unsetenv("TIGEROPEN_PRIVATE_KEY")
			os.Unsetenv("TIGEROPEN_ACCOUNT")
		}()

		cfg, err := NewClientConfig(WithPropertiesFile(path))
		if err != nil {
			t.Fatalf("创建配置失败: %v", err)
		}

		if cfg.TigerID != envTigerID {
			t.Fatalf("TigerID: 期望环境变量值 %q，实际 %q", envTigerID, cfg.TigerID)
		}
		if cfg.PrivateKey != envPrivateKey {
			t.Fatalf("PrivateKey: 期望环境变量值 %q，实际 %q", envPrivateKey, cfg.PrivateKey)
		}
		if cfg.Account != envAccount {
			t.Fatalf("Account: 期望环境变量值 %q，实际 %q", envAccount, cfg.Account)
		}
	})
}
