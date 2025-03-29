package qrcode

import (
	"os"
	"path/filepath"
	"testing"
)

// テスト用のQRコード画像を準備する関数
func getTestImagesPaths() (string, string) {
	// テストデータディレクトリのパス
	testDataDir := "testdata"

	// テスト用の単一QRコード画像のパス
	singleQRPath := filepath.Join(testDataDir, "test_qr.png")

	// テスト用の別のQRコード画像のパス
	helloQRPath := filepath.Join(testDataDir, "hello_qr.png")

	return singleQRPath, helloQRPath
}

// テスト用の画像ファイルを読み込む関数
func loadTestImage(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// TestNew は New 関数のテスト
func TestNew(t *testing.T) {
	detector := New()
	if detector == nil {
		t.Fatal("Failed to create a new detector")
	}
}

// TestDetector_Initialize は Initialize メソッドのテスト
func TestDetector_Initialize(t *testing.T) {
	detector := New()
	err := detector.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize detector: %v", err)
	}
}

// TestDetector_Detect は Detect メソッドのテスト
func TestDetector_Detect(t *testing.T) {
	// テスト用のQRコード画像を準備
	testImagePath, _ := getTestImagesPaths()

	// 画像ファイルの読み込み
	imageData, err := loadTestImage(testImagePath)
	if err != nil {
		t.Fatalf("Failed to load test image: %v", err)
	}

	// QRコードデテクターの初期化
	detector := New()
	err = detector.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize detector: %v", err)
	}

	// QRコード検出の実行
	results, err := detector.Detect(imageData)
	if err != nil {
		t.Fatalf("Detection failed: %v", err)
	}

	// 検出結果の検証
	if len(results) == 0 {
		t.Fatal("Expected to detect QR code, but none found")
	}

	// 検出されたQRコードの内容を確認
	expectedContent := "TEST QR CODE"
	if results[0] != expectedContent {
		t.Fatalf("Expected QR code content '%s', but got '%s'", expectedContent, results[0])
	}

	// テスト後のクリーンアップ
	err = detector.Close()
	if err != nil {
		t.Fatalf("Failed to close detector: %v", err)
	}
}

// TestDetector_Close は Close メソッドのテスト
func TestDetector_Close(t *testing.T) {
	detector := New()
	err := detector.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize detector: %v", err)
	}

	err = detector.Close()
	if err != nil {
		t.Fatalf("Failed to close detector: %v", err)
	}

	// 閉じた後は初期化されていない状態になることを確認
	if detector.IsInitialized {
		t.Error("Detector should not be initialized after Close()")
	}
}

// TestDetector_DetectMultiple は DetectMultiple メソッドのテスト
func TestDetector_DetectMultiple(t *testing.T) {
	// テスト用のQRコード画像を準備
	_, testImagePath := getTestImagesPaths()

	// 画像ファイルの読み込み
	imageData, err := loadTestImage(testImagePath)
	if err != nil {
		t.Fatalf("Failed to load test image: %v", err)
	}

	// QRコードデテクターの初期化
	detector := New()
	err = detector.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize detector: %v", err)
	}

	// 複数QRコード検出の実行
	results, err := detector.DetectMultiple(imageData)
	if err != nil {
		t.Fatalf("Multiple detection failed: %v", err)
	}

	// 少なくとも1つのQRコードが検出されることを確認
	if len(results) == 0 {
		t.Fatal("Expected to detect at least one QR code, but none found")
	}

	// 検出されたQRコードの内容を確認
	expectedContent := "HELLO WORLD"
	if results[0] != expectedContent {
		t.Fatalf("Expected QR code content '%s', but got '%s'", expectedContent, results[0])
	}

	// テスト後のクリーンアップ
	err = detector.Close()
	if err != nil {
		t.Fatalf("Failed to close detector: %v", err)
	}
}
