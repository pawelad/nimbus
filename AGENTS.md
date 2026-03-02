# Project Rules

This file documents the conventions and rules that must be followed when working on this project.

## Conventions

### Makefile Usage
- **Rule**: Always look for a `Makefile` in the current directory or parent directories and use it whenever possible instead of running individual commands.
- **Reason**: Ensures consistency and reduces errors by using predefined workflows.
- **Discovery**: Run `make help` (or just `make` if help is the default) to see a list of available targets and their descriptions.
- **Common Examples**:
    - Root `Makefile`: Used for project-wide tasks like `make check` (linting), `make provision`, and `make deploy`.

### Secrets Management
- **Rule**: NEVER commit plain text secrets. Use Ansible Vault.
- **Tool**: Use `make encrypt-string` to generate encrypted values for Ansible variables.

### YAML Formatting
- **Rule**: All `name` and `notify` values in Ansible tasks/handlers MUST be double-quoted.
- **Reason**: Consistent style and avoids potential YAML parsing issues with special characters.
- **Example**: `- name: "Install package"` instead of `- name: Install package`

### Remote Access
- **Rule**: All commands targeting the production server (e.g., `docker`, `ls`, file operations) MUST be run via SSH using the `nimbus` alias.
- **Example**: `ssh nimbus "docker ps"` instead of `docker ps`.
