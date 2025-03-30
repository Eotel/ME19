# ME19 バイナリビルドと配布ガイド

このドキュメントでは、ME19アプリケーションのバイナリをビルドし、配布する方法について説明します。

## 前提条件

バイナリをビルドするには、以下が必要です：

- Go 1.18以上（1.24.1推奨）
- OpenCVとGoCVのインストール（[SETUP.md](../SETUP.md)を参照）
- Gitがインストールされていること

## ビルドスクリプトの使用

ME19には、複数のプラットフォーム向けにバイナリをビルドするためのスクリプトが含まれています。

```bash
# ビルドスクリプトを実行可能にする
chmod +x ./scripts/build.sh

# スクリプトを実行
./scripts/build.sh
```

このスクリプトは以下のプラットフォーム向けにバイナリをビルドします：

- Linux (amd64)
- macOS (amd64)
- macOS (arm64)
- Windows (amd64)

ビルドされたバイナリは`./build`ディレクトリに保存されます。

## 手動でのクロスコンパイル

スクリプトを使用せずに手動でクロスコンパイルする場合は、以下のコマンドを使用できます：

### Linux向け

```bash
GOOS=linux GOARCH=amd64 go build -o me19_linux_amd64 ./cmd/me19
```

### macOS向け

```bash
# Intel Mac向け
GOOS=darwin GOARCH=amd64 go build -o me19_darwin_amd64 ./cmd/me19

# Apple Silicon (M1/M2)向け
GOOS=darwin GOARCH=arm64 go build -o me19_darwin_arm64 ./cmd/me19
```

### Windows向け

```bash
GOOS=windows GOARCH=amd64 go build -o me19_windows_amd64.exe ./cmd/me19
```

## バイナリの配布

### GitHubリリースの作成

ME19のバイナリを配布する最も簡単な方法は、GitHubリリースを使用することです：

1. GitHubリポジトリで新しいリリースを作成します。
2. セマンティックバージョニングに従ってタグを付けます（例：`v1.0.0`）。
3. ビルドしたバイナリをリリースにアップロードします。
4. SHA256チェックサムファイルもアップロードします。

### リリース自動化

GitHub Actionsを使用して、リリースプロセスを自動化できます：

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'
          
      - name: Build binaries
        run: ./scripts/build.sh
        
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            build/me19_linux_amd64
            build/me19_darwin_amd64
            build/me19_darwin_arm64
            build/me19_windows_amd64.exe
            build/SHA256SUMS.txt
```

## バイナリの検証

配布されたバイナリの整合性を検証するには：

```bash
# チェックサムファイルをダウンロードした同じディレクトリで実行
sha256sum -c SHA256SUMS.txt
```

## Docker配布

Dockerを使用してME19を配布することもできます：

```dockerfile
# Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o me19 ./cmd/me19

FROM alpine:latest

RUN apk add --no-cache opencv

WORKDIR /app
COPY --from=builder /app/me19 .
COPY configs/config.json .

ENTRYPOINT ["./me19"]
```

Dockerイメージをビルドして実行：

```bash
# イメージのビルド
docker build -t me19:latest .

# コンテナの実行（カメラアクセスが必要）
docker run --device=/dev/video0 me19:latest
```

## 注意事項

- OpenCVに依存するため、バイナリを実行するシステムにはOpenCVがインストールされている必要があります。
- 完全に静的にリンクされたバイナリを作成することは、OpenCVの依存関係のため難しい場合があります。
- 各プラットフォームでのテストを忘れないでください。
