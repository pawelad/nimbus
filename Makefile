# Source: http://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
.DEFAULT_GOAL := help

.PHONY: plan
plan: ## Generate a (speculative) Terraform plan
	terraform -chdir=tf plan

.PHONY: apply
apply: ## Generate, confirm and apply a Terraform plan
	terraform -chdir=tf apply

.PHONY: destroy
destroy: ## Destroy infrastructure managed by Terraform
	terraform -chdir=tf destroy

.PHONY: format
format: ## Format Terraform files
	terraform -chdir=tf fmt
	terraform -chdir=tf validate

# Source: https://www.client9.com/self-documenting-makefiles/
.PHONY: help
help: ## Show help message
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {\
	printf "\033[36m%-40s\033[0m %s\n", $$1, $$NF \
	}' $(MAKEFILE_LIST)