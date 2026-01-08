.PHONY: build run test test-e2e test-manual upload-artwork clean init-data install

# Build the server binary
build:
	@echo "Building server..."
	@mkdir -p bin
	go build -o bin/server cmd/server/main.go
	@echo "Copying frontend files..."
	@mkdir -p bin/frontend
	@cp -r frontend/* bin/frontend/
	@echo "Build complete! Frontend files copied to bin/frontend/"

# Run the server
run:
	@echo "Running server..."
	go run cmd/server/main.go

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Initialize data directories
init-data:
	@echo "Initializing data directories..."
	mkdir -p data/categories data/temp

# Run unit tests
test:
	@echo "Running unit tests..."
	go test -v ./...

# Run end-to-end tests (requires server running and API key)
test-e2e:
	@echo "Running end-to-end tests..."
	@echo "Make sure server is running in another terminal (make run)"
	@echo "Place test images in test_images/ folder"
	@echo ""
	go run scripts/test_e2e.go

# Manual testing helper
test-manual:
	@echo "Manual Testing Helper"
	@echo "====================="
	@echo ""
	@echo "Windows: scripts\test_manual.bat [image.jpg] [title] [artist]"
	@echo "Linux/Mac: bash scripts/test_manual.sh [image.jpg] [title] [artist]"
	@echo ""
	@echo "Example:"
	@echo "  scripts\test_manual.bat photo.jpg \"Beach Sunset\" \"John\""

# Interactive artwork upload (for batch uploading from folders)
upload-artwork:
	@echo "Starting Interactive Artwork Upload Agent..."
	@echo "Make sure server is running in another terminal (make run)"
	@echo ""
	go run scripts/interactive_upload.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f data/index.md
	rm -rf data/temp/*
	@echo "Clean complete!"

# Development setup
dev-setup: install init-data
	@echo "Copying .env.example to .env..."
	@if not exist .env copy .env.example .env
	@echo ""
	@echo "Setup complete! Next steps:"
	@echo "1. Edit .env and add your GEMINI_API_KEY"
	@echo "2. Run 'make run' to start the server"

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the server binary (includes web UI)"
	@echo "  run            - Run the server (access web UI at http://localhost:8080)"
	@echo "  install        - Install Go dependencies"
	@echo "  init-data      - Create data directories"
	@echo "  test           - Run unit tests"
	@echo "  test-e2e       - Run end-to-end tests (requires server + API key)"
	@echo "  test-manual    - Show manual testing commands"
	@echo "  upload-artwork - Interactive agent to upload folder of images"
	@echo "  clean          - Clean build artifacts"
	@echo "  dev-setup      - Complete development setup"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Quick start:"
	@echo "  1. make build      - Build server with web UI"
	@echo "  2. cd bin && ./server - Start server"
	@echo "  3. Open http://localhost:8080 in browser"
