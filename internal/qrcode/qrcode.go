package qrcode

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// Detector is responsible for detecting and decoding QR codes from images
type Detector struct {
	// 設定オプション
	IsInitialized bool
	reader        gozxing.Reader
}

// New creates a new QR code detector
func New() *Detector {
	return &Detector{
		IsInitialized: false,
		reader:        nil,
	}
}

// Initialize sets up the QR code detector
func (d *Detector) Initialize() error {
	// QRコードリーダーのインスタンスを作成
	d.reader = qrcode.NewQRCodeReader()
	d.IsInitialized = true
	return nil
}

// Detect finds and decodes QR codes in the provided image data
func (d *Detector) Detect(imageData []byte) ([]string, error) {
	if !d.IsInitialized {
		return nil, errors.New("QR code detector is not initialized")
	}

	// バイトデータから画像を解析
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, errors.New("failed to decode image data: " + err.Error())
	}

	// gozxingのBinaryBitmapに変換
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, errors.New("failed to convert image to bitmap: " + err.Error())
	}

	// QRコードの検出と読み取り
	result, err := d.reader.Decode(bmp, nil)
	if err != nil {
		// QRコードが検出されなかった場合は空のリストを返す（エラーではない）
		return []string{}, nil
	}

	// 検出結果をリストに追加
	results := []string{result.GetText()}

	return results, nil
}

// DetectMultiple finds and decodes multiple QR codes in the provided image data
func (d *Detector) DetectMultiple(imageData []byte) ([]string, error) {
	if !d.IsInitialized {
		return nil, errors.New("QR code detector is not initialized")
	}

	// バイトデータから画像を解析
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, errors.New("failed to decode image data: " + err.Error())
	}

	// gozxingのBinaryBitmapに変換
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, errors.New("failed to convert image to bitmap: " + err.Error())
	}

	// 複数のQRコードを検出するためのマルチフォーマットリーダーを使用
	multiFormatReader := qrcode.NewQRCodeReader()
	// 結果を格納するスライス
	var results []string

	// 画像全体をスキャン（簡易的な実装 - 実際の複数QRコード検出にはもっと複雑なアルゴリズムが必要かもしれません）
	// 注：gozxingは直接複数のQRコードをサポートしていないため、このメソッドは完全な実装ではありません
	result, err := multiFormatReader.Decode(bmp, nil)
	if err == nil {
		// 少なくとも1つのQRコードが見つかった場合
		results = append(results, result.GetText())
	}

	return results, nil
}

// Close releases resources used by the detector
func (d *Detector) Close() error {
	// 特に解放する必要のあるリソースがない場合は、簡単な状態リセットのみを行う
	d.IsInitialized = false
	d.reader = nil
	return nil
}

// BytesToImage converts image byte data to an image.Image interface
func BytesToImage(data []byte) (image.Image, error) {
	// 画像フォーマットを自動判別して読み込み
	reader := bytes.NewReader(data)

	// まずPNGとして読み込みを試みる
	img, err := png.Decode(reader)
	if err == nil {
		return img, nil
	}

	// PNGでないならJPEGとしての読み込みを試みる
	reader.Seek(0, 0) // リーダーをリセット
	img, err = jpeg.Decode(reader)
	if err == nil {
		return img, nil
	}

	// サポートされていない画像フォーマット
	return nil, errors.New("unsupported image format")
}
