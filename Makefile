# Source: http://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
.DEFAULT_GOAL := help

export EXTRA_VARS
export TAGS
# FORCE=1 is a shorthand for Ansible playbooks that injects 'overwrite_config=$(TAGS)'
# into EXTRA_VARS. Example: 'make zapp-deploy TAGS=headplane FORCE=1'
export FORCE

.PHONY: install
install: ## Install necessary dependencies for all components
	$(MAKE) -C src/ansible install
	$(MAKE) -C src/terraform init
	$(MAKE) -C src/tools/tretter-getter tidy

.PHONY: check
check: ## Run code linters
	$(MAKE) -C src/ansible check
	$(MAKE) -C src/terraform format
	yamllint .
	npx dclint -r src/stacks

.PHONY: kif-deploy
kif-deploy: ## Deploy changes to Kif (use EXTRA_VARS for variables, TAGS for tags)
	git push kif main
	$(MAKE) -C src/ansible kif-deploy

.PHONY: kif-provision
kif-provision: ## Provision Kif server (use EXTRA_VARS for variables, TAGS for tags)
	$(MAKE) -C src/ansible kif-provision

.PHONY: zapp-deploy
zapp-deploy: ## Deploy changes to Zapp (use EXTRA_VARS for variables, TAGS for tags)
	git push zapp main
	$(MAKE) -C src/ansible zapp-deploy

.PHONY: zapp-provision
zapp-provision: ## Provision Zapp server (use EXTRA_VARS for variables, TAGS for tags)
	$(MAKE) -C src/ansible zapp-provision

.PHONY: beryl-provision
beryl-provision: ## Provision Beryl AX travel router
	$(MAKE) -C src/ansible beryl-provision

.PHONY: encrypt-string
encrypt-string: ## Encrypt a value with Ansible Vault
	$(MAKE) -C src/ansible encrypt-string

.PHONY: tf-plan
tf-plan: ## Generate a (speculative) Terraform plan
	$(MAKE) -C src/terraform plan

# Source: https://www.client9.com/self-documenting-makefiles/
.PHONY: help
help: ## Show help message
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {\
	printf "\033[36m%-40s\033[0m %s\n", $$1, $$NF \
	}' $(MAKEFILE_LIST)
