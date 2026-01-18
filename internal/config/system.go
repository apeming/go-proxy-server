package config

import (
	"errors"

	"gorm.io/gorm"

	"go-proxy-server/internal/models"
)

// System configuration keys
const (
	KeyAutoStart = "autostart_enabled"
)

// GetSystemConfig gets a system configuration value
func GetSystemConfig(db *gorm.DB, key string) (string, error) {
	var config models.SystemConfig
	err := db.Where("key = ?", key).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil // Not found, return empty string
		}
		return "", err
	}
	return config.Value, nil
}

// SetSystemConfig sets a system configuration value
func SetSystemConfig(db *gorm.DB, key, value string) error {
	var config models.SystemConfig
	err := db.Where("key = ?", key).First(&config).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new config
		config = models.SystemConfig{
			Key:   key,
			Value: value,
		}
		return db.Create(&config).Error
	} else if err != nil {
		return err
	}

	// Update existing config
	config.Value = value
	return db.Save(&config).Error
}

// DeleteSystemConfig deletes a system configuration
func DeleteSystemConfig(db *gorm.DB, key string) error {
	return db.Unscoped().Where("key = ?", key).Delete(&models.SystemConfig{}).Error
}
