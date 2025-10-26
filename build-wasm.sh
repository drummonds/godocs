#!/bin/bash
# Build WebAssembly frontend with version information

set -e

# Get version from git
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
BUILD_DATE=$(date -u +"%Y-%m-%d")

echo "================================================"
echo "  Building goEDMS WebAssembly Frontend"
echo "================================================"
echo "  Version:    $VERSION"
echo "  Build Date: $BUILD_DATE"
echo "================================================"

# Create output directory
mkdir -p web

# Build with version information
GOOS=js GOARCH=wasm go build \
  -ldflags="-X 'github.com/drummonds/goEDMS/webapp.Version=$VERSION' -X 'github.com/drummonds/goEDMS/webapp.BuildDate=$BUILD_DATE'" \
  -o web/app.wasm \
  ./cmd/webapp

# Copy wasm_exec.js
echo "Copying wasm_exec.js..."
GOROOT=$(go env GOROOT)
if [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
    cp "$GOROOT/misc/wasm/wasm_exec.js" web/wasm_exec.js
elif [ -f "/usr/local/go/misc/wasm/wasm_exec.js" ]; then
    cp "/usr/local/go/misc/wasm/wasm_exec.js" web/wasm_exec.js
elif [ -f "/usr/lib/go/misc/wasm/wasm_exec.js" ]; then
    cp "/usr/lib/go/misc/wasm/wasm_exec.js" web/wasm_exec.js
else
    echo "Error: Could not find wasm_exec.js"
    exit 1
fi

echo ""
echo "✓ Build complete!"
echo "  Output: web/app.wasm ($(ls -lh web/app.wasm | awk '{print $5}'))"
echo "  Version in binary: $VERSION"
echo ""
