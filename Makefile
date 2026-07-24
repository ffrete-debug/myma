.PHONY: build test lint clean run docker-build docker-up docker-down tidy fmt help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

check: ## Run vet + fmt check
	@cd server && go vet ./... && test -z "$(gofmt -l .)" || (gofmt -l .; exit 1)
	@echo "check passed"

build: ## Build backend binary
	cd server && go build -o ../bin/ark-commander .

test: ## Run all tests
	cd server && go test ./... -v

lint: ## Run go vet
	cd server && go vet ./...

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
