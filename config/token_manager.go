package config

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultTokenFileName        = "tiger_openapi_token.properties"
	defaultTokenRefreshInterval = 24 * time.Hour
)

// TokenManager Token 管理器
// 支持从文件或自定义来源加载 Token，并可在后台定期刷新。
type TokenManager struct {
	mu              sync.RWMutex
	token           string
	filePath        string
	fileEnabled     bool // true only when WithTokenFilePath was explicitly called
	interval        time.Duration
	stopCh          chan struct{}
	refreshDuration int64
	refreshFn       func() (string, error)
	tokenLoader     func() (string, error)
	tokenWriter     func(token string)
}

// TokenManagerOption TokenManager 配置选项
type TokenManagerOption func(*TokenManager)

// WithTokenFilePath 设置 Token 文件路径，并启用文件持久化。
func WithTokenFilePath(path string) TokenManagerOption {
	return func(m *TokenManager) {
		m.filePath = path
		m.fileEnabled = true
	}
}

// WithTokenRefreshInterval 设置 Token 检查间隔
func WithTokenRefreshInterval(d time.Duration) TokenManagerOption {
	return func(m *TokenManager) { m.interval = d }
}

// TokenLoaderOption 返回设置自定义 token 加载函数的 TokenManagerOption。
func TokenLoaderOption(fn func() (string, error)) TokenManagerOption {
	return func(m *TokenManager) { m.tokenLoader = fn }
}

// TokenWriterOption 返回设置 token 写入回调的 TokenManagerOption。
func TokenWriterOption(fn func(token string)) TokenManagerOption {
	return func(m *TokenManager) { m.tokenWriter = fn }
}

// WithRefreshDuration 设置 Token 刷新阈值（秒），当 token 生成时间超过此值时触发刷新。
// 0 表示不刷新，最小值 30 秒。
func WithRefreshDuration(seconds int64) TokenManagerOption {
	return func(m *TokenManager) {
		if seconds > 0 && seconds < 30 {
			seconds = 30
		}
		m.refreshDuration = seconds
	}
}

// NewTokenManager 创建 Token 管理器
func NewTokenManager(opts ...TokenManagerOption) *TokenManager {
	m := &TokenManager{
		filePath: defaultTokenFileName,
		interval: defaultTokenRefreshInterval,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// LoadToken 加载 Token。
// 若设置了 WithTokenLoader，优先调用自定义加载函数；否则从 properties 文件读取。
func (m *TokenManager) LoadToken() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.tokenLoader != nil {
		token, err := m.tokenLoader()
		if err != nil {
			return "", fmt.Errorf("自定义 token 加载失败: %w", err)
		}
		if token == "" {
			return "", fmt.Errorf("自定义 token 加载返回空值")
		}
		m.token = token
		return token, nil
	}

	props, err := ParsePropertiesFile(m.filePath)
	if err != nil {
		return "", fmt.Errorf("加载 Token 文件失败: %w", err)
	}

	token, ok := props["token"]
	if !ok || token == "" {
		return "", fmt.Errorf("Token 文件中未找到 token 字段")
	}

	m.token = token
	return token, nil
}

// GetToken 获取当前 Token
func (m *TokenManager) GetToken() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.token
}

// SetToken 设置 Token；若启用了文件持久化（WithTokenFilePath），则同步写文件。
// 成功后触发 TokenWriter 回调（如有）。
func (m *TokenManager) SetToken(token string) error {
	m.mu.Lock()
	m.token = token
	var err error
	if m.fileEnabled {
		err = m.saveTokenToFile(token)
	}
	cb := m.tokenWriter
	m.mu.Unlock()

	if err == nil && cb != nil {
		cb(token)
	}
	return err
}

// SyncToken 仅更新内存中的 token，不写文件，不触发回调。
// 用于多个组件共享 token 时的内部同步（如 RefreshToken 同步内部 TokenManager）。
func (m *TokenManager) SyncToken(token string) {
	m.mu.Lock()
	m.token = token
	m.mu.Unlock()
}

// saveTokenToFile 将 Token 保存到 properties 文件
func (m *TokenManager) saveTokenToFile(token string) error {
	dir := filepath.Dir(m.filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}
	content := fmt.Sprintf("token=%s\n", token)
	return os.WriteFile(m.filePath, []byte(content), 0644)
}

// ShouldTokenRefresh returns true when the token's gen_ts is older than refreshDuration.
func (m *TokenManager) ShouldTokenRefresh() bool {
	m.mu.RLock()
	token := m.token
	dur := m.refreshDuration
	m.mu.RUnlock()

	if token == "" || dur == 0 {
		return false
	}

	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return false
	}
	if len(decoded) < 27 {
		return false
	}

	parts := strings.SplitN(string(decoded[:27]), ",", 2)
	if len(parts) < 2 {
		return false
	}

	genTs, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		return false
	}

	return (time.Now().Unix() - genTs/1000) > dur
}

// StartAutoRefresh 启动后台定期刷新。
// 若已有刷新 goroutine 在运行，先停止旧的再启动新的，避免 goroutine 泄漏。
func (m *TokenManager) StartAutoRefresh(refreshFn func() (string, error)) {
	m.mu.Lock()
	if m.stopCh != nil {
		close(m.stopCh)
	}
	m.refreshFn = refreshFn
	m.stopCh = make(chan struct{})
	ch := m.stopCh
	m.mu.Unlock()

	go m.refreshLoopWith(ch)
}

// StopAutoRefresh 停止后台刷新
func (m *TokenManager) StopAutoRefresh() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopCh != nil {
		close(m.stopCh)
		m.stopCh = nil
	}
}

func (m *TokenManager) refreshLoopWith(stopCh <-chan struct{}) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			if !m.ShouldTokenRefresh() {
				continue
			}
			m.mu.RLock()
			fn := m.refreshFn
			m.mu.RUnlock()
			if fn != nil {
				if newToken, err := fn(); err == nil && newToken != "" {
					_ = m.SetToken(newToken)
				} else if err != nil {
					log.Printf("[token_manager] auto-refresh failed: %v", err)
				}
			}
		}
	}
}
