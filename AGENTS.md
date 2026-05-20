# AI Agent Instructions (`AGENTS.md`)

This document defines the strict, non-negotiable guidelines for any AI agent interacting with the `nimbus` codebase. Read it fully before performing any work.

For detailed architecture documentation (networking, DNS, security, etc.), see the [`docs/`](docs/) directory.

---

## Part 1: General Agent Guidelines
*These rules are standard across all projects. Do not modify or bypass them.*

### Tooling & Automation Rules
- **`make` is the Source of Truth**: Never run raw tool commands when a `make` target exists. Run `make help` (or just `make` if help is the default) to see a list of available targets.
- **Pre-flight Validation**: Always run lint/syntax checks (e.g., `make check` or equivalent local checks) before proposing code changes, commits, or deployments.
- **Permission Check**: Always ask the user for explicit permission before modifying infrastructure, executing deployments, or provisioning via `make`.

### Git & Commit Conventions
- **Conventional Commits**: All commit messages must follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification.
  - **Format**: `<type>(<scope>): <description>` (e.g., `feat(mealie): add healthcheck configuration`).
  - **Allowed Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`.
- **No Direct Commits**: The agent must never commit or push code directly. Instead, the agent must suggest a Git commit message conforming to the Conventional Commits specification whenever they complete a task or suggest code changes that should be committed.

### The "Do Not" List (Negative Constraints)
- **Do not hallucinate files or states**: Verify paths exist before reading or writing.
- **Do not use emojis anywhere**: Keep all communication, responses, code comments, commit messages, and PR titles clean and professional.
- **Do not use em dashes (`—`)**: Use a regular dash (`-`) instead.
- **Do not read remote decrypted secrets**: NEVER read production `.env` files or any other files containing decrypted secrets on the remote servers. All configuration should be understood and managed via local Ansible variables and templates.
- **Do not output secrets to terminal**: When generating new secrets, NEVER output them to the terminal. Always pipe the generator output directly into `ansible-vault encrypt_string` (e.g., `openssl rand -base64 24 | tr -d '/+' | ansible-vault encrypt_string --stdin-name <var_name>`) to prevent secrets from being recorded in session history or logs.
- **Do not leave documentation out of date**: If changes affect the project's behavior, infrastructure (DNS, networking, security), or setup, you must update the relevant documentation.

---

## Part 2: Project Context & Conventions
*These rules are specific to the `nimbus` homelab monorepo.*

### Project References

Always consult these directories first for context, standards, and setup instructions:

| File / Directory | Purpose |
|---|---|
| [`README.md`](README.md) | Project overview and general architecture links |
| [`docs/`](docs/) | Detailed architecture documentation (networking, DNS, security, VPN, etc.) |
| [`src/stacks/`](src/stacks/) | Docker Compose stack configurations organized by server |
| [`src/ansible/`](src/ansible/) | Ansible playbooks, roles, and tasks for deployment |
| [`Makefile`](Makefile) | Main developer automation tasks |

### Monorepo Architecture

This is a **homelab monorepo** managing multiple servers:

| Server | Type | Provider | Role |
|---|---|---|---|
| **Zapp** | VPS | Hetzner Cloud | Public gateway (Dokploy, Headscale, Rathole) |
| **Kif** | Home Server | Local | Home automation, media, internal services |

- **Hosting Strategy**: Zapp acts as the public gateway and hosts lightweight web apps. Kif is the powerhouse; all heavy or internal self-hosted services (e.g. Home Assistant, Stremio, media) should be deployed to Kif's resources. Detailed hardware specs are in [`docs/architecture.md`](docs/architecture.md).
- **Traffic Tunneling**: Rathole tunnels traffic from Zapp (public IP) → Kif (behind NAT) for exposing select home services.
- **VPN Mesh**: Headscale/Tailscale provides VPN mesh between servers.
- **Domain & Routing Split**:
  - **Internal Services**: Use `*.pipusznicy.cloud`, accessible only via Tailscale or LAN. Resolved by AdGuard Home on Kif, and reverse-proxied via Kif's Caddy.
  - **External Services**: Publicly accessible via `*.pawelad.me` (resolves to Zapp's public IP). Traffic routes: Public Internet → Zapp Traefik → Rathole Tunnel → Kif Rathole Client → Service Container (bypassing Kif's Caddy).
- **Remote Access**: All commands targeting a production server (e.g., `docker`, `ls`, file operations) MUST be run via SSH using the server alias (e.g. `ssh zapp "docker ps"` or `ssh kif "docker ps"`).

### Infrastructure & Deployment Conventions

#### Deployment Mechanism
- **Git + Ansible Deploy**:
  1. Developers run `make <server>-deploy` to push the `main` branch to the server and deploy stacks.
  2. `git push` updates the working tree via `receive.denyCurrentBranch=updateInstead`.
  3. `ansible-playbook {server}_deploy.yml` runs `docker compose up -d` for each stack.
- **Deployment Rule**: NEVER use `scp` or direct file copying to push changes to a server. ALL changes to files managed in this repository MUST be deployed via `make <server>-deploy` (e.g., `make kif-deploy TAGS=adguard`) to ensure the server repository stays synchronized.
- **Configuration Division**: Configuration (env vars, secrets) is managed by Ansible templates. Application definition (Docker Compose) is managed by Git in `src/stacks`.

#### Data Directory & Organization
- **Data Directory**: All persistent application data MUST be stored in `/data/apps/<app-name>` to maintain a consistent backup and storage structure across servers.
- **Stack Organization**: Stacks are organized under `src/stacks/<server-name>/<stack-name>/compose.yaml`. Stacks shared across multiple servers are placed in `src/stacks/common/<stack-name>/` and symlinked into each server's directory (e.g., `src/stacks/zapp/tailscale-agent` -> `../common/tailscale-agent`).
  - **Server Paths**: The repo is cloned to `/data/nimbus/` and `stacks_dir` resolves to `/data/nimbus/src/stacks/<server-name>/`.
  - **Convenience Symlink**: `/data/stacks` → `stacks_dir`.

#### Docker Compose Rules
- **File Naming**: All Docker Compose files MUST be named `compose.yaml`. `docker-compose.yml` and `docker-compose.yaml` are forbidden.
- **No Standalone Container Management**: NEVER use standalone container management tools or modules (like `community.docker.docker_container` or raw `docker restart` commands) to restart, recreate, or change the state of containers managed by Docker Compose. Always use `community.docker.docker_compose_v2` with `state: restarted` and target the `project_src` instead.
  - *Reason*: Standalone container commands modify the container outside Compose's context, stripping or modifying Compose project labels (e.g. `com.docker.compose.project`). This orphans the container, causing subsequent `docker compose up` deployments to fail with name conflicts.
- **Service Naming**: Service names in `compose.yaml` (keys under `services:`) SHOULD use generic, role/engine-based names (e.g., `backend`, `frontend`, `postgres`, `redis`, `valkey`, `mongo`) rather than stack-prefixed names. When the image itself is named `*-web` (e.g., `multica-web`), prefer `web` as the service name (and `<stack>-web` as the container name).
  - *Exception*: For multi-app stacks (e.g., `stremio`) where a database (e.g., `postgres`) is only used by one specific app (e.g., `comet`), keep the service and container name specific (e.g., `comet-postgres`) to prevent ambiguity.
- **Container Naming**: Every service in `compose.yaml` MUST specify a unique `container_name` prefixed with the stack name (e.g., `container_name: <stack-name>-<role>`). When the stack contains multiple services, the main application container's `container_name` SHOULD be `<stack>-app` (e.g., `monetr-app`).
- **Docker Networks**:
  - Zapp: Uses `dokploy-network` (created by Dokploy, used by stacks that need reverse proxy).
  - Kif: Uses `caddy` network (created by Caddy reverse proxy).
  - Stacks on each server must reference the correct network.
- **Exposed Ports**: Before assigning a host port in a `compose.yaml` file (e.g., `- "0.0.0.0:8082:8080"`), always verify that the host port is not already taken by searching the repository (e.g., searching across `src/stacks`).
- **Docker Compose Healthchecks**:
  - ALWAYS verify which networking tools (`curl`, `wget`, etc.) are actually installed in the container image before committing healthcheck configuration.
  - ALWAYS use loopback IP addresses (like `127.0.0.1`) instead of `localhost` in healthcheck URLs.
- **Docker Image Versions**: Docker images MUST be pinned to specific versions (e.g., `image: binwiederhier/ntfy:v2.16.0`). NEVER use `:latest`.
  - *Exception*: Images that have no versioned tags (e.g., Comet) should be annotated with `# dclint disable-line`.

#### Ansible & YAML Formatting
- **Ansible Exec Modules**: When executing commands inside running Docker containers from Ansible playbooks, ALWAYS use the `community.docker.docker_container_exec` module instead of raw shell/command modules with `docker exec`.
- **Standalone Containers**: When managing standalone Docker containers that are NOT managed by Docker Compose (e.g., `dokploy-traefik`), ALWAYS use the native declarative `community.docker.docker_container` module in Ansible rather than raw command or shell modules.
- **YAML Formatting**: All `name` and `notify` values in Ansible tasks/handlers MUST be double-quoted (e.g., `- name: "Install package"`).
- **Template Organization**: Ansible templates for stacks are organized under `roles/stacks/templates/<server-name>/`. Shared templates are in `roles/stacks/templates/common/`. Task files use `{{ inventory_hostname }}/template.j2` for server-specific templates, or `common/template.j2` for shared ones.
- **Jinja2 Defaults**: NEVER use Jinja2 default values for infrastructure-critical configuration (e.g., domains, IPs, paths). Safe, optional defaults (e.g., `| default('false')`) are allowed to reduce boilerplate.
- **Secrets Management**: Plaintext secrets must never be committed. Use Ansible Vault variables. Use `make encrypt-string` to generate encrypted strings.

#### Caddy Labels (Kif Stacks)
All web-exposed internal Kif stacks MUST use the dual-TLS Caddy label pattern:
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

#### Reusable Ansible Tasks
Two reusable task files in `roles/common/tasks/` form the core deployment building blocks:
- **`ensure_stack`**: Deploys a Docker Compose stack. Pass `stack_name` and optionally `env_template`.
- **`ensure_config`**: Manages templated config files with drift detection. Supports forced overwrite via `make <server>-deploy TAGS=<stack> FORCE=1`.

#### Tooling Rules
- Tools in `src/tools` SHOULD have a `Makefile` with standard targets.
  - **Go tools** (i.e. `tretter-getter`): `make build`, `make check`, `make test`, `make docker-build`, `make help`.
  - **Python tools** (i.e. `gatus_sync`, `uptime_kuma_sync`): `make run`, `make format`, `make help` (run via `uv run`).
- Nested tools are excluded from the root `make check` to keep global linting fast. When modifying code under `src/tools/`, you MUST run the local `make check` (or equivalent) within that specific tool's subdirectory.

### Adding a New Kif Stack
When adding a new service to Kif, the following files must be created/updated:

1. **Compose file**: `src/stacks/kif/<stack-name>/compose.yaml`
   - Join the `caddy` network (external).
   - Add Caddy labels for reverse proxy (see Caddy Labels above).
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

---

## Pre-flight Checklist
*Before concluding any task, verify each item:*

- [ ] Did I run `make check` (or local tool-specific checks) and did it pass cleanly?
- [ ] If I assigned a host port, did I search `src/stacks` to verify it is not already in use?
- [ ] Are all Docker images pinned to specific version tags (no `:latest`), or marked with `# dclint disable-line` if no tags exist?
- [ ] Did I verify that loopback IP addresses (like `127.0.0.1`) instead of `localhost` are used in all healthcheck URLs?
- [ ] Did I check which networking tools (`curl`, `wget`, etc.) are actually installed in the container image before committing a healthcheck?
- [ ] Are all `name` and `notify` fields in Ansible tasks double-quoted?
- [ ] If adding a Kif stack, did I configure the dual-TLS Caddy labels correctly?
- [ ] If the stack needs a custom domain alias, did I update the AdGuard Home template at `src/ansible/roles/stacks/templates/kif/AdGuardHome.yaml.j2`?
- [ ] If changes affect DNS, VPN, networking, security, or stack deployment, did I update the relevant `docs/` files and `AGENTS.md`?
- [ ] Did I suggest a Git commit message conforming to the Conventional Commits specification?
