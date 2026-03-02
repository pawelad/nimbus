# Project Rules

This file documents the conventions and rules that must be followed when working on this project.

## Architecture

### Monorepo Structure
This is a **homelab monorepo** managing multiple servers:

| Server | Type | Provider | Role |
|--------|------|----------|------|
| **Zapp** | VPS (cx22: 2 vCPU, 4GB RAM) | Hetzner Cloud | Public-facing services (Dokploy, Headscale, Rathole server) |
| **Kif** | NUC home server | Local | Home automation, media, internal services (26+ stacks) |

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
│   │   ├── docker/             # Docker installation
│   │   ├── git_deploy/         # Git push-to-deploy setup
│   │   └── stacks/             # Stack deployment (per-host task files). Role name is "stacks".
│   └── tasks/                  # Reusable task files (ensure_stack, ensure_config)
├── stacks/
│   ├── zapp/                   # Zapp's Docker Compose stacks
│   │   ├── dokploy/
│   │   ├── headscale/
│   │   ├── monitoring/
│   │   └── rathole/
│   └── kif/                    # Kif's Docker Compose stacks (TODO: migrate from kif repo)
│       ├── adguard/
│       ├── homepage/
│       ├── komodo/
│       └── ...                 # 26+ stacks
├── terraform/                  # Infrastructure as Code
│   ├── hetzner.tf              # Zapp server
│   ├── cloudflare.tf           # DNS zones
│   └── ...
└── tools/                      # Custom tools (TODO: migrate from kif repo)
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
- **Path on server**: The repo is cloned to `/data/<server-name>/` and `stacks_dir` resolves to `/data/<server-name>/src/stacks/<server-name>/`.
- **Symlink**: `/data/stacks` → `stacks_dir` for convenience.

### Architecture & Deployment
- **Concept**: "Git + Ansible Deploy"
- **Mechanism**:
    1. Developers run `make deploy-<server>` to push the `main` branch to the server and deploy stacks.
    2. `git push` updates the working tree via `receive.denyCurrentBranch=updateInstead`.
    3. `ansible-playbook {server}_deploy.yml` runs `docker compose up -d` for each stack.
- **Rule**: Configuration (env vars, secrets) is managed by Ansible templates. Application definition (Docker Compose) is managed by Git in `src/stacks`.

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
    - Root `Makefile`: Used for project-wide tasks like `make check` (linting), `make provision-zapp`, and `make deploy-zapp`.

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
