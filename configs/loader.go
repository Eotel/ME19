package configs

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadConfig loads the configuration from the specified JSON file
func LoadConfig(filePath string) (Config, error) {
	config := DefaultConfig()

	file, err := os.Open(filePath)
	if err != nil {
		return config, fmt.Errorf("opening config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return config, fmt.Errorf("decoding config file: %w", err)
	}

	return config, nil
}
