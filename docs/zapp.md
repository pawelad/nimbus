# Zapp Server (Public Gateway)

Zapp represents the public-facing gateway and lightweight computing node in the Nimbus architecture. It runs on a Hetzner Cloud VPS (cx23: 2 vCPU, 4GB RAM) with a public static IP address.

## Key Properties
- **Hardware Profile:** Cloud Infrastructure (Hetzner).
- **Network Scope:** Public internet access.
- **Role:** Internet gateway, VPN control plane, and host for lightweight or publicly accessible web applications.

## Core Setup and Management
- **Ansible Driven:** Configured with `zapp.yml` inventory and specific playbook roles. Deployed using `make zapp-deploy`. 
- **Docker Compose Stacks:** Hosted in `src/stacks/zapp/`. Dokploy itself is installed natively on the host (not managed as a stack in this repo); it leverages Docker Swarm concepts internally (`dokploy-postgres`, `dokploy-redis`). The remaining stacks are standard Docker Compose.

## Application Architecture

### Traffic Flow and Routing
- **Cloudflare DNS:** Root domain (`pawelad.me`) and a wildcard (`*.pawelad.me`) point to Zapp's IP. Additionally, `zapp.pipusznicy.cloud` provides a direct DNS shortcut on the homelab domain.
- **Traefik (Reverse Proxy):** Installed natively via Dokploy's install script (running as a `docker run` container instead of a Swarm service) to accurately capture explicit incoming client IP addresses. It automatically intercepts HTTP/HTTPS requests on ports 80/443, provisions Let's Encrypt TLS certificates, and routes traffic over the `dokploy-network` based on Docker labels.

### Noteworthy Services
1. **Dokploy (PaaS):** Orchestrates and manages deployment workflows through a unified interface.
2. **Headscale (Control Plane):** The master authentication and key-distribution node for the private Tailscale VPN mesh. Also runs an embedded DERP relay server (STUN on UDP `3478`) for NAT traversal fallback.
3. **Headplane (Dashboard):** A rich UI interface for managing Headscale's nodes, API keys, and DNS configurations.
4. **Rathole (Server):** Acts as the public termination point for secure reverse tunnels from the `kif` server. It listens publicly (`2333`) for tunnel connections and internally (`3030`) for Traefik routing to forward specific services to the web (e.g., proxying traffic for Stremio addons via `stremio.pawelad.me`).
5. **Monitoring Extensions:** Tailscale Agent, Beszel Agent, Komodo Periphery, Glances.

## Let's Encrypt TLS Certificates
Traefik on `zapp` leverages HTTP-01 challenges to validate domain ownership (`stremio.pawelad.me`, `vpn.pawelad.me`, etc.) and seamlessly establish automatic, public, trusted green-padlock HTTPS encryption.

## Security Context
As a public-facing server, Zapp is hardened against automated internet threats. It utilizes **Fail2ban** for aggressive SSH protection, **UFW-Docker** to ensure Docker containers don't arbitrarily bypass firewall rules, and **CrowdSec** to protect HTTP web apps (like Dokploy-managed applications) against vulnerability scraping and DDoS attacks. 

See the detailed [Security Architecture](security.md) document for more information.
