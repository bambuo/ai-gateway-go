package oauth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ai/gateway/internal/config"
	"ai/gateway/internal/logger"
)

const (
	tokenURL  = "https://platform.claude.com/v1/oauth/token"
	clientID  = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
)

var defaultScopes = []string{
	"user:inference",
	"user:profile",
	"user:sessions:claude_code",
	"user:mcp_servers",
	"user:file_upload",
}

type Manager struct {
	mu           sync.RWMutex
	accessToken  string
	refreshToken string
	expiresAt    time.Time
	stopCh       chan struct{}
	stopped      bool
}

func New(cfg config.OAuthConfig) *Manager {
	return &Manager{
		accessToken:  cfg.AccessToken,
		refreshToken: cfg.RefreshToken,
		expiresAt:    time.Unix(cfg.ExpiresAt, 0),
		stopCh:       make(chan struct{}),
	}
}

func (m *Manager) Init(ctx context.Context) error {
	now := time.Now()
	fiveMinutes := 5 * time.Minute

	if m.accessToken != "" && m.expiresAt.After(now.Add(fiveMinutes)) {
		remaining := int(m.expiresAt.Sub(now).Minutes())
		logger.Info("使用已有的访问令牌", "expires_in_min", remaining)
		m.scheduleRefresh()
		return nil
	}

	if m.accessToken != "" {
		logger.Info("访问令牌已过期，正在刷新...")
	} else {
		logger.Info("未提供访问令牌，正在刷新...")
	}

	if err := m.refresh(ctx); err != nil {
		return fmt.Errorf("初始 OAuth 刷新失败: %w", err)
	}
	m.scheduleRefresh()
	return nil
}

func (m *Manager) AccessToken() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if time.Now().After(m.expiresAt) {
		return ""
	}
	return m.accessToken
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.stopped {
		close(m.stopCh)
		m.stopped = true
	}
}

func (m *Manager) scheduleRefresh() {
	go func() {
		for {
			m.mu.RLock()
			refreshIn := time.Until(m.expiresAt) - 5*time.Minute
			if refreshIn < 30*time.Second {
				refreshIn = 30 * time.Second
			}
			stopCh := m.stopCh
			m.mu.RUnlock()

			select {
			case <-time.After(refreshIn):
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				if err := m.refresh(ctx); err != nil {
					logger.Error("OAuth 刷新失败，30 秒后重试", "error", err)
					time.Sleep(30 * time.Second)
				}
				cancel()
			case <-stopCh:
				return
			}
		}
	}()
}
