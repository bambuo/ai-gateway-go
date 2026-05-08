package proxy

import (
	"encoding/json"
	"net/http"

	"ai/gateway/internal/config"
	"ai/gateway/internal/rewriter"
)

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	clientName, ok := s.auth.Authenticate(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	_ = clientName
	sample := buildVerificationPayload(s.cfg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sample)
}

func buildVerificationPayload(cfg *config.Config) map[string]any {
	type metadata struct {
		DeviceID    string `json:"device_id"`
		AccountUUID string `json:"account_uuid"`
		SessionID   string `json:"session_id"`
	}

	sampleMetadata, _ := json.Marshal(metadata{
		DeviceID:    "REAL_DEVICE_ID_FROM_CLIENT_abc123",
		AccountUUID: "shared-account-uuid",
		SessionID:   "session-xxx",
	})

	sampleInput := map[string]any{
		"metadata": map[string]any{
			"user_id": string(sampleMetadata),
		},
		"system": []map[string]string{
			{"type": "text", "text": "x-anthropic-billing-header: cc_version=2.1.81.a1b; cc_entrypoint=cli;"},
			{"type": "text", "text": "Here is useful information about the environment:\n<env>\nWorking directory: /home/bob/myproject\nPlatform: linux\nShell: bash\nOS Version: Linux 6.5.0-generic\n</env>"},
		},
		"messages": []map[string]any{
			{"role": "user", "content": "hello"},
		},
	}

	body, _ := json.Marshal(sampleInput)
	rw := rewriter.New(cfg)
	rewritten := rw.RewriteBody(body, "/v1/messages")

	var parsed any
	json.Unmarshal(rewritten, &parsed)

	var beforeParsed any
	json.Unmarshal(body, &beforeParsed)

	before := beforeParsed.(map[string]any)
	after := parsed.(map[string]any)

	return map[string]any{
		"_info": "This shows how the gateway rewrites a sample request",
		"before": map[string]any{
			"metadata.user_id":   before["metadata"].(map[string]any)["user_id"],
			"billing_header":     before["system"].([]any)[0].(map[string]any)["text"],
			"system_prompt_env":  before["system"].([]any)[1].(map[string]any)["text"],
			"system_block_count": len(before["system"].([]any)),
		},
		"after": map[string]any{
			"metadata.user_id":   after["metadata"].(map[string]any)["user_id"],
			"billing_header":     "(stripped)",
			"system_prompt_env":  after["system"].([]any)[0].(map[string]any)["text"],
			"system_block_count": len(after["system"].([]any)),
		},
	}
}
