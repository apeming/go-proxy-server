.PHONY: build build-linux build-windows build-windows-gui build-darwin build-resources clean clean-resources test run help

# Binary name
BINARY_NAME=go-proxy-server
MAIN_PATH=./cmd/server
OUTPUT_DIR=bin
RESOURCE_SCRIPT=./scripts/build_resources.sh
SYSO_FILE=$(MAIN_PATH)/resource_windows_amd64.syso

# Build flags
LDFLAGS=-s -w
WINDOWS_GUI_LDFLAGS=-s -w -H=windowsgui

# Default target
all: build

# Build for current platform
build:
	@echo "Building for current platform..."
	@mkdir -p $(OUTPUT_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_DIR)/$(BINARY_NAME)"

# Build for Linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64"

# Build Windows resources (.syso file)
build-resources:
	@echo "Building Windows resources..."
	@bash $(RESOURCE_SCRIPT) || (echo ""; echo "ERROR: Failed to build Windows resources."; echo "See scripts/README.md for installation instructions."; echo ""; exit 1)

# Build for Windows (console mode)
build-windows: build-resources
	@echo "Building for Windows (console mode)..."
	@mkdir -p $(OUTPUT_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_DIR)/$(BINARY_NAME).exe"

# Build for Windows (GUI mode - no console window)
build-windows-gui: build-resources
	@echo "Building for Windows (GUI mode - system tray)..."
	@mkdir -p $(OUTPUT_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(WINDOWS_GUI_LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-gui.exe $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_DIR)/$(BINARY_NAME)-gui.exe"

# Build for macOS
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(OUTPUT_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64"

# Build for all platforms
build-all: build-linux build-windows build-windows-gui build-darwin
	@echo "All builds complete!"

# Clean Windows resources
clean-resources:
	@echo "Cleaning Windows resources..."
	@rm -f $(SYSO_FILE)
	@rm -f $(MAIN_PATH)/rsrc_windows_amd64.syso
	@rm -rf winres/
	@echo "Resources cleaned!"

# Clean build artifacts
clean: clean-resources
	@echo "Cleaning build artifacts..."
	rm -rf $(OUTPUT_DIR)
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run the application
run:
	@echo "Running application..."
	go run $(MAIN_PATH)

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  make build              - Build for current platform"
	@echo "  make build-linux        - Build for Linux"
	@echo "  make build-windows      - Build for Windows (console mode)"
	@echo "  make build-windows-gui  - Build for Windows (GUI/tray mode)"
	@echo "  make build-darwin       - Build for macOS"
	@echo "  make build-resources    - Build Windows resource file (.syso)"
	@echo "  make build-all          - Build for all platforms"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make clean-resources    - Remove Windows resource files"
	@echo "  make test               - Run tests"
	@echo "  make run                - Run the application"
	@echo "  make deps               - Install dependencies"
	@echo "  make fmt                - Format code"
	@echo "  make lint               - Lint code"
	@echo "  make help               - Show this help message"
