package admin

import (
	"encoding/json"
	"net/http"

	"ai/gateway/internal/config"
	"ai/gateway/internal/database"
	"ai/gateway/internal/logger"
	"ai/gateway/internal/model"
)

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	raw, err := database.GetActiveConfigRaw()
	if err != nil {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Success: false, Error: "未找到配置，请先初始化系统"})
		return
	}
	var parsed any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Success: false, Error: "配置数据损坏"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": parsed})
}

func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Config json.RawMessage `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Success: false, Error: "请求体格式错误: " + err.Error()})
		return
	}
	if len(body.Config) == 0 {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Success: false, Error: "配置内容不能为空"})
		return
	}
	if err := config.ValidateConfigJSON(body.Config); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	pretty, err := json.MarshalIndent(body.Config, "", "  ")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Success: false, Error: "配置格式化失败"})
		return
	}
	if err := database.SaveFullConfigRaw(string(pretty)); err != nil {
		logger.Error("保存配置失败", "error", err)
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Success: false, Error: "保存配置失败: " + err.Error()})
		return
	}
	logger.Info("配置已更新")
	s.reloadConfig()
	writeJSON(w, http.StatusOK, model.InitResponse{Success: true, Message: "配置已更新"})
}

func (s *Server) handleReloadConfig(w http.ResponseWriter, r *http.Request) {
	s.reloadConfig()
	writeJSON(w, http.StatusOK, model.InitResponse{Success: true, Message: "配置已重新加载"})
}

func (s *Server) reloadConfig() {
	cfg, err := database.GetActiveFullConfig()
	if err != nil {
		logger.Error("重新加载配置失败", "error", err)
		return
	}
	appCfg := config.FullConfigToAppConfig(cfg)
	if s.onConfigReload != nil {
		s.onConfigReload(appCfg)
	}
}
