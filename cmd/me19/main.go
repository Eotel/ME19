package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/eotel/me19/configs"
	"github.com/eotel/me19/internal/camera"
	"github.com/eotel/me19/internal/fileio"
	"github.com/eotel/me19/internal/qrcode"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "../../configs/config.json", "Path to configuration file")
	flag.Parse()

	fmt.Println("ME19 QR Code Scanner")
	fmt.Println("Press Ctrl+C to exit")

	// Load configuration
	config, err := configs.LoadConfig(*configPath)
	if err != nil {
		log.Printf("Warning: Failed to load config file: %v. Using default configuration.", err)
		config = configs.DefaultConfig()
	}

	// Initialize components
	cam := camera.New()
	detector := qrcode.New()
	writer := fileio.New(config.OutputFile.FilePath)

	// Avoid unused variable warnings during development
	_ = cam
	_ = detector
	_ = writer

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Placeholder for camera initialization
	fmt.Println("Camera initialization would happen here")
	// In the future, this would be: cam.Open()

	// Placeholder for QR code detection
	fmt.Println("QR code detection would happen here")
	// In the future, this would use detector.Detect() and writer.WriteData()

	// Wait for termination signal
	<-sigCh
	fmt.Println("\nShutting down...")

	// Cleanup resources
	// In the future: cam.Close(), detector.Close(), etc.
}
