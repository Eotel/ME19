# ME19 ユーザーガイド

## 概要

ME19 は、カメラから映像を取得し、その映像ストリーム内の QR コードを検出し、デコードしたデータをファイルに書き込む Go アプリケーションです。このガイドでは、ME19 のインストール方法、設定方法、使用方法について説明します。

## インストール方法

### 前提条件

ME19 を使用するには、以下のソフトウェアが必要です：

1. **Go 1.18 以上**（1.24.1 推奨）
2. **OpenCV 4.x**（4.11.0 推奨）
3. **GoCV**（OpenCV の Go バインディング）

詳細なインストール手順については、[SETUP.md](../SETUP.md)を参照してください。

### バイナリからのインストール

最新のリリースバイナリは[GitHub リリースページ](https://github.com/eotel/me19/releases)からダウンロードできます。

各プラットフォーム向けのバイナリが提供されています：

- Windows: `me19_windows_amd64.exe`
- macOS: `me19_darwin_amd64`
- Linux: `me19_linux_amd64`

ダウンロードしたバイナリを実行可能にして、PATH の通ったディレクトリに配置してください。

### ソースからのビルド

ソースコードから ME19 をビルドするには：

```bash
# リポジトリのクローン
git clone https://github.com/eotel/me19.git
cd me19

# 依存関係のインストール
go mod download

# ビルド
go build -o me19 ./cmd/me19

# インストール（オプション）
go install ./cmd/me19
```

## 設定方法

ME19 は、JSON ファイルを使用して設定できます。

### 設定ファイルの検索順序

設定ファイルを明示的に指定しない場合、ME19 は以下の場所を順番に検索します：

1. カレントディレクトリの `config.json`
2. `configs/config.json` サブディレクトリ
3. 親ディレクトリの `../configs/config.json`
4. 親の親ディレクトリの `../../configs/config.json`
5. システム全体の設定 `/etc/me19/config.json`（Linux/macOS のみ）
6. ユーザーホームの設定 `~/.config/me19/config.json`
7. 実行ファイルと同じディレクトリの `config.json`
8. 実行ファイルと同じディレクトリの `configs/config.json`
9. 実行ファイルの親ディレクトリの `../configs/config.json`

有効な設定ファイルが見つからない場合は、アプリケーション内のデフォルト設定が使用されます。

### 設定ファイルの例

```json
{
  "camera": {
    "device_id": 0,
    "width": 1280,
    "height": 720,
    "fps": 30
  },
  "qrcode": {
    "scan_interval_ms": 500
  },
  "output_file": {
    "file_path": "qrcode_output.txt"
  }
}
```

### 設定パラメータ

#### カメラ設定

- `device_id`: カメラデバイスの ID（通常、最初のカメラは 0）
- `width`: キャプチャ解像度の幅（ピクセル）
- `height`: キャプチャ解像度の高さ（ピクセル）
- `fps`: フレームレート（フレーム/秒）

#### QR コード設定

- `scan_interval_ms`: QR コードスキャン間隔（ミリ秒）

#### 出力ファイル設定

- `file_path`: QR コードデータを書き込むファイルのパス

### コマンドライン引数

ME19 は、以下のコマンドライン引数をサポートしています：

```
Usage: me19 [options]

Options:
  -config string   設定ファイルのパス
  -c string        設定ファイルのパス（短縮形）
  -device int      カメラデバイスID
  -d int           カメラデバイスID（短縮形）
  -output string   出力ファイルパス
  -o string        出力ファイルパス（短縮形）
  -h               ヘルプメッセージの表示
```

例：

```bash
# カスタム設定ファイルを使用
me19 -config my_config.json
# または
me19 -c my_config.json

# カメラデバイスIDを指定
me19 -device 1
# または
me19 -d 1

# 出力ファイルを指定
me19 -output $HOME/.local/share/me19/code.txt
# または
me19 -o $HOME/.local/share/me19/code.txt

# 複数のオプションを組み合わせる
me19 -d 1 -o scan_results.txt

# 設定ファイルを指定せずに実行（自動検索）
me19
```

### 環境変数によるオーバーライド

以下の環境変数を設定することで、設定ファイルの値をオーバーライドできます：

```
ME19_CAMERA_DEVICE_ID       - カメラデバイスID
ME19_CAMERA_WIDTH           - キャプチャ幅
ME19_CAMERA_HEIGHT          - キャプチャ高さ
ME19_CAMERA_FPS             - フレームレート
ME19_QRCODE_SCAN_INTERVAL_MS - QRコードスキャン間隔
ME19_OUTPUT_FILE_PATH       - 出力ファイルパス
ME19_TEST_MODE              - テストモード (true/false)
```

例：

```bash
# カメラデバイスIDを環境変数で指定
export ME19_CAMERA_DEVICE_ID=2
me19
```

### 設定の優先順位

ME19 の設定は、以下の優先順位で適用されます（上の方が優先）：

1. 環境変数
2. コマンドライン引数
3. 設定ファイル
4. デフォルト設定

例えば、コマンドライン引数と環境変数の両方でカメラデバイス ID を指定した場合、環境変数の値が使用されます。

## 使用方法

### 基本的な使い方

1. ME19 を起動します：

   ```bash
   me19
   ```

2. アプリケーションがカメラを起動し、QR コードのスキャンを開始します。

3. 検出された QR コードのデータは、設定ファイルで指定されたファイル（デフォルトは`code.txt`）に書き込まれます。

4. プログラムを終了するには、`Ctrl+C`を押します。

### 高度な使用例

#### 特定のカメラを使用

```bash
me19 -device 1
```

#### カスタム設定ファイルを使用

```bash
me19 -config my_settings.json
```

#### 出力ファイルを指定

```bash
me19 -output $HOME/.local/share/me19/code.txt
```

#### テストモードでの実行

```bash
export ME19_TEST_MODE=true
me19
```

## トラブルシューティング

### カメラが見つからない場合

- カメラが正しく接続されていることを確認してください。
- 正しいデバイス ID を指定しているか確認してください。
- 他のアプリケーションがカメラを使用していないか確認してください。

### QR コードが検出されない場合

- QR コードがカメラの視野内にあることを確認してください。
- 十分な照明があることを確認してください。
- QR コードが鮮明で、歪みや反射がないことを確認してください。

### 設定ファイルが見つからない場合

- 特定の設定ファイルを使用する場合は、パスが正しいことを確認してください。
- 設定ファイルが見つからない場合は、デフォルト設定が使用されます。
- 環境変数を使用して設定をオーバーライドすることもできます。

### 出力ファイルについて

- デフォルトでは、出力ファイル（`code.txt`）はプログラムを実行したカレントディレクトリに作成されます。
- 絶対パスを指定するには、コマンドラインオプション `-output` または環境変数 `ME19_OUTPUT_FILE_PATH` を使用してください。

### アプリケーションがクラッシュする場合

- OpenCV と GoCV が正しくインストールされていることを確認してください。
- 最新バージョンの ME19 を使用していることを確認してください。
- 詳細なエラーメッセージを確認し、必要に応じて Issue を報告してください。

## サポート

問題や質問がある場合は、[GitHub の Issue トラッカー](https://github.com/eotel/me19/issues)に報告してください。
