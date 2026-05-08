package database

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"ai/gateway/internal/model"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(bytes), nil
}

func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateJWTSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("生成随机密钥: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func FindAdminByUsername(username string) (*model.Admin, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	var admin model.Admin
	if err := db.Where("username = ?", username).First(&admin).Error; err != nil {
		return nil, fmt.Errorf("查询管理员: %w", err)
	}
	return &admin, nil
}

func GetSystemConfig() (*model.SystemConfig, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	var config model.SystemConfig
	if err := db.Where("initialized = ?", true).First(&config).Error; err != nil {
		return nil, fmt.Errorf("查询系统配置: %w", err)
	}
	return &config, nil
}

func GetAdminInfo() (*model.AdminInfo, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	var admin model.Admin
	if err := db.First(&admin).Error; err != nil {
		return nil, fmt.Errorf("查询管理员: %w", err)
	}
	return &model.AdminInfo{
		Username: admin.Username,
		Email:    admin.Email,
	}, nil
}

func GetJWTSecret() (string, error) {
	cfg, err := GetSystemConfig()
	if err != nil {
		return "", err
	}
	return cfg.JWTSecret, nil
}
