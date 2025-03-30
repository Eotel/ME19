package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func LoadConfig(filePath string) (Config, error) {
	config := DefaultConfig()

	// First check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return config, fmt.Errorf("config file not found: %s", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return config, fmt.Errorf("reading config file: %w", err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("unmarshaling config: %w", err)
	}

	return config, nil
}

func LoadEnvironmentVariables(config *Config) {
	v := viper.New()
	v.SetEnvPrefix("ME19")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if v.IsSet("CAMERA_DEVICE_ID") {
		config.Camera.DeviceID = v.GetInt("CAMERA_DEVICE_ID")
	}
	if v.IsSet("CAMERA_WIDTH") {
		config.Camera.Width = v.GetInt("CAMERA_WIDTH")
	}
	if v.IsSet("CAMERA_HEIGHT") {
		config.Camera.Height = v.GetInt("CAMERA_HEIGHT")
	}
	if v.IsSet("CAMERA_FPS") {
		config.Camera.FPS = v.GetInt("CAMERA_FPS")
	}

	if v.IsSet("QRCODE_SCAN_INTERVAL_MS") {
		config.QRCode.ScanInterval = v.GetInt("QRCODE_SCAN_INTERVAL_MS")
	}

	if v.IsSet("OUTPUT_FILE_PATH") {
		config.OutputFile.FilePath = v.GetString("OUTPUT_FILE_PATH")
	}
}
