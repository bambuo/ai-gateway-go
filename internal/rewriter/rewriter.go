package rewriter

import (
	"encoding/json"

	"ai/gateway/internal/config"
)

func New(cfg *config.Config) *Rewriter {
	return &Rewriter{cfg: cfg}
}

type Rewriter struct {
	cfg *config.Config
}

func (r *Rewriter) RewriteBody(body []byte, path string) []byte {
	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return body
	}

	switch {
	case hasPrefix(path, "/v1/messages"):
		r.rewriteMessages(parsed)
	case contains(path, "/event_logging/batch"):
		r.rewriteEvents(parsed)
	case contains(path, "/policy_limits"), contains(path, "/settings"):
		r.rewriteGenericIdentity(parsed)
	}

	result, err := json.Marshal(parsed)
	if err != nil {
		return body
	}
	return result
}

func (r *Rewriter) RewriteHeaders(h map[string]any) map[string]string {
	out := make(map[string]string, len(h))
	for key, value := range h {
		if value == nil {
			continue
		}

		var v string
		switch val := value.(type) {
		case string:
			v = val
		case []any:
			for i, s := range val {
				if i > 0 {
					v += ", "
				}
				v += toString(s)
			}
		default:
			v = toString(value)
		}

		lower := toLower(key)

		if isHopByHop(lower) {
			continue
		}
		if lower == "authorization" || lower == "proxy-authorization" || lower == "x-api-key" {
			continue
		}
		if lower == "x-anthropic-billing-header" {
			continue
		}
		if lower == "user-agent" {
			out[key] = "claude-code/" + r.cfg.Env.Version + " (external, cli)"
			continue
		}
		out[key] = v
	}
	return out
}
