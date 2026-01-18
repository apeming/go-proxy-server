package web

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"go-proxy-server/internal/auth"
	"go-proxy-server/internal/autostart"
	"go-proxy-server/internal/config"
	"go-proxy-server/internal/models"
	"go-proxy-server/internal/proxy"
	"go-proxy-server/internal/proxyconfig"
	"go-proxy-server/internal/sysconfig"
)

// ProxyServer represents a running proxy server
type ProxyServer struct {
	Type       string // "socks5" or "http"
	Port       int
	BindListen bool
	AutoStart  bool // Whether to auto-start on application launch
	Listener   net.Listener
	Running    bool
	mu         sync.Mutex
}

// Manager manages the web interface and proxy servers
type Manager struct {
	db          *gorm.DB
	socksServer *ProxyServer
	httpServer  *ProxyServer
	mu          sync.RWMutex
	webPort     int
}

// NewManager creates a new web manager
func NewManager(db *gorm.DB, webPort int) *Manager {
	manager := &Manager{
		db:      db,
		webPort: webPort,
		socksServer: &ProxyServer{
			Type: "socks5",
		},
		httpServer: &ProxyServer{
			Type: "http",
		},
	}

	// Load saved configurations from database
	if socksConfig, err := proxyconfig.LoadConfigFromDB(db, "socks5"); err == nil && socksConfig != nil {
		manager.socksServer.Port = socksConfig.Port
		manager.socksServer.BindListen = socksConfig.BindListen
		manager.socksServer.AutoStart = socksConfig.AutoStart
	}

	if httpConfig, err := proxyconfig.LoadConfigFromDB(db, "http"); err == nil && httpConfig != nil {
		manager.httpServer.Port = httpConfig.Port
		manager.httpServer.BindListen = httpConfig.BindListen
		manager.httpServer.AutoStart = httpConfig.AutoStart
	}

	return manager
}

// StartServer starts the web management server
func (wm *Manager) StartServer() error {
	// Setup API routes
	http.HandleFunc("/api/status", wm.handleStatus)
	http.HandleFunc("/api/users", wm.handleUsers)
	http.HandleFunc("/api/whitelist", wm.handleWhitelist)
	http.HandleFunc("/api/proxy/start", wm.handleProxyStart)
	http.HandleFunc("/api/proxy/stop", wm.handleProxyStop)
	http.HandleFunc("/api/proxy/config", wm.handleProxyConfig)
	http.HandleFunc("/api/system/settings", wm.handleSystemSettings)
	http.HandleFunc("/api/timeout", wm.handleTimeout)

	// Static files and SPA fallback (must be last)
	http.HandleFunc("/", wm.handleIndex)

	// Only listen on localhost for security
	addr := fmt.Sprintf("localhost:%d", wm.webPort)
	fmt.Printf("Web management interface started at http://%s\n", addr)
	fmt.Printf("Open your browser and visit: http://%s\n", addr)

	return http.ListenAndServe(addr, nil)
}

// handleIndex serves the static files and SPA fallback
func (wm *Manager) handleIndex(w http.ResponseWriter, r *http.Request) {
	// If requesting API path, return 404
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	// Get embedded static file system
	staticFS, err := GetStaticFS()
	if err != nil {
		http.Error(w, "Failed to load static files", http.StatusInternalServerError)
		return
	}

	// Create file server
	fileServer := http.FileServer(http.FS(staticFS))

	// If requested file doesn't exist, serve index.html (SPA fallback)
	if r.URL.Path != "/" {
		if _, err := staticFS.Open(strings.TrimPrefix(r.URL.Path, "/")); err != nil {
			r.URL.Path = "/"
		}
	}

	fileServer.ServeHTTP(w, r)
}

// handleStatus returns the current status of proxy servers
func (wm *Manager) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	wm.mu.RLock()
	defer wm.mu.RUnlock()

	status := map[string]interface{}{
		"socks5": map[string]interface{}{
			"running":    wm.socksServer.Running,
			"port":       wm.socksServer.Port,
			"bindListen": wm.socksServer.BindListen,
			"autoStart":  wm.socksServer.AutoStart,
		},
		"http": map[string]interface{}{
			"running":    wm.httpServer.Running,
			"port":       wm.httpServer.Port,
			"bindListen": wm.httpServer.BindListen,
			"autoStart":  wm.httpServer.AutoStart,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleUsers handles user management (GET, POST, DELETE)
func (wm *Manager) handleUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// List all users
		var users []models.User
		if err := wm.db.Find(&users).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(users)

	case http.MethodPost:
		// Add new user
		var req struct {
			IP       string `json:"ip"`
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := auth.AddUser(wm.db, req.IP, req.Username, req.Password); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "success"})

	case http.MethodDelete:
		// Delete user
		var req struct {
			Username string `json:"username"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := auth.DeleteUser(wm.db, req.Username); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "success"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleWhitelist handles IP whitelist management
func (wm *Manager) handleWhitelist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// List whitelist IPs
		ips := auth.GetWhitelistIPs()
		json.NewEncoder(w).Encode(ips)

	case http.MethodPost:
		// Add IP to whitelist
		var req struct {
			IP string `json:"ip"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := auth.AddIPToWhitelist(wm.db, req.IP); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "success"})

	case http.MethodDelete:
		// Delete IP from whitelist
		var req struct {
			IP string `json:"ip"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := auth.DeleteIPFromWhitelist(wm.db, req.IP); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "success"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleProxyStart starts a proxy server
func (wm *Manager) handleProxyStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Type       string `json:"type"` // "socks5" or "http"
		Port       int    `json:"port"`
		BindListen bool   `json:"bindListen"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wm.mu.Lock()
	defer wm.mu.Unlock()

	var server *ProxyServer
	if req.Type == "socks5" {
		server = wm.socksServer
	} else if req.Type == "http" {
		server = wm.httpServer
	} else {
		http.Error(w, "Invalid proxy type", http.StatusBadRequest)
		return
	}

	if server.Running {
		http.Error(w, "Proxy already running", http.StatusBadRequest)
		return
	}

	// Start the proxy server
	if err := wm.startProxy(server, req.Port, req.BindListen); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleProxyStop stops a proxy server
func (wm *Manager) handleProxyStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Type string `json:"type"` // "socks5" or "http"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wm.mu.Lock()
	defer wm.mu.Unlock()

	var server *ProxyServer
	if req.Type == "socks5" {
		server = wm.socksServer
	} else if req.Type == "http" {
		server = wm.httpServer
	} else {
		http.Error(w, "Invalid proxy type", http.StatusBadRequest)
		return
	}

	if !server.Running {
		http.Error(w, "Proxy not running", http.StatusBadRequest)
		return
	}

	// Stop the proxy server
	wm.stopProxy(server)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// startProxy starts a proxy server in a goroutine
func (wm *Manager) startProxy(server *ProxyServer, port int, bindListen bool) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	server.Port = port
	server.BindListen = bindListen
	server.Listener = listener
	server.Running = true

	// Save configuration to database
	config := &models.ProxyConfig{
		Type:       server.Type,
		Port:       port,
		BindListen: bindListen,
		AutoStart:  server.AutoStart, // Preserve existing AutoStart setting
	}
	if err := proxyconfig.SaveConfigToDB(wm.db, config); err != nil {
		fmt.Printf("Warning: Failed to save proxy config to database: %v\n", err)
	}

	// Start config reload goroutine if not already running
	go func() {
		for server.Running {
			auth.LoadCredentialsFromDB(wm.db)
			auth.LoadWhitelistFromDB(wm.db)
			time.Sleep(time.Second * 10)
		}
	}()

	// Start accepting connections
	go func() {
		for server.Running {
			conn, err := listener.Accept()
			if err != nil {
				if server.Running {
					fmt.Printf("%s proxy accept error: %v\n", server.Type, err)
				}
				continue
			}

			if server.Type == "socks5" {
				go proxy.HandleSocks5Connection(conn, bindListen)
			} else if server.Type == "http" {
				go proxy.HandleHTTPConnection(conn, bindListen)
			}
		}
	}()

	fmt.Printf("%s proxy started on port %d\n", server.Type, port)
	return nil
}

// stopProxy stops a running proxy server
func (wm *Manager) stopProxy(server *ProxyServer) {
	server.Running = false
	if server.Listener != nil {
		server.Listener.Close()
	}
	fmt.Printf("%s proxy stopped\n", server.Type)
}

// handleProxyConfig handles proxy configuration updates
func (wm *Manager) handleProxyConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Type       string `json:"type"`
		Port       int    `json:"port"`
		BindListen bool   `json:"bindListen"`
		AutoStart  bool   `json:"autoStart"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wm.mu.Lock()
	defer wm.mu.Unlock()

	var server *ProxyServer
	if req.Type == "socks5" {
		server = wm.socksServer
	} else if req.Type == "http" {
		server = wm.httpServer
	} else {
		http.Error(w, "Invalid proxy type", http.StatusBadRequest)
		return
	}

	// Update configuration in memory
	server.AutoStart = req.AutoStart
	if !server.Running {
		// Only update port and bindListen if proxy is not running
		server.Port = req.Port
		server.BindListen = req.BindListen
	}

	// Save configuration to database
	config := &models.ProxyConfig{
		Type:       server.Type,
		Port:       server.Port,
		BindListen: server.BindListen,
		AutoStart:  server.AutoStart,
	}
	if err := proxyconfig.SaveConfigToDB(wm.db, config); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// AutoStartProxy starts a proxy server automatically on application launch
func (wm *Manager) AutoStartProxy(proxyType string, port int, bindListen bool) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	var server *ProxyServer
	if proxyType == "socks5" {
		server = wm.socksServer
	} else if proxyType == "http" {
		server = wm.httpServer
	} else {
		return fmt.Errorf("invalid proxy type: %s", proxyType)
	}

	if server.Running {
		return fmt.Errorf("proxy already running")
	}

	return wm.startProxy(server, port, bindListen)
}

// handleSystemSettings handles system settings (GET, POST)
func (wm *Manager) handleSystemSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// Get current settings
		autostartValue, _ := sysconfig.GetConfig(wm.db, sysconfig.KeyAutoStart)
		autostartEnabled := autostartValue == "true"

		// Check actual registry status (Windows only)
		registryEnabled, _ := autostart.IsEnabled()

		settings := map[string]interface{}{
			"autostartEnabled":   autostartEnabled,
			"registryEnabled":    registryEnabled,
			"autostartSupported": true, // Will be false on non-Windows
		}

		json.NewEncoder(w).Encode(settings)

	case http.MethodPost:
		// Update settings
		var req struct {
			AutostartEnabled bool `json:"autostartEnabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Update registry
		if req.AutostartEnabled {
			if err := autostart.Enable(); err != nil {
				http.Error(w, fmt.Sprintf("Failed to enable autostart: %v", err), http.StatusInternalServerError)
				return
			}
		} else {
			if err := autostart.Disable(); err != nil {
				http.Error(w, fmt.Sprintf("Failed to disable autostart: %v", err), http.StatusInternalServerError)
				return
			}
		}

		// Update database
		value := "false"
		if req.AutostartEnabled {
			value = "true"
		}
		if err := sysconfig.SetConfig(wm.db, sysconfig.KeyAutoStart, value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "success"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTimeout handles timeout configuration (GET, POST)
func (wm *Manager) handleTimeout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// Get current timeout configuration
		timeout := config.GetTimeout()

		response := map[string]interface{}{
			"connect":   int(timeout.Connect.Seconds()),
			"idleRead":  int(timeout.IdleRead.Seconds()),
			"idleWrite": int(timeout.IdleWrite.Seconds()),
		}

		json.NewEncoder(w).Encode(response)

	case http.MethodPost:
		// Update timeout configuration
		var req struct {
			Connect   int `json:"connect"`   // in seconds
			IdleRead  int `json:"idleRead"`  // in seconds
			IdleWrite int `json:"idleWrite"` // in seconds
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate timeout values
		if req.Connect <= 0 || req.Connect > 300 {
			http.Error(w, "Connect timeout must be between 1 and 300 seconds", http.StatusBadRequest)
			return
		}
		if req.IdleRead <= 0 || req.IdleRead > 3600 {
			http.Error(w, "Idle read timeout must be between 1 and 3600 seconds", http.StatusBadRequest)
			return
		}
		if req.IdleWrite <= 0 || req.IdleWrite > 3600 {
			http.Error(w, "Idle write timeout must be between 1 and 3600 seconds", http.StatusBadRequest)
			return
		}

		// Create new timeout configuration
		newTimeout := config.TimeoutConfig{
			Connect:   time.Duration(req.Connect) * time.Second,
			IdleRead:  time.Duration(req.IdleRead) * time.Second,
			IdleWrite: time.Duration(req.IdleWrite) * time.Second,
		}

		// Save to database
		if err := config.SaveTimeoutToDB(wm.db, newTimeout); err != nil {
			http.Error(w, fmt.Sprintf("Failed to save timeout configuration: %v", err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "success"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
