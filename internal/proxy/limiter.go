package proxy

import (
	"sync"
	"sync/atomic"

	"go-proxy-server/internal/constants"
)

// ConnectionLimiter limits the number of concurrent connections globally and per IP
type ConnectionLimiter struct {
	// Global semaphore for total connection limit
	globalSem chan struct{}
	// Per-IP connection counters
	perIPCounters sync.Map // map[string]*int32
	// Current total connections (for metrics)
	totalConnections atomic.Int64
}

// NewConnectionLimiter creates a new connection limiter
func NewConnectionLimiter() *ConnectionLimiter {
	return &ConnectionLimiter{
		globalSem: make(chan struct{}, constants.MaxConcurrentConnections),
	}
}

// Acquire attempts to acquire a connection slot for the given IP
// Returns true if successful, false if the limit is reached
func (cl *ConnectionLimiter) Acquire(clientIP string) bool {
	// Try to acquire global semaphore (non-blocking)
	select {
	case cl.globalSem <- struct{}{}:
		// Global limit not reached, now check per-IP limit
	default:
		// Global limit reached
		return false
	}

	// Check and increment per-IP counter
	counterInterface, _ := cl.perIPCounters.LoadOrStore(clientIP, new(int32))
	counter := counterInterface.(*int32)

	// Atomically increment and check if limit exceeded
	newCount := atomic.AddInt32(counter, 1)
	if newCount > constants.MaxConcurrentConnectionsPerIP {
		// Per-IP limit exceeded, rollback
		atomic.AddInt32(counter, -1)
		// Release global semaphore
		<-cl.globalSem
		return false
	}

	// Successfully acquired, increment total counter
	cl.totalConnections.Add(1)
	return true
}

// Release releases a connection slot for the given IP
func (cl *ConnectionLimiter) Release(clientIP string) {
	// Decrement per-IP counter
	if counterInterface, ok := cl.perIPCounters.Load(clientIP); ok {
		counter := counterInterface.(*int32)
		newCount := atomic.AddInt32(counter, -1)

		// Clean up counter if it reaches zero to prevent memory leak
		if newCount <= 0 {
			cl.perIPCounters.Delete(clientIP)
		}
	}

	// Release global semaphore
	select {
	case <-cl.globalSem:
	default:
		// Should not happen, but handle gracefully
	}

	// Decrement total counter
	cl.totalConnections.Add(-1)
}

// GetTotalConnections returns the current number of active connections
func (cl *ConnectionLimiter) GetTotalConnections() int64 {
	return cl.totalConnections.Load()
}

// GetPerIPConnections returns the current number of connections for a given IP
func (cl *ConnectionLimiter) GetPerIPConnections(clientIP string) int32 {
	if counterInterface, ok := cl.perIPCounters.Load(clientIP); ok {
		counter := counterInterface.(*int32)
		return atomic.LoadInt32(counter)
	}
	return 0
}

// Global connection limiter instances
var (
	socks5Limiter = NewConnectionLimiter()
	httpLimiter   = NewConnectionLimiter()
)

// GetSOCKS5Limiter returns the global SOCKS5 connection limiter
func GetSOCKS5Limiter() *ConnectionLimiter {
	return socks5Limiter
}

// GetHTTPLimiter returns the global HTTP connection limiter
func GetHTTPLimiter() *ConnectionLimiter {
	return httpLimiter
}
