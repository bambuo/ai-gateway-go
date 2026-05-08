package proxy

import (
	"net/http"
)

type TokenProvider interface {
	AccessToken() string
}

type ClientAuthenticator interface {
	Authenticate(r *http.Request) (string, bool)
}

type BodyRewriter interface {
	RewriteBody(body []byte, path string) []byte
}

type HeaderRewriter interface {
	RewriteHeaders(h map[string]any) map[string]string
}
