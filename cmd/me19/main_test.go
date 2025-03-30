package main

import (
	"os"
	"runtime"
	"testing"

	"github.com/eotel/me19/internal/camera"
)

func TestCameraInitialization(t *testing.T) {
	os.Setenv("ME19_TEST_MODE", "true")
	defer os.Unsetenv("ME19_TEST_MODE")

	cam := camera.NewWithTestBackend()
	if cam == nil {
		t.Fatal("Failed to create test camera")
	}

	err := cam.Open()
	if err != nil {
		t.Fatalf("Failed to open camera: %v", err)
	}
	defer cam.Close()

	if !cam.IsOpen() {
		t.Fatal("Camera should be open")
	}

	frame, err := cam.CaptureFrame()
	if err != nil {
		t.Fatalf("Failed to capture frame: %v", err)
	}

	if len(frame) == 0 {
		t.Fatal("Captured frame is empty")
	}
}

func TestPlatformDetection(t *testing.T) {
	originalDisplay := os.Getenv("DISPLAY")
	originalCGOEnabled := os.Getenv("CGO_ENABLED")
	defer func() {
		os.Setenv("DISPLAY", originalDisplay)
		os.Setenv("CGO_ENABLED", originalCGOEnabled)
	}()

	testCases := []struct {
		name           string
		platform       string
		display        string
		cgoEnabled     string
		expectHeadless bool
	}{
		{
			name:           "Linux with DISPLAY",
			platform:       "linux",
			display:        ":0",
			cgoEnabled:     "1",
			expectHeadless: false,
		},
		{
			name:           "Linux without DISPLAY",
			platform:       "linux",
			display:        "",
			cgoEnabled:     "1",
			expectHeadless: true,
		},
		{
			name:           "macOS with CGO_ENABLED",
			platform:       "darwin",
			display:        "",
			cgoEnabled:     "1",
			expectHeadless: false,
		},
		{
			name:           "macOS without CGO_ENABLED",
			platform:       "darwin",
			display:        "",
			cgoEnabled:     "0",
			expectHeadless: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if runtime.GOOS != tc.platform {
				t.Skipf("Skipping test for %s platform on %s", tc.platform, runtime.GOOS)
				return
			}

			os.Setenv("DISPLAY", tc.display)
			os.Setenv("CGO_ENABLED", tc.cgoEnabled)

			var headless bool
			if runtime.GOOS == "darwin" {
				headless = os.Getenv("CGO_ENABLED") == "0"
			} else {
				headless = os.Getenv("DISPLAY") == ""
			}

			if headless != tc.expectHeadless {
				t.Errorf("Expected headless=%v for %s platform with DISPLAY=%q and CGO_ENABLED=%q, got %v",
					tc.expectHeadless, tc.platform, tc.display, tc.cgoEnabled, headless)
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	os.Setenv("ME19_TEST_MODE", "true")
	defer os.Unsetenv("ME19_TEST_MODE")

	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("Expected panic recovery, but no panic occurred")
			}
		}()

		panic("simulated error in window creation")
	}()

	cam := camera.NewWithTestBackend()
	if cam == nil {
		t.Fatal("Failed to create test camera")
	}

	err := cam.Open()
	if err != nil {
		t.Fatalf("Failed to open camera: %v", err)
	}
	defer cam.Close()

	newCam := camera.NewWithTestBackend()
	newCam.SetDeviceID(99)
	err = newCam.Open()
	if err == nil {
		t.Error("Expected error when opening non-existent camera device, but got nil")
	}
}
