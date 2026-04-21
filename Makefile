# Build variables
BINARY_NAME=eink-6color
BUILD_DIR=build
DIST_DIR=$(BUILD_DIR)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

# Go build flags
CGO_ENABLED=0
GOBUILD=CGO_ENABLED=$(CGO_ENABLED) go build

# Supported platforms
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

.PHONY: all build clean test deps help local release check fmt lint vet install $(PLATFORMS)

# Default target
all: build

## deps: Download dependencies
deps:
	go mod download
	go mod tidy

## test: Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## build: Build binaries for all platforms
build: $(PLATFORMS)

# Pattern rule for all platforms
$(PLATFORMS):
	@echo "Building $@..."
	@mkdir -p $(DIST_DIR)
	@$(eval GOOS=$(shell echo "$@" | cut -d/ -f1))
	@$(eval GOARCH=$(shell echo "$@" | cut -d/ -f2))
	@if [ "$(GOOS)/$(GOARCH)" = "linux/amd64" ]; then \
		OUTPUT="$(DIST_DIR)/$(BINARY_NAME)"; \
	elif [ "$(GOOS)/$(GOARCH)" = "linux/arm64" ]; then \
		OUTPUT="$(DIST_DIR)/$(BINARY_NAME)-arm64"; \
	else \
		OUTPUT="$(DIST_DIR)/$(BINARY_NAME)-$(shell echo "$@" | tr / -)"; \
	fi; \
	echo "Output: $$OUTPUT"; \
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) $(LDFLAGS) -trimpath -o "$$OUTPUT" ./main.go
	@if [ "$(shell echo "$@" | cut -d/ -f1)" = "windows" ]; then \
		mv "$(DIST_DIR)/$(BINARY_NAME)-$(shell echo "$@" | tr / -)" \
		   "$(DIST_DIR)/$(BINARY_NAME)-$(shell echo "$@" | tr / -).exe"; \
	fi

## linux: Build Linux binaries (amd64, arm64)
linux: linux/amd64 linux/arm64

## darwin: Build macOS binaries (amd64, arm64)
darwin: darwin/amd64 darwin/arm64

## windows: Build Windows binaries (amd64, arm64)
windows: windows/amd64 windows/arm64

## local: Build for current platform only
local:
	@echo "Building for local platform..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go

## release: Create release with checksums
release: clean build
	@echo "Creating release packages..."
	@cd $(DIST_DIR) && for f in $(BINARY_NAME)-*; do \
		if [ "$${f##*.}" != "sha256" ]; then \
			sha256sum "$$f" > "$$f.sha256"; \
		fi; \
	done
	@echo "Release binaries created in $(DIST_DIR)"
	@ls -lh $(DIST_DIR)

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

## lint: Run linters
lint:
	go vet ./...
	go fmt ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	fi

## vet: Run go vet
vet:
	go vet ./...

## fmt: Format code
fmt:
	go fmt ./...

## check: Run tests and checks
check: fmt vet test

## install: Install binary locally
install: local
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

## help: Show this help message
help:
	@echo "Available targets:"
	@echo "  all        - Build binaries for all platforms (default)"
	@echo "  build      - Build binaries for all platforms"
	@echo "  linux      - Build Linux binaries (amd64, arm64)"
	@echo "  darwin     - Build macOS binaries (amd64, arm64)"
	@echo "  windows    - Build Windows binaries (amd64, arm64)"
	@echo "  local      - Build for current platform only"
	@echo "  release    - Build release with checksums"
	@echo "  deps       - Download dependencies"
	@echo "  test       - Run tests"
	@echo "  lint       - Run linters"
	@echo "  vet        - Run go vet"
	@echo "  fmt        - Format code"
	@echo "  check      - Run tests and checks"
	@echo "  clean      - Remove build artifacts"
	@echo "  install    - Install binary locally"
	@echo "  help       - Show this help message"
	@echo ""
	@echo "Build outputs directory: $(BUILD_DIR)/"
	@echo "Platform-specific variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  BINARY_NAME=$(BINARY_NAME)"
	@echo "  BUILD_DIR=$(BUILD_DIR)"
