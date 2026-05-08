package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"ai/gateway/internal/logger"
)

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func proxyHTTPClient() *http.Client {
	proxyURL := os.Getenv("HTTPS_PROXY")
	if proxyURL == "" {
		proxyURL = os.Getenv("https_proxy")
	}
	if proxyURL == "" {
		proxyURL = os.Getenv("HTTP_PROXY")
	}
	if proxyURL == "" {
		proxyURL = os.Getenv("http_proxy")
	}
	if proxyURL == "" {
		return &http.Client{Timeout: 30 * time.Second}
	}

	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return &http.Client{Timeout: 30 * time.Second}
	}

	logger.Info("OAuth 使用出站代理", "url", proxyURL)
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{Proxy: http.ProxyURL(parsed)},
	}
}

func (m *Manager) refresh(ctx context.Context) error {
	body := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": m.refreshToken,
		"client_id":     clientID,
		"scope":         strings.Join(defaultScopes, " "),
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化 OAuth 请求体: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("创建 OAuth 请求: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := proxyHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("OAuth 刷新请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取 OAuth 响应: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OAuth 刷新失败 (%d): %s", resp.StatusCode, string(respBody))
	}

	var result refreshResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("解码 OAuth 响应: %w", err)
	}

	m.mu.Lock()
	m.accessToken = result.AccessToken
	if result.RefreshToken != "" {
		m.refreshToken = result.RefreshToken
	}
	m.expiresAt = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	m.mu.Unlock()

	logger.Info("OAuth 令牌已刷新", "expires_at", m.expiresAt.Format(time.RFC3339))
	return nil
}
