package admin

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"ai/gateway/internal/database"
	"ai/gateway/internal/logger"
	"ai/gateway/internal/model"
)

var (
	reUpper   = regexp.MustCompile(`[A-Z]`)
	reLower   = regexp.MustCompile(`[a-z]`)
	reDigit   = regexp.MustCompile(`[0-9]`)
	reSpecial = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`)
)

func (s *Server) handleInitStatus(w http.ResponseWriter, r *http.Request) {
	status := database.InitStatus()
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": status})
}

func (s *Server) handleSystemInit(w http.ResponseWriter, r *http.Request) {
	if database.IsInitialized() {
		writeJSON(w, http.StatusConflict, model.ErrorResponse{Success: false, Error: "系统已经初始化，不能重复执行"})
		return
	}
	var req model.InitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Success: false, Error: "请求体格式错误: " + err.Error()})
		return
	}
	if err := validateGatewayConfig(&req.Gateway); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	if err := validateAdminCreate(&req.Admin); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	jwtSecret, err := database.GenerateJWTSecret()
	if err != nil {
		logger.Error("生成 JWT 密钥失败", "error", err)
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Success: false, Error: "系统内部错误，请稍后重试"})
		return
	}
	if err := database.InitSystem(&req, jwtSecret); err != nil {
		logger.Error("系统初始化失败", "error", err)
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Success: false, Error: "系统初始化失败: " + err.Error()})
		return
	}
	logger.Info("系统初始化成功", "admin", req.Admin.Username)
	writeJSON(w, http.StatusOK, model.InitResponse{Success: true, Message: "系统初始化成功"})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Success: false, Error: "请求体格式错误"})
		return
	}
	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Success: false, Error: "用户名和密码不能为空"})
		return
	}
	admin, err := database.FindAdminByUsername(req.Username)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{Success: false, Error: "用户名或密码错误"})
		return
	}
	if !database.VerifyPassword(req.Password, admin.PasswordHash) {
		writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{Success: false, Error: "用户名或密码错误"})
		return
	}
	token, err := generateJWT(admin.Username)
	if err != nil {
		logger.Error("生成 JWT 失败", "error", err)
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Success: false, Error: "生成令牌失败"})
		return
	}
	writeJSON(w, http.StatusOK, model.LoginResponse{
		Success: true,
		Token:   token,
		Admin:   model.AdminInfo{Username: admin.Username, Email: admin.Email},
	})
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	cfg, err := database.GetSystemConfig()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Success: false, Error: "获取系统配置失败"})
		return
	}
	adminInfo, err := database.GetAdminInfo()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Success: false, Error: "获取管理员信息失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": model.DashboardData{
			Gateway: model.GatewayConfig{
				Address:     cfg.Address,
				Port:        cfg.Port,
				Protocol:    cfg.Protocol,
				UpstreamURL: cfg.UpstreamURL,
				TLSCertPath: cfg.TLSCertPath,
				TLSKeyPath:  cfg.TLSKeyPath,
			},
			Admin:   *adminInfo,
			Version: "1.0.0",
		},
	})
}

func validateGatewayConfig(gw *model.GatewayConfig) error {
	if gw.Address == "" {
		return &valErr{"网关地址不能为空"}
	}
	if gw.Port <= 0 || gw.Port > 65535 {
		return &valErr{"端口号必须在 1-65535 之间"}
	}
	if gw.Protocol == "" {
		gw.Protocol = "https"
	}
	if gw.Protocol != "http" && gw.Protocol != "https" {
		return &valErr{"协议只能为 http 或 https"}
	}
	if gw.UpstreamURL == "" {
		gw.UpstreamURL = "https://api.anthropic.com"
	}
	if !strings.HasPrefix(gw.UpstreamURL, "http://") && !strings.HasPrefix(gw.UpstreamURL, "https://") {
		return &valErr{"上游地址必须包含协议前缀 (http:// 或 https://)"}
	}
	return nil
}

func validateAdminCreate(a *model.AdminCreate) error {
	if a.Username == "" {
		return &valErr{"用户名不能为空"}
	}
	if len(a.Username) < 3 {
		return &valErr{"用户名至少需要 3 个字符"}
	}
	if len(a.Username) > 64 {
		return &valErr{"用户名不能超过 64 个字符"}
	}
	if err := checkPassword(a.Password); err != nil {
		return err
	}
	if a.Password != a.ConfirmPassword {
		return &valErr{"两次输入的密码不一致"}
	}
	if a.Email == "" {
		return &valErr{"邮箱不能为空"}
	}
	if !strings.Contains(a.Email, "@") || !strings.Contains(a.Email, ".") {
		return &valErr{"邮箱格式不正确"}
	}
	return nil
}

func checkPassword(p string) error {
	if len(p) < 8 {
		return &valErr{"密码长度至少需要 8 个字符"}
	}
	if len(p) > 128 {
		return &valErr{"密码不能超过 128 个字符"}
	}
	if !reUpper.MatchString(p) {
		return &valErr{"密码必须包含至少一个大写字母"}
	}
	if !reLower.MatchString(p) {
		return &valErr{"密码必须包含至少一个小写字母"}
	}
	if !reDigit.MatchString(p) {
		return &valErr{"密码必须包含至少一个数字"}
	}
	if !reSpecial.MatchString(p) {
		return &valErr{"密码必须包含至少一个特殊字符 (!@#$%^&*()_+-=[]{}等)"}
	}
	return nil
}

type valErr struct{ msg string }

func (e *valErr) Error() string { return e.msg }
