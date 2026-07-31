# Stremio & Media Architecture

Nimbus hosts a robust media processing pipeline on `kif`, integrating Stremio addons with a focus on streaming efficiency and access control. 

## The Core Concept: How Stremio Works
A common misconception is that Stremio addons stream the actual video data. **They do not.**
When you watch a movie, the architecture works like this:
1. **The Catalog Request:** The TV's Stremio app asks the addon (`stremio.pawelad.me`) for a list of available streams for a specific movie.
2. **The Resolution:** The addon (like AIOStreams or Comet) queries an indexer (Prowlarr) or a Debrid provider (Real-Debrid) to find the cached video files.
3. **The Response:** The addon sends back a JSON payload to the TV containing **direct links to the Debrid provider's servers**.
4. **The Stream:** The TV plays the video directly from the public internet (Real-Debrid's high-speed CDN). The video data never touches `zapp` or `kif`. 

## Remote Accessibility via Rathole
Because the actual video streaming handles itself via the public internet, the only piece that needs to be accessible to your TVs (whether at home or in a hotel on the travel router) is the JSON Addon API itself (`aiostreams`).

To achieve this without opening ports on `kif`'s home router:
1. **Rathole Tunnel:** The Rathole Client on `kif` creates a secure outbound TCP connection to the Rathole Server on `zapp`, forwarding traffic directly to the local `aiostreams` Docker container on port `3000` (bypassing the internal Caddy proxy).
2. **Zapp Public Proxy:** The Rathole Server exposes the tunnel securely on its internal Docker network (`3030`). Zapp's Traefik reverse proxy intercepts HTTPS traffic for `stremio.pawelad.me`, attaches a Let's Encrypt certificates natively, and routes it into the Rathole tunnel.

**Result:** You enter `https://stremio.pawelad.me/manifest.json` into Stremio on any device, anywhere in the world, and it seamlessly queries `kif`'s AIOStreams container to fetch the media catalog, and then the TV streams the heavy video files directly from your Debrid provider.

## Local Services
Other supplementary services (Comet, Debrid Media Manager, Prowlarr) are internal — accessible only via the LAN or Tailnet using `*.pipusznicy.cloud` FQDNs (with valid Let's Encrypt wildcard certificates) or local `.home` domain aliases.

> [!NOTE]
> Stream generator addons like Comet require `PUBLIC_BASE_URL` explicitly configured (e.g. `PUBLIC_BASE_URL: "https://comet.pipusznicy.cloud"`) in `compose.yaml`. This ensures that even when aggregated internally via AIOStreams using Docker networking (`http://comet:8000`), Comet generates stream playback URLs using the resolvable FQDN rather than internal Docker hostnames.

