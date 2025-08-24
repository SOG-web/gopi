build:
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

.PHONY: build run clean