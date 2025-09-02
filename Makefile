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
	@echo "  help           - Show this help message"

.PHONY: build run clean dev swag test test-verbose test-race test-coverage test-user test-integration test-short test-clean fmt lint help