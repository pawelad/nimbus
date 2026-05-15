# Kif Server (Home Server)

Kif represents the internal home infrastructure node in the Nimbus architecture. It is a local server (NUC i5-7260U, 16GB RAM) running behind a home router with NAT. It manages heavy workloads, private data, and media applications.

## Key Properties
- **Hardware Profile:** Local hardware (Intel NUC).
- **Network Scope:** Private network segment (e.g., `192.168.1.9`), accessible securely from the outside via the Tailscale/Headscale mesh VPN or via specific tunneled ports.
- **Operating System Packages:** Inherits common features (`acl`, `git`, `htop`, etc.) and specific additions (`bluez`) from Ansible.

## Core Setup and Management
- **Ansible Driven:** Configured with `kif.yml` inventory and specific roles. Deployed using `make kif-deploy`.
- **Docker Compose Stacks:** Hosted in `src/stacks/kif/`, comprising multiple distinct services.

## Application Architecture

### Traffic Flow and Routing
- **Caddy (Reverse Proxy):** Acts as the internal reverse proxy for all web services. Traffic routes through the `caddy` Docker network.
- **Valid HTTPS (DNS-01 Challenge):** Kif uses `acme.sh` in a sidecar container within the Caddy stack to manage a Wildcard Let's Encrypt certificate for `*.pipusznicy.cloud` via the Cloudflare DNS-01 challenge. The certificate files are shared with Caddy via a bind mount (`/data/apps/acme/certs`). This allows fully valid, browser-trusted HTTPS across all internal apps without requiring manual CA trust on devices.
- **Hybrid TLS Strategy:** Stacks use numbered Caddy labels (`caddy_0`, `caddy_1`) to support both the canonical FQDNs with Let's Encrypt certificates (e.g., `komodo.pipusznicy.cloud` via `caddy_0`) and local `.home` / `.pipusznicy` domains with Caddy's internal CA (via `caddy_1`). The ACME cert paths (`ACME_TLS_CERT`, `ACME_TLS_KEY`) are injected into each stack's `.env` file via a shared Ansible template (`common/acme_cert.env.j2`).
- **AdGuard Home (Local DNS):** Handles internal queries by rewriting domain requests. It utilizes a **Dual-IP Resolution** strategy: for each internal domain (`*.home`, `*.pipusznicy`, `*.pipusznicy.cloud`), it returns both the Tailscale IP (listed first, preferred for remote Tailscale clients) and the local LAN IP (listed second, fallback for LAN-only devices). The Tailscale IP is preferred because remote devices (e.g., on a travel router at an Airbnb) could otherwise collide with devices on common subnets like `192.168.1.0/24`. The Tailscale IP is discovered dynamically during deployment using `tailscale ip -4`.

### ACME Certificate Lifecycle
The wildcard Let's Encrypt certificate for `*.pipusznicy.cloud` is managed by an `acme.sh` sidecar container in the `caddy` stack:

1. **Issuance:** On first deploy, the `cert-helper` stack runs `acme.sh --issue` using the Cloudflare DNS-01 challenge (`CF_Token` + `CF_Zone_ID` env vars). The certificate is written to `/data/apps/acme/certs/`.
2. **Renewal:** The `acme.sh` container runs as a daemon (`command: daemon`) and automatically renews the certificate before expiry.
3. **Distribution:** Caddy reads the certificate via a read-only bind mount (`/data/apps/acme/certs:/data/acme-certs:ro`). Each stack's `.env` file receives `ACME_TLS_CERT` and `ACME_TLS_KEY` paths via the shared `common/acme_cert.env.j2` Ansible template.
4. **Usage in stacks:** Services reference the cert in their Caddy labels: `caddy_0.tls: "${ACME_TLS_CERT} ${ACME_TLS_KEY}"`.

## Home Network Setup
- **Router:** Synology RT6600ax — acts as the home gateway, DHCP server, and Wi-Fi access point.
- **DNS Configuration:** The router's DHCP settings push Kif's LAN IP (`192.168.1.9`) as the primary DNS server for all devices on the home network. This means every device on the home Wi-Fi automatically uses AdGuard Home on Kif for DNS resolution — no per-device configuration needed.
- **How it works:** When a LAN device queries `komodo.pipusznicy.cloud` (or `komodo.home`), the query goes to AdGuard on Kif, which rewrites it to Kif's own LAN IP. Caddy then handles the reverse proxy routing and serves the response with a valid Let's Encrypt wildcard certificate (for `.pipusznicy.cloud` domains) or an internal CA certificate (for `.home` domains).
- **Search Domains:** The Synology RT6600ax does not support pushing DHCP search domains, so bare hostnames (e.g., typing just `kif` in a browser) won't automatically resolve on LAN devices. Use the full domain instead (e.g., `komodo.pipusznicy.cloud` or `komodo.home`). Tailscale clients get `home` and `pipusznicy.cloud` as search domains via Headscale's `search_domains` config.

## Networking Nuances
- To maintain privacy, most services are not exposed beyond the internal home network and the Tailscale overlay network.
- **Firewall (UFW)**: Kif is configured to allow all incoming traffic from the `tailscale0` interface, ensuring that devices on the Tailnet can reach internal services and the AdGuard DNS server without friction.
- **Rathole**: A few selected services are exposed to the internet via Rathole. Rathole client on Kif establishes a connection to Zapp (the remote exit node) and forwards specific requests directly to the container, bypassing Caddy entirely.
