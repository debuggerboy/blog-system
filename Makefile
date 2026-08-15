.PHONY: help run build clean generate

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

generate: ## Generate templ files
	@echo "Generating templ files..."
	templ generate ./...

run: generate ## Run the server
	@echo "Starting server..."
	go run cmd/server/main.go

build: generate ## Build the binary
	@echo "Building..."
	go build -o bin/blog-server cmd/server/main.go

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf views/*_templ.go

install: ## Install dependencies
	@echo "Installing dependencies..."
	go mod tidy
	go install github.com/a-h/templ/cmd/templ@latest

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

.DEFAULT_GOAL := help
