# nimbus
My personal homelab monorepo, managing infrastructure across two servers with [Ansible], [Terraform], and [Docker Compose].

## Architecture
The infrastructure is split across two primary nodes:

| Server   | Type                        | Provider        | Role                                                 |
|----------|-----------------------------|-----------------|------------------------------------------------------|
| **Zapp** | VPS (cx23: 2 vCPU, 4GB RAM) | [Hetzner Cloud] | Public gateway, VPN control plane, reverse tunneling |
| **Kif**  | NUC (i5-7260U, 16GB RAM)    | Local           | Home automation, media, internal services            |

- **Zapp** acts as the internet-facing gateway, running [Dokploy], [Headscale], and [Traefik]
- **Kif** hosts app stacks behind [Caddy], accessible via the home LAN or [Tailscale] mesh VPN
- **Rathole** tunnels select services from Kif through Zapp to the public internet
- DNS is managed via [Cloudflare] (public) and [AdGuard Home] (internal), with Headscale split DNS bridging the two

## Makefile
Available `make` commands:

```console
$ make help
install                                  Install necessary dependencies for all components
check                                    Run code linters
kif-deploy                               Deploy changes to Kif (use EXTRA_VARS for variables, TAGS for tags)
kif-provision                            Provision Kif server (use EXTRA_VARS for variables, TAGS for tags)
zapp-deploy                              Deploy changes to Zapp (use EXTRA_VARS for variables, TAGS for tags)
zapp-provision                           Provision Zapp server (use EXTRA_VARS for variables, TAGS for tags)
beryl-provision                          Provision Beryl AX travel router
encrypt-string                           Encrypt a value with Ansible Vault
tf-plan                                  Generate a (speculative) Terraform plan
help                                     Show help message
```

## Authors
Developed and maintained by [Paweł Adamczak][pawelad].

Source code is available at [GitHub][github nimbus].

Released under [Mozilla Public License 2.0][license].


[adguard home]: https://adguard.com/
[ansible]: https://www.ansible.com/
[caddy]: https://caddyserver.com/
[cloudflare]: https://www.cloudflare.com/
[docker compose]: https://docs.docker.com/compose/
[dokploy]: https://dokploy.com/
[github nimbus]: https://github.com/pawelad/nimbus
[headscale]: https://headscale.net/
[hetzner cloud]: https://www.hetzner.com/cloud/
[license]: ./LICENSE
[pawelad]: https://pawelad.me/
[tailscale]: https://tailscale.com/
[terraform]: https://www.terraform.io/
[traefik]: https://traefik.io/
