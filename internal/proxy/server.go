package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"ai/gateway/internal/config"
	"ai/gateway/internal/logger"
)

type ctxKey struct{}

var clientKey = &ctxKey{}

type Server struct {
	cfg      *config.Config
	mux      *http.ServeMux
	auth     ClientAuthenticator
	tokens   TokenProvider
	bodyRW   BodyRewriter
	headerRW HeaderRewriter
	upstream *url.URL
	proxy    *httputil.ReverseProxy
}

func NewServer(cfg *config.Config, auth ClientAuthenticator, tokens TokenProvider, bodyRW BodyRewriter, headerRW HeaderRewriter) *Server {
	upstreamURL, _ := url.Parse(cfg.Upstream.URL)

	s := &Server{
		cfg:      cfg,
		auth:     auth,
		tokens:   tokens,
		bodyRW:   bodyRW,
		headerRW: headerRW,
		upstream: upstreamURL,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /_health", s.handleHealth)
	mux.HandleFunc("GET /_verify", s.handleVerify)
	mux.HandleFunc("/", s.handleProxy)

	s.mux = mux
	s.proxy = s.buildReverseProxy()
	return s
}

func (s *Server) buildReverseProxy() *httputil.ReverseProxy {
	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("upstream error", "error", err)
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": "bad gateway", "detail": err.Error()})
	}

	transport := newRewritingTransport(s.bodyRW, s.cfg)

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			rewritten := s.headerRW.RewriteHeaders(toMap(req.Header))

			req.Header = make(http.Header)
			for k, v := range rewritten {
				req.Header.Set(k, v)
			}

			token := s.tokens.AccessToken()
			if token == "" {
				return
			}
			req.Header.Set("x-api-key", token)

			req.URL.Scheme = s.upstream.Scheme
			req.URL.Host = s.upstream.Host
			req.Host = s.upstream.Host
		},
		ModifyResponse: func(resp *http.Response) error {
			if s.cfg.Logging.Audit {
				if clientName, ok := resp.Request.Context().Value(clientKey).(string); ok {
					logger.Audit(clientName, resp.Request.Method, resp.Request.URL.Path, resp.StatusCode)
				}
			}
			return nil
		},
		Transport:    transport,
		ErrorHandler: errorHandler,
	}
}

func toMap(h http.Header) map[string]any {
	m := make(map[string]any, len(h))
	for k, v := range h {
		if len(v) == 1 {
			m[k] = v[0]
		} else {
			vals := make([]any, len(v))
			for i, s := range v {
				vals[i] = s
			}
			m[k] = vals
		}
	}
	return m
}

func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if idx := strings.LastIndex(clientIP, ":"); idx != -1 {
		clientIP = clientIP[:idx]
	}

	logger.Info("request received", "method", r.Method, "path", r.URL.Path, "client_ip", clientIP)

	clientName, ok := s.auth.Authenticate(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized - provide client token via x-api-key header"})
		logger.Warn("Unauthorized request", "method", r.Method, "path", r.URL.Path, "client_ip", clientIP)
		return
	}

	logger.Info("client request", "client", clientName, "method", r.Method, "path", r.URL.Path)

	ctx := context.WithValue(r.Context(), clientKey, clientName)
	s.proxy.ServeHTTP(w, r.WithContext(ctx))
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Server.Port)

	var srv *http.Server
	tlsEnabled := s.cfg.Server.TLS.Cert != "" && s.cfg.Server.TLS.Key != ""

	if tlsEnabled {
		srv = &http.Server{Addr: addr, Handler: s.mux}
		logger.Info("AI Gateway listening with TLS", "addr", addr)
	} else {
		srv = &http.Server{Addr: addr, Handler: s.mux}
		logger.Warn("Running without TLS - only use for local development", "addr", addr)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("Shutdown error", "error", err)
		}
	}()

	logger.Info("AI Gateway started",
		"upstream", s.cfg.Upstream.URL,
		"device_id", s.cfg.Identity.DeviceID[:8]+"...",
		"clients", len(s.cfg.Auth.Tokens),
	)

	var err error
	if tlsEnabled {
		err = srv.ListenAndServeTLS(s.cfg.Server.TLS.Cert, s.cfg.Server.TLS.Key)
	} else {
		err = srv.ListenAndServe()
	}
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
