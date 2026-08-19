package client

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewVersionedApiRequest_WithStruct verifies a versioned request is created correctly.
func TestNewVersionedApiRequest_WithStruct(t *testing.T) {
	params := map[string]string{"market": "US"}
	req, err := NewVersionedApiRequest("brief", params, "3.0")
	if err != nil {
		t.Fatalf("创建版本化请求失败: %v", err)
	}
	if req.Method != "brief" {
		t.Errorf("期望 method=brief，实际为 %s", req.Method)
	}
	if req.Version != "3.0" {
		t.Errorf("期望 version=3.0，实际为 %s", req.Version)
	}
	if req.BizContent == "" {
		t.Error("biz_content 不应为空")
	}
}

// TestNewVersionedApiRequest_WithString verifies versioned request with string biz_content.
func TestNewVersionedApiRequest_WithString(t *testing.T) {
	req, err := NewVersionedApiRequest("market_state", `{"market":"US"}`, "3.0")
	if err != nil {
		t.Fatalf("创建版本化请求失败: %v", err)
	}
	if req.Version != "3.0" {
		t.Errorf("期望 version=3.0，实际为 %s", req.Version)
	}
	if req.BizContent != `{"market":"US"}` {
		t.Errorf("biz_content 不匹配: %s", req.BizContent)
	}
}

// TestNewVersionedApiRequest_NilParams verifies versioned request with nil params defaults to {}.
func TestNewVersionedApiRequest_NilParams(t *testing.T) {
	req, err := NewVersionedApiRequest("market_state", nil, "3.0")
	if err != nil {
		t.Fatalf("创建版本化请求失败: %v", err)
	}
	if req.BizContent != "{}" {
		t.Errorf("nil 参数应生成 {}，实际为 %s", req.BizContent)
	}
}

// TestNewVersionedApiRequest_SerializationError verifies that an unserializable param returns an error.
func TestNewVersionedApiRequest_SerializationError(t *testing.T) {
	// A channel cannot be JSON-serialized
	ch := make(chan int)
	_, err := NewVersionedApiRequest("test", ch, "3.0")
	if err == nil {
		t.Fatal("不可序列化的参数应返回错误")
	}
}

// TestHttpClient_Close_WithTokenManager verifies Close stops the token manager.
func TestHttpClient_Close_WithTokenManager(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":{"token":"token"},"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cfg.TokenRefreshDuration = 30 * time.Second
	cfg.TokenCheckInterval = 50 * time.Millisecond

	hc := NewHttpClient(cfg)
	if hc.tokenManager == nil {
		t.Fatal("tokenManager 不应为 nil")
	}
	// Close should not panic and should stop the background goroutine
	hc.Close()
}

// TestHttpClient_Close_WithoutTokenManager verifies Close is a no-op when there's no token manager.
func TestHttpClient_Close_WithoutTokenManager(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	hc := NewHttpClient(cfg)
	if hc.tokenManager != nil {
		t.Fatal("未配置 TokenRefreshDuration 时 tokenManager 应为 nil")
	}
	hc.Close() // should not panic
}

// TestHttpClient_SecretKey_ReturnsConfigValue verifies SecretKey returns the config's secret key.
func TestHttpClient_SecretKey_ReturnsConfigValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cfg.SecretKey = "my_secret_key"
	hc := NewHttpClient(cfg)

	if got := hc.SecretKey(); got != "my_secret_key" {
		t.Errorf("期望 SecretKey=my_secret_key，实际为 %s", got)
	}
}

// TestHttpClient_SecretKey_EmptyWhenNotSet verifies SecretKey returns empty when not set.
func TestHttpClient_SecretKey_EmptyWhenNotSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	hc := NewHttpClient(cfg)

	if got := hc.SecretKey(); got != "" {
		t.Errorf("未设置 SecretKey 时应返回空，实际为 %s", got)
	}
}

// TestBuildCommonParams_DeviceID verifies DeviceID is included when set.
func TestBuildCommonParams_DeviceID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cfg.DeviceID = "device_abc123"
	hc := NewHttpClient(cfg)

	params := hc.buildCommonParams("market_state", "{}", "")
	if params["device_id"] != "device_abc123" {
		t.Errorf("期望 device_id=device_abc123，实际为 %s", params["device_id"])
	}
	if params["language"] != "zh_CN" {
		t.Errorf("期望 language=zh_CN，实际为 %s", params["language"])
	}
}

// TestBuildCommonParams_VersionOverride verifies version override works.
func TestBuildCommonParams_VersionOverride(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	hc := NewHttpClient(cfg)

	params := hc.buildCommonParams("brief", "{}", "3.0")
	if params["version"] != "3.0" {
		t.Errorf("期望 version=3.0，实际为 %s", params["version"])
	}

	params2 := hc.buildCommonParams("brief", "{}", "")
	if params2["version"] != DefaultVersion {
		t.Errorf("空 version 应使用默认值 %s，实际为 %s", DefaultVersion, params2["version"])
	}
}

// TestSignParams_InvalidKey verifies signParams returns an error with an invalid private key.
func TestSignParams_InvalidKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cfg.PrivateKey = "invalid-private-key"
	hc := NewHttpClient(cfg)

	_, err := hc.signParams(map[string]string{"method": "test"})
	if err == nil {
		t.Fatal("无效私钥应返回签名错误")
	}
	if !strings.Contains(err.Error(), "签名失败") {
		t.Errorf("错误信息应提示签名失败，实际: %v", err)
	}
}

// TestVerifyResponseSign_WithPublicKey verifies response signature verification with a valid public key.
func TestVerifyResponseSign_WithPublicKey(t *testing.T) {
	// Generate a key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成 RSA 密钥对失败: %v", err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("序列化公钥失败: %v", err)
	}
	pubB64 := base64.StdEncoding.EncodeToString(pubDER)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cfg.TigerPublicKey = pubB64
	hc := NewHttpClient(cfg)

	if hc.publicKey == nil {
		t.Fatal("publicKey 不应为 nil")
	}

	// Sign a timestamp and verify
	timestamp := "2024-01-01 00:00:00"

	// Test case 1: valid signature should pass
	// Sign the timestamp using the private key
	hashed := sha1.Sum([]byte(timestamp))
	sig, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA1, hashed[:])
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}
	sigB64 := base64.StdEncoding.EncodeToString(sig)

	apiResp := &ApiResponse{Sign: sigB64}
	err = hc.verifyResponseSign(apiResp, timestamp)
	if err != nil {
		t.Fatalf("有效签名验签应通过: %v", err)
	}

	// Test case 2: invalid signature should fail
	apiRespBad := &ApiResponse{Sign: base64.StdEncoding.EncodeToString([]byte("invalid-sig"))}
	err = hc.verifyResponseSign(apiRespBad, timestamp)
	if err == nil {
		t.Fatal("无效签名验签应失败")
	}
	if !strings.Contains(err.Error(), "signature verification failed") {
		t.Errorf("错误信息应提示验签失败，实际: %v", err)
	}
}

// TestVerifyResponseSign_SkipsWhenNoPublicKey verifies verification is skipped without a public key.
func TestVerifyResponseSign_SkipsWhenNoPublicKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	hc := NewHttpClient(cfg)
	// publicKey is nil (no TigerPublicKey set)

	apiResp := &ApiResponse{Sign: "some-sign"}
	err := hc.verifyResponseSign(apiResp, "timestamp")
	if err != nil {
		t.Fatalf("无公钥时应跳过验签，实际: %v", err)
	}
}

// TestVerifyResponseSign_SkipsWhenNoSign verifies verification is skipped when response has no sign.
func TestVerifyResponseSign_SkipsWhenNoSign(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubDER, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	pubB64 := base64.StdEncoding.EncodeToString(pubDER)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cfg.TigerPublicKey = pubB64
	hc := NewHttpClient(cfg)

	apiResp := &ApiResponse{Sign: ""}
	err := hc.verifyResponseSign(apiResp, "timestamp")
	if err != nil {
		t.Fatalf("无签名时应跳过验签，实际: %v", err)
	}
}

// TestVerifyResponseSign_SkipsWhenNoTimestamp verifies verification is skipped without a timestamp.
func TestVerifyResponseSign_SkipsWhenNoTimestamp(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubDER, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	pubB64 := base64.StdEncoding.EncodeToString(pubDER)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cfg.TigerPublicKey = pubB64
	hc := NewHttpClient(cfg)

	apiResp := &ApiResponse{Sign: "some-sign"}
	err := hc.verifyResponseSign(apiResp, "")
	if err != nil {
		t.Fatalf("无时间戳时应跳过验签，实际: %v", err)
	}
}

// TestExecuteRaw_EmptyMethodReturnsTigerError verifies ExecuteRaw returns a TigerError for empty method.
func TestExecuteRaw_EmptyMethodReturnsTigerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	hc := NewHttpClient(cfg)

	_, err := hc.ExecuteRaw("", `{"market":"US"}`)
	if err == nil {
		t.Fatal("空 method 应返回错误")
	}
	te, ok := err.(*TigerError)
	if !ok {
		t.Fatalf("期望 TigerError 类型，实际为 %T", err)
	}
	if te.Code != -1 {
		t.Errorf("期望 code=-1，实际为 %d", te.Code)
	}
}

// TestExecuteRaw_InvalidJSONReturnsTigerError verifies ExecuteRaw returns an error for invalid JSON.
func TestExecuteRaw_InvalidJSONReturnsTigerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	hc := NewHttpClient(cfg)

	_, err := hc.ExecuteRaw("market_state", "not valid json")
	if err == nil {
		t.Fatal("无效 JSON 应返回错误")
	}
	te, ok := err.(*TigerError)
	if !ok {
		t.Fatalf("期望 TigerError 类型，实际为 %T", err)
	}
	if te.Code != -1 {
		t.Errorf("期望 code=-1，实际为 %d", te.Code)
	}
}

// TestExecuteRaw_SuccessWithResponse verifies ExecuteRaw returns the raw response body and logs debug.
func TestExecuteRaw_SuccessWithResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":{"key":"value"},"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cl := &capturingLogger{}
	hc := NewHttpClient(cfg, WithLogger(cl))

	body, err := hc.ExecuteRaw("market_state", `{"market":"US"}`)
	if err != nil {
		t.Fatalf("ExecuteRaw 失败: %v", err)
	}
	if !strings.Contains(body, "value") {
		t.Errorf("响应体应包含 value，实际: %s", body)
	}
	// Verify DEBUG log on success
	if len(cl.entries) != 1 || cl.entries[0].level != "DEBUG" {
		t.Errorf("期望 1 条 DEBUG 日志，实际: %v", cl.entries)
	}
}

// TestExecuteRaw_Non2xxError verifies ExecuteRaw surfaces HTTP errors for non-2xx responses.
func TestExecuteRaw_Non2xxError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `internal server error`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	// Use a trade operation so it doesn't retry
	hc := NewHttpClient(cfg)

	_, err := hc.ExecuteRaw("place_order", `{"order":"test"}`)
	if err == nil {
		t.Fatal("非 2xx 应返回错误")
	}
	if !strings.Contains(err.Error(), "HTTP 500") {
		t.Errorf("错误信息应包含 HTTP 500，实际: %v", err)
	}
}

// TestExecuteRaw_RetrySuccess verifies ExecuteRaw retries and succeeds on transient failure.
func TestExecuteRaw_RetrySuccess(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			// First request: simulate stale connection by closing the connection
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("server doesn't support hijacking")
			}
			conn, _, _ := hj.Hijack()
			conn.Close()
			return
		}
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	hc := NewHttpClient(cfg)
	hc.retryPolicy = &RetryPolicy{
		MaxRetries: 2,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   20 * time.Millisecond,
	}

	body, err := hc.ExecuteRaw("market_state", `{"market":"US"}`)
	if err != nil {
		t.Fatalf("重试后应成功，实际: %v", err)
	}
	if !strings.Contains(body, "success") {
		t.Errorf("响应体应包含 success，实际: %s", body)
	}
	if requestCount < 2 {
		t.Errorf("应至少发起 2 次请求，实际 %d 次", requestCount)
	}
}


// TestNewQuoteHttpClient_UsesQuoteServerURL verifies the quote client uses QuoteServerURL.
func TestNewQuoteHttpClient_UsesQuoteServerURL(t *testing.T) {
	quoteServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer quoteServer.Close()

	cfg := newTestConfig(t, "http://127.0.0.1:19999") // different URL
	cfg.QuoteServerURL = quoteServer.URL

	hc := NewQuoteHttpClient(cfg)
	if hc.config.ServerURL != quoteServer.URL {
		t.Errorf("Quote client 应使用 QuoteServerURL，实际 ServerURL=%s", hc.config.ServerURL)
	}
	// Verify TokenRefreshDuration is suppressed
	if hc.config.TokenRefreshDuration != 0 {
		t.Errorf("Quote client 不应启用自动刷新，实际 TokenRefreshDuration=%v", hc.config.TokenRefreshDuration)
	}
	// Verify it can make requests
	_, err := hc.Execute(&ApiRequest{Method: "market_state", BizContent: `{}`})
	if err != nil {
		t.Fatalf("Quote client Execute 失败: %v", err)
	}
}

// TestNewQuoteHttpClient_NoQuoteURL verifies the quote client falls back to ServerURL.
func TestNewQuoteHttpClient_NoQuoteURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	// QuoteServerURL is empty
	hc := NewQuoteHttpClient(cfg)
	if hc.config.ServerURL != server.URL {
		t.Errorf("无 QuoteServerURL 时应使用 ServerURL，实际 %s", hc.config.ServerURL)
	}
}

// TestWithSharedTokenFrom_VerifiesTokenSharing verifies the shared token works between clients.
func TestWithSharedTokenFrom_VerifiesTokenSharing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cfg.Token = "primary_token"

	primary := NewHttpClient(cfg)

	// Create a quote client sharing the primary's token
	quote := NewQuoteHttpClient(cfg, WithSharedTokenFrom(primary))

	// Verify they share the same atomicToken pointer
	if primary.atomicToken != quote.atomicToken {
		t.Fatal("primary 和 quote 客户端应共享同一个 atomicToken 指针")
	}

	// Update token on primary, quote should see it
	primary.setToken("updated_token")
	if got := quote.getToken(); got != "updated_token" {
		t.Errorf("共享 token 后 quote 应看到更新，期望 updated_token，实际 %s", got)
	}
}

// TestCalculateBackoff_NegativeCount verifies negative retry count returns BaseDelay.
func TestCalculateBackoff_NegativeCount(t *testing.T) {
	p := DefaultRetryPolicy()
	backoff := p.CalculateBackoff(-1)
	if backoff != p.BaseDelay {
		t.Errorf("负 retryCount 应返回 BaseDelay=%v，实际 %v", p.BaseDelay, backoff)
	}
}

// TestIsStaleConnectionError_NilError verifies nil error returns false.
func TestIsStaleConnectionError_NilError(t *testing.T) {
	if IsStaleConnectionError(nil) {
		t.Error("nil error 不应被视为 stale connection error")
	}
}

// TestIsStaleConnectionError_ConnectionReset verifies "connection reset by peer" is detected.
func TestIsStaleConnectionError_ConnectionReset(t *testing.T) {
	err := errors.New("read tcp 127.0.0.1:1234->127.0.0.1:5678: connection reset by peer")
	if !IsStaleConnectionError(err) {
		t.Error("connection reset by peer 应被视为 stale connection error")
	}
}

// TestIsStaleConnectionError_BrokenPipe verifies "broken pipe" is detected.
func TestIsStaleConnectionError_BrokenPipe(t *testing.T) {
	err := errors.New("write tcp: broken pipe")
	if !IsStaleConnectionError(err) {
		t.Error("broken pipe 应被视为 stale connection error")
	}
}

// TestIsStaleConnectionError_ClosedNetworkConnection verifies "use of closed network connection" is detected.
func TestIsStaleConnectionError_ClosedNetworkConnection(t *testing.T) {
	err := errors.New("use of closed network connection")
	if !IsStaleConnectionError(err) {
		t.Error("use of closed network connection 应被视为 stale connection error")
	}
}

// TestIsStaleConnectionError_NetOpError verifies net.OpError with EOF is detected.
func TestIsStaleConnectionError_NetOpError(t *testing.T) {
	// Create a net.OpError
	opErr := &net.OpError{
		Op:  "read",
		Net: "tcp",
		Err: errors.New("EOF"),
	}
	if !IsStaleConnectionError(opErr) {
		t.Error("net.OpError with EOF 应被视为 stale connection error")
	}
}

// TestIsStaleConnectionError_NetOpErrorReset verifies net.OpError with reset is detected.
func TestIsStaleConnectionError_NetOpErrorReset(t *testing.T) {
	opErr := &net.OpError{
		Op:  "read",
		Net: "tcp",
		Err: errors.New("connection reset"),
	}
	if !IsStaleConnectionError(opErr) {
		t.Error("net.OpError with reset 应被视为 stale connection error")
	}
}

// TestIsStaleConnectionError_EOF verifies io.EOF is detected.
func TestIsStaleConnectionError_EOF(t *testing.T) {
	if !IsStaleConnectionError(io.EOF) {
		t.Error("io.EOF 应被视为 stale connection error")
	}
}

// TestIsStaleConnectionError_UnexpectedEOF verifies io.ErrUnexpectedEOF is detected.
func TestIsStaleConnectionError_UnexpectedEOF(t *testing.T) {
	if !IsStaleConnectionError(io.ErrUnexpectedEOF) {
		t.Error("io.ErrUnexpectedEOF 应被视为 stale connection error")
	}
}

// TestIsStaleConnectionError_RegularError verifies a regular error is not detected.
func TestIsStaleConnectionError_RegularError(t *testing.T) {
	err := errors.New("some random error")
	if IsStaleConnectionError(err) {
		t.Error("普通错误不应被视为 stale connection error")
	}
}

// TestExecute_RetrySuccess verifies Execute retries and succeeds after a transient failure.
func TestExecute_RetrySuccess(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			// First request: simulate stale connection
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("server doesn't support hijacking")
			}
			conn, _, _ := hj.Hijack()
			conn.Close()
			return
		}
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cl := &capturingLogger{}
	hc := NewHttpClient(cfg, WithLogger(cl))
	hc.retryPolicy = &RetryPolicy{
		MaxRetries: 2,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   20 * time.Millisecond,
	}

	resp, err := hc.Execute(&ApiRequest{Method: "market_state", BizContent: `{}`})
	if err != nil {
		t.Fatalf("重试后应成功，实际: %v", err)
	}
	if resp.Code != 0 {
		t.Errorf("期望 code=0，实际为 %d", resp.Code)
	}
	if requestCount < 2 {
		t.Errorf("应至少发起 2 次请求，实际 %d 次", requestCount)
	}

	// Verify INFO log was emitted for retry success
	foundInfo := false
	for _, e := range cl.entries {
		if e.level == "INFO" && strings.Contains(e.msg, "after retry") {
			foundInfo = true
		}
	}
	if !foundInfo {
		t.Errorf("应有一条 INFO 日志提示 retry 成功，实际: %v", cl.entries)
	}
}

// TestExecute_ParseErrorWithHTTPError verifies parse error path when body is non-JSON and HTTP error occurred.
func TestExecute_ParseErrorWithHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprint(w, `<html>502 Bad Gateway</html>`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	hc := NewHttpClient(cfg)

	// Use a trade operation so it doesn't retry
	resp, err := hc.Execute(&ApiRequest{Method: "place_order", BizContent: `{}`})
	if err == nil {
		t.Fatal("非 JSON 响应应返回错误")
	}
	// The error should contain HTTP status context
	if !strings.Contains(err.Error(), "HTTP 502") {
		t.Errorf("错误信息应包含 HTTP 502，实际: %v", err)
	}
	// resp should still be nil since parsing failed
	if resp != nil {
		t.Errorf("解析失败时 resp 应为 nil，实际 %+v", resp)
	}
}

// TestExecute_ResponseSignVerificationFailure verifies signature verification failure returns an error.
func TestExecute_ResponseSignVerificationFailure(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubDER, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	pubB64 := base64.StdEncoding.EncodeToString(pubDER)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a response with an invalid signature
		fmt.Fprint(w, `{"code":0,"message":"success","data":null,"timestamp":1700000000,"sign":"invalid-signature"}`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	cfg.TigerPublicKey = pubB64
	cl := &capturingLogger{}
	hc := NewHttpClient(cfg, WithLogger(cl))

	resp, err := hc.Execute(&ApiRequest{Method: "market_state", BizContent: `{}`})
	if err == nil {
		t.Fatal("无效签名应返回错误")
	}
	if !strings.Contains(err.Error(), "signature verification failed") {
		t.Errorf("错误信息应提示验签失败，实际: %v", err)
	}
	// resp should still be returned (parsed before verification)
	if resp == nil {
		t.Fatal("验签失败时 resp 不应为 nil（已解析）")
	}

	// Verify ERROR log was emitted
	foundError := false
	for _, e := range cl.entries {
		if e.level == "ERROR" && strings.Contains(e.msg, "sign verify") {
			foundError = true
		}
	}
	if !foundError {
		t.Errorf("应有一条 ERROR 日志提示验签失败，实际: %v", cl.entries)
	}
}

// TestExecuteRaw_RetryExhausted verifies ExecuteRaw exhausts retries and returns the last error.
func TestExecuteRaw_RetryExhausted(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	cfg := newTestConfig(t, "http://"+ln.Addr().String())
	cl := &capturingLogger{}
	hc := NewHttpClient(cfg, WithLogger(cl))
	hc.retryPolicy = &RetryPolicy{
		MaxRetries: 2,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   20 * time.Millisecond,
	}

	_, err = hc.ExecuteRaw("market_state", `{"market":"US"}`)
	if err == nil {
		t.Fatal("重试耗尽后应返回错误")
	}

	// Verify ERROR log for exhaustion
	foundExhausted := false
	for _, e := range cl.entries {
		if e.level == "ERROR" && strings.Contains(e.msg, "exhausted") {
			foundExhausted = true
		}
	}
	if !foundExhausted {
		t.Errorf("应有一条 ERROR 日志提示重试耗尽，实际: %v", cl.entries)
	}
}

// TestExecuteRaw_Non2xxWithValidJSON verifies ExecuteRaw surfaces non-2xx with body.
func TestExecuteRaw_Non2xxWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, `service unavailable`)
	}))
	defer server.Close()

	cfg := newTestConfig(t, server.URL)
	hc := NewHttpClient(cfg)

	_, err := hc.ExecuteRaw("place_order", `{"order":"test"}`)
	if err == nil {
		t.Fatal("非 2xx 应返回错误")
	}
	if !strings.Contains(err.Error(), "HTTP 503") {
		t.Errorf("错误信息应包含 HTTP 503，实际: %v", err)
	}
}
