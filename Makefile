.PHONY: build run test clean init-data install

# Build the server binary
build:
	@echo "Building server..."
	go build -o bin/server cmd/server/main.go

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

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f data/index.md
	rm -rf data/temp/*

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
	@echo "  build      - Build the server binary"
	@echo "  run        - Run the server"
	@echo "  install    - Install Go dependencies"
	@echo "  init-data  - Create data directories"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts"
	@echo "  dev-setup  - Complete development setup"
	@echo "  help       - Show this help message"
