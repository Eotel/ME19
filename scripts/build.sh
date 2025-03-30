
set -e

VERSION=$(git describe --tags 2>/dev/null || echo "dev")
BUILD_DIR="./build"
BINARY_NAME="me19"

mkdir -p $BUILD_DIR

echo "Building ME19 version $VERSION..."

echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags="-X main.Version=$VERSION" -o $BUILD_DIR/${BINARY_NAME}_linux_amd64 ./cmd/me19
echo "✓ Linux build complete"

echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.Version=$VERSION" -o $BUILD_DIR/${BINARY_NAME}_darwin_amd64 ./cmd/me19
echo "✓ macOS (amd64) build complete"

echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.Version=$VERSION" -o $BUILD_DIR/${BINARY_NAME}_darwin_arm64 ./cmd/me19
echo "✓ macOS (arm64) build complete"

echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-X main.Version=$VERSION" -o $BUILD_DIR/${BINARY_NAME}_windows_amd64.exe ./cmd/me19
echo "✓ Windows build complete"

echo "Creating checksums..."
cd $BUILD_DIR
sha256sum ${BINARY_NAME}_* > SHA256SUMS.txt
cd ..

echo "Build complete! Binaries are available in the $BUILD_DIR directory."
echo "Checksums are available in $BUILD_DIR/SHA256SUMS.txt"
