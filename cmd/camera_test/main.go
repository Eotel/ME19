package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gocv.io/x/gocv"
)

func main() {
	fmt.Println("ME19 カメラテスト - 動作確認中...")
	fmt.Println("Ctrl+Cで終了")

	// カメラを開く
	webcam, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		fmt.Printf("エラー: カメラを開けませんでした: %v\n", err)
		os.Exit(1)
	}
	defer webcam.Close()

	// ウィンドウを作成
	window := gocv.NewWindow("ME19 カメラテスト")
	defer window.Close()

	// 画像バッファを作成
	img := gocv.NewMat()
	defer img.Close()

	// 終了シグナル処理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 一定時間後に自動終了（安全対策）
	timer := time.NewTimer(20 * time.Second)

	// フレーム数カウンター
	frameCount := 0
	
	// メインループ
	for {
		select {
		case <-sigCh:
			fmt.Println("\nシグナルを受信しました。終了します。")
			return
		case <-timer.C:
			fmt.Println("\n20秒経過しました。自動的に終了します。")
			fmt.Printf("取得したフレーム数: %d\n", frameCount)
			return
		default:
			// フレーム取得
			if ok := webcam.Read(&img); !ok {
				fmt.Println("フレーム取得エラー")
				continue
			}
			
			if img.Empty() {
				fmt.Println("空のフレームを受信しました")
				continue
			}

			// フレームを表示
			window.IMShow(img)
			
			// フレーム数をカウント
			frameCount++
			if frameCount % 30 == 0 {
				fmt.Printf("フレーム数: %d\n", frameCount)
			}

			// キー入力確認（1ミリ秒待機）
			if window.WaitKey(1) >= 0 {
				fmt.Println("\nキー入力を検出しました。終了します。")
				fmt.Printf("取得したフレーム数: %d\n", frameCount)
				return
			}
		}
	}
}
