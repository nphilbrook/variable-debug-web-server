.PHONY: all clean darwin linux linux-amd64 linux-arm64

BINARY_NAME=variable-debug-server
BUILD_DIR=build

all: darwin linux

# Build for macOS (native)
darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	@echo "macOS binaries built successfully"

# Build for Linux (cross-compile from macOS)
linux: linux-amd64 linux-arm64

linux-amd64:
	@echo "Building for Linux AMD64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	@echo "Linux AMD64 binary built successfully"

linux-arm64:
	@echo "Building for Linux ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 main.go
	@echo "Linux ARM64 binary built successfully"

# Quick build for most common Linux server target
linux-server: linux-amd64

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f server
	@echo "Clean complete"

# Run locally (macOS)
run:
	@go run main.go

# Show help
help:
	@echo "Available targets:"
	@echo "  all           - Build for both macOS and Linux (default)"
	@echo "  darwin        - Build for macOS (AMD64 and ARM64)"
	@echo "  linux         - Build for Linux (AMD64 and ARM64)"
	@echo "  linux-amd64   - Build for Linux AMD64 only"
	@echo "  linux-arm64   - Build for Linux ARM64 only"
	@echo "  linux-server  - Build for most common Linux server (AMD64)"
	@echo "  clean         - Remove build artifacts"
	@echo "  run           - Run the server locally"
	@echo "  help          - Show this help message"
