package auth

import (
	"net/http"
	"strings"

	"ai/gateway/internal/config"
)

type Authenticator struct {
	tokens map[string]string
}

func New(cfg config.AuthConfig) *Authenticator {
	tokens := make(map[string]string, len(cfg.Tokens))
	for _, t := range cfg.Tokens {
		tokens[t.Token] = t.Name
	}
	return &Authenticator{tokens: tokens}
}

func (a *Authenticator) Authenticate(r *http.Request) (clientName string, ok bool) {
	token := r.Header.Get("x-api-key")
	if name, found := a.tokens[token]; found {
		return name, true
	}

	authStr := r.Header.Get("Proxy-Authorization")
	if authStr == "" {
		authStr = r.Header.Get("Authorization")
	}
	if authStr == "" {
		return "", false
	}

	if !strings.HasPrefix(strings.ToLower(authStr), "bearer ") {
		return "", false
	}
	token = strings.TrimSpace(authStr[7:])

	name, found := a.tokens[token]
	return name, found
}
