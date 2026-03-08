# Source: http://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
.DEFAULT_GOAL := help

.PHONY: build
build: build-templ build-sqlc ## Build the binary
	go build -o bin/tretter-getter .

.PHONY: build-templ
build-templ: ## Generate Templ code
	templ generate

.PHONY: build-sqlc
build-sqlc: ## Generate SQLC code
	sqlc generate

.PHONY: run
run: ## Run the tool
	go run .

.PHONY: run-dry
run-dry: ## Run the tool in dry-run mode
	DRY_RUN=true go run .

.PHONY: run-debug
run-debug: ## Run the tool with debug logging
	LOG_LEVEL=debug go run .

.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t tretter-getter:latest .

.PHONY: test
test: ## Run unit tests
	go test -v -cover ./...

.PHONY: check
check: ## Run linters
	go vet ./...
	staticcheck ./...

.PHONY: format
format: ## Format code
	go fmt ./...
	go fix ./...
	goimports -w .

.PHONY: tidy
tidy: ## Tidy dependencies
	go mod tidy

.PHONY: help
help: ## Show help message
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {\
	printf "\033[36m%-40s\033[0m %s\n", $$1, $$NF \
	}' $(MAKEFILE_LIST)
