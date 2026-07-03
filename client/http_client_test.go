package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tigerfintech/openapi-go-sdk/logger"
)

// TestHttpClient_Execute_Success 测试成功的 API 请求
func TestHttpClient_Execute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法和 Content-Type
		if r.Method != "POST" {
			t.Errorf("期望 POST 方法，实际为 %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/json;charset=UTF-8" {
			t.Errorf("期望 Content-Type 为 application/json;charset=UTF-8，实际为 %s", ct)
		}
		// 验证 User-Agent 与 SDKVersion 保持一致
		ua := r.Header.Get("User-Agent")
		expected := UserAgentPrefix + SDKVersion
		if ua != expected {
			t.Errorf("期望 User-Agent 为 %s，实际为 %s", expected, ua)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"code":0,"message":"success","data":{"market":"US","status":"Trading"},"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	client := NewHttpClient(cfg)

	req := &ApiRequest{Method: "market_state", BizContent: `{"market":"US"}`}
	resp, err := client.Execute(req)
	if err != nil {
		t.Fatalf("Execute 失败: %v", err)
	}
	if resp.Code != 0 {
		t.Errorf("期望 code=0，实际为 %d", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("期望 message=success，实际为 %s", resp.Message)
	}
	if resp.Timestamp != 1700000000 {
		t.Errorf("期望 timestamp=1700000000，实际为 %d", resp.Timestamp)
	}
}

// TestHttpClient_Execute_ApiError 测试 API 返回错误码
func TestHttpClient_Execute_ApiError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":1000,"message":"公共参数错误","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	client := NewHttpClient(cfg)

	req := &ApiRequest{Method: "market_state", BizContent: `{}`}
	resp, err := client.Execute(req)
	if err == nil {
		t.Fatal("期望返回错误，但没有")
	}
	tigerErr, ok := err.(*TigerError)
	if !ok {
		t.Fatalf("期望 TigerError 类型，实际为 %T", err)
	}
	if tigerErr.Code != 1000 {
		t.Errorf("期望错误码 1000，实际为 %d", tigerErr.Code)
	}
	if tigerErr.Category != CategoryCommonParam {
		t.Errorf("期望分类 common_param_error，实际为 %s", tigerErr.Category)
	}
	if resp == nil {
		t.Fatal("即使有错误，resp 也不应为 nil")
	}
}

// TestHttpClient_Execute_RequestParams 测试请求参数包含必要字段
func TestHttpClient_Execute_RequestParams(t *testing.T) {
	var receivedParams map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var params map[string]string
		json.NewDecoder(r.Body).Decode(&params)
		receivedParams = params
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	client := NewHttpClient(cfg)

	req := &ApiRequest{Method: "market_state", BizContent: `{"market":"US"}`}
	_, err := client.Execute(req)
	if err != nil {
		t.Fatalf("Execute 失败: %v", err)
	}

	requiredKeys := []string{"tiger_id", "method", "charset", "sign_type", "timestamp", "version", "biz_content", "sign"}
	for _, key := range requiredKeys {
		if _, ok := receivedParams[key]; !ok {
			t.Errorf("缺少必要参数: %s", key)
		}
	}
	if receivedParams["tiger_id"] != "test_tiger_id" {
		t.Errorf("tiger_id 不匹配: %s", receivedParams["tiger_id"])
	}
	if receivedParams["method"] != "market_state" {
		t.Errorf("method 不匹配: %s", receivedParams["method"])
	}
	if receivedParams["biz_content"] != `{"market":"US"}` {
		t.Errorf("biz_content 不匹配: %s", receivedParams["biz_content"])
	}
}

// TestHttpClient_AuthorizationHeader_WithToken 测试设置 Token 时请求携带 Authorization 头
func TestHttpClient_AuthorizationHeader_WithToken(t *testing.T) {
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cfg.Token = "my_test_token_value"
	client := NewHttpClient(cfg)

	req := &ApiRequest{Method: "market_state", BizContent: `{}`}
	_, err := client.Execute(req)
	if err != nil {
		t.Fatalf("Execute 失败: %v", err)
	}
	if authHeader != "my_test_token_value" {
		t.Errorf("期望 Authorization 头为 my_test_token_value，实际为 %q", authHeader)
	}
}

// TestHttpClient_AuthorizationHeader_WithoutToken 测试未设置 Token 时不携带 Authorization 头
func TestHttpClient_AuthorizationHeader_WithoutToken(t *testing.T) {
	var hasAuth bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hasAuth = r.Header["Authorization"]
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cfg.Token = "" // 无 token
	client := NewHttpClient(cfg)

	req := &ApiRequest{Method: "market_state", BizContent: `{}`}
	_, err := client.Execute(req)
	if err != nil {
		t.Fatalf("Execute 失败: %v", err)
	}
	if hasAuth {
		t.Error("未设置 Token 时不应携带 Authorization 头")
	}
}

// capturingLogger 记录所有收到的日志调用，供测试断言使用。
type capturingLogger struct {
	entries []logEntry
}

type logEntry struct {
	level string
	msg   string
}

func (l *capturingLogger) Debug(msg string, args ...interface{}) {
	l.entries = append(l.entries, logEntry{"DEBUG", fmt.Sprintf(msg, args...)})
}
func (l *capturingLogger) Info(msg string, args ...interface{}) {
	l.entries = append(l.entries, logEntry{"INFO", fmt.Sprintf(msg, args...)})
}
func (l *capturingLogger) Warn(msg string, args ...interface{}) {
	l.entries = append(l.entries, logEntry{"WARN", fmt.Sprintf(msg, args...)})
}
func (l *capturingLogger) Error(msg string, args ...interface{}) {
	l.entries = append(l.entries, logEntry{"ERROR", fmt.Sprintf(msg, args...)})
}
func (l *capturingLogger) SetLevel(_ logger.Level) {}

// TestHttpClient_Execute_LogsDebugOnSuccess 验证成功请求写入 DEBUG 日志
func TestHttpClient_Execute_LogsDebugOnSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cl := &capturingLogger{}
	hc := NewHttpClient(cfg, WithLogger(cl))

	_, err := hc.Execute(&ApiRequest{Method: "market_state", BizContent: `{}`})
	if err != nil {
		t.Fatalf("Execute 失败: %v", err)
	}

	if len(cl.entries) != 1 {
		t.Fatalf("期望 1 条日志，实际 %d 条: %v", len(cl.entries), cl.entries)
	}
	entry := cl.entries[0]
	if entry.level != "DEBUG" {
		t.Errorf("期望 DEBUG 级别，实际 %s", entry.level)
	}
	if !contains(entry.msg, "market_state") {
		t.Errorf("日志应包含 method，实际: %s", entry.msg)
	}
}

// TestHttpClient_Execute_LogsWarnOnBusinessError 验证业务错误（code!=0）写入 WARN 日志
func TestHttpClient_Execute_LogsWarnOnBusinessError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":1000,"message":"invalid param","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cl := &capturingLogger{}
	hc := NewHttpClient(cfg, WithLogger(cl))

	_, _ = hc.Execute(&ApiRequest{Method: "place_order", BizContent: `{}`})

	if len(cl.entries) != 1 {
		t.Fatalf("期望 1 条日志，实际 %d 条: %v", len(cl.entries), cl.entries)
	}
	entry := cl.entries[0]
	if entry.level != "WARN" {
		t.Errorf("期望 WARN 级别，实际 %s", entry.level)
	}
	if !contains(entry.msg, "place_order") || !contains(entry.msg, "1000") {
		t.Errorf("日志应包含 method 和 code，实际: %s", entry.msg)
	}
	// 不应含敏感字段
	if contains(entry.msg, "biz_content") || contains(entry.msg, "sign") {
		t.Errorf("日志不应包含敏感字段，实际: %s", entry.msg)
	}
}

// TestHttpClient_Execute_LogsErrorOnNetworkFailure 验证网络错误写入 ERROR 日志
func TestHttpClient_Execute_LogsErrorOnNetworkFailure(t *testing.T) {
	cfg := newTestConfig(t, "http://127.0.0.1:19999") // 无服务
	cl := &capturingLogger{}
	hc := NewHttpClient(cfg, WithLogger(cl))
	// place_order 不重试，直接 ERROR
	_, _ = hc.Execute(&ApiRequest{Method: "place_order", BizContent: `{}`})

	if len(cl.entries) == 0 {
		t.Fatal("网络错误时应有日志输出")
	}
	last := cl.entries[len(cl.entries)-1]
	if last.level != "ERROR" {
		t.Errorf("期望 ERROR 级别，实际 %s", last.level)
	}
	if !contains(last.msg, "place_order") {
		t.Errorf("日志应包含 method，实际: %s", last.msg)
	}
}

// TestHttpClient_Execute_NopLoggerSilences 验证注入 NopLogger 后 Execute 正常运行不崩溃
func TestHttpClient_Execute_NopLoggerSilences(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	// 使用 SDK 内置 NopLogger，完全静默
	hc := NewHttpClient(cfg, WithLogger(&nopLogger{}))

	_, err := hc.Execute(&ApiRequest{Method: "market_state", BizContent: `{}`})
	if err != nil {
		t.Fatalf("注入 NopLogger 后 Execute 不应失败: %v", err)
	}
}

// nopLogger 是一个满足 logger.Logger 接口的空实现，不输出任何内容。
type nopLogger struct{}

func (n *nopLogger) Debug(msg string, args ...interface{}) {}
func (n *nopLogger) Info(msg string, args ...interface{})  {}
func (n *nopLogger) Warn(msg string, args ...interface{})  {}
func (n *nopLogger) Error(msg string, args ...interface{}) {}
func (n *nopLogger) SetLevel(_ logger.Level)               {}

// contains 检查 s 中是否包含子串 sub
func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
