

# Set environment variables to disable CGO (if not needed)
export CGO_ENABLED=0

# Install cross-compilers for ARM
brew install arm-linux-gnueabihf-binutils
brew install arm-linux-gnueabihf-gcc

# Clean build caches
fyne-cross clean

# Run fyne-cross with verbose logging
fyne-cross linux --arch=arm64 -v --image fyneio/fyne-cross:latest

# If CGO is required, ensure proper cross-compilation setup
export CGO_ENABLED=1

# Run fyne-cross again with appropriate settings
fyne-cross linux --arch=arm64 -v
