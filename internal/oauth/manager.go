package oauth

import (
	"context"
	"sync"
	"time"

	"ai/gateway/internal/config"
	"ai/gateway/internal/logger"
)

const (
	tokenURL = "https://platform.claude.com/v1/oauth/token"
	clientID = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
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

func (m *Manager) Init() {
	now := time.Now()
	fiveMinutes := 5 * time.Minute

	if m.accessToken != "" && m.expiresAt.After(now.Add(fiveMinutes)) {
		remaining := int(m.expiresAt.Sub(now).Minutes())
		logger.Info("使用已有访问令牌", "expires_in_min", remaining)
		m.scheduleRefresh()
		return
	}

	if m.accessToken != "" {
		logger.Info("访问令牌即将过期 — 首次请求时按需刷新")
	} else {
		logger.Info("未缓存访问令牌 — 首次请求时按需刷新")
	}

	m.scheduleRefresh()
}

func (m *Manager) AccessToken() string {
	m.mu.RLock()
	token := m.accessToken
	expired := m.accessToken == "" || time.Now().After(m.expiresAt)
	m.mu.RUnlock()

	if expired {
		logger.Info("未找到有效访问令牌，正在按需刷新...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := m.refresh(ctx); err != nil {
			logger.Error("OAuth 按需刷新失败", "error", err)
			return ""
		}
		m.mu.RLock()
		token = m.accessToken
		m.mu.RUnlock()
	}

	return token
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
			expiresAt := m.expiresAt
			hasToken := m.accessToken != ""
			stopCh := m.stopCh
			m.mu.RUnlock()

			var refreshIn time.Duration
			if hasToken {
				refreshIn = time.Until(expiresAt) - 5*time.Minute
				if refreshIn < 30*time.Second {
					refreshIn = 30 * time.Second
				}
			} else {
				refreshIn = 5 * time.Minute
			}

			select {
			case <-time.After(refreshIn):
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				if err := m.refresh(ctx); err != nil {
					logger.Debug("OAuth 刷新推迟 — 将在下次请求时重试", "error", err)
				}
				cancel()
			case <-stopCh:
				return
			}
		}
	}()
}
