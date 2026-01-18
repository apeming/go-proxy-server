#!/bin/bash
# Build Windows resources (.syso file)
# This script compiles the Windows resource file into a .syso file
# that will be automatically embedded by Go build
#
# Supports three methods (in priority order):
# 1. goversioninfo (pure Go, recommended) - most mature and stable
# 2. windres (from mingw-w64) - traditional method
# 3. go-winres (pure Go) - alternative pure Go solution

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Get GOPATH/bin directory
GOPATH_BIN="$(go env GOPATH)/bin"

echo "Building Windows resources..."

# Method 1: Try goversioninfo first (recommended)
if command -v goversioninfo &> /dev/null || [ -f "$GOPATH_BIN/goversioninfo" ]; then
    echo "Using goversioninfo method (recommended)..."

    # Determine which goversioninfo to use
    if command -v goversioninfo &> /dev/null; then
        GOVERSIONINFO="goversioninfo"
    else
        GOVERSIONINFO="$GOPATH_BIN/goversioninfo"
    fi

    echo "Generating resources with goversioninfo..."
    cd "$PROJECT_ROOT"
    "$GOVERSIONINFO" -64 -o cmd/server/resource_windows_amd64.syso assets/versioninfo.json

    SYSO_FILE="$PROJECT_ROOT/cmd/server/resource_windows_amd64.syso"
    if [ -f "$SYSO_FILE" ]; then
        echo "✓ Resource file created: $SYSO_FILE"
        ls -lh "$SYSO_FILE"
        echo "Done! The .syso file will be automatically included in Windows builds."
        exit 0
    else
        echo "✗ Failed to create resource file"
        exit 1
    fi
fi

# Method 2: Try windres (if available)
if command -v x86_64-w64-mingw32-windres &> /dev/null || command -v windres &> /dev/null; then
    echo "Using windres method..."

    if command -v x86_64-w64-mingw32-windres &> /dev/null; then
        WINDRES="x86_64-w64-mingw32-windres"
    else
        WINDRES="windres"
    fi

    RC_FILE="$SCRIPT_DIR/resource.rc"
    SYSO_FILE="$PROJECT_ROOT/cmd/server/resource_windows_amd64.syso"

    echo "Compiling with $WINDRES..."
    cd "$SCRIPT_DIR"
    $WINDRES -i resource.rc -o "$SYSO_FILE" -O coff

    if [ -f "$SYSO_FILE" ]; then
        echo "✓ Resource file created: $SYSO_FILE"
        ls -lh "$SYSO_FILE"
    else
        echo "✗ Failed to create resource file"
        exit 1
    fi

# Method 2: Fall back to go-winres (pure Go solution)
else
    # Check if go-winres is installed
    if ! command -v go-winres &> /dev/null && [ ! -f "$GOPATH_BIN/go-winres" ]; then
        echo ""
        echo "ERROR: go-winres not found."
        echo ""
        echo "Please install go-winres first:"
        echo "  go install github.com/tc-hib/go-winres@latest"
        echo ""
        echo "Or install mingw-w64 to use windres instead:"
        echo "  Ubuntu/Debian: sudo apt-get install mingw-w64"
        echo "  macOS:         brew install mingw-w64"
        echo ""
        exit 1
    fi

    echo "windres not found, using go-winres (pure Go solution)..."

    # Determine which go-winres to use
    if command -v go-winres &> /dev/null; then
        GOWINRES="go-winres"
    else
        GOWINRES="$GOPATH_BIN/go-winres"
    fi

    # Create winres directory if it doesn't exist
    WINRES_DIR="$PROJECT_ROOT/winres"
    mkdir -p "$WINRES_DIR"

    # Create winres.json configuration
    cat > "$WINRES_DIR/winres.json" <<'EOF'
{
  "RT_MANIFEST": {
    "#1": {
      "0409": "manifest.xml"
    }
  },
  "RT_VERSION": {
    "#1": {
      "0000": {
        "fixed": {
          "file_version": "1.0.0.0",
          "product_version": "1.0.0.0"
        },
        "info": {
          "0409": {
            "CompanyName": "Go Proxy Server Project",
            "FileDescription": "SOCKS5 and HTTP Proxy Server",
            "FileVersion": "1.0.0.0",
            "InternalName": "go-proxy-server",
            "LegalCopyright": "Copyright (C) 2025",
            "OriginalFilename": "go-proxy-server.exe",
            "ProductName": "Go Proxy Server",
            "ProductVersion": "1.0.0.0"
          }
        }
      }
    }
  }
}
EOF

    # Copy manifest to winres directory
    cp "$SCRIPT_DIR/manifest.xml" "$WINRES_DIR/"

    # Generate resources
    echo "Generating resources with go-winres..."
    cd "$PROJECT_ROOT"
    "$GOWINRES" make --in "$WINRES_DIR" --out "$PROJECT_ROOT/cmd/server" --arch amd64

    SYSO_FILE="$PROJECT_ROOT/cmd/server/rsrc_windows_amd64.syso"
    if [ -f "$SYSO_FILE" ]; then
        echo "✓ Resource file created: $SYSO_FILE"
        ls -lh "$SYSO_FILE"
        echo "Done! The .syso file will be automatically included in Windows builds."
        exit 0
    else
        echo "✗ Failed to create resource file"
        exit 1
    fi
fi

# No method available - provide installation instructions
echo ""
echo "ERROR: No Windows resource compiler found!"
echo ""
echo "Please install one of the following:"
echo ""
echo "Option 1 (Recommended): goversioninfo (pure Go)"
echo "  go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest"
echo ""
echo "Option 2: go-winres (pure Go, alternative)"
echo "  go install github.com/tc-hib/go-winres@latest"
echo ""
echo "Option 3: windres (requires C compiler)"
echo "  Ubuntu/Debian: sudo apt-get install mingw-w64"
echo "  macOS:         brew install mingw-w64"
echo ""
echo "After installation, run this script again or use 'make build-windows'"
echo ""
exit 1
