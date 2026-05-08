package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"ai/gateway/internal/config"
	"ai/gateway/internal/logger"
)

const (
	defaultPort     = 8080
	shutdownTimeout = 10 * time.Second
)

type Server struct {
	port           int
	mux            *http.ServeMux
	srv            *http.Server
	static         string
	onConfigReload func(*config.Config)
}

func NewServer(port int, staticDir string) *Server {
	if port <= 0 {
		port = defaultPort
	}
	s := &Server{port: port, static: staticDir, mux: http.NewServeMux()}
	s.registerRoutes()
	return s
}

func NewServerWithReload(port int, staticDir string, reloadFn func(*config.Config)) *Server {
	if port <= 0 {
		port = defaultPort
	}
	s := &Server{port: port, static: staticDir, mux: http.NewServeMux(), onConfigReload: reloadFn}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	mux := s.mux

	// API routes (with built-in CORS)
	mux.HandleFunc("GET /api/system/init-status", s.withCORS(s.handleInitStatus))
	mux.HandleFunc("POST /api/system/init", s.withCORS(s.handleSystemInit))
	mux.HandleFunc("POST /api/admin/login", s.withCORS(s.handleLogin))
	mux.HandleFunc("GET /api/admin/dashboard", s.withCORS(s.withAuth(s.handleDashboard)))
	mux.HandleFunc("GET /api/config", s.withCORS(s.withAuth(s.handleGetConfig)))
	mux.HandleFunc("PUT /api/config", s.withCORS(s.withAuth(s.handleUpdateConfig)))
	mux.HandleFunc("POST /api/config/reload", s.withCORS(s.withAuth(s.handleReloadConfig)))

	// CORS preflight for all /api/ routes
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		corsMiddleware(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	})

	// Static file serving
	s.static = strings.TrimRight(s.static, "/")
	if info, err := os.Stat(s.static); err == nil && info.IsDir() {
		fs := http.FileServer(http.Dir(s.static))
		mux.Handle("GET /assets/", fs)
		mux.Handle("GET /favicon.ico", fs)

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			http.ServeFile(w, r, s.static+"/index.html")
		})
	} else {
		logger.Warn("前端静态目录不存在，仅 API 模式运行", "path", s.static)
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/api/") {
				writeJSON(w, http.StatusNotFound, map[string]string{
					"error":   "前端静态文件未构建",
					"message": "请运行 cd web && bun install && bun run build",
				})
			}
		})
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	s.srv = &http.Server{Addr: addr, Handler: s.mux}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Info("正在优雅关闭管理后台...")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		s.srv.Shutdown(ctx)
	}()

	logger.Info("管理后台已启动", "addr", addr, "static", s.static)
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func corsMiddleware(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

func (s *Server) withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		corsMiddleware(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ah := r.Header.Get("Authorization")
		if ah == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "未提供认证令牌"})
			return
		}
		token := strings.TrimPrefix(ah, "Bearer ")
		if token == ah {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "认证格式错误"})
			return
		}
		claims, err := validateJWT(token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "认证令牌无效或已过期"})
			return
		}
		r.Header.Set("X-Admin-Username", claims.Username)
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("JSON 编码失败", "error", err)
	}
}
