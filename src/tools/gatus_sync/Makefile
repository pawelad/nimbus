# Source: http://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
.DEFAULT_GOAL := help

.PHONY: run
run: ## Run the tool
	uv run gatus_sync.py

.PHONY: format
format: ## Format code
	uv format

.PHONY: check
check: format ## Check the code
	uvx ruff check gatus_sync.py
	# Unset VIRTUAL_ENV so 'ty' resolves imports from the ephemeral environment instead of the host's active virtualenv
	env -u VIRTUAL_ENV uvx --with-requirements gatus_sync.py ty check

.PHONY: help
help: ## Show help message
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {\
	printf "\033[36m%-40s\033[0m %s\n", $$1, $$NF \
	}' $(MAKEFILE_LIST)
