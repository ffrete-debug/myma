.PHONY: build test check lint clean run docker-build docker-up docker-down tidy fmt help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

check: ## Run vet + fmt check
	@echo ">> go vet"
	@cd server && go vet ./...
	@echo ">> gofmt -l"
	@cd server && unformatted="$$(gofmt -l .)"; \
		if [ -n "$$unformatted" ]; then \
			echo "The following files are not gofmt-formatted:"; \
			echo "$$unformatted" | sed 's|^|  |'; \
			echo "Run 'make fmt' to fix."; \
			exit 1; \
		fi
	@echo "check passed"

build: ## Build backend binary
	cd server && go build -o ../bin/ark-commander .

test: ## Run all tests
	cd server && go test ./... -v

lint: ## Run golangci-lint (falls back to go vet if not installed)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo ">> golangci-lint run"; \
		cd server && golangci-lint run ./...; \
	else \
		echo "golangci-lint not found; falling back to 'go vet'."; \
		echo "CI runs golangci-lint, so install it to match: https://golangci-lint.run/welcome/install/"; \
		cd server && go vet ./...; \
	fi

clean: ## Remove build artifacts
	rm -rf bin/

run: ## Run backend locally
	cd server && go run .

docker-build: ## Build Docker image
	docker build -t ark-commander .

docker-up: ## Start Docker Compose
	docker-compose up -d

docker-down: ## Stop Docker Compose
	docker-compose down

tidy: ## Tidy Go modules
	cd server && go mod tidy

fmt: ## Format Go code
	cd server && go fmt ./...
