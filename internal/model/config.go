package model

import (
	"encoding/json"
	"time"
)

type AppConfig struct {
	ID        uint      `gorm:"primaryKey"`
	Config    string    `gorm:"type:text;not null"`
	Version   int       `gorm:"not null;default:1"`
	IsActive  bool      `gorm:"not null;default:true"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type ConfigWrapper struct {
	Config []byte
}

func (AppConfig) TableName() string {
	return "app_configs"
}

type FullConfig struct {
	Server    ServerConfigSection    `json:"server"`
	Upstream  UpstreamConfigSection  `json:"upstream"`
	OAuth     OAuthConfigSection     `json:"oauth"`
	Auth      AuthConfigSection      `json:"auth"`
	Identity  IdentityConfigSection  `json:"identity"`
	Env       EnvConfigSection       `json:"env"`
	PromptEnv PromptEnvConfigSection `json:"prompt_env"`
	Process   ProcessConfigSection   `json:"process"`
	Logging   LoggingConfigSection   `json:"logging"`
}

type ServerConfigSection struct {
	Port int              `json:"port"`
	TLS  TLSConfigSection `json:"tls"`
}

type TLSConfigSection struct {
	Cert string `json:"cert"`
	Key  string `json:"key"`
}

type UpstreamConfigSection struct {
	URL string `json:"url"`
}

type OAuthConfigSection struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type AuthConfigSection struct {
	Tokens []TokenEntrySection `json:"tokens"`
}

type TokenEntrySection struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

type IdentityConfigSection struct {
	DeviceID string `json:"device_id"`
	Email    string `json:"email"`
}

type EnvConfigSection struct {
	Platform         string `json:"platform"`
	PlatformRaw      string `json:"platform_raw"`
	Arch             string `json:"arch"`
	NodeVersion      string `json:"node_version"`
	Terminal         string `json:"terminal"`
	PackageManagers  string `json:"package_managers"`
	Runtimes         string `json:"runtimes"`
	IsRunningWithBun bool   `json:"is_running_with_bun"`
	IsCI             bool   `json:"is_ci"`
	IsClaudeAiAuth   bool   `json:"is_claude_ai_auth"`
	Version          string `json:"version"`
	VersionBase      string `json:"version_base"`
	BuildTime        string `json:"build_time"`
	DeploymentEnv    string `json:"deployment_environment"`
	VCS              string `json:"vcs"`
}

type PromptEnvConfigSection struct {
	Platform   string `json:"platform"`
	Shell      string `json:"shell"`
	OSVersion  string `json:"os_version"`
	WorkingDir string `json:"working_dir"`
}

type ProcessConfigSection struct {
	ConstrainedMemory uint64   `json:"constrained_memory"`
	RSSRange          [2]int64 `json:"rss_range"`
	HeapTotalRange    [2]int64 `json:"heap_total_range"`
	HeapUsedRange     [2]int64 `json:"heap_used_range"`
}

type LoggingConfigSection struct {
	Level string `json:"level"`
	Audit bool   `json:"audit"`
}

func DefaultFullConfig() *FullConfig {
	return &FullConfig{
		Server: ServerConfigSection{
			Port: 8443,
			TLS:  TLSConfigSection{},
		},
		Upstream: UpstreamConfigSection{
			URL: "https://api.anthropic.com",
		},
		OAuth: OAuthConfigSection{},
		Auth: AuthConfigSection{
			Tokens: []TokenEntrySection{},
		},
		Identity: IdentityConfigSection{
			Email: "user@example.com",
		},
		Env: EnvConfigSection{
			Platform:         "darwin",
			PlatformRaw:      "darwin",
			Arch:             "arm64",
			NodeVersion:      "v24.3.0",
			Terminal:         "iTerm2.app",
			PackageManagers:  "npm,pnpm",
			Runtimes:         "node",
			IsRunningWithBun: false,
			IsCI:             false,
			IsClaudeAiAuth:   true,
			Version:          "2.1.81",
			VersionBase:      "2.1.81",
			BuildTime:        "2026-03-20T21:26:18Z",
			DeploymentEnv:    "unknown-darwin",
			VCS:              "git",
		},
		PromptEnv: PromptEnvConfigSection{
			Platform:   "darwin",
			Shell:      "zsh",
			OSVersion:  "Darwin 24.4.0",
			WorkingDir: "/Users/user/projects",
		},
		Process: ProcessConfigSection{
			ConstrainedMemory: 34359738368,
			RSSRange:          [2]int64{300000000, 500000000},
			HeapTotalRange:    [2]int64{40000000, 80000000},
			HeapUsedRange:     [2]int64{100000000, 200000000},
		},
		Logging: LoggingConfigSection{
			Level: "info",
			Audit: true,
		},
	}
}

func (f *FullConfig) ToJSON() (string, error) {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func FullConfigFromJSON(data string) (*FullConfig, error) {
	var cfg FullConfig
	if err := json.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
