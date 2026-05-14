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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const (
	defaultPort     = 8080
	shutdownTimeout = 10 * time.Second
)

type Server struct {
	port           int
	router         *chi.Mux
	srv            *http.Server
	static         string
	onConfigReload func(*config.Config)
}

func NewServer(port int, staticDir string) *Server {
	if port <= 0 {
		port = defaultPort
	}
	s := &Server{port: port, static: staticDir, router: chi.NewRouter()}
	s.registerRoutes()
	return s
}

func NewServerWithReload(port int, staticDir string, reloadFn func(*config.Config)) *Server {
	if port <= 0 {
		port = defaultPort
	}
	s := &Server{port: port, static: staticDir, router: chi.NewRouter(), onConfigReload: reloadFn}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	router := s.router
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           86400,
	}))

	// API routes
	router.Route("/api", func(r chi.Router) {
		r.Get("/system/init-status", s.handleInitStatus)
		r.Post("/system/init", s.handleSystemInit)
		r.Post("/admin/login", s.handleLogin)

		r.Group(func(authed chi.Router) {
			authed.Use(s.authMiddleware)
			authed.Get("/admin/dashboard", s.handleDashboard)
			authed.Get("/config", s.handleGetConfig)
			authed.Put("/config", s.handleUpdateConfig)
			authed.Post("/config/reload", s.handleReloadConfig)
		})

		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		})
	})

	// Static file serving
	s.static = strings.TrimRight(s.static, "/")
	if info, err := os.Stat(s.static); err == nil && info.IsDir() {
		fs := http.FileServer(http.Dir(s.static))
		router.Handle("/assets/*", fs)
		router.Handle("/favicon.ico", fs)

		router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			http.ServeFile(w, r, s.static+"/index.html")
		})
	} else {
		logger.Warn("前端静态目录不存在，仅 API 模式运行", "path", s.static)
		router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
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
	s.srv = &http.Server{Addr: addr, Handler: s.router}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Info("正在优雅关闭管理后台...", "service", "admin")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := s.srv.Shutdown(ctx); err != nil {
			logger.Error("管理后台关闭出错", "service", "admin", "error", err)
		}
	}()

	logger.Info("管理后台已启动", "service", "admin", "addr", addr, "static", s.static)
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())
		ah := r.Header.Get("Authorization")
		if ah == "" {
			logger.Warn("管理后台认证失败", "reason", "missing_authorization", "path", r.URL.Path, "request_id", reqID)
			writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "未提供认证令牌"})
			return
		}
		token := strings.TrimPrefix(ah, "Bearer ")
		if token == ah {
			logger.Warn("管理后台认证失败", "reason", "invalid_authorization_format", "path", r.URL.Path, "request_id", reqID)
			writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "认证格式错误"})
			return
		}
		claims, err := validateJWT(token)
		if err != nil {
			logger.Warn("管理后台认证失败", "reason", "invalid_or_expired_token", "path", r.URL.Path, "request_id", reqID)
			writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "认证令牌无效或已过期"})
			return
		}
		r.Header.Set("X-Admin-Username", claims.Username)
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("JSON 编码失败", "error", err)
	}
}
