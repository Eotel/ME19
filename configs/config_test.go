package configs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestViperConfig(t *testing.T) {
	testConfig := `{
		"camera": {
			"device_id": 1,
			"width": 1920,
			"height": 1080,
			"fps": 60
		},
		"qrcode": {
			"scan_interval_ms": 200
		},
		"output_file": {
			"file_path": "test_output.txt"
		}
	}`

	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("一時ディレクトリを作成できませんでした: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	err = os.WriteFile(configPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("設定ファイルを書き込めませんでした: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("設定を読み込めませんでした: %v", err)
	}

	if config.Camera.DeviceID != 1 {
		t.Errorf("Camera.DeviceID: expected 1, got %d", config.Camera.DeviceID)
	}
	if config.Camera.Width != 1920 {
		t.Errorf("Camera.Width: expected 1920, got %d", config.Camera.Width)
	}
	if config.Camera.Height != 1080 {
		t.Errorf("Camera.Height: expected 1080, got %d", config.Camera.Height)
	}
	if config.Camera.FPS != 60 {
		t.Errorf("Camera.FPS: expected 60, got %d", config.Camera.FPS)
	}
	if config.QRCode.ScanInterval != 200 {
		t.Errorf("QRCode.ScanInterval: expected 200, got %d", config.QRCode.ScanInterval)
	}
	if config.OutputFile.FilePath != "test_output.txt" {
		t.Errorf("OutputFile.FilePath: expected test_output.txt, got %s", config.OutputFile.FilePath)
	}
}

func TestEnvOverride(t *testing.T) {
	os.Setenv("ME19_CAMERA_DEVICE_ID", "2")
	os.Setenv("ME19_OUTPUT_FILE_PATH", "env_output.txt")
	defer os.Unsetenv("ME19_CAMERA_DEVICE_ID")
	defer os.Unsetenv("ME19_OUTPUT_FILE_PATH")

	config := DefaultConfig()

	LoadEnvironmentVariables(&config)

	if config.Camera.DeviceID != 2 {
		t.Errorf("Camera.DeviceID: expected 2, got %d", config.Camera.DeviceID)
	}
	if config.OutputFile.FilePath != "env_output.txt" {
		t.Errorf("OutputFile.FilePath: expected env_output.txt, got %s", config.OutputFile.FilePath)
	}
}

func TestConfigNotFound(t *testing.T) {
	config, err := LoadConfig("/path/to/nonexistent/config.json")

	if err == nil {
		t.Error("存在しないファイルでエラーが発生しませんでした")
	}

	defaultConfig := DefaultConfig()
	if config.Camera.DeviceID != defaultConfig.Camera.DeviceID {
		t.Errorf("デフォルト設定が返されていません")
	}
}

func TestFindConfigFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "config-finder-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a configs subdirectory
	configsDir := filepath.Join(tempDir, "configs")
	err = os.Mkdir(configsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create configs directory: %v", err)
	}

	// Create a test config file
	testConfigPath := filepath.Join(configsDir, "config.json")
	err = os.WriteFile(testConfigPath, []byte("{}"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Change to the temp directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}
	defer os.Chdir(originalWd) // Restore original working directory when done

	// Test finding the config file
	foundPath := FindConfigFile()
	expectedPath := filepath.Join("configs", "config.json")

	if foundPath != expectedPath && foundPath != filepath.Join(".", "configs", "config.json") {
		t.Errorf("FindConfigFile() returned %s, want %s", foundPath, expectedPath)
	}

	// Test with no config file
	err = os.Remove(testConfigPath)
	if err != nil {
		t.Fatalf("Failed to remove test config file: %v", err)
	}

	foundPath = FindConfigFile()
	if foundPath != "" {
		t.Errorf("FindConfigFile() returned %s, want empty string when no config exists", foundPath)
	}
}
