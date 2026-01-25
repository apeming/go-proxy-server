package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"go-proxy-server/internal/auth"
	"go-proxy-server/internal/config"
	"go-proxy-server/internal/constants"
	applogger "go-proxy-server/internal/logger"
	"go-proxy-server/internal/metrics"
	"go-proxy-server/internal/models"
	"go-proxy-server/internal/proxy"
	"go-proxy-server/internal/singleinstance"
	"go-proxy-server/internal/tray"
	"go-proxy-server/internal/web"
)

// setupCleanupHandler sets up signal handlers for graceful shutdown
func setupCleanupHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		applogger.Info("Received shutdown signal, cleaning up...")

		// Close all HTTP transport connections
		proxy.CloseAllTransports()
		applogger.Info("All transport connections closed")

		// Close logger
		applogger.Close()

		os.Exit(0)
	}()
}

// startConfigReloader starts a background goroutine to reload configuration periodically
func startConfigReloader(db *gorm.DB) {
	go func() {
		ticker := time.NewTicker(constants.ConfigReloadInterval)
		defer ticker.Stop()

		for range ticker.C {
			auth.LoadCredentialsFromDB(db)
			auth.LoadWhitelistFromDB(db)
		}
	}()
}

// isListenerClosed checks if the error indicates the listener is closed
func isListenerClosed(err error) bool {
	if err == nil {
		return false
	}
	// Check for common listener closed error messages
	errStr := err.Error()
	return strings.Contains(errStr, "use of closed network connection") ||
		strings.Contains(errStr, "listener closed")
}

// runProxyServer runs a proxy server with proper error handling
// Returns error channel that will receive fatal errors
func runProxyServer(proxyType string, port int, bindListen bool, db *gorm.DB) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to start %s listener: %w", proxyType, err)
	}
	defer listener.Close()

	applogger.Info("%s proxy server started on port %d", proxyType, port)

	consecutiveErrors := 0
	for {
		conn, err := listener.Accept()
		if err != nil {
			// Check if listener is closed (normal shutdown)
			if isListenerClosed(err) {
				applogger.Info("%s proxy server stopped", proxyType)
				return nil
			}

			// Log the error
			applogger.Error("%s accept failed: %v", proxyType, err)
			consecutiveErrors++

			// If too many consecutive errors, consider it a fatal error
			if consecutiveErrors >= constants.MaxConsecutiveAcceptErrors {
				return fmt.Errorf("%s proxy: too many consecutive accept errors", proxyType)
			}

			// Backoff before retrying
			time.Sleep(constants.AcceptErrorBackoff)
			continue
		}

		// Reset error counter on successful accept
		consecutiveErrors = 0

		// Handle connection based on proxy type
		if proxyType == "SOCKS5" {
			go proxy.HandleSocks5Connection(conn, bindListen)
		} else if proxyType == "HTTP" {
			go proxy.HandleHTTPConnection(conn, bindListen)
		}
	}
}

func main() {
	// Initialize logger for stdout output
	applogger.InitStdout()

	// Check for single instance (only on Windows, and only in GUI mode without arguments)
	if runtime.GOOS == "windows" && len(os.Args) == 1 {
		isOnly, err := singleinstance.Check("Global\\GoProxyServerInstance")
		if err != nil {
			applogger.Error("Failed to check single instance: %v", err)
			fmt.Printf("警告: 无法检查是否已有实例运行: %v\n", err)
		} else if !isOnly {
			// Another instance is already running
			applogger.Info("Another instance is already running, exiting")
			fmt.Println("======================================")
			fmt.Println("检测到程序已在运行!")
			fmt.Println("Another instance is already running!")
			fmt.Println("======================================")
			fmt.Println()
			fmt.Println("请检查系统托盘（任务栏右下角）是否已有图标。")
			fmt.Println("Please check the system tray (bottom-right of taskbar) for the application icon.")
			fmt.Println()
			fmt.Println("按任意键退出... Press any key to exit...")
			fmt.Scanln()
			return
		}
		defer singleinstance.Release()
		applogger.Info("Single instance check passed")
	}

	err := config.Load()
	if err != nil {
		applogger.Error("Config error: %v", err)
		return
	}

	// Initialize logger (for Windows GUI mode)
	if err := applogger.Init(); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		// Continue anyway
	}
	defer applogger.Close()

	// Setup cleanup handler for graceful shutdown
	setupCleanupHandler()

	applogger.Info("Go Proxy Server starting...")

	dbPath, err := config.GetDbPath()
	if err != nil {
		applogger.Error("Failed to get database path: %v", err)
		return
	}
	applogger.Info("Config loaded - DB: %s", dbPath)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		applogger.Error("Failed to open database: %v", err)
		return
	}
	applogger.Info("Database opened successfully")

	err = db.AutoMigrate(&models.User{}, &models.Whitelist{}, &models.ProxyConfig{}, &models.SystemConfig{}, &models.MetricsSnapshot{}, &models.AlertConfig{}, &models.AlertHistory{})
	if err != nil {
		applogger.Error("Failed to migrate database: %v", err)
		return
	}
	applogger.Info("Database migration completed")

	// Initialize metrics collector (10-second snapshot interval)
	metrics.InitCollector(db, 10*time.Second)
	applogger.Info("Metrics collector initialized")

	// Initialize timeout configuration from database
	if err := config.InitTimeout(db); err != nil {
		applogger.Error("Failed to initialize timeout configuration: %v", err)
		return
	}
	applogger.Info("Timeout configuration initialized")

	// Start timeout configuration reloader
	config.StartTimeoutReloader(db)

	// Initialize connection limiter configuration from database
	if err := config.InitLimiterConfig(db); err != nil {
		applogger.Error("Failed to initialize connection limiter configuration: %v", err)
		return
	}
	applogger.Info("Connection limiter configuration initialized")

	// Initialize security configuration from database
	if err := config.InitSecurityConfig(db); err != nil {
		applogger.Error("Failed to initialize security configuration: %v", err)
		return
	}
	applogger.Info("Security configuration initialized")

	// Configure database connection pool
	sqlDB, err := db.DB()
	if err != nil {
		applogger.Error("Failed to get database connection: %v", err)
		return
	}
	sqlDB.SetMaxIdleConns(constants.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(constants.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(constants.DBConnMaxLifetime)
	applogger.Info("Database connection pool configured")

	flag.Usage = printUsage

	addUserCmd := flag.NewFlagSet("adduser", flag.ExitOnError)
	addUsername := addUserCmd.String("username", "", "Username to add")
	addPassword := addUserCmd.String("password", "", "Password to add")
	addConnectIp := addUserCmd.String("ip", "", "Connect ip")

	listUsersCmd := flag.NewFlagSet("listuser", flag.ExitOnError)

	deleteUserCmd := flag.NewFlagSet("deleteuser", flag.ExitOnError)
	deleteUsername := deleteUserCmd.String("username", "", "Username to delete")

	addIPCmd := flag.NewFlagSet("addip", flag.ExitOnError)
	addIP := addIPCmd.String("ip", "", "Add an IP address to the whitelist")

	delIPCmd := flag.NewFlagSet("delip", flag.ExitOnError)
	listIpCmd := flag.NewFlagSet("listip", flag.ExitOnError)

	socksCmd := flag.NewFlagSet("socks", flag.ExitOnError)
	socksPort := socksCmd.Int("port", 1080, "The port number for the SOCKS5 proxy server")
	socksBindListen := socksCmd.Bool("bind-listen", false, "use connect ip as output ip")

	httpCmd := flag.NewFlagSet("http", flag.ExitOnError)
	httpPort := httpCmd.Int("port", 8080, "The port number for the HTTP proxy server")
	httpBindListen := httpCmd.Bool("bind-listen", false, "use connect ip as output ip")

	bothCmd := flag.NewFlagSet("both", flag.ExitOnError)
	bothSocksPort := bothCmd.Int("socks-port", 1080, "The port number for the SOCKS5 proxy server")
	bothHttpPort := bothCmd.Int("http-port", 8080, "The port number for the HTTP proxy server")
	bothBindListen := bothCmd.Bool("bind-listen", false, "use connect ip as output ip")

	webCmd := flag.NewFlagSet("web", flag.ExitOnError)
	webPort := webCmd.Int("port", 0, "The port number for the web management interface (0 for random port)")

	flag.Parse()

	applogger.Info("Command line arguments: %v", os.Args)
	applogger.Info("Number of arguments: %d", len(os.Args))

	if len(os.Args) == 1 {
		applogger.Info("Starting in default mode (no arguments)")
		applogger.Info("Platform: %s", runtime.GOOS)

		// Default to web mode for portable application
		// On Windows, start system tray application
		// On other platforms, start web server directly
		if runtime.GOOS == "windows" {
			applogger.Info("Windows detected - attempting to start system tray application")

			// Try to start system tray with panic recovery
			trayStarted := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						applogger.Error("System tray panic recovered in main: %v", r)
						trayStarted = false
					}
				}()

				// Attempt to start tray (this blocks if successful)
				tray.Start(db, 0)
				trayStarted = true
			}()

			// If tray failed to start, fallback to web mode
			if !trayStarted {
				applogger.Info("Falling back to web server mode")
				fmt.Println("系统托盘启动失败，切换到Web服务器模式...")
				fmt.Println("System tray failed to start, falling back to web server mode...")

				// Load initial credentials and whitelist
				auth.LoadCredentialsFromDB(db)
				auth.LoadWhitelistFromDB(db)

				// Create and start web manager with random port
				webManager := web.NewManager(db, 0)

				// Auto-start proxies based on saved configuration
				if socksConfig, err := config.LoadProxyConfig(db, "socks5"); err == nil && socksConfig != nil && socksConfig.AutoStart {
					applogger.Info("Auto-starting SOCKS5 proxy on port %d", socksConfig.Port)
					if err := webManager.AutoStartProxy("socks5", socksConfig.Port, socksConfig.BindListen); err != nil {
						applogger.Error("Failed to auto-start SOCKS5 proxy: %v", err)
					}
				}

				if httpConfig, err := config.LoadProxyConfig(db, "http"); err == nil && httpConfig != nil && httpConfig.AutoStart {
					applogger.Info("Auto-starting HTTP proxy on port %d", httpConfig.Port)
					if err := webManager.AutoStartProxy("http", httpConfig.Port, httpConfig.BindListen); err != nil {
						applogger.Error("Failed to auto-start HTTP proxy: %v", err)
					}
				}

				fmt.Println("Starting web management interface on random port...")
				if err := webManager.StartServer(); err != nil {
					applogger.Error("Web server failed: %v", err)
					return
				}
			}
		} else {
			applogger.Info("Non-Windows platform - starting web server directly")
			// Load initial credentials and whitelist
			auth.LoadCredentialsFromDB(db)
			auth.LoadWhitelistFromDB(db)

			// Create and start web manager with random port
			webManager := web.NewManager(db, 0)

			// Auto-start proxies based on saved configuration
			if socksConfig, err := config.LoadProxyConfig(db, "socks5"); err == nil && socksConfig != nil && socksConfig.AutoStart {
				applogger.Info("Auto-starting SOCKS5 proxy on port %d", socksConfig.Port)
				if err := webManager.AutoStartProxy("socks5", socksConfig.Port, socksConfig.BindListen); err != nil {
					applogger.Error("Failed to auto-start SOCKS5 proxy: %v", err)
				}
			}

			if httpConfig, err := config.LoadProxyConfig(db, "http"); err == nil && httpConfig != nil && httpConfig.AutoStart {
				applogger.Info("Auto-starting HTTP proxy on port %d", httpConfig.Port)
				if err := webManager.AutoStartProxy("http", httpConfig.Port, httpConfig.BindListen); err != nil {
					applogger.Error("Failed to auto-start HTTP proxy: %v", err)
				}
			}

			fmt.Println("Starting web management interface on random port...")
			if err := webManager.StartServer(); err != nil {
				applogger.Error("Web server failed: %v", err)
				return
			}
		}
		return
	} else {
		switch os.Args[1] {
		case "addip":
			addIPCmd.Parse(os.Args[2:])
			err := auth.AddIPToWhitelist(db, *addIP)
			if err != nil {
				applogger.Error("Failed to add whiteip: %v", err)
			}
			fmt.Println("Whiteip added successfully!")
			return
		case "delip":
			delIPCmd.Parse(os.Args[2:])
			return
		case "listip":
			listIpCmd.Parse(os.Args[2:])
			return
		case "adduser":
			addUserCmd.Parse(os.Args[2:])
			if *addUsername == "" || *addPassword == "" {
				fmt.Println("Usage: proxy-server adduser -username [username] -password [password]")
				return
			}
			err := auth.AddUser(db, *addConnectIp, *addUsername, *addPassword)
			if err != nil {
				applogger.Error("Failed to add user: %v", err)
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println("User added successfully!")
			return
		case "deluser":
			deleteUserCmd.Parse((os.Args[2:]))
			if *deleteUsername == "" {
				fmt.Println("Usage: proxy-server deluser -username [username]")
				return
			}
			err := auth.DeleteUser(db, *deleteUsername)
			if err != nil {
				applogger.Error("Failed to delete user: %v", err)
				return
			}
			fmt.Println("User deleted successfully!")
			return
		case "listuser":
			listUsersCmd.Parse(os.Args[2:])
			err := auth.ListUsers(db)
			if err != nil {
				applogger.Error("Failed to list users: %v", err)
				return
			}
		case "socks":
			socksCmd.Parse(os.Args[2:])

			// Start configuration reloader
			startConfigReloader(db)

			// Run SOCKS5 proxy server
			if err := runProxyServer("SOCKS5", *socksPort, *socksBindListen, db); err != nil {
				applogger.Error("SOCKS5 proxy server failed: %v", err)
				return
			}
		case "http":
			httpCmd.Parse(os.Args[2:])

			// Start configuration reloader
			startConfigReloader(db)

			// Run HTTP proxy server
			if err := runProxyServer("HTTP", *httpPort, *httpBindListen, db); err != nil {
				applogger.Error("HTTP proxy server failed: %v", err)
				return
			}
		case "both":
			bothCmd.Parse(os.Args[2:])

			// Start configuration reloader (shared by both servers)
			startConfigReloader(db)

			// Channel to receive errors from goroutines
			errChan := make(chan error, 2)
			var socksStarted atomic.Bool

			// Start SOCKS5 server in a goroutine
			go func() {
				socksStarted.Store(true)
				err := runProxyServer("SOCKS5", *bothSocksPort, *bothBindListen, db)
				if err != nil {
					errChan <- fmt.Errorf("SOCKS5: %w", err)
				}
			}()

			// Wait a bit to ensure SOCKS5 started successfully
			time.Sleep(100 * time.Millisecond)
			if !socksStarted.Load() {
				applogger.Error("SOCKS5 proxy failed to start")
				return
			}

			// Start HTTP server in a goroutine
			go func() {
				err := runProxyServer("HTTP", *bothHttpPort, *bothBindListen, db)
				if err != nil {
					errChan <- fmt.Errorf("HTTP: %w", err)
				}
			}()

			// Wait for any server to fail
			err := <-errChan
			applogger.Error("Proxy server failed: %v", err)
			return
		case "web":
			webCmd.Parse(os.Args[2:])

			// Initialize credentials and whitelist
			auth.LoadCredentialsFromDB(db)
			auth.LoadWhitelistFromDB(db)

			// Create web manager
			webManager := web.NewManager(db, *webPort)

			// Auto-start proxies based on saved configuration
			if socksConfig, err := config.LoadProxyConfig(db, "socks5"); err == nil && socksConfig != nil && socksConfig.AutoStart {
				applogger.Info("Auto-starting SOCKS5 proxy on port %d", socksConfig.Port)
				if err := webManager.AutoStartProxy("socks5", socksConfig.Port, socksConfig.BindListen); err != nil {
					applogger.Error("Failed to auto-start SOCKS5 proxy: %v", err)
				}
			}

			if httpConfig, err := config.LoadProxyConfig(db, "http"); err == nil && httpConfig != nil && httpConfig.AutoStart {
				applogger.Info("Auto-starting HTTP proxy on port %d", httpConfig.Port)
				if err := webManager.AutoStartProxy("http", httpConfig.Port, httpConfig.BindListen); err != nil {
					applogger.Error("Failed to auto-start HTTP proxy: %v", err)
				}
			}

			// Start web server
			if err := webManager.StartServer(); err != nil {
				applogger.Error("Web server failed: %v", err)
				return
			}
		default:
			printUsage()
			return
		}
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  adduser -username <username> -password <password>")
	fmt.Println("  deluser -username <username>")
	fmt.Println("  listuser")
	fmt.Println("  addip -ip <ip_to_add>")
	fmt.Println("  socks -port <port_number> [-bind-listen]")
	fmt.Println("  http -port <port_number> [-bind-listen]")
	fmt.Println("  both -socks-port <port_number> -http-port <port_number> [-bind-listen]")
	fmt.Println("  web [-port <port_number>]  (default: 9090)")
}
