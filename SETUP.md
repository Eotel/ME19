# ME19 Setup Guide

This document provides instructions for setting up the ME19 project environment.

## Prerequisites

### Go Installation
- Install Go 1.18 or later (this project uses Go 1.24.1)
- Set up GOPATH and GOROOT environment variables
- For version management, we recommend using mise:
  ```bash
  # Install mise
  curl https://mise.run | sh
  
  # Add mise to your shell profile for persistent activation
  echo 'eval "$(mise activate bash)"' >> ~/.bashrc
  source ~/.bashrc
  
  # Configure Go version in mise.toml
  cat > mise.toml << EOF
  [tools]
  go = "latest"
  EOF
  
  # Activate mise in current shell
  eval "$(mise activate bash)"
  ```

### OpenCV Installation

GoCV requires OpenCV 4.x to be installed.

#### Ubuntu/Debian
```bash
# Install OpenCV 4.x and dependencies
sudo apt-get update
sudo apt-get install -y build-essential cmake pkg-config
sudo apt-get install -y libjpeg-dev libtiff-dev libpng-dev
sudo apt-get install -y libavcodec-dev libavformat-dev libswscale-dev
sudo apt-get install -y libgtk2.0-dev libcairo2-dev
sudo apt-get install -y libgtkglext1-dev libgtkglext1
sudo apt-get install -y libatlas-base-dev gfortran

# Install OpenCV 4.x
sudo apt-get install -y libopencv-dev

# Verify installation
pkg-config --modversion opencv4
```

#### macOS
```bash
# Install OpenCV 4.x
brew install opencv

# Verify installation
pkg-config --modversion opencv4
```

#### Windows
Follow the instructions at [GoCV Windows installation guide](https://gocv.io/getting-started/windows/).

### Troubleshooting OpenCV Installation

If you encounter build errors related to OpenCV, you may need to install OpenCV from source:

```bash
# Clone OpenCV repositories
git clone https://github.com/opencv/opencv.git
git clone https://github.com/opencv/opencv_contrib.git

# Create build directory
cd opencv
mkdir build
cd build

# Configure and build OpenCV
cmake -D CMAKE_BUILD_TYPE=RELEASE \
      -D CMAKE_INSTALL_PREFIX=/usr/local \
      -D OPENCV_EXTRA_MODULES_PATH=../../opencv_contrib/modules \
      -D BUILD_EXAMPLES=OFF \
      -D BUILD_opencv_apps=OFF \
      -D BUILD_DOCS=OFF \
      -D BUILD_PERF_TESTS=OFF \
      -D BUILD_TESTS=OFF \
      -D BUILD_opencv_java=OFF \
      -D BUILD_opencv_python=OFF \
      -D BUILD_opencv_python2=OFF \
      -D BUILD_opencv_python3=OFF \
      -D WITH_FFMPEG=ON \
      ..

make -j$(nproc)
sudo make install
sudo ldconfig

# Verify installation
pkg-config --modversion opencv4
pkg-config --cflags --libs opencv4
```

### Installing GoCV

After installing OpenCV, you need to install GoCV:

```bash
# Install GoCV
go install gocv.io/x/gocv@latest

# Navigate to the GoCV package directory
cd $GOPATH/src/gocv.io/x/gocv

# If the directory doesn't exist, find it using:
cd $(go env GOPATH)/pkg/mod/gocv.io/x/gocv@v0.41.0  # Use the version you installed

# Install GoCV with OpenCV
make install

# Verify installation
# You should see output showing the GoCV version and OpenCV lib version
```

**Important Note**: Ensure that Go is available in your PATH before running `make install`. If you're using mise for Go version management, make sure mise is activated in your shell profile:

```bash
# Add to ~/.bashrc or ~/.zshrc
echo 'eval "$(mise activate bash)"' >> ~/.bashrc
source ~/.bashrc
```

This ensures that the `go` command is available globally, which is required by the GoCV Makefile during installation.

## Project Setup

1. Clone the repository:
```bash
git clone https://github.com/Eotel/ME19.git
cd ME19
```

2. Install dependencies:
```bash
go mod download
```

3. Verify OpenCV and GoCV installation:
```bash
# Verify OpenCV installation
pkg-config --modversion opencv4
pkg-config --cflags --libs opencv4

# Verify GoCV installation
go run -v ./cmd/camera_test/main.go
```

4. Run tests to verify setup:
```bash
go test ./...
```

## Development Workflow

This project follows git-flow for version control:

1. Create feature branches from develop:
```bash
git flow feature start feature-name
```

2. Follow TDD (Test-Driven Development) practices:
   - Write tests first
   - Implement code to pass tests
   - Refactor as needed

3. Use conventional commits for commit messages:
```
feat: add new feature
fix: fix a bug
docs: update documentation
test: add or update tests
refactor: refactor code without changing functionality
```

4. Complete feature and merge back to develop:
```bash
git flow feature finish feature-name
```
