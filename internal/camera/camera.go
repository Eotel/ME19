package camera

import (
	"errors"
	"gocv.io/x/gocv"
)

// Device defines the interface for camera operations
type Device interface {
	// Open initializes the camera device
	Open() error
	// Close releases resources used by the camera
	Close() error
	// CaptureFrame captures a single frame from the camera
	CaptureFrame() ([]byte, error)
	// SetDeviceID sets the camera device ID to use
	SetDeviceID(id int)
}

type Camera struct {
	deviceID int           // ID of the camera device to use (typically 0 for the first camera)
	isOpen   bool          // Flag indicating if the camera is currently open
	backend  CameraBackend // The implementation that handles actual camera operations
}

// CameraBackend defines the interface for actual camera operations
// This allows us to swap implementations for testing
type CameraBackend interface {
	Open(deviceID int) error
	Close() error
	Read() ([]byte, error)
	IsOpened() bool
}

// DefaultBackend returns the appropriate camera backend based on environment
func DefaultBackend() CameraBackend {
	// u30c6u30b9u30c8u74b0u5883u306eu5834u5408u306fu30e2u30c3u30afu30d0u30c3u30afu30a8u30f3u30c9u3092u4f7fu7528
	if forTesting {
		return newMockBackend()
	}

	// u30d7u30edu30c0u30afu30b7u30e7u30f3u74b0u5883u3067u306fOpenCVu30d0u30c3u30afu30a8u30f3u30c9u3092u4f7fu7528
	return newOpenCVBackend()
}

// New creates a new Camera instance with default values
func New() *Camera {
	return &Camera{
		deviceID: 0, // Default to first camera (usually built-in webcam)
		isOpen:   false,
		backend:  DefaultBackend(),
	}
}

// NewWithBackend creates a new Camera with a specific backend implementation
func NewWithBackend(backend CameraBackend) *Camera {
	return &Camera{
		deviceID: 0,
		isOpen:   false,
		backend:  backend,
	}
}

// CaptureFrameMat は直接Mat形式でフレームを取得します
func (c *Camera) CaptureFrameMat() (gocv.Mat, error) {
	if !c.isOpen {
		return gocv.NewMat(), errors.New("camera is not open")
	}

	// OpenCVバックエンドへの直接アクセス
	if backend, ok := c.backend.(*opencvBackend); ok {
		mat := gocv.NewMat()
		if !backend.camera.Read(&mat) || mat.Empty() {
			mat.Close()
			return gocv.NewMat(), errors.New("failed to read frame")
		}
		return mat, nil
	}

	// 他のバックエンド（モックなど）の場合は従来の方法
	frameBytes, err := c.backend.Read()
	if err != nil {
		return gocv.NewMat(), err
	}

	mat, err := gocv.IMDecode(frameBytes, gocv.IMReadColor)
	if err != nil {
		return gocv.NewMat(), err
	}

	return mat, nil
}

// SetDeviceID sets the camera device ID to use
func (c *Camera) SetDeviceID(id int) {
	c.deviceID = id
}

// Open initializes the camera
func (c *Camera) Open() error {
	if c.isOpen {
		return errors.New("camera is already open")
	}

	// Initialize the camera using the backend
	err := c.backend.Open(c.deviceID)
	if err != nil {
		return err
	}

	c.isOpen = true
	return nil
}

// Close releases the camera resources
func (c *Camera) Close() error {
	if !c.isOpen {
		return errors.New("camera is not open")
	}

	err := c.backend.Close()
	if err != nil {
		return err
	}

	c.isOpen = false
	return nil
}

// CaptureFrame captures a single frame from the camera
func (c *Camera) CaptureFrame() ([]byte, error) {
	if !c.isOpen {
		return nil, errors.New("camera is not open")
	}

	return c.backend.Read()
}

func (c *Camera) IsOpen() bool {
	return c.isOpen
}

// GetDeviceID returns the current camera device ID
func (c *Camera) GetDeviceID() int {
	return c.deviceID
}
