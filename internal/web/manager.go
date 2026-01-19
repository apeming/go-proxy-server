package web

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"gorm.io/gorm"

	"go-proxy-server/internal/auth"
	"go-proxy-server/internal/config"
	"go-proxy-server/internal/models"
	"go-proxy-server/internal/proxy"
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
	db             *gorm.DB
	socksServer    *ProxyServer
	httpServer     *ProxyServer
	mu             sync.RWMutex
	webPort        int
	actualPort     int // Actual port being used (after binding)
	webHttpServer  *http.Server
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

// NewManager creates a new web manager
func NewManager(db *gorm.DB, webPort int) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		db:             db,
		webPort:        webPort,
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
		socksServer: &ProxyServer{
			Type: "socks5",
		},
		httpServer: &ProxyServer{
			Type: "http",
		},
	}

	// Load saved configurations from database
	if socksConfig, err := config.LoadProxyConfig(db, "socks5"); err == nil && socksConfig != nil {
		manager.socksServer.Port = socksConfig.Port
		manager.socksServer.BindListen = socksConfig.BindListen
		manager.socksServer.AutoStart = socksConfig.AutoStart
	}

	if httpConfig, err := config.LoadProxyConfig(db, "http"); err == nil && httpConfig != nil {
		manager.httpServer.Port = httpConfig.Port
		manager.httpServer.BindListen = httpConfig.BindListen
		manager.httpServer.AutoStart = httpConfig.AutoStart
	}

	return manager
}

// startProxy starts a proxy server
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
	proxyConfig := &models.ProxyConfig{
		Type:       server.Type,
		Port:       port,
		BindListen: bindListen,
		AutoStart:  server.AutoStart, // Preserve existing AutoStart setting
	}
	if err := config.SaveProxyConfig(wm.db, proxyConfig); err != nil {
		fmt.Printf("Warning: Failed to save proxy config to database: %v\n", err)
	}

	// Start config reload goroutine with shutdown context
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-wm.shutdownCtx.Done():
				return
			case <-ticker.C:
				if !server.Running {
					return
				}
				auth.LoadCredentialsFromDB(wm.db)
				auth.LoadWhitelistFromDB(wm.db)
			}
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
		return fmt.Errorf("%s proxy is already running", proxyType)
	}

	return wm.startProxy(server, port, bindListen)
}

// GetActualPort returns the actual port being used by the web server
func (wm *Manager) GetActualPort() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.actualPort
}

// SetActualPort sets the actual port being used by the web server
func (wm *Manager) SetActualPort(port int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.actualPort = port
}

// StopAllProxies stops all running proxy servers
func (wm *Manager) StopAllProxies() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.socksServer.Running {
		wm.stopProxy(wm.socksServer)
	}

	if wm.httpServer.Running {
		wm.stopProxy(wm.httpServer)
	}
}

// Shutdown gracefully shuts down the web server
func (wm *Manager) Shutdown() error {
	// Cancel the shutdown context to stop all goroutines
	if wm.shutdownCancel != nil {
		wm.shutdownCancel()
	}

	// Shutdown the HTTP server gracefully
	if wm.webHttpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return wm.webHttpServer.Shutdown(ctx)
	}

	return nil
}

// ShutdownApplication gracefully shuts down the entire application
func (wm *Manager) ShutdownApplication() error {
	// Stop all proxy servers first
	wm.StopAllProxies()

	// Close all HTTP transport connections
	proxy.CloseAllTransports()

	// Shutdown the web server
	if err := wm.Shutdown(); err != nil {
		return err
	}

	return nil
}
