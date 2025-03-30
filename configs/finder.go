package configs

import (
	"os"
	"path/filepath"
)

// FindConfigFile searches for a config file in standard locations
// Returns the path to the first config file found, or an empty string if none found
func FindConfigFile() string {
	// Check in several standard locations
	standardPaths := []string{
		"config.json",               // Current directory
		"configs/config.json",       // configs subdirectory
		"./configs/config.json",     // Explicit current directory
		"../configs/config.json",    // One level up
		"../../configs/config.json", // Two levels up
		"/etc/me19/config.json",     // System-wide config (Linux/macOS)
		filepath.Join(os.ExpandEnv("$HOME"), ".config/me19/config.json"), // User config
	}

	// Get executable directory path
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		// Add paths relative to the executable
		standardPaths = append(standardPaths,
			filepath.Join(exeDir, "config.json"),
			filepath.Join(exeDir, "configs", "config.json"),
			filepath.Join(exeDir, "..", "configs", "config.json"),
		)
	}

	for _, path := range standardPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// No config file found, return empty string
	return ""
}
