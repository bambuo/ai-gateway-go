package rewriter

import (
	"encoding/base64"
	"encoding/json"

	"ai/gateway/internal/config"
	"ai/gateway/internal/logger"
)

func (r *Rewriter) rewriteEvents(body any) {
	obj, ok := body.(map[string]any)
	if !ok {
		return
	}
	events, ok := obj["events"].([]any)
	if !ok {
		return
	}

	for _, ev := range events {
		event, ok := ev.(map[string]any)
		if !ok {
			continue
		}
		data, ok := event["event_data"].(map[string]any)
		if !ok {
			continue
		}

		data["device_id"] = r.cfg.Identity.DeviceID
		data["email"] = r.cfg.Identity.Email

		if _, exists := data["env"]; exists {
			data["env"] = buildCanonicalEnv(r.cfg)
		}

		if proc, exists := data["process"]; exists {
			data["process"] = buildCanonicalProcess(proc, r.cfg)
		}

		delete(data, "baseUrl")
		delete(data, "base_url")
		delete(data, "gateway")

		if meta, ok := data["additional_metadata"].(string); ok {
			data["additional_metadata"] = rewriteAdditionalMeta(meta)
		}

		logger.Debug("已重写事件", "event_name", data["event_name"])
	}
}

func (r *Rewriter) rewriteGenericIdentity(body any) {
	obj, ok := body.(map[string]any)
	if !ok {
		return
	}
	if _, ok := obj["device_id"]; ok {
		obj["device_id"] = r.cfg.Identity.DeviceID
	}
	if _, ok := obj["email"]; ok {
		obj["email"] = r.cfg.Identity.Email
	}
}

func buildCanonicalEnv(cfg *config.Config) map[string]any {
	return map[string]any{
		"platform":              cfg.Env.Platform,
		"platform_raw":          cfg.Env.PlatformRaw,
		"arch":                  cfg.Env.Arch,
		"node_version":          cfg.Env.NodeVersion,
		"terminal":              cfg.Env.Terminal,
		"package_managers":      cfg.Env.PackageManagers,
		"runtimes":              cfg.Env.Runtimes,
		"is_running_with_bun":   cfg.Env.IsRunningWithBun,
		"is_ci":                 false,
		"is_claubbit":           false,
		"is_claude_code_remote": false,
		"is_local_agent_mode":   false,
		"is_conductor":          false,
		"is_github_action":      false,
		"is_claude_code_action": false,
		"is_claude_ai_auth":     cfg.Env.IsClaudeAiAuth,
		"version":               cfg.Env.Version,
		"version_base":          cfg.Env.VersionBase,
		"build_time":            cfg.Env.BuildTime,
		"deployment_environment": cfg.Env.DeploymentEnv,
		"vcs":                   cfg.Env.VCS,
	}
}

func buildCanonicalProcess(proc any, cfg *config.Config) any {
	switch v := proc.(type) {
	case string:
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return v
		}
		var fields map[string]any
		if err := json.Unmarshal(decoded, &fields); err != nil {
			return v
		}
		rewritten := rewriteProcessFields(fields, cfg)
		encoded, err := json.Marshal(rewritten)
		if err != nil {
			return v
		}
		return base64.StdEncoding.EncodeToString(encoded)
	case map[string]any:
		return rewriteProcessFields(v, cfg)
	default:
		return v
	}
}

func rewriteProcessFields(proc map[string]any, cfg *config.Config) map[string]any {
	p := cfg.Process
	proc["constrainedMemory"] = p.ConstrainedMemory
	proc["rss"] = randomInRange(p.RSSRange[0], p.RSSRange[1])
	proc["heapTotal"] = randomInRange(p.HeapTotalRange[0], p.HeapTotalRange[1])
	proc["heapUsed"] = randomInRange(p.HeapUsedRange[0], p.HeapUsedRange[1])
	return proc
}

func rewriteAdditionalMeta(encoded string) string {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return encoded
	}
	var fields map[string]any
	if err := json.Unmarshal(decoded, &fields); err != nil {
		return encoded
	}
	delete(fields, "baseUrl")
	delete(fields, "base_url")
	delete(fields, "gateway")
	result, err := json.Marshal(fields)
	if err != nil {
		return encoded
	}
	return base64.StdEncoding.EncodeToString(result)
}
