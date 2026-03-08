# Source: http://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
.DEFAULT_GOAL := help

.PHONY: run
run: ## Run the tool
	uv run uptime_kuma_sync.py

.PHONY: run-dry
run-dry: ## Run the tool in dry-run mode
	uv run uptime_kuma_sync.py --dry-run

.PHONY: format
format: ## Format code
	uv format

.PHONY: help
help: ## Show help message
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {\
	printf "\033[36m%-40s\033[0m %s\n", $$1, $$NF \
	}' $(MAKEFILE_LIST)
