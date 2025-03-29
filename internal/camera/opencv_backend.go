package camera

import (
	"bytes"
	"errors"
	"image/jpeg"

	"gocv.io/x/gocv"
)

// opencvBackend implements the CameraBackend interface using OpenCV
type opencvBackend struct {
	camera *gocv.VideoCapture
	isOpen bool
}

// newOpenCVBackend creates a new OpenCV-based camera backend
func newOpenCVBackend() CameraBackend {
	return &opencvBackend{}
}

// Open initializes the camera with OpenCV
func (o *opencvBackend) Open(deviceID int) error {
	if o.isOpen {
		return errors.New("camera is already open")
	}

	camera, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		return err
	}

	o.camera = camera
	o.isOpen = true
	return nil
}

// Close releases OpenCV camera resources
func (o *opencvBackend) Close() error {
	if !o.isOpen || o.camera == nil {
		return errors.New("camera not open")
	}

	err := o.camera.Close()
	if err != nil {
		return err
	}

	o.camera = nil
	o.isOpen = false
	return nil
}

// Read captures a frame from the OpenCV camera
func (o *opencvBackend) Read() ([]byte, error) {
	if !o.isOpen || o.camera == nil {
		return nil, errors.New("camera not open")
	}

	// Capture a frame
	img := gocv.NewMat()
	defer img.Close()

	if ok := o.camera.Read(&img); !ok {
		return nil, errors.New("could not read from camera")
	}

	if img.Empty() {
		return nil, errors.New("captured frame is empty")
	}

	// Convert to JPEG for easier handling
	rgbImg, err := img.ToImage()
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, rgbImg, &jpeg.Options{Quality: 75})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// IsOpened returns whether the OpenCV camera is open
func (o *opencvBackend) IsOpened() bool {
	return o.isOpen && o.camera != nil
}
