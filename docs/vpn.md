# Virtual Private Network (VPN)

The Nimbus infrastructure utilizes Headscale and Tailscale to create a secure, peer-to-peer mesh VPN overlay network.

## Core Components
- **Headscale (`zapp`):** The self-hosted Tailscale control server. It handles authentication, assigns IPs (`100.64.0.0/10`), and distributes public keys (the "Netmap"). It also runs an **embedded DERP relay server** (see below). It **does not** route regular VPN traffic itself. MagicDNS base domain is `net.pipusznicy.cloud`, so nodes are reachable as `kif.net.pipusznicy.cloud`, `zapp.net.pipusznicy.cloud`, etc.
- **Headplane (`zapp`):** User-friendly web dashboard for managing Headscale devices, ACLs, and DNS settings.
- **Tailscale Agents:** The actual VPN clients running on servers (`kif`, `zapp`), personal endpoints (laptops, phones), and travel routers. They use the Netmap to establish direct, peer-to-peer Wireguard tunnels.

## DERP Relay
Tailscale normally creates direct peer-to-peer Wireguard tunnels between devices. However, when both peers are behind strict/symmetric NATs and direct connection fails, traffic falls back to a **DERP relay server**. Zapp runs an embedded DERP relay (region `zapp`, STUN on UDP port `3478`) so that relay traffic stays within our own infrastructure instead of routing through Tailscale's public DERP nodes. Tailscale's public DERP map is also included as a fallback. Only authenticated Tailnet clients can use the relay (`verify_clients: true`).

## Authentication & Enrollment
Devices join the Tailnet using two primary methods:
1. **Interactive Login (Personal Devices):** Running `tailscale up --login-server https://vpn.pawelad.me` generates a URL. You open it in a browser, log in via Headplane, and approve the device. Best for laptops and phones.
2. **Pre-Auth Keys (Servers/Automated):** Reusable auth keys (`tailscale_authkey` in Ansible Vault) allow automated deployments (like `kif`'s ansible setup) to join the network without manual intervention.

## Routing and DNS

### Split Tunneling (Internal vs External Traffic)
Tailscale defaults to split tunneling, which creates a clear distinction between traffic types:
- **Internal Traffic**: Traffic destined for other Tailscale devices (`100.64.x.y`) or specific routed subnets is encrypted and routed over the peer-to-peer VPN mesh.
- **External Traffic**: General internet traffic (e.g., Netflix, YouTube) uses the device's local network gateway directly. This preserves bandwidth and ensures that VPN overhead doesn't slow down regular browsing.

### Split DNS & Hybrid Resolution
Instead of forcing 100% of your internet's DNS queries through the VPN (which could add significant global latency if you are traveling globally), Headscale utilizes a highly efficient **Split DNS Architecture**:
- When devices connect to the Tailnet, their general internet queries (like `google.com`) use the fastest local DNS provider globally (`1.1.1.1` or the native ISP).
- However, we dynamically inject Kif's Tailscale IP into Headscale's `dns.split` block for two zones: `home` and `pipusznicy.cloud`. This means queries for `*.home` and `*.pipusznicy.cloud` are routed to Kif's AdGuard over the VPN.
- **Dual-IP Strategy**: To ensure reliability, AdGuard Home on **Kif** returns both the Tailscale IP (`100.x.x.x`, listed first) and the local LAN IP (`192.168.1.9`, listed second) for internal domains. The Tailscale IP is prioritized because remote devices (e.g., on a travel router at an Airbnb with a `192.168.1.0/24` subnet) would otherwise try to connect to a random device on the foreign network before falling back. LAN-only devices at home (smart TVs, IoT) gracefully fall back to the LAN IP.
- When travel devices ask for internal domains (i.e. `*.home` or `*.pipusznicy.cloud`), Headscale seamlessly intercepts the request and fires it down the secure Wireguard tunnel exclusively to Kif's AdGuard server.
- **Extra Records**: Headscale's `extra_records` feature provides `kif.pipusznicy.cloud` as a tailnet-only DNS shortcut pointing to Kif's Tailscale IP. This is resolved directly by Headscale's MagicDNS for all tailnet clients.
- **Search Domains**: `home` and `pipusznicy.cloud` are configured as search domains, allowing users to reach apps via their short name (e.g. `komodo.home` or just `komodo` from a Tailscale device).

## Client Devices Setup

### Home Server (`kif`)
Deployed via the `tailscale-agent` stack through Ansible and authenticated seamlessly via reusable pre-auth keys.

### Travel Router (Beryl)
The travel router (GL-MT3000) acts as a transparent gateway for all connected devices.
- **NAT Masquerade**: Traffic from the router's Wi-Fi clients is "masked" as the router's own VPN IP (`100.x.x.x`). This ensures high reliability and zero routing conflicts even if hotel Wi-Fi subnets change.
- **Full Transparency**: Devices connected to the Beryl's Wi-Fi (TVs, Laptops, Phones) can access `*.home` services instantly without needing Tailscale installed natively.
- **Provisioning**: Configuration is automated via `make beryl-provision`.

### Personal Devices (Phones / Macs)
Installing the official Tailscale client directly on end devices is recommended for best performance. It allows direct Wireguard peer-to-peer connections (bypassing the router's CPU bottleneck) and remains connected everywhere (coffee shops, cell data) without needing the travel router.
