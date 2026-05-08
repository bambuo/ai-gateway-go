package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"ai/gateway/internal/config"
	"ai/gateway/internal/logger"
)

type rewritingTransport struct {
	bodyRW BodyRewriter
	cfg    *config.Config
	inner  http.RoundTripper
}

func newRewritingTransport(rw BodyRewriter, cfg *config.Config) *rewritingTransport {
	proxyURL := os.Getenv("HTTPS_PROXY")
	if proxyURL == "" {
		proxyURL = os.Getenv("HTTP_PROXY")
	}
	var inner http.RoundTripper = http.DefaultTransport
	if proxyURL != "" {
		parsed, err := url.Parse(proxyURL)
		if err == nil {
			inner = &http.Transport{Proxy: http.ProxyURL(parsed)}
			logger.Info("Using outbound proxy", "url", proxyURL)
		}
	}

	return &rewritingTransport{bodyRW: rw, cfg: cfg, inner: inner}
}

func (t *rewritingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body == nil || req.Body == http.NoBody {
		return t.inner.RoundTrip(req)
	}

	body, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)
	}

	body = t.bodyRW.RewriteBody(body, req.URL.Path)

	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))

	return t.inner.RoundTrip(req)
}
