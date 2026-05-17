# Source: http://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
.DEFAULT_GOAL := help

.PHONY: init
init: ## Initialize Terraform
	terraform init

.PHONY: plan
plan: ## Generate a (speculative) Terraform plan
	terraform plan -lock=false

.PHONY: apply
apply: ## Generate, confirm and apply a Terraform plan
	terraform apply

.PHONY: upgrade
upgrade: ## Upgrade Terraform providers
	terraform init -upgrade

.PHONY: destroy
destroy: ## Destroy infrastructure managed by Terraform
	terraform destroy

.PHONY: format
format: ## Format Terraform files
	terraform fmt
	terraform validate

# Source: https://www.client9.com/self-documenting-makefiles/
.PHONY: help
help: ## Show help message
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {\
	printf "\033[36m%-40s\033[0m %s\n", $$1, $$NF \
	}' $(MAKEFILE_LIST)
