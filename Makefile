build: test
	go build -o bin/go-server cmd/api/main.go

run:
	./bin/go-server

clean:
	rm -f bin/go-server

.PHONY: dev
dev:
	@if command -v air >/dev/null 2>&1; then \
		air -c .air.toml; \
	else \
		if [ -x ./tmp/bin/air ]; then \
			./tmp/bin/air -c .air.toml; \
		else \
			echo "Installing Air locally..."; \
			mkdir -p ./tmp/bin; \
			curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b ./tmp/bin; \
			./tmp/bin/air -c .air.toml; \
		fi; \
	fi

.PHONY: swag
swag:
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api/main.go -o ./docs; \
	else \
		if [ -x ./tmp/bin/swag ]; then \
			./tmp/bin/swag init -g cmd/api/main.go -o ./docs; \
		else \
			echo "Installing swag locally..."; \
			mkdir -p ./tmp/bin; \
			GOBIN=$(CURDIR)/tmp/bin go install github.com/swaggo/swag/cmd/swag@latest; \
			./tmp/bin/swag init -g cmd/api/main.go -o ./docs; \
		fi; \
	fi

# Test targets
.PHONY: test test-verbose test-race test-coverage test-user test-integration test-short

# Run all tests
test:
	go test ./tests/...

# Run all tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with race detection
test-race:
	go test -race ./...

# Run tests with coverage report
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run only user module tests
test-user:
	go test -v ./tests/user

# Run integration tests (if any exist)
test-integration:
	go test -v ./tests/integration

# Run only short tests (skip long-running tests)
test-short:
	go test -short ./...

# Clean test artifacts
test-clean:
	rm -f coverage.out coverage.html

# Development helpers
.PHONY: fmt lint help

# Format Go code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Windows build and packaging targets
.PHONY: build-windows package-windows clean-windows

# Build Windows executable
build-windows:
	@echo "Building Windows executable..."
	GOOS=windows GOARCH=amd64 go build -o gopi-windows.exe ./cmd/api
	@echo "Windows executable created: gopi-windows.exe"

# Create Windows package with all necessary files
package-windows: build-windows
	@echo "Creating Windows package..."
	mkdir -p windows-package
	cp gopi-windows.exe windows-package/
	@echo "# GoPadi Backend Configuration for Windows Development" > windows-package/env.example.txt
	@echo "# Copy this file and customize as needed" >> windows-package/env.example.txt
	@echo "" >> windows-package/env.example.txt
	@echo "# Server Configuration" >> windows-package/env.example.txt
	@echo "PUBLIC_HOST=http://localhost" >> windows-package/env.example.txt
	@echo "PORT=8080" >> windows-package/env.example.txt
	@echo "" >> windows-package/env.example.txt
	@echo "# Database Configuration (SQLite for development)" >> windows-package/env.example.txt
	@echo "DB_DRIVER=sqlite" >> windows-package/env.example.txt
	@echo "DB_HOST=127.0.0.1" >> windows-package/env.example.txt
	@echo "DB_PORT=3306" >> windows-package/env.example.txt
	@echo "DB_USER=root" >> windows-package/env.example.txt
	@echo "DB_PASSWORD=password" >> windows-package/env.example.txt
	@echo "DB_NAME=gopi_dev.db" >> windows-package/env.example.txt
	@echo "" >> windows-package/env.example.txt
	@echo "# Email Configuration (LOCAL EMAIL SERVICE ENABLED)" >> windows-package/env.example.txt
	@echo "# Set to true for local development (logs emails instead of sending)" >> windows-package/env.example.txt
	@echo "USE_LOCAL_EMAIL=true" >> windows-package/env.example.txt
	@echo "" >> windows-package/env.example.txt
	@echo "# Email log file path (only used when USE_LOCAL_EMAIL=true)" >> windows-package/env.example.txt
	@echo "EMAIL_LOG_PATH=./logs/emails.log" >> windows-package/env.example.txt
	@echo "" >> windows-package/env.example.txt
	@echo "# Email settings (not needed when using local service)" >> windows-package/env.example.txt
	@echo "EMAIL_HOST=smtp.gmail.com" >> windows-package/env.example.txt
	@echo "EMAIL_PORT=587" >> windows-package/env.example.txt
	@echo "EMAIL_USERNAME=your-email@gmail.com" >> windows-package/env.example.txt
	@echo "EMAIL_PASSWORD=your-app-password" >> windows-package/env.example.txt
	@echo "EMAIL_FROM=noreply@gopadi.com" >> windows-package/env.example.txt
	@echo "" >> windows-package/env.example.txt
	@echo "# JWT Configuration" >> windows-package/env.example.txt
	@echo "JWT_SECRET=dev-jwt-secret-change-me-in-production-key-for-windows-dev" >> windows-package/env.example.txt
	@echo "USE_DATABASE_JWT=false" >> windows-package/env.example.txt
	@echo "" >> windows-package/env.example.txt
	@echo "# Password Reset Configuration" >> windows-package/env.example.txt
	@echo "USE_DATABASE_PWRESET=false" >> windows-package/env.example.txt
	@echo "" >> windows-package/env.example.txt
	@echo "# Redis Configuration (optional for development)" >> windows-package/env.example.txt
	@echo "REDIS_ADDR=localhost:6379" >> windows-package/env.example.txt
	@echo "REDIS_PASSWORD=" >> windows-package/env.example.txt
	@echo "REDIS_DB=0" >> windows-package/env.example.txt
	@echo "" >> windows-package/env.example.txt
	@echo "# Logging Configuration" >> windows-package/env.example.txt
	@echo "LOG_LEVEL=info" >> windows-package/env.example.txt
	@echo "LOG_FILE=logs/app.log" >> windows-package/env.example.txt
	@echo "LOG_FILE_ENABLED=true" >> windows-package/env.example.txt
	@echo "GIN_MODE=debug" >> windows-package/env.example.txt
	@echo "" >> windows-package/env.example.txt
	@echo "# Session Configuration" >> windows-package/env.example.txt
	@echo "SESSION_SECRET=dev-session-secret-windows" >> windows-package/env.example.txt
	@echo "SESSION_NAME=gopi_session" >> windows-package/env.example.txt
	@echo "SESSION_SECURE=false" >> windows-package/env.example.txt
	@echo "SESSION_DOMAIN=" >> windows-package/env.example.txt
	@echo "SESSION_MAX_AGE=86400" >> windows-package/env.example.txt
	@echo "" >> windows-package/env.example.txt
	@echo "# Storage Configuration" >> windows-package/env.example.txt
	@echo "STORAGE_BACKEND=local" >> windows-package/env.example.txt
	@echo "UPLOAD_BASE_DIR=./uploads" >> windows-package/env.example.txt
	@echo "UPLOAD_PUBLIC_BASE_URL=/uploads" >> windows-package/env.example.txt
	@echo "" >> windows-package/env.example.txt
	@echo "# S3 Configuration (if using S3 storage - not needed for local dev)" >> windows-package/env.example.txt
	@echo "S3_ENDPOINT=" >> windows-package/env.example.txt
	@echo "S3_REGION=us-east-1" >> windows-package/env.example.txt
	@echo "S3_BUCKET=" >> windows-package/env.example.txt
	@echo "S3_ACCESS_KEY_ID=" >> windows-package/env.example.txt
	@echo "S3_SECRET_ACCESS_KEY=" >> windows-package/env.example.txt
	@echo "S3_USE_SSL=true" >> windows-package/env.example.txt
	@echo "S3_FORCE_PATH_STYLE=false" >> windows-package/env.example.txt
	@echo "S3_PUBLIC_BASE_URL=" >> windows-package/env.example.txt

	@echo "GoPadi Backend - Windows Development Package" > windows-package/README_Windows.txt
	@echo "=============================================" >> windows-package/README_Windows.txt
	@echo "" >> windows-package/README_Windows.txt
	@echo "This package contains the GoPadi backend server compiled for Windows, with local email service enabled for frontend development." >> windows-package/README_Windows.txt
	@echo "" >> windows-package/README_Windows.txt
	@echo "📁 Package Contents:" >> windows-package/README_Windows.txt
	@echo "- gopi-windows.exe    - The main server executable" >> windows-package/README_Windows.txt
	@echo "- env.example.txt     - Sample environment configuration file" >> windows-package/README_Windows.txt
	@echo "- README_Windows.txt  - This file" >> windows-package/README_Windows.txt
	@echo "- start-server.bat    - Easy startup script" >> windows-package/README_Windows.txt
	@echo "" >> windows-package/README_Windows.txt
	@echo "🚀 Quick Start" >> windows-package/README_Windows.txt
	@echo "==============" >> windows-package/README_Windows.txt
	@echo "" >> windows-package/README_Windows.txt
	@echo "1. Extract all files to a folder on your Windows machine" >> windows-package/README_Windows.txt
	@echo "2. Copy env.example.txt to .env and customize if needed" >> windows-package/README_Windows.txt
	@echo "3. Open Command Prompt or PowerShell in the folder" >> windows-package/README_Windows.txt
	@echo "4. Run the server: gopi-windows.exe" >> windows-package/README_Windows.txt
	@echo "" >> windows-package/README_Windows.txt
	@echo "The server will start on http://localhost:8080" >> windows-package/README_Windows.txt
	@echo "" >> windows-package/README_Windows.txt
	@echo "📧 Email Testing" >> windows-package/README_Windows.txt
	@echo "===============" >> windows-package/README_Windows.txt
	@echo "" >> windows-package/README_Windows.txt
	@echo "When USE_LOCAL_EMAIL=true (which is set by default), all emails are logged to:" >> windows-package/README_Windows.txt
	@echo "./logs/emails.log" >> windows-package/README_Windows.txt
	@echo "" >> windows-package/README_Windows.txt
	@echo "Instead of being sent via SMTP. This means:" >> windows-package/README_Windows.txt
	@echo "" >> windows-package/README_Windows.txt
	@echo "✅ No email server configuration needed" >> windows-package/README_Windows.txt
	@echo "✅ No spam during development" >> windows-package/README_Windows.txt
	@echo "✅ OTP codes and reset links are clearly logged" >> windows-package/README_Windows.txt
	@echo "✅ Perfect for frontend testing" >> windows-package/README_Windows.txt
	@echo "" >> windows-package/README_Windows.txt
	@echo "Example log output:" >> windows-package/README_Windows.txt
	@echo "=========================================" >> windows-package/README_Windows.txt
	@echo "OTP EMAIL REQUEST" >> windows-package/README_Windows.txt
	@echo "=========================================" >> windows-package/README_Windows.txt
	@echo "To: user@example.com" >> windows-package/README_Windows.txt
	@echo "Name: John Doe" >> windows-package/README_Windows.txt
	@echo "OTP CODE: 123456" >> windows-package/README_Windows.txt
	@echo "Timestamp: 2025-09-03T15:30:25Z" >> windows-package/README_Windows.txt
	@echo "=========================================" >> windows-package/README_Windows.txt
	@echo "COPY THIS OTP CODE FOR TESTING:" >> windows-package/README_Windows.txt
	@echo "OTP: 123456" >> windows-package/README_Windows.txt
	@echo "=========================================" >> windows-package/README_Windows.txt

	@echo "@echo off" > windows-package/start-server.bat
	@echo "echo ========================================" >> windows-package/start-server.bat
	@echo "echo     GoPadi Backend Server - Windows" >> windows-package/start-server.bat
	@echo "echo ========================================" >> windows-package/start-server.bat
	@echo "echo." >> windows-package/start-server.bat
	@echo "echo Starting server on http://localhost:8080" >> windows-package/start-server.bat
	@echo "echo." >> windows-package/start-server.bat
	@echo "echo Email logs will be saved to: ./logs/emails.log" >> windows-package/start-server.bat
	@echo "echo." >> windows-package/start-server.bat
	@echo "echo Press Ctrl+C to stop the server" >> windows-package/start-server.bat
	@echo "echo." >> windows-package/start-server.bat
	@echo "echo ========================================" >> windows-package/start-server.bat
	@echo "echo." >> windows-package/start-server.bat
	@echo "echo." >> windows-package/start-server.bat
	@echo "gopi-windows.exe" >> windows-package/start-server.bat
	@echo "echo." >> windows-package/start-server.bat
	@echo "echo Server stopped." >> windows-package/start-server.bat
	@echo "pause" >> windows-package/start-server.bat

	@echo "Creating ZIP package..."
	zip -r gopi-backend-windows.zip windows-package/
	@echo "Package created: gopi-backend-windows.zip"
	@echo "Package contents:"
	@ls -la windows-package/
	@echo "ZIP file size:"
	@ls -lh gopi-backend-windows.zip

# Clean Windows build artifacts
clean-windows:
	@echo "Cleaning Windows build artifacts..."
	rm -f gopi-windows.exe
	rm -rf windows-package/
	rm -f gopi-backend-windows.zip

# Docker targets
.PHONY: docker-build docker-run docker-stop docker-logs docker-clean

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t gopi-backend .

# Run Docker container
docker-run: docker-build
	@echo "Starting Docker container..."
	docker run -d --name gopi-backend \
		-p 8080:8080 \
		-v $(PWD)/logs:/app/logs \
		-v $(PWD)/uploads:/app/uploads \
		--env-file .env \
		gopi-backend

# Stop Docker container
docker-stop:
	@echo "Stopping Docker container..."
	docker stop gopi-backend || true
	docker rm gopi-backend || true

# Show Docker container logs
docker-logs:
	docker logs -f gopi-backend

# Docker Compose targets
.PHONY: docker-compose-up docker-compose-down docker-compose-build docker-compose-logs

# Start all services with Docker Compose
docker-compose-up:
	@echo "Starting all services with Docker Compose..."
	docker compose up -d

# Stop all services with Docker Compose
docker-compose-down:
	@echo "Stopping all services with Docker Compose..."
	docker compose down

# Build all services with Docker Compose
docker-compose-build:
	@echo "Building all services with Docker Compose..."
	docker compose build

# Show Docker Compose logs
docker-compose-logs:
	docker compose logs -f

# Clean Docker artifacts
docker-clean:
	@echo "Cleaning Docker artifacts..."
	docker system prune -f
	docker volume prune -f

# Show help
help:
	@echo "Available commands:"
	@echo "  build          - Build the Go server (runs user tests first)"
	@echo "  run            - Run the built server"
	@echo "  dev            - Start development server with hot reload (requires Air)"
	@echo "  swag           - Generate Swagger documentation"
	@echo "  test           - Run all tests"
	@echo "  test-verbose   - Run all tests with verbose output"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-user      - Run only user module tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-short     - Run only short tests"
	@echo "  test-clean     - Clean test artifacts"
	@echo "  fmt            - Format Go code"
	@echo "  lint           - Run linter (requires golangci-lint)"
	@echo "  clean          - Clean build artifacts"
	@echo "  build-windows  - Build Windows executable"
	@echo "  package-windows - Build Windows executable and create complete package"
	@echo "  clean-windows  - Clean Windows build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-stop    - Stop Docker container"
	@echo "  docker-logs    - Show Docker container logs"
	@echo "  docker-compose-up    - Start all services with Docker Compose"
	@echo "  docker-compose-down  - Stop all services with Docker Compose"
	@echo "  docker-compose-build - Build all services with Docker Compose"
	@echo "  docker-compose-logs  - Show Docker Compose logs"
	@echo "  docker-clean   - Clean Docker artifacts"
	@echo "  help           - Show this help message"

.PHONY: build run clean dev swag test test-verbose test-race test-coverage test-user test-integration test-short test-clean fmt lint help build-windows package-windows clean-windows docker-build docker-run docker-stop docker-logs docker-clean docker-compose-up docker-compose-down docker-compose-build docker-compose-logs