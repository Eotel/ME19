package signal

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestHandleSignals(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		HandleSignals(ctx, cancel)
		close(done)
	}()

	cleanupCalled := false
	RegisterCleanupFunc(func() {
		cleanupCalled = true
	})

	time.Sleep(100 * time.Millisecond)

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("プロセスを見つけられませんでした: %v", err)
	}
	err = p.Signal(syscall.SIGINT)
	if err != nil {
		t.Fatalf("シグナルを送信できませんでした: %v", err)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("シグナルハンドリングがタイムアウトしました")
	}

	select {
	case <-ctx.Done():
	default:
		t.Fatal("コンテキストがキャンセルされていません")
	}

	if !cleanupCalled {
		t.Fatal("クリーンアップ関数が呼び出されていません")
	}
}
