package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
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

	if os.Getenv("ME19_TEST_MODE") == "true" {
		cam = camera.NewWithTestBackend()
		log.Println("Using mock camera backend (test mode enabled via environment variable)")
	} else {
		cam = camera.New()
		log.Println("Using real camera backend")
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

	var headless bool

	if runtime.GOOS == "darwin" {
		headless = os.Getenv("CGO_ENABLED") == "0"
	} else {
		headless = os.Getenv("DISPLAY") == ""
	}

	if headless {
		log.Println("Running in headless mode - camera preview window disabled")
	} else {
		log.Printf("Running with display enabled on %s platform", runtime.GOOS)
	}

	var window *gocv.Window
	if !headless {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Warning: Failed to create window: %v", r)
				log.Println("Continuing in headless mode")
				headless = true
			}
		}()

		window = gocv.NewWindow("ME19 QR Code Scanner")
		defer window.Close()
	}

	currentDeviceID := config.Camera.DeviceID

	switchCamera := func(newDeviceID int) {
		if newDeviceID == currentDeviceID {
			return
		}

		log.Printf("Switching to camera device ID: %d", newDeviceID)

		if cam.IsOpen() {
			if err := cam.Close(); err != nil {
				log.Printf("Warning: Error closing camera: %v", err)
			}
		}

		var newCam *camera.Camera
		if os.Getenv("ME19_TEST_MODE") == "true" {
			newCam = camera.NewWithTestBackend()
		} else {
			newCam = camera.New()
		}
		newCam.SetDeviceID(newDeviceID)

		err := newCam.Open()
		if err != nil {
			log.Printf("Error opening camera with device ID %d: %v", newDeviceID, err)
			if os.Getenv("ME19_TEST_MODE") == "true" {
				cam = camera.NewWithTestBackend()
			} else {
				cam = camera.New()
			}
			cam.SetDeviceID(currentDeviceID)
			err = cam.Open()
			if err != nil {
				log.Printf("Error reopening original camera: %v", err)
			}
			return
		}

		cam = newCam
		currentDeviceID = newDeviceID
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

				if !headless && window != nil {
					func() {
						defer func() {
							if r := recover(); r != nil {
								log.Printf("Warning: Error in frame processing: %v", r)
								if runtime.GOOS == "darwin" {
									log.Println("Switching to headless mode due to display error on macOS")
									headless = true
								}
							}
						}()

						frame, err := gocv.IMDecode(frameBytes, gocv.IMReadColor)
						if err != nil {
							log.Printf("Error decoding frame: %v", err)
							return
						}
						defer frame.Close()

						window.IMShow(frame)

						key := window.WaitKey(1)
						if key >= 48 && key <= 57 { // 0-9のキー
							newDeviceID := key - 48 // ASCII値から数値に変換
							go switchCamera(newDeviceID)
						}
					}()
				}
			}
		}
	}()

	<-ctx.Done()
}
