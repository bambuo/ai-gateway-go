package proxy

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	oauthOk := s.tokens.AccessToken() != ""
	status := http.StatusOK
	if !oauthOk {
		status = http.StatusServiceUnavailable
	}

	clientNames := make([]string, len(s.cfg.Auth.Tokens))
	for i, t := range s.cfg.Auth.Tokens {
		clientNames[i] = t.Name
	}

	body := map[string]any{
		"status":             mapCond(oauthOk, "ok", "degraded"),
		"oauth":              mapCond(oauthOk, "valid", "expired/refreshing"),
		"canonical_device":   s.cfg.Identity.DeviceID[:8] + "...",
		"canonical_platform": s.cfg.Env.Platform,
		"upstream":           s.cfg.Upstream.URL,
		"clients":            clientNames,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func mapCond[T any](t bool, yes, no T) T {
	if t {
		return yes
	}
	return no
}
