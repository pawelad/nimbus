# Nimbus Architecture

This document provides a high-level overview of the Nimbus homelab monorepo architecture. 

The infrastructure is split across two primary nodes: a public-facing VPS (`zapp`) and a powerful local home server (`kif`). This split allows for secure public exposure of select services while ensuring heavy workloads and private data remain within the home network.

## The Nodes

### 1. Zapp (The Gateway)
- **Role:** Public-facing gateway and lightweight app host.
- **Provider:** Hetzner Cloud VPS.
- **Key Services:**
  - **Dokploy:** PaaS for managing lightweight web apps.
  - **Traefik:** Reverse proxy handling public traffic routing and Let's Encrypt TLS certificates. Runs standalone (via `docker run`) to preserve true client IP addresses instead of Docker Swarm's ingress network masking.
  - **Headscale & Headplane:** The control plane and web UI for the private Tailscale VPN.
  - **Rathole (Server):** Secure tunneling software used to expose specific services from the private `kif` server to the public internet (e.g., Stremio add-ons).

### 2. Kif (The Powerhouse)
- **Role:** Home automation, media, and internal services host.
- **Provider:** Local NUC home server (behind NAT).
- **Key Services:**
  - **Caddy:** Internal reverse proxy automatically routing traffic for local services. Paired with an `acme.sh` sidecar that manages a Let's Encrypt wildcard certificate for `*.pipusznicy.cloud` via Cloudflare DNS-01 challenges. Uses a **dual-TLS label strategy**: each service gets two Caddy site blocks (`caddy_0` for the FQDN with the LE wildcard cert, `caddy_1` for local `.home`/`.pipusznicy` domains with Caddy's internal CA).
  - **AdGuard Home:** Local DNS server handling DNS rewrites for custom domains (`*.pipusznicy.cloud`, `*.home`, `*.pipusznicy`). Uses a **dual-IP resolution** strategy, returning both the LAN IP (preferred) and Tailscale IP for each domain.
  - **Application Stacks:** Home automation (Home Assistant), monitoring (Gatus, Komodo, Beszel), notifications (ntfy, Apprise), utilities (Stirling PDF, MeTube, Ollama), and more.
  - **Rathole (Client):** Connects to `zapp` to tunnel local services (like `aiostreams` on port `3000`) so they can be exposed publicly by Traefik.

## Networking & DNS
The connectivity and name resolution bridge the two worlds:

1. **Public DNS (Cloudflare):** Three zones are managed via Terraform: `pawelad.me` (public-facing services + personal), `pawelad.dev` (open-source projects), and `pipusznicy.cloud` (homelab infrastructure). A wildcard `*.pawelad.me` DNS record points to `zapp`'s public IP address, and Traefik routes incoming traffic based on the subdomain (e.g., `vpn.`, `stremio.`). `zapp.pipusznicy.cloud` also points to Zapp's public IP. There is no public A record for `*.pipusznicy.cloud` — these domains resolve only via internal DNS.
2. **Private Network (Tailscale/Headscale):** All devices (servers, laptops, phones, travel routers) join a private mesh VPN overlay network via Headscale, establishing secure, direct, peer-to-peer Wireguard tunnels. MagicDNS base domain is `net.pipusznicy.cloud` (e.g., `kif.net.pipusznicy.cloud`). A Headscale `extra_records` entry provides `kif.pipusznicy.cloud` as a tailnet-only shortcut.
3. **Internal DNS (AdGuard + Headscale Split DNS):** Kif serves as the primary DNS server for the home network. AdGuard Home on Kif handles DNS rewrites for custom zones. To prevent conflicts with Tailscale's MagicDNS (`net.pipusznicy.cloud`), AdGuard uses an exception rule (`@@||net.pipusznicy.cloud^$dnsrewrite`) that explicitly excludes the MagicDNS domain from any rewrites, ensuring VPN-internal addresses are correctly passed through to Headscale.
4. **Domain Split Philosophy:** `pipusznicy.cloud` is the canonical domain for internal/tailnet-only homelab services (e.g., `komodo.pipusznicy.cloud`), while `pawelad.me` handles public-facing services and personal sites (e.g., `mealie.pawelad.me`). Local `.home` and `.pipusznicy` domains are preserved as aliases for backward compatibility.

## Deployment Strategy
The monorepo embraces a "Git + Ansible Deploy" philosophy:
1. Application configurations (Docker Compose files) are stored in `src/stacks/<server>` (or `src/stacks/common` with symlinks).
2. Secrets are managed securely via Ansible Vault (`group_vars` and `host_vars`).
3. Deployments use a Makefile (`make zapp-deploy`, `make kif-deploy`), which pushes the Git repository to the server and runs Ansible playbooks to template `.env` files and run `docker compose up -d`.

## Infrastructure as Code (Terraform)
The public infrastructure is managed via Terraform, deployed through [Spacelift](https://spacelift.io/). Three Cloudflare DNS zones (`pawelad.me`, `pawelad.dev`, `pipusznicy.cloud`) and the Hetzner VPS (`zapp`) are defined in `src/terraform/`. Changes are planned with `make tf-plan` and applied automatically by Spacelift on merge.

## Security
Nimbus employs several layers of security to protect both public and private infrastructure, including **Fail2ban** (SSH brute-force protection), **UFW-Docker** (Firewall enforcement for containers), and **CrowdSec** (Collaborative HTTP application protection). 

For detailed information on these implementations, see the [Security Architecture](security.md) documentation.
