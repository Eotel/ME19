package qrcode

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// テスト用変数
var forTesting bool = false

// EnableTestMode テストモードを有効にする
func EnableTestMode() {
	forTesting = true
}

// DisableTestMode テストモードを無効にする
func DisableTestMode() {
	forTesting = false
}

// IsTestMode テストモードかどうかを返す
func IsTestMode() bool {
	return forTesting
}

// GenerateTestImage テスト用の単純な画像を生成する
// 注: これは実際にはQRコードではなく、テスト用の簡易画像です
func GenerateTestImage(width, height int) ([]byte, error) {
	// 新しい画像を作成
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 単純なパターンで塗りつぶす
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if (x+y)%2 == 0 {
				img.Set(x, y, color.Black)
			} else {
				img.Set(x, y, color.White)
			}
		}
	}

	// PNG形式にエンコード
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
