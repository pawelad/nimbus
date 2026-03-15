# Source: http://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
.DEFAULT_GOAL := help

EXTRA_VARS ?=
TAGS ?=
FORCE ?=

# FORCE=1 is a shorthand that injects overwrite_config=$(TAGS) into EXTRA_VARS
_EXTRA_VARS = $(strip $(EXTRA_VARS) $(if $(FORCE),overwrite_config=$(TAGS)))

.PHONY: install
install: ## Install Ansible collections from requirements.yml
	ansible-galaxy collection install -r requirements.yml -p ./collections

.PHONY: check
check: ## Run Ansible syntax check and linting
	ansible-playbook --syntax-check playbooks/*.yml
	ansible-lint --yamllint-file=../../.yamllint --exclude=collections/

.PHONY: kif-deploy
kif-deploy: ## Deploy changes to Kif (use EXTRA_VARS for variables, TAGS for tags)
	ansible-playbook playbooks/kif_deploy.yml $(if $(_EXTRA_VARS),-e '$(_EXTRA_VARS)') $(if $(TAGS),--tags '$(TAGS)')

.PHONY: kif-provision
kif-provision: ## Provision Kif server (use EXTRA_VARS for variables, TAGS for tags)
	ansible-playbook playbooks/kif_provision.yml $(if $(_EXTRA_VARS),-e '$(_EXTRA_VARS)') $(if $(TAGS),--tags '$(TAGS)')

.PHONY: zapp-deploy
zapp-deploy: ## Deploy changes to Zapp (use EXTRA_VARS for variables, TAGS for tags)
	ansible-playbook playbooks/zapp_deploy.yml $(if $(_EXTRA_VARS),-e '$(_EXTRA_VARS)') $(if $(TAGS),--tags '$(TAGS)')

.PHONY: zapp-provision
zapp-provision: ## Provision Zapp server (use EXTRA_VARS for variables, TAGS for tags)
	ansible-playbook playbooks/zapp_provision.yml $(if $(_EXTRA_VARS),-e '$(_EXTRA_VARS)') $(if $(TAGS),--tags '$(TAGS)')

.PHONY: beryl-provision
beryl-provision: ## Provision Beryl AX travel router
	ansible-playbook playbooks/beryl_provision.yml

.PHONY: encrypt-string
encrypt-string: ## Encrypt a value with Ansible Vault
	@read -p "Enter variable name: " name; \
	echo "Enter secret value (press Ctrl+D to end):"; \
	ansible-vault encrypt_string --name "$$name"

# Source: https://www.client9.com/self-documenting-makefiles/
.PHONY: help
help: ## Show help message
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {\
	printf "\033[36m%-40s\033[0m %s\n", $$1, $$NF \
	}' $(MAKEFILE_LIST)
