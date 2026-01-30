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

// MetricsSnapshot stores historical metrics data
type MetricsSnapshot struct {
	gorm.Model
	Timestamp            int64   // Unix timestamp
	ActiveConnections    int     // Number of active connections
	MaxActiveConnections int     // Maximum active connections since start
	TotalConnections     int64   // Total connections since start
	BytesReceived        int64   // Total bytes received
	BytesSent            int64   // Total bytes sent
	UploadSpeed          float64 // Upload speed in bytes/sec
	DownloadSpeed        float64 // Download speed in bytes/sec
	MaxUploadSpeed       float64 // Maximum upload speed since start
	MaxDownloadSpeed     float64 // Maximum download speed since start
	ErrorCount           int64   // Total error count
}

// AlertConfig stores alert configuration
type AlertConfig struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex"` // Alert name
	Metric      string // Metric to monitor (connections, bandwidth, errors)
	Operator    string // Comparison operator (gt, lt, eq)
	Threshold   float64 // Threshold value
	Duration    int    // Duration in seconds before triggering
	Enabled     bool   // Whether alert is enabled
	NotifyEmail string // Email for notifications (optional)
}

// AlertHistory stores alert trigger history
type AlertHistory struct {
	gorm.Model
	AlertConfigID uint    // Reference to AlertConfig
	Timestamp     int64   // When alert was triggered
	MetricValue   float64 // Value that triggered the alert
	Message       string  // Alert message
	Resolved      bool    // Whether alert has been resolved
	ResolvedAt    *int64  // When alert was resolved
}
