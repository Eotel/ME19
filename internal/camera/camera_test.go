package camera

import (
	"testing"
)

func TestNew(t *testing.T) {
	// テスト用のモックバックエンドを使ったカメラインスタンスを作成
	cam := NewWithTestBackend()
	if cam == nil {
		t.Fatal("NewWithTestBackend() returned nil")
	}

	// デフォルトではカメラは初期化されていないはず
	// Close()を呼び出して間接的にテスト - "not open" エラーが発生するはず
	err := cam.Close()
	if err == nil || err.Error() != "camera is not open" {
		t.Error("New camera should not be initialized by default")
	}

	// deviceIDはプライベートフィールドなので直接テストできない
	// 代わりに設定後の振る舞いでテスト
}

func TestOpen(t *testing.T) {
	// テスト用のモックバックエンドを使ったカメラインスタンスを作成
	cam := NewWithTestBackend()

	// カメラが正常に開くことをテスト
	err := cam.Open()
	if err != nil {
		t.Fatalf("Failed to open camera: %v", err)
	}

	// カメラが開いていることを確認するため、再度開こうとする - "already open" エラーが発生するはず
	err = cam.Open()
	if err == nil || err.Error() != "camera is already open" {
		t.Error("Camera should be marked as open after initialization")
	}

	// クリーンアップ
	err = cam.Close()
	if err != nil {
		t.Fatalf("Failed to close camera: %v", err)
	}
}

func TestClose(t *testing.T) {
	cam := NewWithTestBackend()

	// 最初にカメラを開く
	err := cam.Open()
	if err != nil {
		t.Fatalf("Failed to open camera: %v", err)
	}

	// 正常に閉じることをテスト
	err = cam.Close()
	if err != nil {
		t.Fatalf("Failed to close camera: %v", err)
	}

	// カメラが閉じていることを確認するため、再度閉じようとする - "not open" エラーが発生するはず
	err = cam.Close()
	if err == nil || err.Error() != "camera is not open" {
		t.Error("Camera should be marked as closed after Close()")
	}
}

func TestCaptureFrame(t *testing.T) {
	cam := NewWithTestBackend()

	// 最初にカメラを開く
	err := cam.Open()
	if err != nil {
		t.Fatalf("Failed to open camera: %v", err)
	}

	// フレームを正常にキャプチャできることをテスト
	frame, err := cam.CaptureFrame()
	if err != nil {
		t.Fatalf("Failed to capture frame: %v", err)
	}

	// フレームがnilでなく、内容があることを確認
	if frame == nil {
		t.Fatal("Captured frame is nil")
	}
	if len(frame) == 0 {
		t.Error("Captured frame is empty")
	}

	// クリーンアップ
	err = cam.Close()
	if err != nil {
		t.Fatalf("Failed to close camera: %v", err)
	}
}

func TestSetDeviceID(t *testing.T) {
	// deviceIDはプライベートフィールドなので直接テストできない
	// 代わりに非デフォルト値を設定した際の振る舞いをテスト

	// テスト用のモックバックエンドを使用
	cam := NewWithTestBackend()

	// 存在しないカメラID（モックでは99で失敗するように実装）
	constantID := 99
	cam.SetDeviceID(constantID)

	// 存在しないカメラを開こうとするとエラーになるはず
	// これにより間接的にdeviceIDが変更されたことをテスト
	err := cam.Open()
	if err == nil {
		// 予期せず成功した場合はクリーンアップ
		cam.Close()
		t.Error("Expected error when opening non-existent camera ID 99")
	}
	// エラーが発生すれば期待通り、deviceIDが設定されたと判断できる
}
