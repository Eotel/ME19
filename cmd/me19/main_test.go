package main

import (
	"os"
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
