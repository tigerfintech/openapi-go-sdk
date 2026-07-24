package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	// Default values
	defaultLanguage  = "zh_CN"
	defaultTimeout   = 15 * time.Second
	defaultServerURL = "https://openapi.tigerfintech.com/gateway"

	// Tiger public key for response signature verification
	tigerPublicKey = "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDNF3G8SoEcCZh2rshUbayDgLLrj6rKgzNMxDL2HSnKcB0+GPOsndqSv+a4IBu9+I3fyBp5hkyMMG2+AXugd9pMpy6VxJxlNjhX1MYbNTZJUT4nudki4uh+LMOkIBHOceGNXjgB+cXqmlUnjlqha/HgboeHSnSgpM3dKSJQlIOsDwIDAQAB"

	// Environment variable names
	envTigerID    = "TIGEROPEN_TIGER_ID"
	envPrivateKey = "TIGEROPEN_PRIVATE_KEY"
	envAccount    = "TIGEROPEN_ACCOUNT"
	envSecretKey  = "TIGEROPEN_SECRET_KEY"
	envToken      = "TIGEROPEN_TOKEN"
	envTokenFile  = "TIGEROPEN_TOKEN_FILE"
)

// ClientConfig holds authentication credentials and runtime parameters.
type ClientConfig struct {
	TigerID              string                 `json:"tiger_id"`
	PrivateKey           string                 `json:"-"`
	Account              string                 `json:"account"`
	SecretKey            string                 `json:"-"`
	License              string                 `json:"license"`
	Language             string                 `json:"language"`
	Timezone             string                 `json:"timezone"`
	DeviceID             string                 `json:"device_id"`
	Timeout              time.Duration          `json:"-"`
	Token                string                 `json:"-"`
	TokenRefreshDuration time.Duration          `json:"-"`
	TokenCheckInterval   time.Duration          `json:"-"` // 后台检查间隔，默认 5m，仅 TokenRefreshDuration>0 时生效
	TokenWriter          func(token string)     `json:"-"` // token 刷新写入后的可选回调
	TokenLoader          func() (string, error) `json:"-"` // 自定义 token 加载函数，替代默认的文件加载
	ServerURL            string                 `json:"-"`
	QuoteServerURL       string                 `json:"-"`
	EnableDynamicDomain  bool                   `json:"-"`
	TigerPublicKey       string                 `json:"-"`
	// explicitPropertiesFile is set when WithPropertiesFile was called, so auto-discovery
	// skips the local ./tiger_openapi_config.properties to prevent accidental contamination.
	explicitPropertiesFile bool
}

// Option 配置选项函数类型
type Option func(*ClientConfig)

// WithTigerID 设置开发者 ID
func WithTigerID(id string) Option {
	return func(c *ClientConfig) { c.TigerID = id }
}

// WithPrivateKey 设置 RSA 私钥
func WithPrivateKey(key string) Option {
	return func(c *ClientConfig) { c.PrivateKey = key }
}

// WithAccount 设置交易账户
func WithAccount(account string) Option {
	return func(c *ClientConfig) { c.Account = account }
}

// WithSecretKey 设置机构账户 Secret Key（机构用户交易鉴权使用）。
func WithSecretKey(key string) Option {
	return func(c *ClientConfig) { c.SecretKey = key }
}

// WithLicense 设置牌照类型
func WithLicense(license string) Option {
	return func(c *ClientConfig) { c.License = license }
}

// WithLanguage 设置语言（zh_CN/zh_TW/en_US）
func WithLanguage(lang string) Option {
	return func(c *ClientConfig) { c.Language = lang }
}

// WithTimezone 设置时区
func WithTimezone(tz string) Option {
	return func(c *ClientConfig) { c.Timezone = tz }
}

// WithTimeout sets the request timeout duration.
func WithTimeout(d time.Duration) Option {
	return func(c *ClientConfig) { c.Timeout = d }
}

// WithToken sets the TBHK license token.
func WithToken(token string) Option {
	return func(c *ClientConfig) { c.Token = token }
}

// WithTokenRefreshDuration 设置 Token 刷新阈值。
// Token 生成时间超过此值时触发刷新；默认 0 表示不刷新。
func WithTokenRefreshDuration(d time.Duration) Option {
	return func(c *ClientConfig) { c.TokenRefreshDuration = d }
}

// WithTokenCheckInterval 设置后台 token 刷新检查间隔，默认 5 分钟。
func WithTokenCheckInterval(d time.Duration) Option {
	return func(c *ClientConfig) { c.TokenCheckInterval = d }
}

// WithTokenWriter 设置 token 刷新写入后的回调函数。
// 仅在 TokenRefreshDuration > 0 时生效，由 NewHttpClient 自动注册到内部 TokenManager。
func WithTokenWriter(fn func(token string)) Option {
	return func(c *ClientConfig) { c.TokenWriter = fn }
}

// WithTokenLoader 设置自定义 token 加载函数，替代默认的文件加载。
// 由 NewHttpClient 自动注册到内部 TokenManager，LoadToken 时优先执行此函数。
func WithTokenLoader(fn func() (string, error)) Option {
	return func(c *ClientConfig) { c.TokenLoader = fn }
}

// WithDeviceID sets the device identifier (e.g. MAC address).
func WithDeviceID(id string) Option {
	return func(c *ClientConfig) { c.DeviceID = id }
}

// WithTigerPublicKey overrides the Tiger server public key used for response signature verification.
// Passing an empty string is a no-op (the default key is kept).
func WithTigerPublicKey(key string) Option {
	return func(c *ClientConfig) {
		if key != "" {
			c.TigerPublicKey = key
		}
	}
}

// WithServerURL sets a dedicated trade/common server URL (overrides the default production gateway).
func WithServerURL(url string) Option {
	return func(c *ClientConfig) { c.ServerURL = url }
}

// WithQuoteServerURL sets a dedicated quote server URL.
func WithQuoteServerURL(url string) Option {
	return func(c *ClientConfig) { c.QuoteServerURL = url }
}

// WithEnableDynamicDomain 设置是否启用动态域名获取（默认启用）
func WithEnableDynamicDomain(enable bool) Option {
	return func(c *ClientConfig) { c.EnableDynamicDomain = enable }
}

// WithPropertiesFile 从 properties 配置文件加载配置。
// 调用后，auto-discovery 会跳过本地 ./tiger_openapi_config.properties，
// 防止开发目录中的测试配置文件污染显式指定的配置。
func WithPropertiesFile(path string) Option {
	return func(c *ClientConfig) {
		props, err := ParsePropertiesFile(path)
		if err != nil {
			// 文件加载失败时静默跳过，后续校验会捕获必填字段缺失
			return
		}
		applyProperties(c, props)
		c.explicitPropertiesFile = true
	}
}

// applyProperties 将 properties 键值对应用到配置对象
func applyProperties(c *ClientConfig, props map[string]string) {
	if v, ok := props["tiger_id"]; ok && c.TigerID == "" {
		c.TigerID = v
	}
	// 私钥优先级：private_key > private_key_pk8 > private_key_pk1
	if c.PrivateKey == "" {
		if v, ok := props["private_key"]; ok {
			c.PrivateKey = v
		} else if v, ok := props["private_key_pk8"]; ok {
			c.PrivateKey = v
		} else if v, ok := props["private_key_pk1"]; ok {
			c.PrivateKey = v
		}
	}
	if v, ok := props["account"]; ok && c.Account == "" {
		c.Account = v
	}
	if v, ok := props["secret_key"]; ok && c.SecretKey == "" {
		c.SecretKey = v
	}
	if v, ok := props["server_url"]; ok && c.ServerURL == "" {
		c.ServerURL = v
	}
	if v, ok := props["quote_server_url"]; ok && c.QuoteServerURL == "" {
		c.QuoteServerURL = v
	}
	if v, ok := props["license"]; ok && c.License == "" {
		c.License = v
	}
	if v, ok := props["language"]; ok && c.Language == "" {
		c.Language = v
	}
	if v, ok := props["timezone"]; ok && c.Timezone == "" {
		c.Timezone = v
	}
}

// NewClientConfig creates a client config.
// Priority: environment variables > explicit options (incl. WithPropertiesFile) > auto-discovered config file > defaults.
// Returns an error if required fields tiger_id or private_key are empty.
func NewClientConfig(opts ...Option) (*ClientConfig, error) {
	cfg := &ClientConfig{
		EnableDynamicDomain: true, // default: enable dynamic domain
	}

	// Apply explicit options first (code settings and WithPropertiesFile)
	for _, opt := range opts {
		opt(cfg)
	}

	// Auto-discover config file for any fields still empty.
	// Searches: ./tiger_openapi_config.properties, then ~/.tigeropen/tiger_openapi_config.properties.
	// When WithPropertiesFile was used, skip ./tiger_openapi_config.properties to prevent local
	// development configs from contaminating explicitly specified credentials.
	var defaultPaths []string
	if !cfg.explicitPropertiesFile {
		defaultPaths = append(defaultPaths, "tiger_openapi_config.properties")
	}
	if home, err := os.UserHomeDir(); err == nil {
		defaultPaths = append(defaultPaths, filepath.Join(home, ".tigeropen", "tiger_openapi_config.properties"))
	}
	for _, p := range defaultPaths {
		if _, err := os.Stat(p); err == nil {
			props, err := ParsePropertiesFile(p)
			if err == nil {
				applyProperties(cfg, props)
			}
			break
		}
	}

	// 环境变量覆盖（最高优先级）
	if v := os.Getenv(envTigerID); v != "" {
		cfg.TigerID = v
	}
	if v := os.Getenv(envPrivateKey); v != "" {
		cfg.PrivateKey = v
	}
	if v := os.Getenv(envAccount); v != "" {
		cfg.Account = v
	}
	if v := os.Getenv(envSecretKey); v != "" {
		cfg.SecretKey = v
	}

	// Load token: env var > WithTokenLoader > token file
	if cfg.Token == "" {
		if v := os.Getenv(envToken); v != "" {
			cfg.Token = v
		} else if cfg.TokenLoader != nil {
			if token, err := cfg.TokenLoader(); err == nil && token != "" {
				cfg.Token = token
			}
		} else {
			tokenFilePath := os.Getenv(envTokenFile)
			if tokenFilePath == "" {
				tokenFilePath = defaultTokenFileName
			}
			tm := NewTokenManager(WithTokenFilePath(tokenFilePath))
			if token, err := tm.LoadToken(); err == nil && token != "" {
				cfg.Token = token
			}
		}
	}

	// 设置默认值
	if cfg.Language == "" {
		cfg.Language = defaultLanguage
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}

	// Auto-detect device ID from MAC address if not set
	if cfg.DeviceID == "" {
		cfg.DeviceID = detectDeviceID()
	}

	// 确定服务器地址：动态域名 > 默认
	var domainConf map[string]interface{}
	if cfg.ServerURL == "" || cfg.QuoteServerURL == "" {
		if cfg.EnableDynamicDomain {
			domainConf = QueryDomains(cfg.License)
		}
	}

	if cfg.ServerURL == "" {
		if domainConf != nil {
			if dynamicURL := resolveDynamicServerURL(domainConf, cfg.License); dynamicURL != "" {
				cfg.ServerURL = dynamicURL
			}
		}
		// 动态域名获取失败或未启用，使用默认地址
		if cfg.ServerURL == "" {
			cfg.ServerURL = defaultServerURL
		}
	}

	// Resolve quote server URL from dynamic domains
	if cfg.QuoteServerURL == "" && domainConf != nil {
		if quoteURL := resolveDynamicQuoteServerURL(domainConf, cfg.License); quoteURL != "" {
			cfg.QuoteServerURL = quoteURL
		}
	}
	// Fallback: use ServerURL if QuoteServerURL is still empty
	if cfg.QuoteServerURL == "" {
		cfg.QuoteServerURL = cfg.ServerURL
	}

	// Set default tiger public key for response signature verification
	if cfg.TigerPublicKey == "" {
		cfg.TigerPublicKey = tigerPublicKey
	}

	// 校验必填字段
	if cfg.TigerID == "" {
		return nil, fmt.Errorf("tiger_id 不能为空，请通过 WithTigerID 或环境变量 %s 设置", envTigerID)
	}
	if cfg.PrivateKey == "" {
		return nil, fmt.Errorf("private_key 不能为空，请通过 WithPrivateKey 或环境变量 %s 设置", envPrivateKey)
	}

	return cfg, nil
}

// detectDeviceID tries to find the MAC address of the network interface
// used for the default route. Falls back to the first interface with a MAC.
func detectDeviceID() string {
	// Try to find MAC of the interface used for default route
	conn, err := net.DialTimeout("udp", "8.8.8.8:80", 100*time.Millisecond)
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		interfaces, _ := net.Interfaces()
		for _, iface := range interfaces {
			if len(iface.HardwareAddr) == 0 {
				continue
			}
			addrs, _ := iface.Addrs()
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.Equal(localAddr.IP) {
					return iface.HardwareAddr.String()
				}
			}
		}
	}
	// Fallback: first interface with a MAC
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		if len(iface.HardwareAddr) > 0 {
			return iface.HardwareAddr.String()
		}
	}
	return ""
}
