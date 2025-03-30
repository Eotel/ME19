# ME19 ローカル環境での動作確認とビルド手順

このドキュメントでは、ME19アプリケーションをローカル環境でビルド、実行、テストする方法について説明します。

## 1. 環境セットアップの確認

### 前提条件の確認

以下のコマンドを実行して、必要なツールとライブラリがインストールされていることを確認します：

```bash
# Goのバージョン確認
go version

# OpenCVのバージョン確認
pkg-config --modversion opencv4

# GoCVのバージョン確認
cd $GOPATH/pkg/mod/gocv.io/x/gocv@v0.41.0
make version
```

### miseの設定確認

miseを使用している場合は、以下のコマンドで設定を確認します：

```bash
# miseのアクティベーション確認
which mise

# Goのバージョン確認（mise経由）
mise exec go -- version
```

## 2. ビルド手順

### 依存関係のインストール

```bash
# リポジトリのクローン
git clone https://github.com/sports-time-machine/ME19.git
cd ME19

# 依存関係のダウンロード
go mod download
```

### 開発用ビルド

```bash
# 開発用ビルド（デバッグ情報付き）
go build -gcflags=all="-N -l" -o me19_debug ./cmd/me19
```

### リリース用ビルド

```bash
# リリース用ビルド（最適化あり）
go build -ldflags="-s -w" -o me19 ./cmd/me19
```

### クロスプラットフォームビルド

提供されているビルドスクリプトを使用して、複数のプラットフォーム向けにビルドできます：

```bash
# ビルドスクリプトを実行可能にする
chmod +x ./scripts/build.sh

# スクリプトを実行
./scripts/build.sh
```

ビルドされたバイナリは`./build`ディレクトリに保存されます。

## 3. 実行方法

### 基本的な実行

```bash
# デフォルト設定で実行
./me19

# または、ビルドせずに直接実行
go run ./cmd/me19
```

### カスタム設定での実行

```bash
# カスタム設定ファイルを指定
./me19 --config configs/custom_config.json

# カメラデバイスIDを指定
./me19 --device 1

# 出力ファイルを指定
./me19 --output qrcode_data.txt
```

### カメラテストユーティリティの実行

カメラの動作だけを確認したい場合：

```bash
go run ./cmd/camera_test
```

## 4. テスト実行

### ユニットテストの実行

```bash
# すべてのユニットテストを実行
go test ./...

# 特定のパッケージのテストを実行
go test ./internal/camera
go test ./internal/qrcode
go test ./internal/fileio
go test ./internal/signal
go test ./configs

# 詳細な出力でテストを実行
go test -v ./...

# カバレッジレポートの生成
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 統合テストの実行

```bash
# 統合テストのみを実行
go test ./tests/integration

# 詳細な出力で統合テストを実行
go test -v ./tests/integration
```

### ベンチマークの実行

```bash
# ベンチマークテストを実行
go test -bench=. ./...

# 特定のパッケージのベンチマークを実行
go test -bench=. ./internal/qrcode
```

## 5. 動作確認

### QRコード検出の確認

1. QRコードを含む画像を用意します（テスト用画像は`internal/qrcode/testdata`にあります）
2. アプリケーションを実行します：
   ```bash
   ./me19 --output test_output.txt
   ```
3. カメラにQRコードを表示します
4. `test_output.txt`ファイルを確認して、QRコードが正しく検出されたか確認します：
   ```bash
   cat test_output.txt
   ```

### エンドツーエンドの動作確認

1. アプリケーションを実行します：
   ```bash
   ./me19
   ```
2. 以下を確認します：
   - カメラが正常に起動するか
   - QRコードが検出されるか
   - 検出結果がファイルに書き込まれるか
   - Ctrl+Cでアプリケーションが正常に終了するか

## 6. トラブルシューティング

### カメラが見つからない場合

```bash
# カメラデバイスの一覧を確認（Linuxの場合）
ls -l /dev/video*

# 別のカメラデバイスIDを試す
./me19 --device 1
```

### OpenCVエラーが発生する場合

```bash
# OpenCVのインストール状態を確認
pkg-config --libs opencv4
pkg-config --cflags opencv4

# GoCVの再インストール
go install gocv.io/x/gocv@latest
cd $GOPATH/pkg/mod/gocv.io/x/gocv@v0.41.0
make install
```

### ビルドエラーが発生する場合

```bash
# 依存関係を再ダウンロード
go mod tidy

# キャッシュをクリア
go clean -cache

# 再ビルド
go build ./cmd/me19
```

### テストが失敗する場合

```bash
# 詳細なテスト出力を確認
go test -v ./...

# 特定のテストのみを実行
go test -v -run TestName ./package/path
```

## 7. パフォーマンス最適化

### プロファイリング

```bash
# CPUプロファイリングを有効にして実行
go test -cpuprofile cpu.prof -bench=. ./...

# プロファイルの分析
go tool pprof cpu.prof
```

### メモリ使用量の確認

```bash
# メモリプロファイリングを有効にして実行
go test -memprofile mem.prof -bench=. ./...

# メモリプロファイルの分析
go tool pprof mem.prof
```

## 8. CI/CD環境での実行

CI/CD環境でME19をビルドおよびテストする場合は、以下のコマンドを使用できます：

```bash
# 依存関係のインストール
go mod download

# ビルド
go build -o me19 ./cmd/me19

# テスト
go test -v ./...

# クロスプラットフォームビルド
./scripts/build.sh
```

これらのコマンドは、GitHub ActionsやJenkinsなどのCI/CDパイプラインで使用できます。
