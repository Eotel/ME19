package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/eotel/me19/configs"
	"github.com/eotel/me19/internal/camera"
	"github.com/eotel/me19/internal/fileio"
	"github.com/eotel/me19/internal/qrcode"
	"github.com/eotel/me19/internal/signal"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "../../configs/config.json", "Path to configuration file")
	flag.Parse()

	fmt.Println("ME19 QR Code Scanner")
	fmt.Println("Press Ctrl+C to exit")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	config, err := configs.LoadConfig(*configPath)
	if err != nil {
		log.Printf("Warning: Failed to load config file: %v. Using default configuration.", err)
		config = configs.DefaultConfig()
	}

	configs.LoadEnvironmentVariables(&config)

	// Initialize components
	cam := camera.New()
	detector := qrcode.New()
	_ = fileio.New(config.OutputFile.FilePath) // Will be used in future implementation

	go signal.HandleSignals(ctx, cancel)

	signal.RegisterCleanupFunc(func() {
		fmt.Println("\nShutting down...")
		if cam != nil {
			cam.Close()
		}
		if detector != nil {
			detector.Close()
		}
	})

	// Placeholder for camera initialization
	fmt.Println("Camera initialization would happen here")
	// In the future, this would be: cam.Open()

	// Placeholder for QR code detection
	fmt.Println("QR code detection would happen here")
	// In the future, this would use detector.Detect() and writer.WriteData()

	<-ctx.Done()
}
