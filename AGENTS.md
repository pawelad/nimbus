# Project Rules

This file documents the conventions and rules that must be followed when working on this project.

For detailed architecture documentation (networking, DNS, security, etc.), see the [`docs/`](docs/) directory.

## Architecture

### Monorepo Structure
This is a **homelab monorepo** managing multiple servers:

| Server | Type                               | Provider | Role |
|--------|------------------------------------|----------|------|
| **Zapp** | VPS (cx23: 2 vCPU, 4GB RAM)        | Hetzner Cloud | Public-facing services (Dokploy, Headscale, Rathole server) |
| **Kif** | NUC home server (i5-7260U, 16GB RAM) | Local | Home automation, media, internal services |

The relationship and hosting strategy between them:
- **Hosting Strategy**: Zapp acts as the public gateway and hosts lightweight web apps. Kif is the powerhouse; all heavy or internal self-hosted services (e.g. Home Assistant, Stremio, media) should be deployed to Kif's 16GB pool.
- **Rathole** tunnels traffic from Zapp (public IP) → Kif (behind NAT) for exposing select home services
- **Headscale/Tailscale** provides VPN mesh between servers
- **Beszel + Komodo** monitor both servers

## Conventions

### Data Directory
- **Rule**: All persistent application data MUST be stored in `/data/apps/<app-name>`.
- **Reason**: To maintain a consistent backup and storage structure across servers.

### Stack Organization
- **Rule**: Stacks are organized under `src/stacks/<server-name>/<stack-name>/compose.yaml`.
- **Rule (Unified)**: Stacks shared across multiple servers are placed in `src/stacks/common/<stack-name>/` and symlinked into each server's directory (e.g., `src/stacks/zapp/tailscale-agent` -> `../common/tailscale-agent`).
- **Path on server**: The repo is cloned to `/data/nimbus/` and `stacks_dir` resolves to `/data/nimbus/src/stacks/<server-name>/`.
- **Symlink**: `/data/stacks` → `stacks_dir` for convenience.

### Architecture & Deployment
- **Concept**: "Git + Ansible Deploy"
- **Mechanism**:
    1. Developers run `make <server>-deploy` to push the `main` branch to the server and deploy stacks.
    2. `git push` updates the working tree via `receive.denyCurrentBranch=updateInstead`.
    3. `ansible-playbook {server}_deploy.yml` runs `docker compose up -d` for each stack.
- **Rule**: Configuration (env vars, secrets) is managed by Ansible templates. Application definition (Docker Compose) is managed by Git in `src/stacks`.
- **Rule**: NEVER use `scp` or direct file copying to push changes to a server. ALL changes to files managed in this repository MUST be deployed via `make <server>-deploy` (e.g., `make kif-deploy TAGS=adguard`) to ensure the server repository stays synchronized.
- **Rule**: Always ask the user for permission before modifying infrastructure, deploying code, or provisioning via `make`.
- **Rule**: Always run `make check` to validate syntax and linting before doing any deploy, provision, or git commit actions.

### Docker Compose
- **Rule**: All Docker Compose files MUST be named `compose.yaml`.
- **Reason**: Official Docker recommendation and matches current project consistency.
- **Forbidden**: `docker-compose.yml`, `docker-compose.yaml`.
- **Rule**: NEVER use standalone container management tools or modules (like `community.docker.docker_container` or raw `docker restart` commands) to restart, recreate, or change the state of containers managed by Docker Compose. Always use `community.docker.docker_compose_v2` with `state: restarted` and target the `project_src` instead.
- **Reason**: Standalone container commands modify the container outside of Compose's context, stripping or modifying Compose project labels (e.g. `com.docker.compose.project`). This orphans the container, causing Subsequent `docker compose up` deployments to fail with name conflicts (e.g. `Conflict. The container name "/<name>" is already in use`).

### Docker Networks
- **Zapp**: Uses `dokploy-network` (created by Dokploy, used by stacks that need reverse proxy).
- **Kif**: Uses `caddy` network (created by Caddy reverse proxy).
- Stacks on each server must reference the correct network.

### Makefile Usage
- **Rule**: Always look for a `Makefile` in the current directory or parent directories and use it whenever possible instead of running individual commands.
- **Reason**: Ensures consistency and reduces errors by using predefined workflows.
- **Discovery**: Run `make help` (or just `make` if help is the default) to see a list of available targets and their descriptions.
- **Common Examples**:
    - Root `Makefile`: Used for project-wide tasks like `make check` (linting), `make zapp-provision`, `make zapp-deploy`, `make kif-provision`, `make kif-deploy`.
    - `src/tools/tretter-getter/Makefile`: Used for tool-specific tasks like `make build`, `make test`.

### Docker Compose Exposed Ports
- **Rule**: Before assigning a host port in a `compose.yaml` file (e.g., `- "0.0.0.0:8082:8080"`), always verify that the host port is not already taken by searching the repository (e.g., searching across `src/stacks`).
- **Reason**: Prevents port conflicts when deploying multiple stacks to the same host.

### Secrets Management
- **Rule**: NEVER commit plain text secrets. Use Ansible Vault.
- **Rule**: When generating new secrets, NEVER output them to the terminal. Always pipe the generator output directly into `ansible-vault encrypt_string` (e.g., `openssl rand -base64 24 | tr -d '/+' | ansible-vault encrypt_string --stdin-name <var_name>`) to prevent secrets from being recorded in session history or logs.
- **Rule**: NEVER read production `.env` files or any other files containing decrypted secrets on the remote servers. All configuration should be understood and managed via local Ansible variables and templates.
- **Tool**: Use `make encrypt-string` to generate encrypted values for Ansible variables.

### YAML Formatting
- **Rule**: All `name` and `notify` values in Ansible tasks/handlers MUST be double-quoted.
- **Reason**: Consistent style and avoids potential YAML parsing issues with special characters.
- **Example**: `- name: "Install package"` instead of `- name: Install package`

### Remote Access
- **Rule**: All commands targeting a production server (e.g., `docker`, `ls`, file operations) MUST be run via SSH using the server alias.
- **Example**: `ssh zapp "docker ps"`, `ssh kif "docker ps"`, or `ssh beryl "tailscale status"`.

### Docker Compose Healthchecks
- **Rule**: When writing healthchecks, ALWAYS verify which networking tools (`curl`, `wget`, etc.) are actually installed in the container image before committing the code.
- **Reason**: Prevents false unhealthy states due to missing commands.
- **Rule**: ALWAYS use loopback IP addresses (like `127.0.0.1`) instead of `localhost` in healthcheck URLs.
- **Reason**: BusyBox/Alpine-based images often prioritize IPv6 loopback (`::1`) when resolving `localhost`, which fails with `Connection refused` if the service (e.g. Next.js/Node) only listens on the IPv4 loopback interface (`127.0.0.1`).

### Template Organization
- **Rule**: Ansible templates for stacks are organized under `roles/stacks/templates/<server-name>/`. Shared templates are in `roles/stacks/templates/common/`.
- **Reference**: Task files use `{{ inventory_hostname }}/template.j2` for server-specific templates, or `common/template.j2` for shared ones.

### Tooling
- **Rule**: Tools in `src/tools` SHOULD have a `Makefile` with standard targets.
- **Go tools** (`tretter-getter`): `make build`, `make check`, `make test`, `make docker-build`, `make help`.
- **Python tools** (`gatus_sync`, `uptime_kuma_sync`): `make run`, `make format`, `make help`. Run via `uv run`.
- **Reason**: Consistent developer experience across different tools.

### Documentation
- **Rule**: When adding a new stack, server, or changing infrastructure (DNS, networking, security), update the relevant `docs/` files and this `AGENTS.md`.
- **DNS/domain changes**: Update `docs/architecture.md` (Networking & DNS section) and `docs/vpn.md` (Split DNS).
- **Security changes**: Update `docs/security.md`.
- **Reason**: Documentation drifts from reality quickly. Keeping it in sync during the change is far easier than auditing later.

### Git & Commits
- **Rule**: All commits to this repository MUST follow the **Conventional Commits** specification.
- **Format**: `<type>(<scope>): <description>` (e.g., `feat(mealie): add healthcheck configuration`).
- **Allowed Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`.
- **Reason**: Maintains a clean, searchable git history and facilitates automated changelog generation.

### Adding a New Kif Stack
When adding a new service to Kif, the following files must be created/updated:

1. **Compose file**: `src/stacks/kif/<stack-name>/compose.yaml`
   - Join the `caddy` network (external).
   - Add Caddy labels for reverse proxy (see Caddy Labels below).
   - Add Homepage service discovery labels.
   - Pin the Docker image version.
   - Add a healthcheck.
   - Store persistent data in `/data/apps/<stack-name>`.

2. **Ansible task file**: `src/ansible/roles/stacks/tasks/kif/<stack_name>.yml`
   - Use `import_role: common / ensure_stack` to deploy the stack.
   - If the stack needs a templated `.env`, set `env_template`. For stacks that only need ACME cert paths, use `env_template: "common/acme_cert.env.j2"`. If the stack has its own `.env.j2` template but also needs ACME cert paths (e.g., for Caddy labels), include the shared template at the top of the file using `{% include 'common/acme_cert.env.j2' %}`.
   - If the stack needs templated config files, use `import_role: common / ensure_config` before `ensure_stack`.

3. **Register in deployment**: Add an `import_tasks` entry in `src/ansible/roles/stacks/tasks/kif_stacks.yml` with the appropriate tag.

4. **Ansible template** (if needed): `src/ansible/roles/stacks/templates/kif/<template>.j2`

5. **AdGuard DNS** (if the stack needs a custom domain alias): Update the AdGuard Home template at `src/ansible/roles/stacks/templates/kif/AdGuardHome.yaml.j2`.

### Caddy Labels (Kif Stacks)
- **Rule**: All web-exposed Kif stacks MUST use the dual-TLS Caddy label pattern:
  ```yaml
  labels:
    # FQDN with Let's Encrypt wildcard cert
    caddy_0: <service>.pipusznicy.cloud
    caddy_0.tls: ${ACME_TLS_CERT} ${ACME_TLS_KEY}
    caddy_0.reverse_proxy: "{{upstreams <port>}}"
    # Local domains with Caddy internal CA
    caddy_1: <service>.home <service>.pipusznicy
    caddy_1.tls: internal
    caddy_1.reverse_proxy: "{{upstreams <port>}}"
  ```
- **Reason**: Ensures services are accessible via both canonical FQDNs (valid LE cert for browsers) and local short domains (internal CA).

### Reusable Ansible Tasks
Two reusable task files in `roles/common/tasks/` form the core deployment building blocks:

- **`ensure_stack`**: Deploys a Docker Compose stack. Pass `stack_name` and optionally `env_template`. Supports `force_deploy` and `healthcheck_wait` parameters. Sets a global `stack_changed` flag when containers are recreated.
- **`ensure_config`**: Manages templated config files with drift detection. Seeds the file on first deploy, warns on drift during subsequent deploys, and supports forced overwrite via `make <server>-deploy TAGS=<stack> FORCE=1`.

### Ansible Templates
- **Rule**: NEVER use Jinja2 default values for infrastructure-critical configuration (e.g., domains, IPs, paths). 
- **Exception**: Safe, optional defaults (e.g., `| default('false')` for toggles or `| default('')` for optional API keys) are allowed to reduce boilerplate in `host_vars`.
- **Reason**: Critical configuration should be explicit to avoid hidden drift, while optional settings can stay concise.

### Docker Image Versions
- **Rule**: Docker images MUST be pinned to specific versions (e.g., `image: binwiederhier/ntfy:v2.16.0`). NEVER use `:latest`.
- **Exception**: Images that have no versioned tags (e.g., Comet) should be annotated with `# dclint disable-line`.
- **Reason**: Reproducible deployments. Prevents unexpected breaking changes during `docker compose up`.
