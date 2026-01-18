.PHONY: build build-linux build-windows build-windows-gui build-darwin build-resources clean clean-resources test run help frontend-build frontend-deps frontend-dev frontend-clean

# Binary name
BINARY_NAME=go-proxy-server
MAIN_PATH=./cmd/server
OUTPUT_DIR=bin
RESOURCE_SCRIPT=./scripts/build_resources.sh
SYSO_FILE=$(MAIN_PATH)/resource_windows_amd64.syso

# Frontend
FRONTEND_DIR=web-ui
FRONTEND_DIST=$(FRONTEND_DIR)/dist

# Build flags
LDFLAGS=-s -w
WINDOWS_GUI_LDFLAGS=-s -w -H=windowsgui

# Default target
all: build

# Check if npm is installed
check-npm:
	@which npm > /dev/null || (echo "Error: npm is not installed. Please install Node.js and npm first." && exit 1)

# Install frontend dependencies
frontend-deps: check-npm
	@echo "Installing frontend dependencies..."
	@cd $(FRONTEND_DIR) && npm install

# Build frontend for production
frontend-build: check-npm
	@echo "Checking frontend dependencies..."
	@if [ ! -d "$(FRONTEND_DIR)/node_modules" ]; then \
		echo "node_modules not found, installing dependencies..."; \
		cd $(FRONTEND_DIR) && npm install; \
	fi
	@echo "Building frontend..."
	@cd $(FRONTEND_DIR) && npm run build
	@echo "Copying frontend build to internal/web/dist..."
	@rm -rf internal/web/dist
	@cp -r $(FRONTEND_DIST) internal/web/dist
	@echo "Frontend build complete: $(FRONTEND_DIST)"

# Clean frontend build
frontend-clean:
	@echo "Cleaning frontend build..."
	@rm -rf $(FRONTEND_DIST)
	@rm -rf $(FRONTEND_DIR)/node_modules

# Development: run frontend dev server
frontend-dev: check-npm
	@echo "Starting frontend dev server..."
	@cd $(FRONTEND_DIR) && npm run dev

# Build for current platform
build: frontend-build
	@echo "Building for current platform..."
	@mkdir -p $(OUTPUT_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_DIR)/$(BINARY_NAME)"

# Build for Linux
build-linux: frontend-build
	@echo "Building for Linux..."
	@mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64"

# Build Windows resources (.syso file)
build-resources:
	@echo "Building Windows resources..."
	@bash $(RESOURCE_SCRIPT) || (echo ""; echo "ERROR: Failed to build Windows resources."; echo "See scripts/README.md for installation instructions."; echo ""; exit 1)

# Build for Windows (console mode)
build-windows: frontend-build build-resources
	@echo "Building for Windows (console mode)..."
	@mkdir -p $(OUTPUT_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_DIR)/$(BINARY_NAME).exe"

# Build for Windows (GUI mode - no console window)
build-windows-gui: frontend-build build-resources
	@echo "Building for Windows (GUI mode - system tray)..."
	@mkdir -p $(OUTPUT_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(WINDOWS_GUI_LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-gui.exe $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_DIR)/$(BINARY_NAME)-gui.exe"

# Build for macOS
build-darwin: frontend-build
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
clean: clean-resources frontend-clean
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
	@echo "  make build              - Build for current platform (includes frontend)"
	@echo "  make build-linux        - Build for Linux (includes frontend)"
	@echo "  make build-windows      - Build for Windows console mode (includes frontend)"
	@echo "  make build-windows-gui  - Build for Windows GUI/tray mode (includes frontend)"
	@echo "  make build-darwin       - Build for macOS (includes frontend)"
	@echo "  make build-resources    - Build Windows resource file (.syso)"
	@echo "  make build-all          - Build for all platforms (includes frontend)"
	@echo "  make frontend-build     - Build frontend only"
	@echo "  make frontend-dev       - Start frontend dev server (port 3000)"
	@echo "  make frontend-deps      - Install frontend dependencies"
	@echo "  make frontend-clean     - Clean frontend build and dependencies"
	@echo "  make clean              - Remove all build artifacts"
	@echo "  make clean-resources    - Remove Windows resource files"
	@echo "  make test               - Run tests"
	@echo "  make run                - Run the application"
	@echo "  make deps               - Install dependencies"
	@echo "  make fmt                - Format code"
	@echo "  make lint               - Lint code"
	@echo "  make help               - Show this help message"
