# Project Rules

This file documents the conventions and rules that must be followed when working on this project.

## Architecture

### Monorepo Structure
This is a **homelab monorepo** managing multiple servers:

| Server | Type                               | Provider | Role |
|--------|------------------------------------|----------|------|
| **Zapp** | VPS (cx22: 2 vCPU, 4GB RAM)        | Hetzner Cloud | Public-facing services (Dokploy, Headscale, Rathole server) |
| **Kif** | NUC home server (i5-7260U, 16GB RAM | Local | Home automation, media, internal services (35 stacks) |

The relationship between them:
- **Rathole** tunnels traffic from Zapp (public IP) → Kif (behind NAT) for exposing select home services
- **Headscale/Tailscale** provides VPN mesh between servers
- **Beszel + Komodo** monitor both servers

### Directory Layout
```
src/
├── ansible/                    # All Ansible configuration
│   ├── inventories/            # Per-host inventory files (zapp.ini, kif.ini)
│   ├── group_vars/
│   │   ├── all.yml             # Shared variables & cross-server secrets
│   │   ├── zapp.yml            # Zapp-specific variables & secrets
│   │   └── kif.yml             # Kif-specific variables & secrets
│   ├── playbooks/              # Playbooks, named as {host}_{action}.yml
│   ├── roles/
│   │   ├── common/             # Shared server setup (users, packages, UFW, dirs)
│   │   │   └── tasks/          # main.yml + composable: bluetooth, data_disk, journald, disable_default_dns
│   │   ├── docker/             # Docker + containerd installation
│   │   ├── git_deploy/         # Git push-to-deploy setup
│   │   ├── samba/              # Samba file sharing (kif only)
│   │   └── stacks/             # Stack deployment (per-host task files)
│   │       └── templates/      # Per-server: templates/{zapp,kif}/*.j2
│   └── tasks/                  # Reusable task files (ensure_stack, ensure_config)
├── stacks/
│   ├── zapp/                   # Zapp's Docker Compose stacks
│   │   ├── dokploy/
│   │   ├── headscale/
│   │   └── rathole/
│   └── kif/                    # Kif's Docker Compose stacks
│       ├── adguard/
│       ├── homepage/
│       ├── komodo/
│       └── ...                 # 35 stacks
├── terraform/                  # Infrastructure as Code
│   ├── hetzner.tf              # Zapp server
│   ├── cloudflare.tf           # DNS zones
│   └── ...
└── tools/                      # Custom tools
    ├── tretter-getter/         # Go app
    ├── gatus_sync/             # Python script
    └── uptime_kuma_sync/       # Python script
```

## Conventions

### Data Directory
- **Rule**: All persistent application data MUST be stored in `/data/apps/<app-name>`.
- **Reason**: To maintain a consistent backup and storage structure across servers.

### Stack Organization
- **Rule**: Stacks are organized under `src/stacks/<server-name>/<stack-name>/compose.yaml`.
- **Path on server**: The repo is cloned to `/data/nimbus/` and `stacks_dir` resolves to `/data/nimbus/src/stacks/<server-name>/`.
- **Symlink**: `/data/stacks` → `stacks_dir` for convenience.

### Architecture & Deployment
- **Concept**: "Git + Ansible Deploy"
- **Mechanism**:
    1. Developers run `make deploy-<server>` to push the `main` branch to the server and deploy stacks.
    2. `git push` updates the working tree via `receive.denyCurrentBranch=updateInstead`.
    3. `ansible-playbook {server}_deploy.yml` runs `docker compose up -d` for each stack.
- **Rule**: Configuration (env vars, secrets) is managed by Ansible templates. Application definition (Docker Compose) is managed by Git in `src/stacks`.
- **Rule**: NEVER use `scp` or direct file copying to push changes to a server. ALL changes to files managed in this repository MUST be deployed via `make deploy-<server>` (e.g., `make deploy-kif TAGS=adguard`) to ensure the server repository stays synchronized.

### Docker Compose
- **Rule**: All Docker Compose files MUST be named `compose.yaml`.
- **Reason**: Official Docker recommendation and matches current project consistency.
- **Forbidden**: `docker-compose.yml`, `docker-compose.yaml`.

### Docker Networks
- **Zapp**: Uses `dokploy-network` (created by Dokploy, used by stacks that need reverse proxy).
- **Kif**: Uses `caddy` network (created by Caddy reverse proxy).
- Stacks on each server must reference the correct network.

### Makefile Usage
- **Rule**: Always look for a `Makefile` in the current directory or parent directories and use it whenever possible instead of running individual commands.
- **Reason**: Ensures consistency and reduces errors by using predefined workflows.
- **Discovery**: Run `make help` (or just `make` if help is the default) to see a list of available targets and their descriptions.
- **Common Examples**:
    - Root `Makefile`: Used for project-wide tasks like `make check` (linting), `make provision-zapp`, `make deploy-zapp`, `make provision-kif`, `make deploy-kif`.
    - `src/tools/tretter-getter/Makefile`: Used for tool-specific tasks like `make build`, `make test`.

### Docker Compose Exposed Ports
- **Rule**: Before assigning a host port in a `compose.yaml` file (e.g., `- "0.0.0.0:8082:8080"`), always verify that the host port is not already taken by searching the repository (e.g., searching across `src/stacks`).
- **Reason**: Prevents port conflicts when deploying multiple stacks to the same host.

### Secrets Management
- **Rule**: NEVER commit plain text secrets. Use Ansible Vault.
- **Tool**: Use `make encrypt-string` to generate encrypted values for Ansible variables.

### YAML Formatting
- **Rule**: All `name` and `notify` values in Ansible tasks/handlers MUST be double-quoted.
- **Reason**: Consistent style and avoids potential YAML parsing issues with special characters.
- **Example**: `- name: "Install package"` instead of `- name: Install package`

### Remote Access
- **Rule**: All commands targeting a production server (e.g., `docker`, `ls`, file operations) MUST be run via SSH using the server alias.
- **Example**: `ssh zapp "docker ps"` or `ssh kif "docker ps"`.

### Docker Compose Healthchecks
- **Rule**: When writing healthchecks, always verify which networking tools (`curl`, `wget`, etc.) are actually installed in the container image.
- **Reason**: Prevents false unhealthy states due to missing commands.

### Template Organization
- **Rule**: Ansible templates for stacks are organized per-server under `roles/stacks/templates/<server-name>/`.
- **Reference**: Task files use `{{ inventory_hostname }}/template.j2` so the same task works for any server.

### Tooling
- **Rule**: Tools in `src/tools` SHOULD have a `Makefile` with standard targets.
- **Go tools** (`tretter-getter`): `make build`, `make check`, `make test`, `make docker-build`, `make help`.
- **Python tools** (`gatus_sync`, `uptime_kuma_sync`): `make run`, `make format`, `make help`. Run via `uv run`.
- **Reason**: Consistent developer experience across different tools.
