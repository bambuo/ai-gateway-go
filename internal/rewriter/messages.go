package rewriter

import (
	"encoding/json"
	"regexp"
	"strings"

	"ai/gateway/internal/config"
	"ai/gateway/internal/logger"
)

var (
	reCCVersion  = regexp.MustCompile(`cc_version=[\d.]+\.[a-f0-9]{3}`)
	rePlatform   = regexp.MustCompile(`Platform:\s*\S+`)
	reShell      = regexp.MustCompile(`Shell:\s*\S+`)
	reOSVersion  = regexp.MustCompile(`OS Version:\s*[^\n<]+`)
	reWorkingDir = regexp.MustCompile(`((?:Primary )?[Ww]orking directory:\s*)/\S+`)
	reHomePath   = regexp.MustCompile(`/(?:Users|home)/[^/\s]+/`)
	reSysReminder = regexp.MustCompile(`(<system-reminder>)([\s\S]*?)(</system-reminder>)`)
	reBillingHeader = regexp.MustCompile(`^\s*x-anthropic-billing-header:`)
)

func (r *Rewriter) rewriteMessages(body any) {
	obj, ok := body.(map[string]any)
	if !ok {
		return
	}

	// Step 1: Rewrite metadata.user_id
	if md, ok := obj["metadata"].(map[string]any); ok {
		if uid, ok := md["user_id"].(string); ok {
			var userID map[string]any
			if err := json.Unmarshal([]byte(uid), &userID); err == nil {
				userID["device_id"] = r.cfg.Identity.DeviceID
				if b, err := json.Marshal(userID); err == nil {
					md["user_id"] = string(b)
					logger.Debug("已重写 metadata.user_id 中的 device_id")
				}
			}
		}
	}

	// Step 2: Rewrite <system-reminder> blocks in messages
	if msgs, ok := obj["messages"].([]any); ok {
		for _, msg := range msgs {
			if m, ok := msg.(map[string]any); ok {
				switch content := m["content"].(type) {
				case string:
					m["content"] = rewriteSystemReminders(content, r.cfg)
				case []any:
					for i, block := range content {
						if b, ok := block.(map[string]any); ok {
							if text, ok := b["text"].(string); ok {
								b["text"] = rewriteSystemReminders(text, r.cfg)
								content[i] = b
							}
						}
					}
				}
			}
		}
	}

	// Step 3: Extract first user message
	firstUserText := extractFirstUserMessage(obj["messages"])
	version := toString(r.cfg.Env.Version)

	// Step 4: Compute CCH hash
	hash := fallbackHash
	if firstUserText != "" {
		hash = computeCCH(firstUserText, version)
	}
	logger.Debug("已计算 CCH 哈希", "hash", hash, "message_len", len(firstUserText))

	// Step 5: Strip billing header + rewrite system prompt
	if system, ok := obj["system"]; ok {
		switch s := system.(type) {
		case []any:
			var filtered []any
			for _, item := range s {
				switch it := item.(type) {
				case map[string]any:
					if text, ok := it["text"].(string); ok {
						if reBillingHeader.MatchString(text) {
							logger.Debug("已从系统提示中移除计费头信息")
							continue
						}
						it["text"] = rewritePromptText(text, r.cfg.PromptEnv, version, hash)
						filtered = append(filtered, it)
					} else {
						filtered = append(filtered, it)
					}
				case string:
					if reBillingHeader.MatchString(it) {
						logger.Debug("Stripped billing header block from system prompt")
						continue
					}
					filtered = append(filtered, rewritePromptText(it, r.cfg.PromptEnv, version, hash))
				default:
					filtered = append(filtered, it)
				}
			}
			obj["system"] = filtered

		case string:
			s = reBillingHeader.ReplaceAllString(s, "")
			obj["system"] = rewritePromptText(s, r.cfg.PromptEnv, version, hash)
		}
	}
}

func rewritePromptText(text string, pe config.PromptEnvConfig, version, hash string) string {
	result := text

	if hash != "" {
		result = reCCVersion.ReplaceAllString(result, "cc_version="+version+"."+hash)
	}

	result = rePlatform.ReplaceAllString(result, "Platform: "+pe.Platform)
	result = reShell.ReplaceAllString(result, "Shell: "+pe.Shell)
	result = reOSVersion.ReplaceAllString(result, "OS Version: "+pe.OSVersion)
	result = reWorkingDir.ReplaceAllString(result, "${1}"+pe.WorkingDir)

	parts := strings.SplitN(strings.TrimLeft(pe.WorkingDir, "/"), "/", 3)
	homePrefix := "/Users/user/"
	if len(parts) >= 2 {
		homePrefix = "/" + parts[0] + "/" + parts[1] + "/"
	}
	result = reHomePath.ReplaceAllString(result, homePrefix)

	return result
}

func rewriteSystemReminders(text string, cfg *config.Config) string {
	return reSysReminder.ReplaceAllStringFunc(text, func(match string) string {
		parts := reSysReminder.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}
		rewritten := rewritePromptText(parts[2], cfg.PromptEnv, "", "")
		return parts[1] + rewritten + parts[3]
	})
}

func extractFirstUserMessage(messages any) string {
	msgs, ok := messages.([]any)
	if !ok {
		return ""
	}
	for _, msg := range msgs {
		m, ok := msg.(map[string]any)
		if !ok {
			continue
		}
		if role, ok := m["role"].(string); !ok || role != "user" {
			continue
		}
		switch content := m["content"].(type) {
		case string:
			return content
		case []any:
			for _, block := range content {
				if b, ok := block.(map[string]any); ok {
					if text, ok := b["text"].(string); ok {
						return text
					}
				}
			}
		}
	}
	return ""
}
