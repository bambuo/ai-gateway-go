package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port int       `yaml:"port"`
	TLS  TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type UpstreamConfig struct {
	URL string `yaml:"url"`
}

type OAuthConfig struct {
	AccessToken  string `yaml:"access_token"`
	RefreshToken string `yaml:"refresh_token"`
	ExpiresAt    int64  `yaml:"expires_at"`
}

type TokenEntry struct {
	Name  string `yaml:"name"`
	Token string `yaml:"token"`
}

type AuthConfig struct {
	Tokens []TokenEntry `yaml:"tokens"`
}

type IdentityConfig struct {
	DeviceID string `yaml:"device_id"`
	Email    string `yaml:"email"`
}

type EnvConfig struct {
	Platform         string `yaml:"platform"`
	PlatformRaw      string `yaml:"platform_raw"`
	Arch             string `yaml:"arch"`
	NodeVersion      string `yaml:"node_version"`
	Terminal         string `yaml:"terminal"`
	PackageManagers  string `yaml:"package_managers"`
	Runtimes         string `yaml:"runtimes"`
	IsRunningWithBun bool   `yaml:"is_running_with_bun"`
	IsCI             bool   `yaml:"is_ci"`
	IsClaudeAiAuth   bool   `yaml:"is_claude_ai_auth"`
	Version          string `yaml:"version"`
	VersionBase      string `yaml:"version_base"`
	BuildTime        string `yaml:"build_time"`
	DeploymentEnv    string `yaml:"deployment_environment"`
	VCS              string `yaml:"vcs"`
}

type PromptEnvConfig struct {
	Platform   string `yaml:"platform"`
	Shell      string `yaml:"shell"`
	OSVersion  string `yaml:"os_version"`
	WorkingDir string `yaml:"working_dir"`
}

type ProcessConfig struct {
	ConstrainedMemory uint64   `yaml:"constrained_memory"`
	RSSRange          [2]int64 `yaml:"rss_range"`
	HeapTotalRange    [2]int64 `yaml:"heap_total_range"`
	HeapUsedRange     [2]int64 `yaml:"heap_used_range"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	Audit bool   `yaml:"audit"`
}

type AdminConfig struct {
	DBPath string `yaml:"db_path"`
}

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Upstream  UpstreamConfig  `yaml:"upstream"`
	OAuth     OAuthConfig     `yaml:"oauth"`
	Auth      AuthConfig      `yaml:"auth"`
	Identity  IdentityConfig  `yaml:"identity"`
	Env       EnvConfig       `yaml:"env"`
	PromptEnv PromptEnvConfig `yaml:"prompt_env"`
	Process   ProcessConfig   `yaml:"process"`
	Logging   LoggingConfig   `yaml:"logging"`
	Admin     AdminConfig     `yaml:"admin"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件 %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件: %w", err)
	}

	if len(cfg.Identity.DeviceID) != 64 ||
		strings.Contains(cfg.Identity.DeviceID, "0000000000") {
		return nil, fmt.Errorf("配置错误: identity.device_id 必须是一个有效的 64 位十六进制值。请运行: gateway gen-identity")
	}
	if len(cfg.Auth.Tokens) == 0 {
		return nil, fmt.Errorf("配置错误: auth.tokens 至少需要包含一个条目")
	}
	if cfg.OAuth.RefreshToken == "" {
		return nil, fmt.Errorf("配置错误: oauth.refresh_token 是必需的。请在管理机器上通过浏览器完成 OAuth 登录，然后从 ~/.claude/.credentials.json 中复制 refresh token，或运行: gateway gen-config")
	}

	return &cfg, nil
}
