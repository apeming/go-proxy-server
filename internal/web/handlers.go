package web

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"go-proxy-server/internal/auth"
	"go-proxy-server/internal/autostart"
	"go-proxy-server/internal/config"
	"go-proxy-server/internal/metrics"
	"go-proxy-server/internal/models"
)

// StartServer starts the web management server
func (wm *Manager) StartServer() error {
	// Create a new ServeMux for this server
	mux := http.NewServeMux()

	// Setup API routes
	mux.HandleFunc("/api/status", wm.handleStatus)
	mux.HandleFunc("/api/users", wm.handleUsers)
	mux.HandleFunc("/api/whitelist", wm.handleWhitelist)
	mux.HandleFunc("/api/proxy/start", wm.handleProxyStart)
	mux.HandleFunc("/api/proxy/stop", wm.handleProxyStop)
	mux.HandleFunc("/api/proxy/config", wm.handleProxyConfig)
	mux.HandleFunc("/api/config", wm.handleConfig)
	mux.HandleFunc("/api/metrics/realtime", wm.handleMetricsRealtime)
	mux.HandleFunc("/api/metrics/history", wm.handleMetricsHistory)
	mux.HandleFunc("/api/shutdown", wm.handleShutdown)

	// Static files and SPA fallback (must be last)
	mux.HandleFunc("/", wm.handleIndex)

	// Create listener
	addr := fmt.Sprintf("localhost:%d", wm.webPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start web server: %w", err)
	}

	// Get actual port (useful when port is 0 for random assignment)
	actualPort := listener.Addr().(*net.TCPAddr).Port
	wm.SetActualPort(actualPort)

	// Print URL with actual port
	fmt.Printf("Web management interface started at http://localhost:%d\n", actualPort)
	fmt.Printf("Open your browser and visit: http://localhost:%d\n", actualPort)

	// Create HTTP server with graceful shutdown support
	wm.webHttpServer = &http.Server{
		Handler: mux,
	}

	// Start serving (this will block until Shutdown is called)
	if err := wm.webHttpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
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
	proxyConfig := &models.ProxyConfig{
		Type:       server.Type,
		Port:       server.Port,
		BindListen: server.BindListen,
		AutoStart:  server.AutoStart,
	}
	if err := config.SaveProxyConfig(wm.db, proxyConfig); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleConfig handles unified configuration (GET, POST)
// Includes: timeout, connection limiter, and system settings
func (wm *Manager) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// Get current timeout configuration
		timeout := config.GetTimeout()

		// Get current limiter configuration
		limiterConfig := config.GetLimiterConfig()

		// Get autostart settings
		autostartValue, _ := config.GetSystemConfig(wm.db, config.KeyAutoStart)
		autostartEnabled := autostartValue == "true"

		// Check actual registry status (Windows only)
		registryEnabled, _ := autostart.IsEnabled()

		response := map[string]interface{}{
			"timeout": map[string]interface{}{
				"connect":   int(timeout.Connect.Seconds()),
				"idleRead":  int(timeout.IdleRead.Seconds()),
				"idleWrite": int(timeout.IdleWrite.Seconds()),
			},
			"limiter": map[string]interface{}{
				"maxConcurrentConnections":      limiterConfig.MaxConcurrentConnections,
				"maxConcurrentConnectionsPerIP": limiterConfig.MaxConcurrentConnectionsPerIP,
			},
			"system": map[string]interface{}{
				"autostartEnabled":   autostartEnabled,
				"registryEnabled":    registryEnabled,
				"autostartSupported": true,
			},
			"security": map[string]interface{}{
				"allowPrivateIPAccess": config.GetAllowPrivateIPAccess(),
			},
		}

		json.NewEncoder(w).Encode(response)

	case http.MethodPost:
		// Update configuration
		var req struct {
			Timeout *struct {
				Connect   int `json:"connect"`
				IdleRead  int `json:"idleRead"`
				IdleWrite int `json:"idleWrite"`
			} `json:"timeout"`
			Limiter *struct {
				MaxConcurrentConnections      int32 `json:"maxConcurrentConnections"`
				MaxConcurrentConnectionsPerIP int32 `json:"maxConcurrentConnectionsPerIP"`
			} `json:"limiter"`
			System *struct {
				AutostartEnabled bool `json:"autostartEnabled"`
			} `json:"system"`
			Security *struct {
				AllowPrivateIPAccess bool `json:"allowPrivateIPAccess"`
			} `json:"security"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Update timeout configuration if provided
		if req.Timeout != nil {
			// Validate timeout values
			if req.Timeout.Connect <= 0 || req.Timeout.Connect > 300 {
				http.Error(w, "Connect timeout must be between 1 and 300 seconds", http.StatusBadRequest)
				return
			}
			if req.Timeout.IdleRead <= 0 || req.Timeout.IdleRead > 3600 {
				http.Error(w, "Idle read timeout must be between 1 and 3600 seconds", http.StatusBadRequest)
				return
			}
			if req.Timeout.IdleWrite <= 0 || req.Timeout.IdleWrite > 3600 {
				http.Error(w, "Idle write timeout must be between 1 and 3600 seconds", http.StatusBadRequest)
				return
			}

			// Create new timeout configuration
			newTimeout := config.TimeoutConfig{
				Connect:   time.Duration(req.Timeout.Connect) * time.Second,
				IdleRead:  time.Duration(req.Timeout.IdleRead) * time.Second,
				IdleWrite: time.Duration(req.Timeout.IdleWrite) * time.Second,
			}

			// Save to database
			if err := config.SaveTimeoutToDB(wm.db, newTimeout); err != nil {
				http.Error(w, fmt.Sprintf("Failed to save timeout configuration: %v", err), http.StatusInternalServerError)
				return
			}
		}

		// Update limiter configuration if provided
		if req.Limiter != nil {
			// Validate limiter values
			if req.Limiter.MaxConcurrentConnections <= 0 || req.Limiter.MaxConcurrentConnections > 1000000 {
				http.Error(w, "Max concurrent connections must be between 1 and 1000000", http.StatusBadRequest)
				return
			}
			if req.Limiter.MaxConcurrentConnectionsPerIP <= 0 || req.Limiter.MaxConcurrentConnectionsPerIP > 100000 {
				http.Error(w, "Max concurrent connections per IP must be between 1 and 100000", http.StatusBadRequest)
				return
			}

			// Update limiter configuration in database and memory
			if err := config.UpdateLimiterConfig(wm.db, req.Limiter.MaxConcurrentConnections, req.Limiter.MaxConcurrentConnectionsPerIP); err != nil {
				http.Error(w, fmt.Sprintf("Failed to update limiter configuration: %v", err), http.StatusInternalServerError)
				return
			}
		}

		// Update system settings if provided
		if req.System != nil {
			// Update registry
			if req.System.AutostartEnabled {
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
			if req.System.AutostartEnabled {
				value = "true"
			}
			if err := config.SetSystemConfig(wm.db, config.KeyAutoStart, value); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Update security settings if provided
		if req.Security != nil {
			if err := config.UpdateAllowPrivateIPAccess(wm.db, req.Security.AllowPrivateIPAccess); err != nil {
				http.Error(w, fmt.Sprintf("Failed to update security configuration: %v", err), http.StatusInternalServerError)
				return
			}
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "success"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleShutdown handles application shutdown request
func (wm *Manager) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Send success response first
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Application is shutting down..."})

	// Flush the response to ensure client receives it
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Shutdown the application in a goroutine to allow the response to be sent
	go func() {
		// Give the response time to be sent
		time.Sleep(500 * time.Millisecond)

		// Gracefully shutdown the application
		if err := wm.ShutdownApplication(); err != nil {
			fmt.Printf("Error during shutdown: %v\n", err)
		}

		// Exit the application
		fmt.Println("Application shutdown complete")
		time.Sleep(100 * time.Millisecond)
		os.Exit(0)
	}()
}

// handleMetricsRealtime returns real-time metrics snapshot
func (wm *Manager) handleMetricsRealtime(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	collector := metrics.GetCollector()
	if collector == nil {
		http.Error(w, "Metrics collector not initialized", http.StatusInternalServerError)
		return
	}

	snapshot := collector.GetSnapshot()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

// handleMetricsHistory returns historical metrics data
func (wm *Manager) handleMetricsHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	collector := metrics.GetCollector()
	if collector == nil {
		http.Error(w, "Metrics collector not initialized", http.StatusInternalServerError)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	startTime := query.Get("startTime")
	endTime := query.Get("endTime")
	limit := query.Get("limit")

	// Default values
	var start, end int64
	var limitInt int = 100

	if startTime != "" {
		fmt.Sscanf(startTime, "%d", &start)
	} else {
		// Default to last 24 hours
		start = time.Now().Add(-24 * time.Hour).Unix()
	}

	if endTime != "" {
		fmt.Sscanf(endTime, "%d", &end)
	} else {
		end = time.Now().Unix()
	}

	if limit != "" {
		fmt.Sscanf(limit, "%d", &limitInt)
	}

	// Get historical snapshots from database
	snapshots, err := collector.GetHistoricalSnapshots(start, end, limitInt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve metrics: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshots)
}
