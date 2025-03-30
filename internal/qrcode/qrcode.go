package qrcode

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	_ "image/jpeg" // Register JPEG format
	"image/png"
	_ "image/png" // Register PNG format

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// Detector is responsible for detecting and decoding QR codes from images
type Detector struct {
	// IsInitialized indicates whether the detector has been properly initialized
	IsInitialized bool
	qrReader      gozxing.Reader
}

// New creates a new QR code detector
func New() *Detector {
	return &Detector{
		IsInitialized: false,
		qrReader:      nil,
	}
}

// Initialize sets up the QR code detector
func (d *Detector) Initialize() error {
	// QRコードリーダーのインスタンスを作成
	d.qrReader = qrcode.NewQRCodeReader()
	d.IsInitialized = true
	return nil
}

// Detect finds and decodes a single QR code in the provided image data
func (d *Detector) Detect(imageData []byte) ([]string, error) {
	if !d.IsInitialized {
		return nil, errors.New("QR code detector is not initialized")
	}

	// バイトデータから画像を解析
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, err
	}

	// gozxingのBinaryBitmapに変換
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, err
	}

	// QRコードの検出と読み取り
	result, err := d.qrReader.Decode(bmp, nil)
	if err != nil {
		// QRコードが検出されなかった場合は空のリストを返す（エラーではない）
		return []string{}, nil
	}

	// 検出結果をリストに追加
	return []string{result.GetText()}, nil
}

// DetectMultiple finds and decodes multiple QR codes in the provided image data
func (d *Detector) DetectMultiple(imageData []byte) ([]string, error) {
	if !d.IsInitialized {
		return nil, errors.New("QR code detector is not initialized")
	}

	// 単一QRコード検出の結果を配列に格納して返す簡易実装
	// 将来的に複数QRコード検出に拡張予定
	// 現在はDetect()メソッドの結果をそのまま返す
	return d.Detect(imageData)
}

// Close releases resources used by the detector
func (d *Detector) Close() error {
	// このシンプルな実装では特別なリソース解放は必要ないが、
	// 将来的な拡張性のためにメソッドを提供しています
	d.qrReader = nil
	d.IsInitialized = false
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
