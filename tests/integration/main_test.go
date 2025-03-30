package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eotel/me19/configs"
	"github.com/eotel/me19/internal/camera"
	"github.com/eotel/me19/internal/fileio"
	"github.com/eotel/me19/internal/qrcode"
)

func TestEndToEndFlow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "me19-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputFilePath := filepath.Join(tmpDir, "qrcode_output.txt")

	cam := camera.NewWithTestBackend()
	detector := qrcode.New()
	writer := fileio.New(outputFilePath)

	err = cam.Open()
	if err != nil {
		t.Fatalf("Failed to open camera: %v", err)
	}
	defer cam.Close()

	err = detector.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize QR code detector: %v", err)
	}
	defer detector.Close()

	frame, err := cam.CaptureFrame()
	if err != nil {
		t.Fatalf("Failed to capture frame: %v", err)
	}

	results, err := detector.Detect(frame)
	if err != nil {
		t.Fatalf("Failed to detect QR codes: %v", err)
	}

	if len(results) == 0 {
		t.Log("No QR codes detected in test frame, this is expected with mock data")
	} else {
		t.Logf("Detected %d QR codes", len(results))
	}

	testData := "TEST QR CODE DATA"
	err = writer.WriteData(testData)
	if err != nil {
		t.Fatalf("Failed to write data to file: %v", err)
	}

	content, err := os.ReadFile(outputFilePath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if string(content) != testData {
		t.Errorf("File content mismatch. Expected: %s, Got: %s", testData, string(content))
	}
}

func TestConfigIntegration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "me19-config-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	testConfig := `{
		"camera": {
			"device_id": 2,
			"width": 1280,
			"height": 720,
			"fps": 30
		},
		"qrcode": {
			"scan_interval_ms": 500
		},
		"output_file": {
			"file_path": "integration_test_output.txt"
		}
	}`

	err = os.WriteFile(configPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	config, err := configs.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cam := camera.NewWithTestBackend()
	cam.SetDeviceID(config.Camera.DeviceID)

	if config.Camera.DeviceID != 2 {
		t.Errorf("Camera device ID not set correctly. Expected: 2, Got: %d", config.Camera.DeviceID)
	}

	if config.QRCode.ScanInterval != 500 {
		t.Errorf("QR code scan interval not set correctly. Expected: 500, Got: %d", config.QRCode.ScanInterval)
	}

	if config.OutputFile.FilePath != "integration_test_output.txt" {
		t.Errorf("Output file path not set correctly. Expected: integration_test_output.txt, Got: %s", config.OutputFile.FilePath)
	}
}

func TestConcurrentOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "me19-concurrent-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputFilePath := filepath.Join(tmpDir, "concurrent_output.txt")
	writer := fileio.New(outputFilePath)

	cam := camera.NewWithTestBackend()
	detector := qrcode.New()

	err = cam.Open()
	if err != nil {
		t.Fatalf("Failed to open camera: %v", err)
	}
	defer cam.Close()

	err = detector.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize QR code detector: %v", err)
	}
	defer detector.Close()

	done := make(chan bool)
	errCh := make(chan error, 3)

	go func() {
		for i := 0; i < 5; i++ {
			_, err := cam.CaptureFrame()
			if err != nil {
				errCh <- err
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 5; i++ {
			err := writer.AppendData("Concurrent test data " + time.Now().String() + "\n")
			if err != nil {
				errCh <- err
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
		done <- true
	}()

	for i := 0; i < 2; i++ {
		select {
		case err := <-errCh:
			t.Fatalf("Error in concurrent operation: %v", err)
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent operations timed out")
		}
	}

	content, err := os.ReadFile(outputFilePath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Output file is empty, expected content from concurrent operations")
	}
}
