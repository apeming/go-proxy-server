package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/bcrypt"

	"go-proxy-server/internal/constants"
)

// authCacheEntry stores authentication results with expiration time
type authCacheEntry struct {
	authenticated bool
	expiresAt     time.Time
}

var (
	// Dummy hash for timing attack protection (generated at init)
	dummyHash []byte
	// Authentication cache for SOCKS5 (key: hash(clientIP+username), value: authCacheEntry)
	authCache sync.Map
	// Auth cache cleanup started flag
	authCacheCleanupStarted atomic.Bool
)

func init() {
	// Generate dummy hash at initialization for timing attack protection
	// This prevents attackers from distinguishing between valid and invalid usernames
	var err error
	dummyHash, err = bcrypt.GenerateFromPassword([]byte(""), bcrypt.DefaultCost)
	if err != nil {
		// Fallback to a pre-computed hash if generation fails
		dummyHash = []byte("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy")
	}
}

// generateAuthCacheKey generates a cache key for authentication
func generateAuthCacheKey(clientIP, username string) string {
	hash := sha256.Sum256([]byte(clientIP + ":" + username))
	return hex.EncodeToString(hash[:])
}

// CheckAuthCache checks if authentication is cached and still valid
func CheckAuthCache(clientIP, username string) bool {
	key := generateAuthCacheKey(clientIP, username)
	if cached, ok := authCache.Load(key); ok {
		if entry, ok := cached.(authCacheEntry); ok {
			if time.Now().Before(entry.expiresAt) && entry.authenticated {
				return true
			}
			// Expired or not authenticated, remove from cache
			authCache.Delete(key)
		}
	}
	return false
}

// SetAuthCache caches authentication result
func SetAuthCache(clientIP, username string, authenticated bool) {
	key := generateAuthCacheKey(clientIP, username)
	entry := authCacheEntry{
		authenticated: authenticated,
		expiresAt:     time.Now().Add(constants.AuthCacheTTL),
	}
	authCache.Store(key, entry)
}

// cleanupAuthCache periodically removes expired entries from the auth cache
func cleanupAuthCache() {
	ticker := time.NewTicker(constants.AuthCacheCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		authCache.Range(func(key, value interface{}) bool {
			if entry, ok := value.(authCacheEntry); ok {
				if now.After(entry.expiresAt) {
					authCache.Delete(key)
				}
			} else {
				// Invalid entry type, delete it
				authCache.Delete(key)
			}
			return true
		})
	}
}

// VerifyCredentials verifies username and password against stored credentials
// Uses constant-time comparison to prevent timing attacks
func VerifyCredentials(username string, password []byte) error {
	// Lock-free read using atomic.Value - no type assertion overhead
	creds := getCredentials()
	expectedPassword, ok := creds[username]

	// To prevent timing attacks, always perform bcrypt comparison
	// even if username doesn't exist. Use the dynamically generated dummy hash.
	if !ok {
		// Use the dummy hash generated at init time
		// This ensures consistent timing regardless of username existence
		bcrypt.CompareHashAndPassword(dummyHash, password)
		return fmt.Errorf("invalid credentials")
	}

	// Compare the received password with the expected password
	if err := bcrypt.CompareHashAndPassword(expectedPassword, password); err != nil {
		return fmt.Errorf("invalid credentials")
	}

	return nil
}

// VerifyCredentialsWithCache verifies credentials with caching support
// This reduces bcrypt overhead for repeated authentication attempts
func VerifyCredentialsWithCache(clientIP, username string, password []byte) error {
	// Start auth cache cleanup goroutine on first call
	if authCacheCleanupStarted.CompareAndSwap(false, true) {
		go cleanupAuthCache()
	}

	// Check cache first
	if CheckAuthCache(clientIP, username) {
		return nil
	}

	// Verify credentials
	err := VerifyCredentials(username, password)

	// Cache the result (only cache successful authentications)
	if err == nil {
		SetAuthCache(clientIP, username, true)
	}

	return err
}
