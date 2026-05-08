package database

import (
	"fmt"
	"os"
	"path/filepath"

	"ai/gateway/internal/logger"
	"ai/gateway/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var db *gorm.DB

func Init(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建数据库目录 %s: %w", dir, err)
		}
	}

	var err error
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return fmt.Errorf("打开数据库 %s: %w", dbPath, err)
	}

	if err := db.AutoMigrate(&model.SystemConfig{}, &model.Admin{}); err != nil {
		return fmt.Errorf("自动迁移: %w", err)
	}

	logger.Info("数据库已初始化", "path", dbPath)
	return nil
}

func IsInitialized() bool {
	if db == nil {
		return false
	}
	var config model.SystemConfig
	if err := db.Where("initialized = ?", true).First(&config).Error; err != nil {
		return false
	}
	var admin model.Admin
	if err := db.First(&admin).Error; err != nil {
		return false
	}
	return true
}

func InitStatus() *model.InitStatus {
	status := &model.InitStatus{}
	if db == nil {
		return status
	}
	var config model.SystemConfig
	if err := db.Where("initialized = ?", true).First(&config).Error; err == nil {
		status.HasGatewayConfig = true
	}
	var admin model.Admin
	if err := db.First(&admin).Error; err == nil {
		status.HasAdmin = true
	}
	status.Initialized = status.HasGatewayConfig && status.HasAdmin
	return status
}

func InitSystem(req *model.InitRequest, jwtSecret string) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		config := model.SystemConfig{
			Address:     req.Gateway.Address,
			Port:        req.Gateway.Port,
			Protocol:    req.Gateway.Protocol,
			UpstreamURL: req.Gateway.UpstreamURL,
			TLSCertPath: req.Gateway.TLSCertPath,
			TLSKeyPath:  req.Gateway.TLSKeyPath,
			Initialized: true,
			JWTSecret:   jwtSecret,
		}
		if err := tx.Create(&config).Error; err != nil {
			return fmt.Errorf("保存系统配置: %w", err)
		}
		passwordHash, err := hashPassword(req.Admin.Password)
		if err != nil {
			return fmt.Errorf("密码加密: %w", err)
		}
		admin := model.Admin{
			Username:     req.Admin.Username,
			PasswordHash: passwordHash,
			Email:        req.Admin.Email,
		}
		if err := tx.Create(&admin).Error; err != nil {
			return fmt.Errorf("创建管理员: %w", err)
		}
		return nil
	})
}
