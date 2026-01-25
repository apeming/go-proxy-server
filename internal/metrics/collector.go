package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"go-proxy-server/internal/models"
	"gorm.io/gorm"
)

// Collector collects and aggregates proxy metrics
type Collector struct {
	mu                sync.RWMutex
	db                *gorm.DB
	startTime         time.Time
	activeConnections int32
	totalConnections  int64
	bytesReceived     int64
	bytesSent         int64
	errorCount        int64

	// For speed calculation
	lastSnapshot      time.Time
	lastBytesReceived int64
	lastBytesSent     int64
	uploadSpeed       float64
	downloadSpeed     float64

	// Background aggregation
	stopChan          chan struct{}
	snapshotInterval  time.Duration
}

var (
	globalCollector *Collector
	once            sync.Once
)

// InitCollector initializes the global metrics collector
func InitCollector(db *gorm.DB, snapshotInterval time.Duration) *Collector {
	once.Do(func() {
		globalCollector = &Collector{
			db:               db,
			startTime:        time.Now(),
			lastSnapshot:     time.Now(),
			stopChan:         make(chan struct{}),
			snapshotInterval: snapshotInterval,
		}

		// Start background aggregation
		go globalCollector.backgroundAggregation()
	})
	return globalCollector
}

// GetCollector returns the global collector instance
func GetCollector() *Collector {
	return globalCollector
}

// RecordConnection increments connection counters
func (c *Collector) RecordConnection() {
	atomic.AddInt32(&c.activeConnections, 1)
	atomic.AddInt64(&c.totalConnections, 1)
}

// RecordDisconnection decrements active connection counter
func (c *Collector) RecordDisconnection() {
	atomic.AddInt32(&c.activeConnections, -1)
}

// RecordBytesReceived adds to bytes received counter
func (c *Collector) RecordBytesReceived(bytes int64) {
	atomic.AddInt64(&c.bytesReceived, bytes)
}

// RecordBytesSent adds to bytes sent counter
func (c *Collector) RecordBytesSent(bytes int64) {
	atomic.AddInt64(&c.bytesSent, bytes)
}

// RecordError increments error counter
func (c *Collector) RecordError() {
	atomic.AddInt64(&c.errorCount, 1)
}

// GetSnapshot returns current metrics snapshot
func (c *Collector) GetSnapshot() *MetricsSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &MetricsSnapshot{
		Timestamp:         time.Now().Unix(),
		ActiveConnections: int(atomic.LoadInt32(&c.activeConnections)),
		TotalConnections:  atomic.LoadInt64(&c.totalConnections),
		BytesReceived:     atomic.LoadInt64(&c.bytesReceived),
		BytesSent:         atomic.LoadInt64(&c.bytesSent),
		UploadSpeed:       c.uploadSpeed,
		DownloadSpeed:     c.downloadSpeed,
		ErrorCount:        atomic.LoadInt64(&c.errorCount),
		Uptime:            int64(time.Since(c.startTime).Seconds()),
	}
}

// MetricsSnapshot represents a point-in-time metrics snapshot
type MetricsSnapshot struct {
	Timestamp         int64   `json:"timestamp"`
	ActiveConnections int     `json:"activeConnections"`
	TotalConnections  int64   `json:"totalConnections"`
	BytesReceived     int64   `json:"bytesReceived"`
	BytesSent         int64   `json:"bytesSent"`
	UploadSpeed       float64 `json:"uploadSpeed"`
	DownloadSpeed     float64 `json:"downloadSpeed"`
	ErrorCount        int64   `json:"errorCount"`
	Uptime            int64   `json:"uptime"`
}

// backgroundAggregation periodically calculates speeds and saves snapshots
func (c *Collector) backgroundAggregation() {
	ticker := time.NewTicker(c.snapshotInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.calculateSpeeds()
			c.saveSnapshot()
		case <-c.stopChan:
			return
		}
	}
}

// calculateSpeeds calculates upload and download speeds
func (c *Collector) calculateSpeeds() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(c.lastSnapshot).Seconds()

	if elapsed > 0 {
		currentBytesReceived := atomic.LoadInt64(&c.bytesReceived)
		currentBytesSent := atomic.LoadInt64(&c.bytesSent)

		c.downloadSpeed = float64(currentBytesReceived-c.lastBytesReceived) / elapsed
		c.uploadSpeed = float64(currentBytesSent-c.lastBytesSent) / elapsed

		c.lastBytesReceived = currentBytesReceived
		c.lastBytesSent = currentBytesSent
		c.lastSnapshot = now
	}
}

// saveSnapshot saves current metrics to database
func (c *Collector) saveSnapshot() {
	if c.db == nil {
		return
	}

	snapshot := &models.MetricsSnapshot{
		Timestamp:         time.Now().Unix(),
		ActiveConnections: int(atomic.LoadInt32(&c.activeConnections)),
		TotalConnections:  atomic.LoadInt64(&c.totalConnections),
		BytesReceived:     atomic.LoadInt64(&c.bytesReceived),
		BytesSent:         atomic.LoadInt64(&c.bytesSent),
		UploadSpeed:       c.uploadSpeed,
		DownloadSpeed:     c.downloadSpeed,
		ErrorCount:        atomic.LoadInt64(&c.errorCount),
	}

	c.db.Create(snapshot)
}

// GetHistoricalSnapshots retrieves historical metrics from database
func (c *Collector) GetHistoricalSnapshots(startTime, endTime int64, limit int) ([]models.MetricsSnapshot, error) {
	var snapshots []models.MetricsSnapshot

	query := c.db.Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Order("timestamp ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&snapshots).Error
	return snapshots, err
}

// Stop stops the background aggregation
func (c *Collector) Stop() {
	close(c.stopChan)
}

// Reset resets all metrics counters
func (c *Collector) Reset() {
	atomic.StoreInt32(&c.activeConnections, 0)
	atomic.StoreInt64(&c.totalConnections, 0)
	atomic.StoreInt64(&c.bytesReceived, 0)
	atomic.StoreInt64(&c.bytesSent, 0)
	atomic.StoreInt64(&c.errorCount, 0)

	c.mu.Lock()
	c.startTime = time.Now()
	c.lastSnapshot = time.Now()
	c.lastBytesReceived = 0
	c.lastBytesSent = 0
	c.uploadSpeed = 0
	c.downloadSpeed = 0
	c.mu.Unlock()
}
