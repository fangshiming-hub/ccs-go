package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetHomeDir returns the user's home directory
func GetHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return home, nil
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// FormatFileSize formats bytes into a human-readable string
func FormatFileSize(bytes int64) string {
	if bytes == 0 {
		return "0 Bytes"
	}

	const unit = 1024
	sizes := []string{"Bytes", "KB", "MB", "GB"}

	i := 0
	fb := float64(bytes)
	for fb >= unit && i < len(sizes)-1 {
		fb /= unit
		i++
	}

	return fmt.Sprintf("%.2f %s", fb, sizes[i])
}

// GetConfigDir returns the default config directory path
func GetConfigDir() (string, error) {
	home, err := GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude-config-switch"), nil
}

// GetConfigFilePath returns the default config file path
func GetConfigFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "claudeEnvConfig.json"), nil
}

// GetSettingsFilePath returns the target settings.json path
func GetSettingsFilePath() (string, error) {
	home, err := GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
}