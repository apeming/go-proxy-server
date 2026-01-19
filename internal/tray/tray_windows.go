//go:build windows
// +build windows

package tray

import (
	_ "embed"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/getlantern/systray"
	"gorm.io/gorm"

	"go-proxy-server/internal/auth"
	"go-proxy-server/internal/config"
	"go-proxy-server/internal/logger"
	"go-proxy-server/internal/web"
)

//go:embed icon.ico
var iconData []byte

var globalDB *gorm.DB
var globalWebManager *web.Manager
var actualWebPort int // Store actual port after binding

// Start starts the system tray application
func Start(db *gorm.DB, webPort int) {
	globalDB = db
	logger.Info("Starting system tray application...")
	systray.Run(onReady(webPort), onExit)
}

func onReady(webPort int) func() {
	return func() {
		logger.Info("Initializing system tray...")

		// Set icon (using a simple icon data)
		systray.SetIcon(getIcon())
		systray.SetTitle("Go Proxy Server")
		systray.SetTooltip("Go Proxy Server - 代理服务器管理")

		logger.Info("Tray icon initialized")

		// Load initial credentials and whitelist
		auth.LoadCredentialsFromDB(globalDB)
		auth.LoadWhitelistFromDB(globalDB)

		logger.Info("Starting web server on port %d...", webPort)

		// Start web server in background
		globalWebManager = web.NewManager(globalDB, webPort)

		// Auto-start proxies based on saved configuration
		if socksConfig, err := config.LoadProxyConfig(globalDB, "socks5"); err == nil && socksConfig != nil && socksConfig.AutoStart {
			logger.Info("Auto-starting SOCKS5 proxy on port %d", socksConfig.Port)
			if err := globalWebManager.AutoStartProxy("socks5", socksConfig.Port, socksConfig.BindListen); err != nil {
				logger.Error("Failed to auto-start SOCKS5 proxy: %v", err)
			}
		}

		if httpConfig, err := config.LoadProxyConfig(globalDB, "http"); err == nil && httpConfig != nil && httpConfig.AutoStart {
			logger.Info("Auto-starting HTTP proxy on port %d", httpConfig.Port)
			if err := globalWebManager.AutoStartProxy("http", httpConfig.Port, httpConfig.BindListen); err != nil {
				logger.Error("Failed to auto-start HTTP proxy: %v", err)
			}
		}

		go func() {
			logger.Info("Web management interface starting on port %d (0 = random)", webPort)
			if err := globalWebManager.StartServer(); err != nil {
				logger.Error("Web server failed: %v", err)
			}
		}()

			// Wait for server to bind to port (StartServer sets actualPort before blocking on http.Serve)
			time.Sleep(200 * time.Millisecond)
			actualWebPort = globalWebManager.GetActualPort()
			logger.Info("Web management interface started on http://localhost:%d", actualWebPort)

		logger.Info("Adding tray menu items...")

		// Add menu items
		mOpen := systray.AddMenuItem("打开管理界面", "在浏览器中打开管理界面")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("退出", "退出程序")

		logger.Info("System tray application ready!")

		// Handle menu clicks
		go func() {
			for {
				select {
				case <-mOpen.ClickedCh:
					logger.Info("Opening browser...")
					// Use actual port
					port := actualWebPort
					if port == 0 {
						port = webPort
					}
					openBrowser(fmt.Sprintf("http://localhost:%d", port))
				case <-mQuit.ClickedCh:
					logger.Info("Quit requested by user")
					systray.Quit()
					return
				}
			}
		}()
	}
}

func onExit() {
	// Cleanup code here
	logger.Info("System tray application exiting...")
}

// openBrowser opens the default browser with the given URL
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default: // "linux", "freebsd", "openbsd", "netbsd"
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		logger.Error("Failed to open browser: %v", err)
	} else {
		logger.Info("Browser opened successfully")
	}
}

// getIcon returns the embedded icon data (ICO format)
// Icon: Blue circular network/proxy theme with transparent background
// Embedded sizes: 256x256, 128x128, 96x96, 64x64, 48x48, 32x32, 16x16
func getIcon() []byte {
	return iconData
}
