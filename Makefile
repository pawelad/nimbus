# Source: http://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
.DEFAULT_GOAL := help

.PHONY: plan
plan: ## Generate a (speculative) Terraform plan
	terraform -chdir=src plan

.PHONY: apply
apply: ## Generate, confirm and apply a Terraform plan
	terraform -chdir=src apply

.PHONY: upgrade
upgrade: ## Upgrade Terraform providers
	terraform -chdir=src init -upgrade

.PHONY: destroy
destroy: ## Destroy infrastructure managed by Terraform
	terraform -chdir=src destroy

.PHONY: format
format: ## Format Terraform files
	terraform -chdir=src fmt
	terraform -chdir=src validate

# Source: https://www.client9.com/self-documenting-makefiles/
.PHONY: help
help: ## Show help message
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {\
	printf "\033[36m%-40s\033[0m %s\n", $$1, $$NF \
	}' $(MAKEFILE_LIST)
