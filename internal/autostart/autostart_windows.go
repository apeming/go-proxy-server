//go:build windows
// +build windows

package autostart

import (
	"fmt"
	"os"
	"path/filepath"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

const (
	appName = "GoProxyServer.lnk"
)

// getStartupFolder returns the Windows Startup folder path
func getStartupFolder() (string, error) {
	// Use APPDATA environment variable
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("APPDATA environment variable not set")
	}
	return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup"), nil
}

// getShortcutPath returns the full path to the shortcut file
func getShortcutPath() (string, error) {
	startupFolder, err := getStartupFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(startupFolder, appName), nil
}

// IsEnabled checks if autostart is enabled
func IsEnabled() (bool, error) {
	shortcutPath, err := getShortcutPath()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(shortcutPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Enable enables autostart by creating a shortcut in the Startup folder
// Uses pure Go implementation via go-ole library to avoid VBScript execution
// This significantly reduces antivirus false positives
func Enable() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Resolve symlinks
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %v", err)
	}

	shortcutPath, err := getShortcutPath()
	if err != nil {
		return err
	}

	// Create shortcut using COM interface (pure Go, no VBScript)
	err = createShortcut(exePath, shortcutPath, filepath.Dir(exePath))
	if err != nil {
		return fmt.Errorf("failed to create shortcut: %v", err)
	}

	return nil
}

// Disable disables autostart by removing the shortcut
func Disable() error {
	shortcutPath, err := getShortcutPath()
	if err != nil {
		return err
	}

	err = os.Remove(shortcutPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Already disabled
		}
		return fmt.Errorf("failed to remove shortcut: %v", err)
	}

	return nil
}

// createShortcut creates a Windows shortcut (.lnk) file using COM interface
// This is a pure Go implementation that doesn't require VBScript or external tools
func createShortcut(targetPath, shortcutPath, workingDir string) error {
	// Initialize COM
	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	if err != nil {
		return fmt.Errorf("failed to initialize COM: %v", err)
	}
	defer ole.CoUninitialize()

	// Create WScript.Shell object
	oleShellObject, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return fmt.Errorf("failed to create WScript.Shell object: %v", err)
	}
	defer oleShellObject.Release()

	// Get IDispatch interface
	wshell, err := oleShellObject.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("failed to query IDispatch interface: %v", err)
	}
	defer wshell.Release()

	// Call CreateShortcut method
	cs, err := oleutil.CallMethod(wshell, "CreateShortcut", shortcutPath)
	if err != nil {
		return fmt.Errorf("failed to call CreateShortcut: %v", err)
	}

	// Get shortcut IDispatch
	idispatch := cs.ToIDispatch()
	defer idispatch.Release()

	// Set shortcut properties
	_, err = oleutil.PutProperty(idispatch, "TargetPath", targetPath)
	if err != nil {
		return fmt.Errorf("failed to set TargetPath: %v", err)
	}

	_, err = oleutil.PutProperty(idispatch, "WorkingDirectory", workingDir)
	if err != nil {
		return fmt.Errorf("failed to set WorkingDirectory: %v", err)
	}

	_, err = oleutil.PutProperty(idispatch, "Description", "Go Proxy Server")
	if err != nil {
		return fmt.Errorf("failed to set Description: %v", err)
	}

	// Save the shortcut
	_, err = oleutil.CallMethod(idispatch, "Save")
	if err != nil {
		return fmt.Errorf("failed to save shortcut: %v", err)
	}

	return nil
}
