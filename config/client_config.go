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
	envToken      = "TIGEROPEN_TOKEN"
	envTokenFile  = "TIGEROPEN_TOKEN_FILE"
)

// ClientConfig holds authentication credentials and runtime parameters.
type ClientConfig struct {
	TigerID              string        `json:"tiger_id"`
	PrivateKey           string        `json:"private_key"`
	Account              string        `json:"account"`
	License              string        `json:"license"`
	Language             string        `json:"language"`
	Timezone             string        `json:"timezone"`
	DeviceID             string        `json:"device_id"`
	Timeout              time.Duration `json:"-"`
	Token                string        `json:"-"`
	TokenRefreshDuration time.Duration `json:"-"`
	ServerURL            string        `json:"-"`
	QuoteServerURL       string        `json:"-"`
	EnableDynamicDomain  bool          `json:"-"`
	TigerPublicKey       string        `json:"-"`
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

// WithTokenRefreshDuration 设置 Token 刷新间隔
func WithTokenRefreshDuration(d time.Duration) Option {
	return func(c *ClientConfig) { c.TokenRefreshDuration = d }
}

// WithDeviceID sets the device identifier (e.g. MAC address).
func WithDeviceID(id string) Option {
	return func(c *ClientConfig) { c.DeviceID = id }
}

// WithQuoteServerURL sets a dedicated quote server URL.
func WithQuoteServerURL(url string) Option {
	return func(c *ClientConfig) { c.QuoteServerURL = url }
}

// WithEnableDynamicDomain 设置是否启用动态域名获取（默认启用）
func WithEnableDynamicDomain(enable bool) Option {
	return func(c *ClientConfig) { c.EnableDynamicDomain = enable }
}

// WithPropertiesFile 从 properties 配置文件加载配置
func WithPropertiesFile(path string) Option {
	return func(c *ClientConfig) {
		props, err := ParsePropertiesFile(path)
		if err != nil {
			// 文件加载失败时静默跳过，后续校验会捕获必填字段缺失
			return
		}
		applyProperties(c, props)
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
	// Searches: ./tiger_openapi_config.properties, then ~/.tigeropen/tiger_openapi_config.properties
	defaultPaths := []string{
		"tiger_openapi_config.properties",
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

	// Load token from environment variable or token file
	if cfg.Token == "" {
		if v := os.Getenv(envToken); v != "" {
			cfg.Token = v
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
	conn, err := net.Dial("udp", "8.8.8.8:80")
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
