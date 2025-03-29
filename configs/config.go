package configs

// Config holds the application configuration
type Config struct {
	Camera     CameraConfig     `json:"camera"`
	QRCode     QRCodeConfig     `json:"qrcode"`
	OutputFile OutputFileConfig `json:"output_file"`
}

// CameraConfig holds camera-related configuration
type CameraConfig struct {
	DeviceID int `json:"device_id"`
	Width    int `json:"width"`
	Height   int `json:"height"`
	FPS      int `json:"fps"`
}

// QRCodeConfig holds QR code detection configuration
type QRCodeConfig struct {
	ScanInterval int `json:"scan_interval_ms"` // Interval between scans in milliseconds
}

// OutputFileConfig holds file output configuration
type OutputFileConfig struct {
	FilePath string `json:"file_path"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Camera: CameraConfig{
			DeviceID: 0,
			Width:    1280,
			Height:   720,
			FPS:      30,
		},
		QRCode: QRCodeConfig{
			ScanInterval: 500,
		},
		OutputFile: OutputFileConfig{
			FilePath: "code.txt",
		},
	}
}
