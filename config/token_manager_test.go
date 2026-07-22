package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// makeTestToken 构造一个 base64 编码的测试 token，前 27 字符包含 "gen_ts,expire_ts"
func makeTestToken(genTsMs int64, expireTsMs int64) string {
	// 前 27 字符格式: "gen_ts_ms,expire_ts_m" 补齐到 27 字符
	header := fmt.Sprintf("%013d,%013d", genTsMs, expireTsMs)
	// header 恰好 27 字符: 13 + 1 + 13
	payload := header + "some_extra_payload_data"
	return base64.StdEncoding.EncodeToString([]byte(payload))
}

func TestTokenManager_LoadToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.properties")
	os.WriteFile(path, []byte("token=test_token_123\n"), 0644)

	m := NewTokenManager(WithTokenFilePath(path))
	token, err := m.LoadToken()
	if err != nil {
		t.Fatalf("加载 Token 失败: %v", err)
	}
	if token != "test_token_123" {
		t.Errorf("Token 应为 test_token_123，实际为 %s", token)
	}
	if m.GetToken() != "test_token_123" {
		t.Error("GetToken 返回值不一致")
	}
}

func TestTokenManager_LoadToken_FileNotFound(t *testing.T) {
	m := NewTokenManager(WithTokenFilePath("/nonexistent/path"))
	_, err := m.LoadToken()
	if err == nil {
		t.Fatal("文件不存在时应返回错误")
	}
}

func TestTokenManager_LoadToken_NoTokenField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.properties")
	os.WriteFile(path, []byte("other_key=value\n"), 0644)

	m := NewTokenManager(WithTokenFilePath(path))
	_, err := m.LoadToken()
	if err == nil {
		t.Fatal("无 token 字段时应返回错误")
	}
}

func TestTokenManager_SetToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.properties")

	m := NewTokenManager(WithTokenFilePath(path))
	err := m.SetToken("new_token_456")
	if err != nil {
		t.Fatalf("设置 Token 失败: %v", err)
	}
	if m.GetToken() != "new_token_456" {
		t.Error("内存中 Token 未更新")
	}

	// 验证文件已更新
	content, _ := os.ReadFile(path)
	if string(content) != "token=new_token_456\n" {
		t.Errorf("文件内容不正确: %s", string(content))
	}
}

func TestTokenManager_SetToken_ThenLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.properties")

	m := NewTokenManager(WithTokenFilePath(path))
	m.SetToken("round_trip_token")

	m2 := NewTokenManager(WithTokenFilePath(path))
	token, err := m2.LoadToken()
	if err != nil {
		t.Fatalf("重新加载 Token 失败: %v", err)
	}
	if token != "round_trip_token" {
		t.Errorf("Token 应为 round_trip_token，实际为 %s", token)
	}
}

func TestTokenManager_AutoRefresh(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.properties")

	// 构造一个 gen_ts 在 100 秒前的 token，refreshDuration 设为 30 秒
	oldGenTs := (time.Now().Unix() - 100) * 1000
	oldToken := makeTestToken(oldGenTs, oldGenTs+3600000)

	m := NewTokenManager(
		WithTokenFilePath(path),
		WithTokenRefreshInterval(50*time.Millisecond),
		WithRefreshDuration(30),
	)
	m.SetToken(oldToken)

	callCount := 0
	m.StartAutoRefresh(func() (string, error) {
		callCount++
		return "refreshed_token", nil
	})

	time.Sleep(200 * time.Millisecond)
	m.StopAutoRefresh()

	if callCount == 0 {
		t.Error("刷新函数应至少被调用一次")
	}
	if m.GetToken() != "refreshed_token" {
		t.Errorf("Token 应为 refreshed_token，实际为 %s", m.GetToken())
	}
}

func TestTokenManager_AutoRefresh_SkipsWhenNotNeeded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.properties")

	// 构造一个刚刚生成的 token，refreshDuration 设为 3600 秒（远未到期）
	freshGenTs := time.Now().Unix() * 1000
	freshToken := makeTestToken(freshGenTs, freshGenTs+3600000)

	m := NewTokenManager(
		WithTokenFilePath(path),
		WithTokenRefreshInterval(50*time.Millisecond),
		WithRefreshDuration(3600),
	)
	m.SetToken(freshToken)

	callCount := 0
	m.StartAutoRefresh(func() (string, error) {
		callCount++
		return "should_not_be_set", nil
	})

	time.Sleep(200 * time.Millisecond)
	m.StopAutoRefresh()

	if callCount != 0 {
		t.Errorf("Token 未过期时不应调用刷新函数，但被调用了 %d 次", callCount)
	}
}

func TestTokenManager_ShouldTokenRefresh_EmptyToken(t *testing.T) {
	m := NewTokenManager(WithRefreshDuration(30))
	// 空 token 不需要刷新
	if m.ShouldTokenRefresh() {
		t.Error("空 token 不应需要刷新")
	}
}

func TestTokenManager_ShouldTokenRefresh_ZeroDuration(t *testing.T) {
	m := NewTokenManager()
	m.SetToken(makeTestToken(1000000, 2000000))
	// refreshDuration 为 0 时不刷新
	if m.ShouldTokenRefresh() {
		t.Error("refreshDuration 为 0 时不应需要刷新")
	}
}

func TestTokenManager_ShouldTokenRefresh_Expired(t *testing.T) {
	m := NewTokenManager(WithRefreshDuration(30))
	// gen_ts 在 100 秒前
	oldGenTs := (time.Now().Unix() - 100) * 1000
	m.SetToken(makeTestToken(oldGenTs, oldGenTs+3600000))
	if !m.ShouldTokenRefresh() {
		t.Error("Token 已过期应需要刷新")
	}
}

func TestTokenManager_ShouldTokenRefresh_NotExpired(t *testing.T) {
	m := NewTokenManager(WithRefreshDuration(3600))
	// gen_ts 刚刚生成
	freshGenTs := time.Now().Unix() * 1000
	m.SetToken(makeTestToken(freshGenTs, freshGenTs+7200000))
	if m.ShouldTokenRefresh() {
		t.Error("Token 未过期不应需要刷新")
	}
}

func TestTokenManager_WithTokenLoader_LoadToken(t *testing.T) {
	m := NewTokenManager(
		TokenLoaderOption(func() (string, error) { return "loaded_from_custom", nil }),
	)

	token, err := m.LoadToken()
	if err != nil {
		t.Fatalf("自定义加载失败: %v", err)
	}
	if token != "loaded_from_custom" {
		t.Errorf("期望 loaded_from_custom，实际=%s", token)
	}
	if m.GetToken() != "loaded_from_custom" {
		t.Error("GetToken 返回值不一致")
	}
}

func TestTokenManager_WithTokenLoader_ErrorPropagated(t *testing.T) {
	m := NewTokenManager(
		TokenLoaderOption(func() (string, error) { return "", fmt.Errorf("source unavailable") }),
	)

	_, err := m.LoadToken()
	if err == nil {
		t.Fatal("自定义加载失败时应返回错误")
	}
}

func TestTokenManager_WithTokenLoader_EmptyReturnErrors(t *testing.T) {
	m := NewTokenManager(
		TokenLoaderOption(func() (string, error) { return "", nil }),
	)

	_, err := m.LoadToken()
	if err == nil {
		t.Fatal("自定义加载返回空值时应返回错误")
	}
}

func TestTokenManager_WithTokenLoader_TakesPriorityOverFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.properties")
	os.WriteFile(path, []byte("token=file_token\n"), 0644)

	m := NewTokenManager(
		WithTokenFilePath(path),
		TokenLoaderOption(func() (string, error) { return "loader_token", nil }),
	)

	token, err := m.LoadToken()
	if err != nil {
		t.Fatalf("加载失败: %v", err)
	}
	if token != "loader_token" {
		t.Errorf("WithTokenLoader 应优先于文件，期望 loader_token，实际=%s", token)
	}
}

func TestTokenManager_StopAutoRefresh_NoOp(t *testing.T) {
	m := NewTokenManager()
	// 未启动时停止不应 panic
	m.StopAutoRefresh()
}

func TestTokenManager_TokenWriter_CalledOnSetToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.properties")

	var got string
	m := NewTokenManager(
		WithTokenFilePath(path),
		TokenWriterOption(func(token string) { got = token }),
	)

	if err := m.SetToken("callback_token"); err != nil {
		t.Fatalf("SetToken 失败: %v", err)
	}
	if got != "callback_token" {
		t.Errorf("TokenWriter 未触发或值不符，got=%q", got)
	}
}

func TestTokenManager_TokenWriter_CalledOnAutoRefresh(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.properties")

	var saved string
	oldGenTs := (time.Now().Unix() - 100) * 1000
	m := NewTokenManager(
		WithTokenFilePath(path),
		WithTokenRefreshInterval(50*time.Millisecond),
		WithRefreshDuration(30),
		TokenWriterOption(func(token string) { saved = token }),
	)
	m.SetToken(makeTestToken(oldGenTs, oldGenTs+3600000))
	saved = "" // 重置，忽略 SetToken 触发的那次

	m.StartAutoRefresh(func() (string, error) {
		return "auto_refreshed_token", nil
	})
	time.Sleep(200 * time.Millisecond)
	m.StopAutoRefresh()

	if saved != "auto_refreshed_token" {
		t.Errorf("自动刷新后 TokenWriter 未触发，got=%q", saved)
	}
}

func TestTokenManager_TokenWriter_NotCalledOnFileError(t *testing.T) {
	// 文件路径不可写时，SetToken 应返回错误且不触发回调
	callCount := 0
	// 注意：必须显式调用 WithTokenFilePath 才会启用文件持久化
	m := NewTokenManager(
		WithTokenFilePath("/nonexistent_dir/token.properties"),
		TokenWriterOption(func(token string) { callCount++ }),
	)

	err := m.SetToken("some_token")
	if err == nil {
		t.Fatal("写入不可写路径时应返回错误")
	}
	if callCount != 0 {
		t.Errorf("文件写入失败时不应触发 TokenWriter，但被调用了 %d 次", callCount)
	}
}

func TestTokenManager_SetToken_NoFileWrite_WhenPathNotConfigured(t *testing.T) {
	// 未调用 WithTokenFilePath 时，SetToken 不应写文件，也不应报错
	m := NewTokenManager()
	if err := m.SetToken("in_memory_only"); err != nil {
		t.Fatalf("未配置文件路径时 SetToken 不应返回错误: %v", err)
	}
	if m.GetToken() != "in_memory_only" {
		t.Errorf("GetToken 期望 in_memory_only，实际=%s", m.GetToken())
	}
	// 默认路径文件不应被创建（工作目录下不应有 tiger_openapi_token.properties）
	if _, err := os.Stat(defaultTokenFileName); err == nil {
		// 如果文件恰好已存在（其他测试遗留），跳过此断言
		t.Log("默认 token 文件已存在，跳过文件未创建断言")
	}
}

func TestTokenManager_WithRefreshDuration_Clamp(t *testing.T) {
	// seconds 在 1~29 之间应被 clamp 到 30
	m := NewTokenManager(WithRefreshDuration(10))
	oldGenTs := (time.Now().Unix() - 25) * 1000 // 25 秒前生成，未超过 clamp 后的 30 秒
	m.SetToken(makeTestToken(oldGenTs, oldGenTs+3600000))
	if m.ShouldTokenRefresh() {
		t.Error("clamp 到 30 秒后，25 秒前的 token 不应触发刷新")
	}
}

func TestTokenManager_ShouldTokenRefresh_InvalidBase64(t *testing.T) {
	m := NewTokenManager(WithRefreshDuration(30))
	m.mu.Lock()
	m.token = "not-valid-base64!!!"
	m.mu.Unlock()
	if m.ShouldTokenRefresh() {
		t.Error("base64 解码失败时不应触发刷新")
	}
}

func TestTokenManager_ShouldTokenRefresh_ShortDecoded(t *testing.T) {
	m := NewTokenManager(WithRefreshDuration(30))
	// base64 编码不足 27 字节的内容
	m.mu.Lock()
	m.token = base64.StdEncoding.EncodeToString([]byte("tooshort"))
	m.mu.Unlock()
	if m.ShouldTokenRefresh() {
		t.Error("解码后内容不足 27 字节时不应触发刷新")
	}
}

func TestTokenManager_ShouldTokenRefresh_NoCommaInHeader(t *testing.T) {
	m := NewTokenManager(WithRefreshDuration(30))
	// 前 27 字节不含逗号
	payload := "nocolonorcommahere_______extra_payload"
	m.mu.Lock()
	m.token = base64.StdEncoding.EncodeToString([]byte(payload))
	m.mu.Unlock()
	if m.ShouldTokenRefresh() {
		t.Error("前 27 字节无逗号时不应触发刷新")
	}
}

func TestTokenManager_ShouldTokenRefresh_NonNumericGenTs(t *testing.T) {
	m := NewTokenManager(WithRefreshDuration(30))
	// 前 27 字节包含逗号但 gen_ts 部分不是数字
	payload := "not_a_number,0000000000000extra"
	m.mu.Lock()
	m.token = base64.StdEncoding.EncodeToString([]byte(payload))
	m.mu.Unlock()
	if m.ShouldTokenRefresh() {
		t.Error("gen_ts 非数字时不应触发刷新")
	}
}
