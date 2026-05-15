# Nimbus Security Architecture

This document outlines the security mechanisms implemented across the Nimbus monorepo to protect the public gateway (`zapp`) and the local home server (`kif`).

## 1. SSH Protection (Fail2ban)

Because `zapp` is a public-facing VPS, its SSH port (22) is constantly exposed to automated botnets and brute-force scanning.
To mitigate this, **Fail2ban** is installed on all servers via the `common` Ansible role.
- **Aggressive Mode:** The `sshd` jail is configured in `aggressive` mode. This means Fail2ban not only blocks IPs after multiple failed login attempts but also bans IPs that exhibit suspicious behavior (like connecting and dropping immediately, a common hallmark of botnet scanners).
- **Log Parsing:** It reads `/var/log/auth.log` and dynamically injects `iptables`/UFW rules to ban malicious IPs.

*Note:* While `kif` is not directly exposed to the internet, Fail2ban is active there as a defense-in-depth measure against internal network compromise.

## 2. Docker Firewall Isolation (UFW-Docker)

By default, Docker directly modifies `iptables` to route traffic to containers. This completely bypasses UFW (Uncomplicated Firewall) rules. If a container binds to `0.0.0.0:8080`, it is publicly accessible even if UFW is set to deny all incoming traffic.

To secure this, we use [ufw-docker](https://github.com/chaifeng/ufw-docker), which forces Docker traffic to respect your UFW rules.
- **Zapp:** Ensures that only explicitly allowed ports (like 80, 443 for Traefik) are accessible from the internet. Any new containers spun up by Dokploy will remain private unless explicitly routed.
- **Kif:** Prevents dozens of internal services (like Homepage, AdGuard) from being blindly exposed to the entire network. To ensure local access continues working (e.g., via `*.home` domains), UFW is configured to explicitly allow routing from the local LAN subnet (`192.168.1.0/24`) and the Tailscale interface (`tailscale0`).

## 3. HTTP Application Protection (CrowdSec)

To protect web applications deployed on `zapp` (like apps managed via Dokploy) from application-level attacks (e.g., vulnerability scraping, SQL Injection, XSS), we deploy **CrowdSec**.

CrowdSec is a modern, collaborative intrusion prevention system:
- **Engine (`src/stacks/zapp/crowdsec`):** Runs as a Docker container, parsing Traefik's access logs and system authentication logs to detect malicious behavior. 
- **WAF / AppSec:** The engine is configured with the **CrowdSec AppSec (WAF)** component. This allows for deep inspection of HTTP requests beyond simple IP-based banning, protecting apps against specific exploit payloads (e.g., SQLi, Path Traversal). 
- **Dynamic Acquisition:** Ansible dynamically retrieves the unique Docker log path of the `dokploy-traefik` container to ensure the engine only monitors relevant web traffic, avoiding recursive parsing of its own logs.
- **Log Formatting:** Traefik is configured to use the **Common Log Format** (text-based) for access logs. This ensures maximum compatibility and performance with CrowdSec's high-speed Grok parsers and avoids common JSON unmarshaling errors.
- **Traefik Integration:** Integrates with Dokploy's Traefik via the `crowdsec-bouncer-traefik-plugin`. It intercepts requests at the gateway—if the engine or the WAF flags an IP or a payload, Traefik instantly drops the connection with a `403 Forbidden` error.
- **Global Intelligence:** It syncs with a community-driven global blocklist, preemptively banning decentralized botnets that have attacked other CrowdSec users worldwide.
