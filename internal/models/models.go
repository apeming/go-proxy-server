package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	IP       string // For audit/logging only
	Username string `gorm:"uniqueIndex"` // Globally unique
	Password []byte
}

type Whitelist struct {
	gorm.Model
	IP string `gorm:"uniqueIndex"`
}

// ProxyConfig stores proxy server configuration
type ProxyConfig struct {
	gorm.Model
	Type       string `gorm:"uniqueIndex"` // "socks5" or "http"
	Port       int
	BindListen bool
	AutoStart  bool // Whether to auto-start on application launch
}

// SystemConfig stores system-level configuration
type SystemConfig struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex"` // Configuration key
	Value string // Configuration value
}
