package config

import (
	"encoding/json"
	"fmt"

	"ai/gateway/internal/model"
)

func FullConfigToAppConfig(fc *model.FullConfig) *Config {
	appCfg := &Config{
		Server: ServerConfig{
			Port: fc.Server.Port,
			TLS: TLSConfig{
				Cert: fc.Server.TLS.Cert,
				Key:  fc.Server.TLS.Key,
			},
		},
		Upstream: UpstreamConfig{
			URL: fc.Upstream.URL,
		},
		OAuth: OAuthConfig{
			AccessToken:  fc.OAuth.AccessToken,
			RefreshToken: fc.OAuth.RefreshToken,
			ExpiresAt:    fc.OAuth.ExpiresAt,
		},
		Identity: IdentityConfig{
			DeviceID: fc.Identity.DeviceID,
			Email:    fc.Identity.Email,
		},
		Env: EnvConfig{
			Platform:         fc.Env.Platform,
			PlatformRaw:      fc.Env.PlatformRaw,
			Arch:             fc.Env.Arch,
			NodeVersion:      fc.Env.NodeVersion,
			Terminal:         fc.Env.Terminal,
			PackageManagers:  fc.Env.PackageManagers,
			Runtimes:         fc.Env.Runtimes,
			IsRunningWithBun: fc.Env.IsRunningWithBun,
			IsCI:             fc.Env.IsCI,
			IsClaudeAiAuth:   fc.Env.IsClaudeAiAuth,
			Version:          fc.Env.Version,
			VersionBase:      fc.Env.VersionBase,
			BuildTime:        fc.Env.BuildTime,
			DeploymentEnv:    fc.Env.DeploymentEnv,
			VCS:              fc.Env.VCS,
		},
		PromptEnv: PromptEnvConfig{
			Platform:   fc.PromptEnv.Platform,
			Shell:      fc.PromptEnv.Shell,
			OSVersion:  fc.PromptEnv.OSVersion,
			WorkingDir: fc.PromptEnv.WorkingDir,
		},
		Process: ProcessConfig{
			ConstrainedMemory: fc.Process.ConstrainedMemory,
			RSSRange:          fc.Process.RSSRange,
			HeapTotalRange:    fc.Process.HeapTotalRange,
			HeapUsedRange:     fc.Process.HeapUsedRange,
		},
		Logging: LoggingConfig{
			Level: fc.Logging.Level,
			Audit: fc.Logging.Audit,
		},
	}
	for _, t := range fc.Auth.Tokens {
		appCfg.Auth.Tokens = append(appCfg.Auth.Tokens, TokenEntry{
			Name:  t.Name,
			Token: t.Token,
		})
	}
	return appCfg
}

func AppConfigToFullConfig(appCfg *Config) *model.FullConfig {
	fc := &model.FullConfig{
		Server: model.ServerConfigSection{
			Port: appCfg.Server.Port,
			TLS: model.TLSConfigSection{
				Cert: appCfg.Server.TLS.Cert,
				Key:  appCfg.Server.TLS.Key,
			},
		},
		Upstream: model.UpstreamConfigSection{
			URL: appCfg.Upstream.URL,
		},
		OAuth: model.OAuthConfigSection{
			AccessToken:  appCfg.OAuth.AccessToken,
			RefreshToken: appCfg.OAuth.RefreshToken,
			ExpiresAt:    appCfg.OAuth.ExpiresAt,
		},
		Identity: model.IdentityConfigSection{
			DeviceID: appCfg.Identity.DeviceID,
			Email:    appCfg.Identity.Email,
		},
		Env: model.EnvConfigSection{
			Platform:         appCfg.Env.Platform,
			PlatformRaw:      appCfg.Env.PlatformRaw,
			Arch:             appCfg.Env.Arch,
			NodeVersion:      appCfg.Env.NodeVersion,
			Terminal:         appCfg.Env.Terminal,
			PackageManagers:  appCfg.Env.PackageManagers,
			Runtimes:         appCfg.Env.Runtimes,
			IsRunningWithBun: appCfg.Env.IsRunningWithBun,
			IsCI:             appCfg.Env.IsCI,
			IsClaudeAiAuth:   appCfg.Env.IsClaudeAiAuth,
			Version:          appCfg.Env.Version,
			VersionBase:      appCfg.Env.VersionBase,
			BuildTime:        appCfg.Env.BuildTime,
			DeploymentEnv:    appCfg.Env.DeploymentEnv,
			VCS:              appCfg.Env.VCS,
		},
		PromptEnv: model.PromptEnvConfigSection{
			Platform:   appCfg.PromptEnv.Platform,
			Shell:      appCfg.PromptEnv.Shell,
			OSVersion:  appCfg.PromptEnv.OSVersion,
			WorkingDir: appCfg.PromptEnv.WorkingDir,
		},
		Process: model.ProcessConfigSection{
			ConstrainedMemory: appCfg.Process.ConstrainedMemory,
			RSSRange:          appCfg.Process.RSSRange,
			HeapTotalRange:    appCfg.Process.HeapTotalRange,
			HeapUsedRange:     appCfg.Process.HeapUsedRange,
		},
		Logging: model.LoggingConfigSection{
			Level: appCfg.Logging.Level,
			Audit: appCfg.Logging.Audit,
		},
	}
	for _, t := range appCfg.Auth.Tokens {
		fc.Auth.Tokens = append(fc.Auth.Tokens, model.TokenEntrySection{
			Name:  t.Name,
			Token: t.Token,
		})
	}
	return fc
}

func NewDefaultConfig() *Config {
	return FullConfigToAppConfig(model.DefaultFullConfig())
}

func ValidateConfigJSON(raw json.RawMessage) error {
	var fc model.FullConfig
	if err := json.Unmarshal(raw, &fc); err != nil {
		return fmt.Errorf("配置 JSON 格式错误: %w", err)
	}
	if fc.Server.Port <= 0 || fc.Server.Port > 65535 {
		return fmt.Errorf("服务器端口必须在 1-65535 之间")
	}
	if fc.Upstream.URL == "" {
		return fmt.Errorf("上游地址不能为空")
	}
	return nil
}
