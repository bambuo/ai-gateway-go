package model

import "time"

type SystemConfig struct {
	ID          uint      `gorm:"primaryKey"`
	Address     string    `gorm:"size:255;not null"`
	Port        int       `gorm:"not null"`
	Protocol    string    `gorm:"size:10;not null;default:https"`
	UpstreamURL string    `gorm:"size:500;not null"`
	TLSCertPath string    `gorm:"size:500"`
	TLSKeyPath  string    `gorm:"size:500"`
	Initialized bool      `gorm:"not null;default:false"`
	JWTSecret   string    `gorm:"size:128;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

type Admin struct {
	ID           uint      `gorm:"primaryKey"`
	Username     string    `gorm:"uniqueIndex;size:64;not null"`
	PasswordHash string    `gorm:"size:256;not null"`
	Email        string    `gorm:"size:256;not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

type InitRequest struct {
	Gateway GatewayConfig `json:"gateway"`
	Admin   AdminCreate   `json:"admin"`
}

type GatewayConfig struct {
	Address     string `json:"address"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	UpstreamURL string `json:"upstream_url"`
	TLSCertPath string `json:"tls_cert_path,omitempty"`
	TLSKeyPath  string `json:"tls_key_path,omitempty"`
}

type AdminCreate struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Email           string `json:"email"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool      `json:"success"`
	Token   string    `json:"token,omitempty"`
	Admin   AdminInfo `json:"admin,omitempty"`
}

type AdminInfo struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type InitStatus struct {
	Initialized      bool `json:"initialized"`
	HasGatewayConfig bool `json:"has_gateway_config"`
	HasAdmin         bool `json:"has_admin"`
}

type InitResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type DashboardData struct {
	Gateway GatewayConfig `json:"gateway"`
	Admin   AdminInfo     `json:"admin"`
	Version string        `json:"version"`
}
