package fileio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriter_WriteData(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "fileio_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// テスト用のファイルパスを作成
	testFilePath := filepath.Join(tmpDir, "qrcode.txt")

	// テストケース
	tests := []struct {
		name     string
		data     string
		expected string
		wantErr  bool
	}{
		{
			name:     "write simple data",
			data:     "test qr code data",
			expected: "test qr code data",
			wantErr:  false,
		},
		{
			name:     "write empty data",
			data:     "",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "write multiline data",
			data:     "line1\nline2\nline3",
			expected: "line1\nline2\nline3",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用のライターを作成
			writer := New(testFilePath)

			// データを書き込む
			err := writer.WriteData(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// ファイルの内容を読み取って検証
			content, err := os.ReadFile(testFilePath)
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			if string(content) != tt.expected {
				t.Errorf("WriteData() wrote %q, want %q", string(content), tt.expected)
			}
		})
	}
}

func TestWriter_AppendData(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "fileio_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// テスト用のファイルパスを作成
	testFilePath := filepath.Join(tmpDir, "qrcode.txt")

	// テストケース
	tests := []struct {
		name           string
		initialContent string
		dataToAppend   string
		expected       string
		wantErr        bool
	}{
		{
			name:           "append to empty file",
			initialContent: "",
			dataToAppend:   "appended data",
			expected:       "appended data",
			wantErr:        false,
		},
		{
			name:           "append to existing content",
			initialContent: "initial data\n",
			dataToAppend:   "appended data",
			expected:       "initial data\nappended data",
			wantErr:        false,
		},
		{
			name:           "append empty data",
			initialContent: "existing data",
			dataToAppend:   "",
			expected:       "existing data",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 初期データが存在する場合は先に書き込む
			if tt.initialContent != "" {
				err := os.WriteFile(testFilePath, []byte(tt.initialContent), 0644)
				if err != nil {
					t.Fatalf("Failed to write initial content: %v", err)
				}
			}

			// テスト用のライターを作成
			writer := New(testFilePath)

			// データを追加書き込みする
			err := writer.AppendData(tt.dataToAppend)
			if (err != nil) != tt.wantErr {
				t.Errorf("AppendData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// ファイルの内容を読み取って検証
			content, err := os.ReadFile(testFilePath)
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			if string(content) != tt.expected {
				t.Errorf("AppendData() wrote %q, want %q", string(content), tt.expected)
			}
		})
	}
}
