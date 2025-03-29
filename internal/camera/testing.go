package camera

import (
	"os"
	"testing"
)

// forTesting は現在テスト中かどうかを判断する
var forTesting bool

func init() {
	// testing.TB が使用可能かどうかでテスト実行中か判断する
	// テスト実行中は特殊な名前の実行ファイルになる
	forTesting = isTestBinary()
}

// isTestBinary は実行バイナリがテスト用かどうかをチェックする
func isTestBinary() bool {
	return len(os.Args) > 0 && (testing.Testing() || isTestArg())
}

// isTestArg はコマンドライン引数からテスト実行を検出する
func isTestArg() bool {
	for _, arg := range os.Args {
		if arg == "-test.v" || arg == "-test.run" {
			return true
		}
	}
	return false
}

// NewWithTestBackend はテスト用にモックバックエンドを使用したカメラインスタンスを作成する
func NewWithTestBackend() *Camera {
	return NewWithBackend(newMockBackend())
}
