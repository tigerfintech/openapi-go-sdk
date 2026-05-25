package client

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/signer"
)

const (
	// UserAgentPrefix User-Agent 前缀
	UserAgentPrefix = "openapi-go-sdk-"
	// SDKVersion SDK 版本号
	SDKVersion = "0.3.6"
	// DefaultCharset 默认字符集
	DefaultCharset = "UTF-8"
	// DefaultSignType 默认签名类型
	DefaultSignType = "RSA"
	// DefaultVersion 默认 API 版本
	DefaultVersion = "2.0"

	// methodTokenRefresh Tiger OpenAPI token 刷新接口
	methodTokenRefresh = "user_token_refresh"
)

// HttpClient wraps HTTP requests with signing, retry, and timeout support.
type HttpClient struct {
	config       *config.ClientConfig
	httpClient   *http.Client
	retryPolicy  *RetryPolicy
	publicKey    *rsa.PublicKey
	tokenManager *config.TokenManager // non-nil when auto-refresh is active
}

// NewHttpClient creates an HttpClient instance.
// It loads the tiger public key for response signature verification if configured.
// If cfg.TokenRefreshDuration > 0, background token auto-refresh is started automatically.
func NewHttpClient(cfg *config.ClientConfig) *HttpClient {
	hc := &HttpClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		retryPolicy: DefaultRetryPolicy(),
	}

	// Load tiger public key for response signature verification
	if cfg.TigerPublicKey != "" {
		pubKey, err := signer.LoadPublicKey(cfg.TigerPublicKey)
		if err == nil {
			hc.publicKey = pubKey
		}
	}

	// Auto-start token refresh when TokenRefreshDuration is configured (mirrors Python SDK behaviour)
	if cfg.TokenRefreshDuration > 0 {
		interval := cfg.TokenCheckInterval
		if interval <= 0 {
			interval = 5 * time.Minute
		}
		opts := []config.TokenManagerOption{
			config.WithRefreshDuration(int64(cfg.TokenRefreshDuration.Seconds())),
			config.WithTokenRefreshInterval(interval),
		}
		if cfg.TokenWriter != nil {
			opts = append(opts, config.TokenWriterOption(cfg.TokenWriter))
		}
		if cfg.TokenLoader != nil {
			opts = append(opts, config.TokenLoaderOption(cfg.TokenLoader))
		}
		hc.tokenManager = hc.StartTokenAutoRefresh(nil, opts...)
	}

	return hc
}

// Close stops the background token auto-refresh goroutine (if running).
// Call this when the HttpClient is no longer needed to avoid goroutine leaks.
func (c *HttpClient) Close() {
	if c.tokenManager != nil {
		c.tokenManager.StopAutoRefresh()
	}
}

// buildCommonParams constructs common request parameters.
// If version is non-empty, it overrides the default API version.
func (c *HttpClient) buildCommonParams(apiMethod string, bizContent string, version string) map[string]string {
	v := DefaultVersion
	if version != "" {
		v = version
	}
	params := map[string]string{
		"tiger_id":    c.config.TigerID,
		"method":      apiMethod,
		"charset":     DefaultCharset,
		"sign_type":   DefaultSignType,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     v,
		"biz_content": bizContent,
	}
	if c.config.Language != "" {
		params["language"] = c.config.Language
	}
	if c.config.DeviceID != "" {
		params["device_id"] = c.config.DeviceID
	}
	return params
}

// signParams 对参数进行签名
func (c *HttpClient) signParams(params map[string]string) (string, error) {
	content := signer.GetSignContent(params)
	sign, err := signer.SignWithRSA(c.config.PrivateKey, content)
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}
	return sign, nil
}

// verifyResponseSign verifies the response signature using the tiger public key.
// If the public key is not loaded or the response has no sign field, verification is skipped.
func (c *HttpClient) verifyResponseSign(resp *ApiResponse, requestTimestamp string) error {
	if c.publicKey == nil || resp.Sign == "" || requestTimestamp == "" {
		return nil
	}
	if err := signer.VerifyWithRSA(c.publicKey, requestTimestamp, resp.Sign); err != nil {
		return fmt.Errorf("response signature verification failed: %w", err)
	}
	return nil
}

// Execute 执行结构化 API 请求，返回解析后的 ApiResponse
func (c *HttpClient) Execute(request *ApiRequest) (*ApiResponse, error) {
	params := c.buildCommonParams(request.Method, request.BizContent, request.Version)
	requestTimestamp := params["timestamp"]
	sign, err := c.signParams(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign

	var lastErr error
	maxAttempts := 1
	if c.retryPolicy.ShouldRetry(request.Method) {
		maxAttempts = c.retryPolicy.MaxRetries + 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			backoff := c.retryPolicy.CalculateBackoff(attempt - 1)
			time.Sleep(backoff)
		}

		resp, err := c.doHTTPPost(params)
		if err != nil {
			lastErr = err
			if !c.retryPolicy.ShouldRetry(request.Method) {
				return nil, err
			}
			continue
		}

		apiResp, parseErr := ParseApiResponse(resp)
		if parseErr != nil {
			return apiResp, parseErr
		}

		// Verify response signature
		if err := c.verifyResponseSign(apiResp, requestTimestamp); err != nil {
			return apiResp, err
		}

		return apiResp, nil
	}

	return nil, lastErr
}

// ExecuteRaw 通用 API 调用方法
// apiMethod: API 方法名（如 "market_state"、"place_order"）
// requestJSON: 原始 biz_content JSON 字符串
// 返回原始 response JSON 字符串，不做任何解析
func (c *HttpClient) ExecuteRaw(apiMethod string, requestJSON string) (string, error) {
	// 参数校验
	if apiMethod == "" {
		return "", &TigerError{Code: -1, Message: "api_method 不能为空", Category: CategoryUnknown}
	}
	if !json.Valid([]byte(requestJSON)) {
		return "", &TigerError{Code: -1, Message: "request_json 不是有效的 JSON", Category: CategoryUnknown}
	}

	params := c.buildCommonParams(apiMethod, requestJSON, "")
	sign, err := c.signParams(params)
	if err != nil {
		return "", err
	}
	params["sign"] = sign

	var lastErr error
	maxAttempts := 1
	if c.retryPolicy.ShouldRetry(apiMethod) {
		maxAttempts = c.retryPolicy.MaxRetries + 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			backoff := c.retryPolicy.CalculateBackoff(attempt - 1)
			time.Sleep(backoff)
		}

		body, err := c.doHTTPPost(params)
		if err != nil {
			lastErr = err
			if !c.retryPolicy.ShouldRetry(apiMethod) {
				return "", err
			}
			continue
		}

		return string(body), nil
	}

	return "", lastErr
}

// doHTTPPost 发送 HTTP POST 请求
func (c *HttpClient) doHTTPPost(params map[string]string) ([]byte, error) {
	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("序列化请求参数失败: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.ServerURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("User-Agent", UserAgentPrefix+SDKVersion)
	if c.config.Token != "" {
		req.Header.Set("Authorization", c.config.Token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	return body, nil
}

// QueryToken 调用 user_token_refresh 接口获取新 token，返回新 token 字符串。
// 此方法仅查询，不修改 config 中的 token。
func (c *HttpClient) QueryToken() (string, error) {
	req, err := NewApiRequest(methodTokenRefresh, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.Execute(req)
	if err != nil {
		return "", fmt.Errorf("token 刷新请求失败: %w", err)
	}
	var data struct {
		Token string `json:"token"`
	}
	if err := UnmarshalData(resp.Data, &data); err != nil {
		return "", fmt.Errorf("解析 token 响应失败: %w", err)
	}
	if data.Token == "" {
		return "", fmt.Errorf("服务端返回空 token")
	}
	return data.Token, nil
}

// RefreshToken 刷新 token：调用接口获取新 token 后更新 config.Token 并持久化到文件。
// tokenManager 可为 nil，此时只更新内存中的 config.Token，不写文件。
// 若 NewHttpClient 已自动启动内部 TokenManager，其内存 token 也会同步更新。
func (c *HttpClient) RefreshToken(tokenManager *config.TokenManager) error {
	oldToken := c.config.Token
	newToken, err := c.QueryToken()
	if err != nil {
		return err
	}
	log.Printf("[token] refreshed (len %d -> %d)", len(oldToken), len(newToken))
	c.config.Token = newToken
	// Keep internal tokenManager in sync so ShouldTokenRefresh stays accurate.
	if c.tokenManager != nil && tokenManager != c.tokenManager {
		c.tokenManager.SyncToken(newToken)
	}
	if tokenManager != nil {
		return tokenManager.SetToken(newToken)
	}
	return nil
}

// StartTokenAutoRefresh 启动后台 token 自动刷新。
//
// tokenManager 可为 nil：SDK 内部自动创建一个仅内存的 TokenManager，
// 适合直接通过 WithToken 在代码里设置 token、不使用文件的场景。
// opts 用于配置内部 TokenManager（仅在 tokenManager 为 nil 时生效），
// 常见用途是注入 WithTokenWriter 回调或调整 WithRefreshDuration/WithTokenRefreshInterval。
//
// 传入 tokenManager 时，opts 被忽略；刷新后 token 同时写入文件并触发其上注册的回调。
// 无论哪种方式，必须保证 TokenManager 上的 WithRefreshDuration > 0，否则不会触发刷新。
func (c *HttpClient) StartTokenAutoRefresh(tokenManager *config.TokenManager, opts ...config.TokenManagerOption) *config.TokenManager {
	if tokenManager == nil {
		tokenManager = config.NewTokenManager(opts...)
		// 从 config.Token 同步当前 token，使 ShouldTokenRefresh 能正确判断
		if c.config.Token != "" {
			_ = tokenManager.SetToken(c.config.Token)
		}
	}
	tokenManager.StartAutoRefresh(func() (string, error) {
		newToken, err := c.QueryToken()
		if err != nil {
			return "", err
		}
		c.config.Token = newToken
		return newToken, nil
	})
	return tokenManager
}

func safePrefix(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// NewQuoteHttpClient creates an HttpClient that uses the QuoteServerURL.
// It shares the token auto-refresh goroutine with the original config rather than
// starting a second one — the caller is responsible for managing the lifecycle via
// the primary HttpClient's Close().
func NewQuoteHttpClient(cfg *config.ClientConfig) *HttpClient {
	cloned := *cfg
	if cloned.QuoteServerURL != "" {
		cloned.ServerURL = cloned.QuoteServerURL
	}
	// Suppress auto-start inside NewHttpClient; the primary client already owns the
	// refresh goroutine (if any). Token updates via config.Token pointer are shared
	// because both clients hold a pointer to the same underlying config.
	cloned.TokenRefreshDuration = 0
	return NewHttpClient(&cloned)
}
