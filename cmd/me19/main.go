package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eotel/me19/configs"
	"github.com/eotel/me19/internal/camera"
	"github.com/eotel/me19/internal/fileio"
	"github.com/eotel/me19/internal/qrcode"
	"github.com/eotel/me19/internal/signal"
	"gocv.io/x/gocv"
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
	var cam *camera.Camera
	if _, err := os.Stat("/dev/video0"); os.IsNotExist(err) {
		cam = camera.NewWithTestBackend()
		log.Println("Using mock camera backend for testing")
	} else {
		cam = camera.New()
	}
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

	err = cam.Open()
	if err != nil {
		log.Fatalf("Error opening camera: %v", err)
	}

	headless := os.Getenv("DISPLAY") == ""
	if headless {
		log.Println("Running in headless mode - camera preview window disabled")
	}

	var window *gocv.Window
	if !headless {
		window = gocv.NewWindow("ME19 QR Code Scanner")
		defer window.Close()
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				frameBytes, err := cam.CaptureFrame()
				if err != nil {
					log.Printf("Error capturing frame: %v", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}

				if !headless {
					frame, err := gocv.IMDecode(frameBytes, gocv.IMReadColor)
					if err != nil {
						log.Printf("Error decoding frame: %v", err)
						continue
					}

					window.IMShow(frame)
					window.WaitKey(1)
					frame.Close()
				}

			}
		}
	}()

	<-ctx.Done()
}
