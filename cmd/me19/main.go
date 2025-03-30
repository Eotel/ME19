package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/eotel/me19/configs"
	"github.com/eotel/me19/internal/camera"
	"github.com/eotel/me19/internal/fileio"
	"github.com/eotel/me19/internal/qrcode"
	"gocv.io/x/gocv"
)

func init() {
	// Lock the main thread for proper macOS UI handling
	runtime.LockOSThread()
}

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	deviceID := flag.Int("device", -1, "Camera device ID")
	outputFile := flag.String("output", "", "Path to output file")

	// 短縮形のフラグも追加
	flag.StringVar(configPath, "c", "", "Path to configuration file (shorthand)")
	flag.IntVar(deviceID, "d", -1, "Camera device ID (shorthand)")
	flag.StringVar(outputFile, "o", "", "Path to output file (shorthand)")
	flag.Parse()

	fmt.Println("ME19 QR Code Scanner")
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println("Press keys 0-9 to switch camera")

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	setupSignalHandler(cancel)

	// Load configuration
	var config configs.Config
	var err error

	if *configPath != "" {
		// User specified a config file, try to load it
		config, err = configs.LoadConfig(*configPath)
		if err != nil {
			log.Printf("Warning: Failed to load specified config file '%s': %v", *configPath, err)
			log.Println("Using default configuration.")
			config = configs.DefaultConfig()
		} else {
			log.Printf("Using configuration from specified file: %s", *configPath)
		}
	} else {
		// Try to find a config file in standard locations
		foundConfigPath := configs.FindConfigFile()
		if foundConfigPath != "" {
			config, err = configs.LoadConfig(foundConfigPath)
			if err != nil {
				log.Printf("Warning: Found config file at %s but failed to load it: %v", foundConfigPath, err)
				log.Println("Using default configuration.")
				config = configs.DefaultConfig()
			} else {
				log.Printf("Using configuration from: %s", foundConfigPath)
			}
		} else {
			// No config file found, use default without error message
			log.Println("No configuration file found. Using default configuration.")
			config = configs.DefaultConfig()
		}
	}

	// Override config with command line arguments if provided
	if *deviceID >= 0 {
		log.Printf("Overriding camera device ID from command line: %d", *deviceID)
		config.Camera.DeviceID = *deviceID
	}

	if *outputFile != "" {
		log.Printf("Overriding output file path from command line: %s", *outputFile)
		config.OutputFile.FilePath = *outputFile
	}

	// Load environment variables (which override both config file and command line)
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
	defer cam.Close()

	// カメラのデバイスIDを設定
	cam.SetDeviceID(config.Camera.DeviceID)

	// QRコード検出器を作成
	detector := qrcode.New()
	if err := detector.Initialize(); err != nil {
		log.Fatalf("Failed to initialize QR code detector: %v", err)
	}
	defer detector.Close()

	// ファイル書き込みオブジェクトを作成して参照を保持
	writer := fileio.New(config.OutputFile.FilePath)
	log.Printf("QR code data will be written to: %s", config.OutputFile.FilePath)

	// ヘッドレスモードで実行するかどうかを確認
	var headless bool
	if runtime.GOOS == "darwin" {
		headless = os.Getenv("CGO_ENABLED") == "0"
	} else {
		headless = os.Getenv("DISPLAY") == ""
	}

	if headless {
		log.Println("Running in headless mode - camera preview window disabled")
		runHeadless(ctx, cam, detector, writer)
	} else {
		log.Printf("Running with display enabled on %s platform", runtime.GOOS)
		runWithDisplay(ctx, cam, detector, writer)
	}
}

// setupSignalHandler creates a signal handler for graceful shutdown
func setupSignalHandler(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nShutting down...")
		cancel()
	}()
}

// runHeadless runs the application without UI
func runHeadless(ctx context.Context, cam *camera.Camera, detector *qrcode.Detector, writer *fileio.Writer) {
	// Open the camera
	if err := cam.Open(); err != nil {
		log.Fatalf("Error opening camera: %v", err)
	}

	lastCode := "" // 前回検出したQRコードを追跡

	// Just keep the camera running
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
			// フレームを取得
			frame, err := cam.CaptureFrame()
			if err != nil {
				log.Printf("Error capturing frame: %v", err)
				continue
			}

			// QRコードを検出
			codes, err := detector.Detect(frame)
			if err != nil {
				log.Printf("Error detecting QR codes: %v", err)
				continue
			}

			// 検出したQRコードをファイルに書き込み
			processDetectedCodes(codes, writer, &lastCode)
		}
	}
}

// processDetectedCodes processes detected QR codes and writes new codes to file
func processDetectedCodes(codes []string, writer *fileio.Writer, lastCode *string) {
	for _, code := range codes {
		// 新しいコードの場合のみファイルに書き込み
		if code != *lastCode && code != "" {
			if err := writer.WriteData(code); err != nil {
				log.Printf("Error writing QR code data to file: %v", err)
			} else {
				log.Printf("Detected new QR code and wrote to file: %s", code)
				*lastCode = code
			}
		}
	}
}

// tryOpenCamera attempts to open the camera with the specified device ID
// Returns true if successful, false otherwise
func tryOpenCamera(cam *camera.Camera, deviceID int) bool {
	// First close the current camera if it's open
	if cam.IsOpen() {
		if err := cam.Close(); err != nil {
			log.Printf("Error closing current camera: %v", err)
			return false
		}
	}

	// Set new device ID
	cam.SetDeviceID(deviceID)

	// Try to open with new device ID
	err := cam.Open()
	if err != nil {
		log.Printf("Failed to open camera with device ID %d: %v", deviceID, err)
		return false
	}

	return true
}

// runWithDisplay runs the application with UI
func runWithDisplay(ctx context.Context, cam *camera.Camera, detector *qrcode.Detector, writer *fileio.Writer) {
	// Open the camera
	if err := cam.Open(); err != nil {
		log.Fatalf("Error opening camera: %v", err)
	}
	lastCode := "" // 前回検出したQRコードを追跡

	// Create window on the main thread
	window := gocv.NewWindow("ME19 QR Code Scanner")
	window.SetWindowProperty(gocv.WindowPropertyAutosize, gocv.WindowAutosize)
	defer window.Close()

	// Current device ID
	currentDeviceID := cam.GetDeviceID()
	log.Printf("Initial camera device ID: %d", currentDeviceID)

	// Frame counter for periodic logging
	frameCount := 0

	fmt.Println("Window is open. Click on the window and press keys 0-9 to switch cameras")

	// Main display loop
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Capture frame directly as Mat for display
			mat, err := cam.CaptureFrameMat()
			if err != nil {
				log.Printf("Error capturing frame: %v", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if mat.Empty() {
				log.Println("Empty frame received")
				mat.Close()
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Matをバッファに変換してからQRコード検出に使用
			// 画像をエンコードして一時ファイルに保存
			tempFile := os.TempDir() + "/me19_frame.jpg"
			success := gocv.IMWrite(tempFile, mat)
			if !success {
				log.Printf("Error saving image to temp file")
			} else {
				// 一時ファイルからバイトデータを読み込み
				frameBytes, err := os.ReadFile(tempFile)
				if err != nil {
					log.Printf("Error reading frame from temp file: %v", err)
				} else {
					// QRコードを検出
					codes, err := detector.Detect(frameBytes)
					if err != nil {
						// エラーが頻発するため、ログ出力は抑制
						// log.Printf("Error detecting QR codes: %v", err)
					} else if len(codes) > 0 {
						// 検出したQRコードを処理
						for _, code := range codes {
							if code != lastCode && code != "" {
								if err := writer.WriteData(code); err != nil {
									log.Printf("Error writing QR code data to file: %v", err)
								} else {
									log.Printf("Detected new QR code and wrote to file: %s", code)
									lastCode = code

									// QRコードが検出されたことを表示
									gocv.PutText(&mat, fmt.Sprintf("QR: %s", code), image.Point{X: 10, Y: 60}, gocv.FontHersheyPlain, 1.2, color.RGBA{R: 255, G: 0, B: 0, A: 255}, 2)
								}
							}
						}
					}

					// 一時ファイルを削除（エラー処理は省略）
					os.Remove(tempFile)
				}
			}

			// Draw current device ID text on the frame
			gocv.PutText(&mat,
				fmt.Sprintf("Device ID: %d (Press 0-9 to switch)", currentDeviceID),
				image.Point{X: 10, Y: 30},
				gocv.FontHersheyPlain, 1.2,
				color.RGBA{0, 255, 0, 255}, 2)

			// Show the image in the window
			window.IMShow(mat)
			mat.Close()

			key := window.WaitKey(1)

			// Log key presses for debugging
			if key >= 0 {
				log.Printf("Key pressed: %d", key)
			}

			// Log frame info occasionally
			frameCount++
			if frameCount%100 == 0 {
				log.Printf("Processed %d frames, current device: %d", frameCount, currentDeviceID)
			}

			// Handle numeric key presses (both standard and numpad)
			// ASCII: 0-9 are 48-57, numpad 0-9 are typically 96-105 on some systems
			if (key >= 48 && key <= 57) || (key >= 96 && key <= 105) {
				var newDeviceID int
				if key >= 96 && key <= 105 {
					newDeviceID = key - 96 // Convert numpad keys
				} else {
					newDeviceID = key - 48 // Convert standard number keys
				}

				log.Printf("Key %d pressed - attempting to switch to camera device ID: %d", key, newDeviceID)

				if newDeviceID != currentDeviceID {
					log.Printf("Switching from device ID %d to %d", currentDeviceID, newDeviceID)

					if tryOpenCamera(cam, newDeviceID) {
						log.Printf("Successfully switched to camera device ID: %d", newDeviceID)
						currentDeviceID = newDeviceID
					} else {
						log.Printf("Failed to switch camera. Reopening original camera (device ID: %d)", currentDeviceID)
						if tryOpenCamera(cam, currentDeviceID) {
							log.Println("Successfully reopened original camera")
						} else {
							log.Println("Failed to reopen original camera. Exiting.")
							return
						}
					}
				}
			}
		}
	}
}
