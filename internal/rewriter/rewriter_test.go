package rewriter

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"ai/gateway/internal/config"
)

func testConfig() *config.Config {
	return &config.Config{
		Server:   config.ServerConfig{Port: 8443},
		Upstream: config.UpstreamConfig{URL: "https://api.anthropic.com"},
		Auth:     config.AuthConfig{Tokens: []config.TokenEntry{{Name: "test", Token: "test-token"}}},
		OAuth:    config.OAuthConfig{RefreshToken: "test-refresh"},
		Identity: config.IdentityConfig{
			DeviceID: "canonical_device_id_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			Email:    "canonical@example.com",
		},
		Env: config.EnvConfig{
			Platform:    "darwin",
			PlatformRaw: "darwin",
			Arch:        "arm64",
			NodeVersion: "v24.3.0",
			Terminal:    "iTerm2.app",
			Runtimes:    "node",
			Version:     "2.1.81",
			VersionBase: "2.1.81",
			BuildTime:   "2026-03-20T21:26:18Z",
			DeploymentEnv: "unknown-darwin",
			VCS:         "git",
		},
		PromptEnv: config.PromptEnvConfig{
			Platform:   "darwin",
			Shell:      "zsh",
			OSVersion:  "Darwin 24.4.0",
			WorkingDir: "/Users/jack/projects",
		},
		Process: config.ProcessConfig{
			ConstrainedMemory: 34359738368,
			RSSRange:          [2]int64{300000000, 500000000},
			HeapTotalRange:    [2]int64{40000000, 80000000},
			HeapUsedRange:     [2]int64{100000000, 200000000},
		},
		Logging: config.LoggingConfig{Level: "error", Audit: false},
	}
}

func rewriteJSON(t *testing.T, r *Rewriter, path string, bodyJSON string) map[string]any {
	t.Helper()
	raw := r.RewriteBody([]byte(bodyJSON), path)
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	return result
}

func TestRewriteMetadataUserID(t *testing.T) {
	r := New(testConfig())
	body := `{
		"metadata": {
			"user_id": "{\"device_id\":\"original_device_id\",\"account_uuid\":\"acct-123\",\"session_id\":\"sess-456\"}"
		},
		"messages": [{"role": "user", "content": "hello"}]
	}`
	result := rewriteJSON(t, r, "/v1/messages", body)

	uidRaw := result["metadata"].(map[string]any)["user_id"].(string)
	var userID map[string]string
	json.Unmarshal([]byte(uidRaw), &userID)

	if userID["device_id"] != testConfig().Identity.DeviceID {
		t.Errorf("device_id = %q, want %q", userID["device_id"], testConfig().Identity.DeviceID)
	}
	if userID["account_uuid"] != "acct-123" {
		t.Errorf("account_uuid = %q, want acct-123", userID["account_uuid"])
	}
	if userID["session_id"] != "sess-456" {
		t.Errorf("session_id = %q, want sess-456", userID["session_id"])
	}
}

func TestRewritePlatformInSystemPrompt(t *testing.T) {
	r := New(testConfig())
	body := `{
		"system": [{"type": "text", "text": "Platform: linux\nShell: bash\nOS Version: Linux 6.5.0"}],
		"messages": []
	}`
	result := rewriteJSON(t, r, "/v1/messages", body)

	system := result["system"].([]any)
	text := system[0].(map[string]any)["text"].(string)

	if !strings.Contains(text, "Platform: darwin") {
		t.Errorf("missing Platform: darwin in %q", text)
	}
	if !strings.Contains(text, "Shell: zsh") {
		t.Errorf("missing Shell: zsh in %q", text)
	}
	if !strings.Contains(text, "OS Version: Darwin 24.4.0") {
		t.Errorf("missing OS Version: Darwin 24.4.0 in %q", text)
	}
}

func TestRewriteWorkingDirectory(t *testing.T) {
	r := New(testConfig())
	body := `{
		"system": "Primary working directory: /home/bob/myproject",
		"messages": []
	}`
	result := rewriteJSON(t, r, "/v1/messages", body)

	system := result["system"].(string)
	if !strings.Contains(system, "/Users/jack/projects") {
		t.Errorf("missing canonical path in %q", system)
	}
	if strings.Contains(system, "/home/bob/") {
		t.Errorf("original path should be replaced, got %q", system)
	}
}

func TestStripBillingHeaderString(t *testing.T) {
	r := New(testConfig())
	body := `{
		"system": "x-anthropic-billing-header: cc_version=2.1.81.a1b; cc_entrypoint=cli; cch=00000;\nOther content here.",
		"messages": []
	}`
	result := rewriteJSON(t, r, "/v1/messages", body)

	system := result["system"].(string)
	if strings.Contains(system, "billing-header") {
		t.Errorf("billing header should be stripped, got %q", system)
	}
	if !strings.Contains(system, "Other content") {
		t.Errorf("non-billing content should remain, got %q", system)
	}
}

func TestStripBillingHeaderArray(t *testing.T) {
	r := New(testConfig())
	body := `{
		"system": [
			{"type": "text", "text": "x-anthropic-billing-header: cc_version=2.1.81.a1b; cc_entrypoint=cli;"},
			{"type": "text", "text": "Platform: linux\nShell: bash"}
		],
		"messages": []
	}`
	result := rewriteJSON(t, r, "/v1/messages", body)

	system := result["system"].([]any)
	if len(system) != 1 {
		t.Fatalf("expected 1 system block after stripping billing header, got %d", len(system))
	}
	text := system[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "Platform: darwin") {
		t.Errorf("remaining block should be rewritten, got %q", text)
	}
}

func TestRewriteHomePathsInSystemReminder(t *testing.T) {
	r := New(testConfig())
	body := `{
		"system": "",
		"messages": [{"role": "user", "content": "<system-reminder>Working directory: /home/alice/code</system-reminder>"}]
	}`
	result := rewriteJSON(t, r, "/v1/messages", body)

	msgs := result["messages"].([]any)
	content := msgs[0].(map[string]any)["content"].(string)
	if strings.Contains(content, "/home/alice/") {
		t.Errorf("home path should be rewritten, got %q", content)
	}
}

func TestRewriteDeviceIDAndEmailInEvents(t *testing.T) {
	r := New(testConfig())
	body := `{
		"events": [{
			"event_type": "ClaudeCodeInternalEvent",
			"event_data": {
				"device_id": "real_device_id",
				"email": "real@email.com",
				"event_name": "tengu_init",
				"env": {"platform": "linux", "arch": "x64"}
			}
		}]
	}`
	result := rewriteJSON(t, r, "/api/event_logging/batch", body)

	events := result["events"].([]any)
	data := events[0].(map[string]any)["event_data"].(map[string]any)

	if data["device_id"] != testConfig().Identity.DeviceID {
		t.Errorf("device_id = %q, want %q", data["device_id"], testConfig().Identity.DeviceID)
	}
	if data["email"] != testConfig().Identity.Email {
		t.Errorf("email = %q, want %q", data["email"], testConfig().Identity.Email)
	}
}

func TestReplaceEnvWithCanonical(t *testing.T) {
	r := New(testConfig())
	body := `{
		"events": [{
			"event_type": "ClaudeCodeInternalEvent",
			"event_data": {
				"device_id": "x",
				"env": {
					"platform": "linux",
					"arch": "x64",
					"node_version": "v20.0.0",
					"terminal": "xterm",
					"is_ci": true,
					"deployment_environment": "unknown-linux"
				}
			}
		}]
	}`
	result := rewriteJSON(t, r, "/api/event_logging/batch", body)

	events := result["events"].([]any)
	env := events[0].(map[string]any)["event_data"].(map[string]any)["env"].(map[string]any)

	if env["platform"] != "darwin" {
		t.Errorf("platform = %q, want darwin", env["platform"])
	}
	if env["arch"] != "arm64" {
		t.Errorf("arch = %q, want arm64", env["arch"])
	}
	if env["node_version"] != "v24.3.0" {
		t.Errorf("node_version = %q, want v24.3.0", env["node_version"])
	}
	if env["terminal"] != "iTerm2.app" {
		t.Errorf("terminal = %q, want iTerm2.app", env["terminal"])
	}
	if env["is_ci"] != false {
		t.Errorf("is_ci = %v, want false", env["is_ci"])
	}
}

func TestStripBaseURL(t *testing.T) {
	r := New(testConfig())
	body := `{
		"events": [{
			"event_type": "ClaudeCodeInternalEvent",
			"event_data": {
				"device_id": "x",
				"baseUrl": "https://gateway.office.com:8443",
				"gateway": "custom"
			}
		}]
	}`
	result := rewriteJSON(t, r, "/api/event_logging/batch", body)

	events := result["events"].([]any)
	data := events[0].(map[string]any)["event_data"].(map[string]any)

	if _, ok := data["baseUrl"]; ok {
		t.Error("baseUrl should be stripped")
	}
	if _, ok := data["gateway"]; ok {
		t.Error("gateway should be stripped")
	}
}

func TestRewriteProcessMetricsBase64(t *testing.T) {
	r := New(testConfig())
	processData := map[string]any{
		"uptime":            float64(100),
		"rss":               float64(999999999),
		"heapTotal":         float64(999999999),
		"heapUsed":          float64(999999999),
		"constrainedMemory": float64(68719476736),
		"cpuUsage":          map[string]any{"user": float64(1000), "system": float64(500)},
	}
	procBytes, _ := json.Marshal(processData)
	procB64 := base64.StdEncoding.EncodeToString(procBytes)

	body := `{
		"events": [{
			"event_type": "ClaudeCodeInternalEvent",
			"event_data": {
				"device_id": "x",
				"process": "` + procB64 + `"
			}
		}]
	}`

	cfg := testConfig()
	r = New(cfg)
	result := rewriteJSON(t, r, "/api/event_logging/batch", body)

	events := result["events"].([]any)
	procStr := events[0].(map[string]any)["event_data"].(map[string]any)["process"].(string)
	decoded, _ := base64.StdEncoding.DecodeString(procStr)

	var proc map[string]any
	json.Unmarshal(decoded, &proc)

	if proc["constrainedMemory"] != float64(34359738368) {
		t.Errorf("constrainedMemory = %v, want 34359738368", proc["constrainedMemory"])
	}
	if proc["uptime"] != float64(100) {
		t.Errorf("uptime should be preserved, got %v", proc["uptime"])
	}
	rss := proc["rss"].(float64)
	if rss < 300000000 || rss > 500000000 {
		t.Errorf("rss %v out of range [300000000, 500000000]", rss)
	}
}

func TestRewriteUserAgent(t *testing.T) {
	r := New(testConfig())
	headers := r.RewriteHeaders(map[string]any{
		"user-agent": "claude-code/2.0.50 (external, cli)",
		"x-app":      "cli",
	})

	if headers["user-agent"] != "claude-code/2.1.81 (external, cli)" {
		t.Errorf("user-agent = %q, want claude-code/2.1.81 (external, cli)", headers["user-agent"])
	}
	if headers["x-app"] != "cli" {
		t.Errorf("x-app = %q, want cli", headers["x-app"])
	}
}

func TestStripAuthorizationHeader(t *testing.T) {
	r := New(testConfig())
	headers := r.RewriteHeaders(map[string]any{
		"authorization": "Bearer client-placeholder-token",
		"x-app":         "cli",
	})

	if _, ok := headers["authorization"]; ok {
		t.Error("authorization should be stripped")
	}
	if headers["x-app"] != "cli" {
		t.Errorf("x-app should be preserved, got %q", headers["x-app"])
	}
}

func TestStripProxyAuthorizationHeader(t *testing.T) {
	r := New(testConfig())
	headers := r.RewriteHeaders(map[string]any{
		"proxy-authorization": "Bearer proxy-token",
	})

	if _, ok := headers["proxy-authorization"]; ok {
		t.Error("proxy-authorization should be stripped")
	}
}

func TestStripXApiKeyHeader(t *testing.T) {
	r := New(testConfig())
	headers := r.RewriteHeaders(map[string]any{
		"x-api-key": "client-gateway-token",
		"x-app":     "cli",
	})

	if _, ok := headers["x-api-key"]; ok {
		t.Error("x-api-key should be stripped")
	}
	if headers["x-app"] != "cli" {
		t.Errorf("x-app should be preserved, got %q", headers["x-app"])
	}
}

func TestStripXBillingHeader(t *testing.T) {
	r := New(testConfig())
	headers := r.RewriteHeaders(map[string]any{
		"x-anthropic-billing-header": "cc_version=2.1.81.a1b; cc_entrypoint=cli;",
	})

	if _, ok := headers["x-anthropic-billing-header"]; ok {
		t.Error("x-anthropic-billing-header should be stripped")
	}
}

func TestNonJSONPassthrough(t *testing.T) {
	r := New(testConfig())
	raw := r.RewriteBody([]byte("not json content"), "/v1/messages")
	if string(raw) != "not json content" {
		t.Errorf("non-JSON body should pass through unchanged, got %q", string(raw))
	}
}

func TestMain(m *testing.M) {
	// Suppress log output during tests
	os.Exit(m.Run())
}
