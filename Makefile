# Source: http://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
.DEFAULT_GOAL := help

EXTRA_VARS ?=
TAGS ?=
FORCE ?=

# FORCE=1 is a shorthand that injects overwrite_config=$(TAGS) into EXTRA_VARS, e.g.:
#   make zapp-deploy TAGS=headplane FORCE=1
# is equivalent to:
#   make zapp-deploy EXTRA_VARS='overwrite_config=headplane' TAGS=headplane
_EXTRA_VARS = $(strip $(EXTRA_VARS) $(if $(FORCE),overwrite_config=$(TAGS)))

.PHONY: kif-deploy
kif-deploy: ## Deploy changes to Kif (use EXTRA_VARS for variables, TAGS for tags)
	git push kif main
	cd src/ansible && ansible-playbook playbooks/kif_deploy.yml $(if $(_EXTRA_VARS),-e '$(_EXTRA_VARS)') $(if $(TAGS),--tags '$(TAGS)')

.PHONY: kif-provision
kif-provision: ## Provision Kif server with Ansible (use EXTRA_VARS for variables, TAGS for tags)
	cd src/ansible && ansible-playbook playbooks/kif_setup.yml $(if $(_EXTRA_VARS),-e '$(_EXTRA_VARS)') $(if $(TAGS),--tags '$(TAGS)')

.PHONY: zapp-deploy
zapp-deploy: ## Deploy changes to Zapp (use EXTRA_VARS for variables, TAGS for tags)
	git push zapp main
	cd src/ansible && ansible-playbook playbooks/zapp_deploy.yml $(if $(_EXTRA_VARS),-e '$(_EXTRA_VARS)') $(if $(TAGS),--tags '$(TAGS)')

.PHONY: zapp-provision
zapp-provision: ## Provision Zapp server with Ansible (use EXTRA_VARS for variables, TAGS for tags)
	cd src/ansible && ansible-playbook playbooks/zapp_setup.yml $(if $(_EXTRA_VARS),-e '$(_EXTRA_VARS)') $(if $(TAGS),--tags '$(TAGS)')

.PHONY: check
check: tf-format ## Run code linters
	cd src/ansible && ansible-playbook --syntax-check playbooks/*.yml
	cd src/ansible && ansible-lint --yamllint-file=../../.yamllint  --exclude=collections/
	yamllint .
	npx dclint --fix -r src/stacks

.PHONY: encrypt-string
encrypt-string: ## Encrypt a value with Ansible Vault
	@read -p "Enter variable name: " name; \
	echo "Enter secret value (press Ctrl+D to end):"; \
	cd src/ansible && ansible-vault encrypt_string --name "$$name"

.PHONY: tf-plan
tf-plan: ## Generate a (speculative) Terraform plan
	terraform -chdir=src/terraform plan

.PHONY: tf-apply
tf-apply: ## Generate, confirm and apply a Terraform plan
	terraform -chdir=src/terraform apply

.PHONY: tf-upgrade
tf-upgrade: ## Upgrade Terraform providers
	terraform -chdir=src/terraform init -upgrade

.PHONY: tf-destroy
tf-destroy: ## Destroy infrastructure managed by Terraform
	terraform -chdir=src/terraform destroy

.PHONY: tf-format
tf-format: ## Format Terraform files
	terraform -chdir=src/terraform fmt
	terraform -chdir=src/terraform validate

# Source: https://www.client9.com/self-documenting-makefiles/
.PHONY: help
help: ## Show help message
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {\
	printf "\033[36m%-40s\033[0m %s\n", $$1, $$NF \
	}' $(MAKEFILE_LIST)
