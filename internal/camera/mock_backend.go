package camera

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
)

// mockBackend implements the CameraBackend interface for testing
type mockBackend struct {
	isOpen bool
}

// newMockBackend creates a new mock camera backend for testing
func newMockBackend() CameraBackend {
	return &mockBackend{isOpen: false}
}

// Open simulates opening a camera device
func (m *mockBackend) Open(deviceID int) error {
	// If device ID is 99, simulate a non-existent camera
	if deviceID == 99 {
		return errors.New("device not found")
	}
	m.isOpen = true
	return nil
}

// Close simulates closing a camera device
func (m *mockBackend) Close() error {
	if !m.isOpen {
		return errors.New("camera not open")
	}
	m.isOpen = false
	return nil
}

// Read simulates reading a frame from the camera
func (m *mockBackend) Read() ([]byte, error) {
	if !m.isOpen {
		return nil, errors.New("camera not open")
	}

	// Generate a test image (a simple gray rectangle)
	return createTestImage(640, 480)
}

// IsOpened returns whether the mock camera is open
func (m *mockBackend) IsOpened() bool {
	return m.isOpen
}

// createTestImage generates a test image for the mock camera
func createTestImage(width, height int) ([]byte, error) {
	// Create a new grayscale image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill the image with a gradient
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Simple gradient based on position
			gray := uint8((x + y) % 256)
			img.Set(x, y, color.RGBA{gray, gray, gray, 255})
		}
	}

	// Add a QR code-like pattern in the center for testing
	centerX, centerY := width/2, height/2
	patternSize := 100
	for y := centerY - patternSize/2; y < centerY+patternSize/2; y++ {
		for x := centerX - patternSize/2; x < centerX+patternSize/2; x++ {
			if (x/10+y/10)%2 == 0 {
				img.Set(x, y, color.RGBA{0, 0, 0, 255}) // Black
			} else {
				img.Set(x, y, color.RGBA{255, 255, 255, 255}) // White
			}
		}
	}

	// Encode to JPEG
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 75})
	return buf.Bytes(), err
}
