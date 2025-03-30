package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/eotel/me19/configs"
	"github.com/eotel/me19/internal/camera"
	"github.com/eotel/me19/internal/fileio"
	"github.com/eotel/me19/internal/qrcode"
	"gocv.io/x/gocv"
)

// QRCodeResult は検出されたQRコードの結果を表す構造体
type QRCodeResult struct {
	Code string
	Time time.Time
}

// FrameData は処理のためのフレームデータを表す構造体
type FrameData struct {
	Mat  gocv.Mat
	Time time.Time
}

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
	// 検出器を初期化
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

	// QRコード検出結果を共有するためのチャネル
	resultChan := make(chan QRCodeResult, 10)

	// QRコード検出用のゴルーチンを起動
	frameChannel := make(chan gocv.Mat, 5)
	go detectQRCodesFromFrames(ctx, detector, frameChannel, resultChan)

	// 最後に検出したQRコード
	var lastCode string

	// フレーム取得と結果処理ループ
	for {
		select {
		case <-ctx.Done():
			// すべての送信済みMatを閉じる
			close(frameChannel)
			return

		case result := <-resultChan:
			// 新しいコードであれば記録
			if result.Code != lastCode && result.Code != "" {
				if err := writer.WriteData(result.Code); err != nil {
					log.Printf("Error writing QR code data to file: %v", err)
				} else {
					log.Printf("Detected new QR code and wrote to file: %s", result.Code)
					lastCode = result.Code
				}
			}

		default:
			// メインスレッドでフレームを取得
			mat, err := cam.CaptureFrameMat()
			if err != nil || mat.Empty() {
				if mat.Ptr() != nil {
					mat.Close()
				}
				time.Sleep(10 * time.Millisecond)
				continue
			}

			// フレームを検出チャネルに送信（コピーを作成）
			clone := mat.Clone()
			select {
			case frameChannel <- clone:
				// フレームが正常に送信された
			default:
				// チャネルがいっぱいの場合はフレームを破棄
				clone.Close()
			}

			// 元のMatを閉じる
			mat.Close()

			time.Sleep(10 * time.Millisecond)
		}
	}
}

// detectQRCodesFromFrames はMatチャネルからQRコードを検出する
func detectQRCodesFromFrames(ctx context.Context, detector *qrcode.Detector, frameChan <-chan gocv.Mat, resultChan chan<- QRCodeResult) {
	for {
		select {
		case <-ctx.Done():
			return

		case mat, ok := <-frameChan:
			if !ok {
				// チャネルが閉じられた
				return
			}

			// MatからQRコードを検出
			codes, err := detectQRCodesFromMat(mat, detector)

			// 使用済みのMatは必ず閉じる
			mat.Close()

			if err != nil {
				continue
			}

			// 検出されたQRコードを結果チャネルに送信
			for _, code := range codes {
				if code != "" {
					resultChan <- QRCodeResult{
						Code: code,
						Time: time.Now(),
					}
				}
			}
		}
	}
}

// detectQRCodesFromMat はMatからQRコードを検出する
func detectQRCodesFromMat(mat gocv.Mat, detector *qrcode.Detector) ([]string, error) {
	// MatをImageに変換
	img, err := mat.ToImage()
	if err != nil {
		return nil, err
	}

	// ImageをJPEGにエンコード
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 75}); err != nil {
		return nil, err
	}

	// JPEGバイトデータからQRコードを検出
	return detector.Detect(buf.Bytes())
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

	// 検出されたQRコードの結果を受け取るチャネル
	resultChan := make(chan QRCodeResult, 10)

	// フレーム処理チャネル
	frameChannel := make(chan gocv.Mat, 5)

	// QRコード検出用のゴルーチンを起動
	go detectQRCodesFromFrames(ctx, detector, frameChannel, resultChan)

	// 現在のQRコード情報を保持する
	type displayInfo struct {
		code string
		time time.Time
		mu   sync.RWMutex
	}

	currentQRCode := &displayInfo{}

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

	// 最後に書き込んだコード
	var lastWrittenCode string

	// Main display loop
	for {
		select {
		case <-ctx.Done():
			// すべての送信済みMatを閉じる
			close(frameChannel)
			return

		case result := <-resultChan:
			// 新しいコードであれば記録
			if result.Code != lastWrittenCode && result.Code != "" {
				if err := writer.WriteData(result.Code); err != nil {
					log.Printf("Error writing QR code data to file: %v", err)
				} else {
					log.Printf("Detected new QR code and wrote to file: %s", result.Code)
					lastWrittenCode = result.Code

					// 表示用の情報を更新
					currentQRCode.mu.Lock()
					currentQRCode.code = result.Code
					currentQRCode.time = result.Time
					currentQRCode.mu.Unlock()
				}
			}

		default:
			// Capture frame directly as Mat for display
			mat, err := cam.CaptureFrameMat()
			if err != nil {
				log.Printf("Error capturing frame: %v", err)
				time.Sleep(10 * time.Millisecond)
				continue
			}

			if mat.Empty() {
				log.Println("Empty frame received")
				mat.Close()
				time.Sleep(10 * time.Millisecond)
				continue
			}

			// QRコード検出用にMatのコピーを作成
			clone := mat.Clone()
			select {
			case frameChannel <- clone:
				// フレームが正常に送信された
			default:
				// チャネルがいっぱいの場合はフレームを破棄
				clone.Close()
			}

			// Draw current device ID text on the frame
			gocv.PutText(&mat,
				fmt.Sprintf("Device ID: %d (Press 0-9 to switch)", currentDeviceID),
				image.Point{X: 10, Y: 30},
				gocv.FontHersheyPlain, 1.2,
				color.RGBA{0, 255, 0, 255}, 2)

			// 検出されたQRコードがあれば表示
			currentQRCode.mu.RLock()
			if currentQRCode.code != "" {
				// QRコードの検出から一定時間以内なら表示
				if time.Since(currentQRCode.time) < 3*time.Second {
					gocv.PutText(&mat,
						fmt.Sprintf("QR: %s", currentQRCode.code),
						image.Point{X: 10, Y: 60},
						gocv.FontHersheyPlain, 1.2,
						color.RGBA{R: 255, G: 0, B: 0, A: 255}, 2)
				}
			}
			currentQRCode.mu.RUnlock()

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

			// フレームレート調整
			time.Sleep(10 * time.Millisecond)
		}
	}
}
