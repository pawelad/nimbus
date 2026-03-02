# Source: http://clarkgrubb.com/makefile-style-guide
MAKEFLAGS += --warn-undefined-variables
.DEFAULT_GOAL := help

.PHONY: check
check: ## Run code linters
	cd src/ansible && ansible-playbook --syntax-check playbooks/*.yml
	cd src/ansible && ansible-lint --yamllint-file=../../.yamllint  --exclude=collections/
	yamllint .
	npx dclint --fix -r src/stacks

.PHONY: provision
provision: ## Provision server with Ansible (use EXTRA_VARS for variables, TAGS for tags)
	cd src/ansible && ansible-playbook playbooks/server_setup.yml $(if $(EXTRA_VARS),-e '$(EXTRA_VARS)') $(if $(TAGS),--tags '$(TAGS)')

.PHONY: deploy
deploy: ## Deploy changes to production (use EXTRA_VARS for variables, TAGS for tags)
	git push nimbus main
	cd src/ansible && ansible-playbook playbooks/deploy_stacks.yml $(if $(EXTRA_VARS),-e '$(EXTRA_VARS)') $(if $(TAGS),--tags '$(TAGS)')

.PHONY: encrypt-string
encrypt-string: ## Encrypt a value with Ansible Vault
	@read -p "Enter variable name: " name; \
	echo "Enter secret value (press Ctrl+D to end):"; \
	cd src/ansible && ansible-vault encrypt_string --name "$$name"

.PHONY: server-reboot
server-reboot: ## Reboot the server
	cd src/ansible && ansible all -m ansible.builtin.reboot --become

.PHONY: server-shutdown
server-shutdown: ## Shutdown the server
	cd src/ansible && ansible all -a "/usr/bin/systemctl poweroff" --become

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
