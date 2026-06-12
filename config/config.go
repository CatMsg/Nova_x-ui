// Package config provides configuration management utilities for the Nova_x-ui panel,
// including version information, logging levels, database paths, and environment variable handling.
package config

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed version
var version string

//go:embed name
var name string

// LogLevel represents the logging level for the application.
type LogLevel string

// Logging level constants
const (
	Debug   LogLevel = "debug"
	Info    LogLevel = "info"
	Notice  LogLevel = "notice"
	Warning LogLevel = "warning"
	Error   LogLevel = "error"
)

// GetVersion returns the version string of the Nova_x-ui application.
func GetVersion() string {
	return strings.TrimSpace(version)
}

// GetName returns the name of the Nova_x-ui application.
func GetName() string {
	return strings.TrimSpace(name)
}

// GetLogLevel returns the current logging level based on environment variables or defaults to Info.
func GetLogLevel() LogLevel {
	if IsDebug() {
		return Debug
	}
	logLevel := os.Getenv("XUI_LOG_LEVEL")
	if logLevel == "" {
		return Info
	}
	return LogLevel(logLevel)
}

// IsDebug returns true if debug mode is enabled via the XUI_DEBUG environment variable.
func IsDebug() bool {
	return os.Getenv("XUI_DEBUG") == "true"
}

// GetBinFolderPath returns the path to the binary folder.
// On macOS and Windows it defaults to a repo-local bin/ directory; on Linux it stays relative.
func GetBinFolderPath() string {
	binFolderPath := os.Getenv("XUI_BIN_FOLDER")
	if binFolderPath == "" {
		if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
			binFolderPath = filepath.Join(getBaseDir(), "bin")
		} else {
			binFolderPath = "bin"
		}
	}
	return binFolderPath
}

func getBaseDir() string {
	exePath, err := os.Executable()
	if err != nil {
		if wd, wdErr := os.Getwd(); wdErr == nil {
			return wd
		}
		return "."
	}
	exeDir := filepath.Dir(exePath)
	exeDirLower := strings.ToLower(filepath.ToSlash(exeDir))
	if strings.Contains(exeDirLower, "/appdata/local/temp/") ||
		strings.Contains(exeDirLower, "/private/var/folders/") ||
		strings.Contains(exeDirLower, "/var/folders/") ||
		strings.Contains(exeDirLower, "/go-build") ||
		strings.Contains(exeDirLower, "/tmp/") {
		wd, err := os.Getwd()
		if err != nil {
			return "."
		}
		return wd
	}
	return exeDir
}

// GetDBFolderPath returns the path to the database folder based on environment variables or platform defaults.
func GetDBFolderPath() string {
	dbFolderPath := os.Getenv("XUI_DB_FOLDER")
	if dbFolderPath != "" {
		return dbFolderPath
	}
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return filepath.Join(getBaseDir(), "db")
	}
	return "/etc/Nova_x-ui"
}

// GetDBPath returns the full path to the database file.
func GetDBPath() string {
	return fmt.Sprintf("%s/%s.db", GetDBFolderPath(), GetName())
}

// GetLogFolder returns the path to the log folder based on environment variables or platform defaults.
func GetLogFolder() string {
	logFolderPath := os.Getenv("XUI_LOG_FOLDER")
	if logFolderPath != "" {
		return logFolderPath
	}
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return filepath.Join(getBaseDir(), "log")
	}
	return "/var/log/Nova_x-ui"
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Sync()
}

func init() {
	if runtime.GOOS != "windows" {
		return
	}
	if os.Getenv("XUI_DB_FOLDER") != "" {
		return
	}
	oldDBFolder := "/etc/Nova_x-ui"
	oldDBPath := fmt.Sprintf("%s/%s.db", oldDBFolder, GetName())
	newDBFolder := GetDBFolderPath()
	newDBPath := fmt.Sprintf("%s/%s.db", newDBFolder, GetName())
	_, err := os.Stat(newDBPath)
	if err == nil {
		return // new exists
	}
	_, err = os.Stat(oldDBPath)
	if os.IsNotExist(err) {
		return // old does not exist
	}
	_ = copyFile(oldDBPath, newDBPath) // ignore error
}
