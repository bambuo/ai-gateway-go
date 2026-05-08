package database

import (
	"fmt"

	"ai/gateway/internal/model"

	"gorm.io/gorm"
)

func SaveFullConfig(cfg *model.FullConfig) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	jsonData, err := cfg.ToJSON()
	if err != nil {
		return fmt.Errorf("序列化配置: %w", err)
	}
	return SaveFullConfigRaw(jsonData)
}

func SaveFullConfigRaw(jsonConfig string) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.AppConfig{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
			return fmt.Errorf("停用旧配置: %w", err)
		}
		var lastVersion int
		tx.Model(&model.AppConfig{}).Select("COALESCE(MAX(version), 0)").Scan(&lastVersion)
		ac := model.AppConfig{
			Config:   jsonConfig,
			Version:  lastVersion + 1,
			IsActive: true,
		}
		if err := tx.Create(&ac).Error; err != nil {
			return fmt.Errorf("保存配置: %w", err)
		}
		return nil
	})
}

func GetActiveFullConfig() (*model.FullConfig, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	var ac model.AppConfig
	if err := db.Where("is_active = ?", true).Last(&ac).Error; err != nil {
		return nil, fmt.Errorf("查询活动配置: %w", err)
	}
	return model.FullConfigFromJSON(ac.Config)
}

func GetActiveConfigRaw() (string, error) {
	if db == nil {
		return "", fmt.Errorf("数据库未初始化")
	}
	var ac model.AppConfig
	if err := db.Where("is_active = ?", true).Last(&ac).Error; err != nil {
		return "", fmt.Errorf("查询活动配置: %w", err)
	}
	return ac.Config, nil
}

func HasAppConfig() bool {
	if db == nil {
		return false
	}
	var count int64
	db.Model(&model.AppConfig{}).Where("is_active = ?", true).Count(&count)
	return count > 0
}
