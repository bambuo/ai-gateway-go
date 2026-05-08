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
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if len(cfg.Identity.DeviceID) != 64 ||
		strings.Contains(cfg.Identity.DeviceID, "0000000000") {
		return nil, fmt.Errorf("config: identity.device_id must be a real 64-char hex value. Run: gateway gen-identity")
	}
	if len(cfg.Auth.Tokens) == 0 {
		return nil, fmt.Errorf("config: auth.tokens must have at least one entry")
	}
	if cfg.OAuth.RefreshToken == "" {
		return nil, fmt.Errorf("config: oauth.refresh_token is required. Do a browser OAuth login on the admin machine, then copy the refresh token from ~/.claude/.credentials.json, or run: gateway gen-config")
	}

	return &cfg, nil
}
