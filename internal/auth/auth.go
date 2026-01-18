package auth

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"go-proxy-server/internal/constants"
	"go-proxy-server/internal/logger"
	"go-proxy-server/internal/models"
)

type Credentials map[string][]byte

// dnsCacheEntry stores DNS lookup results with expiration time
type dnsCacheEntry struct {
	ips       []net.IP
	expiresAt time.Time
	key       string // Store key for LRU eviction
}

// authCacheEntry stores authentication results with expiration time
type authCacheEntry struct {
	authenticated bool
	expiresAt     time.Time
}

// lruCache implements a simple LRU cache for DNS entries
type lruCache struct {
	mu       sync.Mutex
	capacity int
	cache    map[string]*list.Element
	lruList  *list.List
}

type lruEntry struct {
	key   string
	value dnsCacheEntry
}

// Optimized whitelist and credentials storage to avoid type assertion overhead
type whitelistMap struct {
	data map[string]bool
}

type credentialsMap struct {
	data Credentials
}

var (
	// Use atomic.Value for lock-free reads in high-concurrency scenarios
	// This eliminates read lock contention and improves performance
	ipWhitelistAtomic atomic.Value // stores *whitelistMap
	credentialsAtomic atomic.Value // stores *credentialsMap
	// Mutex only needed for write operations (periodic reload and manual add/delete)
	whitelistWriteLock sync.Mutex
	credWriteLock      sync.Mutex
	// Dummy hash for timing attack protection (generated at init)
	dummyHash []byte
	// DNS cache with LRU eviction
	dnsLRUCache *lruCache
	// Authentication cache for SOCKS5 (key: hash(clientIP+username), value: authCacheEntry)
	authCache sync.Map
	// Auth cache cleanup started flag
	authCacheCleanupStarted atomic.Bool
)

// newLRUCache creates a new LRU cache with the specified capacity
func newLRUCache(capacity int) *lruCache {
	return &lruCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		lruList:  list.New(),
	}
}

// Get retrieves a value from the LRU cache
func (c *lruCache) Get(key string) (dnsCacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		entry := elem.Value.(*lruEntry)
		// Check if expired
		if time.Now().After(entry.value.expiresAt) {
			// Remove expired entry
			c.lruList.Remove(elem)
			delete(c.cache, key)
			return dnsCacheEntry{}, false
		}
		// Move to front (most recently used)
		c.lruList.MoveToFront(elem)
		return entry.value, true
	}
	return dnsCacheEntry{}, false
}

// Put adds or updates a value in the LRU cache
func (c *lruCache) Put(key string, value dnsCacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing entry
	if elem, ok := c.cache[key]; ok {
		c.lruList.MoveToFront(elem)
		elem.Value.(*lruEntry).value = value
		return
	}

	// Add new entry
	entry := &lruEntry{key: key, value: value}
	elem := c.lruList.PushFront(entry)
	c.cache[key] = elem

	// Evict least recently used if over capacity
	if c.lruList.Len() > c.capacity {
		oldest := c.lruList.Back()
		if oldest != nil {
			c.lruList.Remove(oldest)
			delete(c.cache, oldest.Value.(*lruEntry).key)
		}
	}
}

// CleanExpired removes all expired entries from the cache
func (c *lruCache) CleanExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0

	// Iterate through all entries and remove expired ones
	for elem := c.lruList.Back(); elem != nil; {
		entry := elem.Value.(*lruEntry)
		prev := elem.Prev()

		if now.After(entry.value.expiresAt) {
			c.lruList.Remove(elem)
			delete(c.cache, entry.key)
			removed++
		}

		elem = prev
	}

	return removed
}

func init() {
	// Initialize atomic values with empty maps wrapped in structs
	ipWhitelistAtomic.Store(&whitelistMap{data: make(map[string]bool)})
	credentialsAtomic.Store(&credentialsMap{data: make(Credentials)})

	// Initialize DNS LRU cache
	dnsLRUCache = newLRUCache(constants.DNSCacheMaxSize)

	// Generate dummy hash at initialization for timing attack protection
	// This prevents attackers from distinguishing between valid and invalid usernames
	var err error
	dummyHash, err = bcrypt.GenerateFromPassword([]byte(""), bcrypt.DefaultCost)
	if err != nil {
		// Fallback to a pre-computed hash if generation fails
		dummyHash = []byte("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy")
	}
}

func CheckIPWhitelist(clientIP string) bool {
	// Lock-free read using atomic.Value - no type assertion overhead
	whitelist := ipWhitelistAtomic.Load().(*whitelistMap)
	return whitelist.data[clientIP]
}

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func LoadWhitelistFromDB(db *gorm.DB) error {
	var whitelist []models.Whitelist

	err := db.Find(&whitelist).Error
	if err != nil {
		return err
	}

	tempWhitelist := make(map[string]bool)
	for _, item := range whitelist {
		tempWhitelist[item.IP] = true
	}

	// Atomic store - no read lock needed, lock-free reads continue to work
	whitelistWriteLock.Lock()
	ipWhitelistAtomic.Store(&whitelistMap{data: tempWhitelist})
	whitelistWriteLock.Unlock()

	return nil
}

func AddIPToWhitelist(db *gorm.DB, ip string) error {
	if !isValidIP(ip) {
		return fmt.Errorf("invalid ip")
	}

	// Directly insert and rely on database unique constraint
	// This prevents race conditions in concurrent scenarios
	whitelist := models.Whitelist{IP: ip}
	err := db.Create(&whitelist).Error
	if err != nil {
		// Check if error is due to unique constraint violation
		if strings.Contains(err.Error(), "UNIQUE constraint failed") ||
			strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("IP already in whitelist")
		}
		return err
	}

	// Reload whitelist from database
	// If reload fails, rollback the database insertion to maintain consistency
	if err := LoadWhitelistFromDB(db); err != nil {
		// Rollback: delete the just-inserted record
		db.Unscoped().Where("ip = ?", ip).Delete(&models.Whitelist{})
		return fmt.Errorf("failed to reload whitelist after insertion: %w", err)
	}

	return nil
}

func DeleteIPFromWhitelist(db *gorm.DB, ip string) error {
	// Use Unscoped to permanently delete the record (hard delete)
	err := db.Unscoped().Where("ip = ?", ip).Delete(&models.Whitelist{}).Error
	if err != nil {
		return err
	}

	// Reload whitelist from database
	if err := LoadWhitelistFromDB(db); err != nil {
		logger.Error("Failed to reload whitelist after deletion: %v", err)
		return err
	}

	return nil
}

func LoadCredentialsFromDB(db *gorm.DB) error {
	var users []models.User

	err := db.Find(&users).Error

	if err != nil {
		return err
	}

	tempCred := make(Credentials)

	for _, user := range users {
		// Username should be globally unique due to database constraint
		// If duplicate found, it indicates data corruption
		if _, exists := tempCred[user.Username]; exists {
			return fmt.Errorf("data corruption: duplicate username '%s' found in database", user.Username)
		}
		tempCred[user.Username] = user.Password
	}

	// Atomic store - no read lock needed, lock-free reads continue to work
	credWriteLock.Lock()
	credentialsAtomic.Store(&credentialsMap{data: tempCred})
	credWriteLock.Unlock()

	return nil
}

func AddUser(db *gorm.DB, ip, username, password string) error {
	// Validate password strength
	if err := validatePasswordStrength(password); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := models.User{
		IP:       ip,
		Username: username,
		Password: hashedPassword,
	}

	// Directly insert and rely on database unique constraint
	// This prevents race conditions in concurrent scenarios
	err = db.Create(&user).Error
	if err != nil {
		// Check if error is due to unique constraint violation
		if strings.Contains(err.Error(), "UNIQUE constraint failed") ||
			strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("Username '%s' already exists", username)
		}
		return err
	}

	// Update the userCredentials map by re-syncing from the database
	// If reload fails, rollback the database insertion to maintain consistency
	if err := LoadCredentialsFromDB(db); err != nil {
		// Rollback: delete the just-inserted record
		db.Unscoped().Where("username = ?", username).Delete(&models.User{})
		return fmt.Errorf("failed to reload credentials after insertion: %w", err)
	}

	return nil
}

// validatePasswordStrength checks if the password meets minimum security requirements
func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if len(password) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}

	// Check for at least one letter and one number
	hasLetter := false
	hasDigit := false
	for _, char := range password {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
		}
		if char >= '0' && char <= '9' {
			hasDigit = true
		}
		if hasLetter && hasDigit {
			break
		}
	}

	if !hasLetter {
		return fmt.Errorf("password must contain at least one letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	return nil
}

func DeleteUser(db *gorm.DB, username string) error {
	// Use Unscoped to permanently delete the record (hard delete)
	// Username is globally unique, so we only need to check username
	err := db.Unscoped().Where("username = ?", username).Delete(&models.User{}).Error
	if err != nil {
		return err
	}

	// Update the userCredentials map by re-syncing from the database
	if err := LoadCredentialsFromDB(db); err != nil {
		logger.Error("Failed to reload credentials after deletion: %v", err)
		return err
	}

	return nil
}

func ListUsers(db *gorm.DB) error {
	var users []models.User
	err := db.Find(&users).Error
	if err != nil {
		logger.Error("Failed to list users: %v", err)
		return err
	}

	fmt.Println("Username")
	fmt.Println("----------")

	for _, user := range users {
		fmt.Printf("%-15s\t\n", user.Username)
	}

	return nil
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

func VerifyCredentials(username string, password []byte) error {
	// Lock-free read using atomic.Value - no type assertion overhead
	creds := credentialsAtomic.Load().(*credentialsMap)
	expectedPassword, ok := creds.data[username]

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

func GetWhitelistIPs() []string {
	// Lock-free read using atomic.Value - no type assertion overhead
	whitelist := ipWhitelistAtomic.Load().(*whitelistMap)

	ips := make([]string, 0, len(whitelist.data))
	for ip := range whitelist.data {
		ips = append(ips, ip)
	}
	return ips
}

// IsPrivateIP checks if an IP address is private/internal
// Uses Go standard library methods for reliable detection
func IsPrivateIP(ip net.IP) bool {
	// Check for loopback addresses (127.0.0.0/8, ::1)
	if ip.IsLoopback() {
		return true
	}

	// Check for link-local addresses (169.254.0.0/16, fe80::/10)
	if ip.IsLinkLocalUnicast() {
		return true
	}

	// Check for private addresses (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, fc00::/7)
	// Note: IsPrivate() was added in Go 1.17
	if ip.IsPrivate() {
		return true
	}

	return false
}

// CheckSSRF validates that the target host is not a private/internal address
// Returns error if the host is private or cannot be resolved
// Note: This is the initial check before connection. Use VerifyConnectedIP() after
// establishing connection to prevent DNS rebinding attacks.
func CheckSSRF(host string) error {
	// Parse host to extract IP or hostname
	// host can be "example.com:80" or "192.168.1.1:80" or just "example.com"
	hostOnly := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostOnly = h
	}

	// Try to parse as IP first
	if ip := net.ParseIP(hostOnly); ip != nil {
		if IsPrivateIP(ip) {
			return fmt.Errorf("access to private IP addresses is not allowed")
		}
		return nil
	}

	// If not an IP, resolve the hostname with caching
	var ips []net.IP

	// Check DNS LRU cache first
	if entry, ok := dnsLRUCache.Get(hostOnly); ok {
		// Cache hit and not expired (Get already checks expiration)
		ips = entry.ips
	}

	// Cache miss or expired, perform DNS lookup
	if ips == nil {
		resolver := &net.Resolver{}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var err error
		ips, err = resolver.LookupIP(ctx, "ip", hostOnly)
		if err != nil {
			// DNS resolution failure could be used to bypass SSRF protection
			// Return error to prevent potential security bypass
			// Note: Don't log the hostname or error details to avoid leaking user's target destinations
			logger.Warn("DNS resolution failed during SSRF check")
			return fmt.Errorf("failed to resolve hostname: %v", err)
		}

		// Store in LRU cache with TTL
		dnsLRUCache.Put(hostOnly, dnsCacheEntry{
			ips:       ips,
			expiresAt: time.Now().Add(constants.DNSCacheTTL),
			key:       hostOnly,
		})
	}

	// Check all resolved IPs
	for _, ip := range ips {
		if IsPrivateIP(ip) {
			return fmt.Errorf("hostname resolves to private IP address, access not allowed")
		}
	}

	return nil
}

// VerifyConnectedIP verifies that the actual connected IP is not private
// This prevents DNS rebinding attacks where DNS resolves to public IP initially
// but later resolves to private IP when connection is established
func VerifyConnectedIP(conn net.Conn) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	remoteAddr := conn.RemoteAddr()
	if remoteAddr == nil {
		return fmt.Errorf("remote address is nil")
	}

	// Extract IP from remote address
	var ip net.IP
	switch addr := remoteAddr.(type) {
	case *net.TCPAddr:
		ip = addr.IP
	case *net.UDPAddr:
		ip = addr.IP
	default:
		// Try to parse address string
		host, _, err := net.SplitHostPort(remoteAddr.String())
		if err != nil {
			return fmt.Errorf("failed to parse remote address: %v", err)
		}
		ip = net.ParseIP(host)
		if ip == nil {
			return fmt.Errorf("failed to parse IP from remote address")
		}
	}

	// Verify the connected IP is not private
	if IsPrivateIP(ip) {
		return fmt.Errorf("connected to private IP address: %s (possible DNS rebinding attack)", ip.String())
	}

	return nil
}
