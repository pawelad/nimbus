# Uptime Kuma Sync
Python tool that automates the synchronization of Uptime Kuma monitors by scanning `docker-compose.yaml` files in your `stacks` directory (used by [Dockge][dockge]).

It ensures your Uptime Kuma instance is always aware of the services you have deployed, creating new monitors and updating existing ones as your infrastructure evolves.

## Usage
As long as you have `uv` installed, you should be able to run the script with:

```console
❯ uv run uptime_kuma_sync.py --help
Usage: uptime_kuma_sync.py [ARGS]

Sync Uptime Kuma monitors from compose.yaml stacks.

╭─ Commands ───────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ --help (-h)  Display this message and exit.                                                                          │
│ --version    Display application version.                                                                            │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭─ Parameters ─────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ URL --url                           Uptime Kuma URL [env var: UPTIME_KUMA_URL] [default: http://localhost:3001]      │
│ USER --user                         Uptime Kuma username [env var: UPTIME_KUMA_USER] [default: admin]                │
│ PASSWORD --password                 Uptime Kuma password [env var: UPTIME_KUMA_PASSWORD] [default: admin]            │
│ CHECK --check --dry-run --no-check  Dry run mode (do not make changes) [default: False]                              │
│   --no-dry-run                                                                                                       │
│ STACKS-DIR --stacks-dir             Path to stacks directory [default: /Users/pawelad/Dev/Misc/kif/src/stacks]       │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

## Configuration
Services in your `compose.yaml` files are discovered if they have appropriate labels. The tool looks for:
- `homepage.name`: The name of the monitor.
- `homepage.href`: The URL to monitor.

Example service definition:

```yaml
services:
  my-app:
    image: my-app:latest
    labels:
      homepage.name: "My App"
      homepage.href: "https://myapp.example.com"
```

## Makefile
Available `make` commands:

```console
$ make help
setup                                     Install dependencies
run                                       Run the tool
run-dry                                   Run the tool in dry-run mode
format                                    Format code
help                                      Show help message
```

## Authors
Developed and maintained by [Paweł Adamczak][pawelad].

Source code is available at [GitHub][github kif].

Released under [Mozilla Public License 2.0][license].


[github kif]: https://github.com/pawelad/kif
[license]: ./LICENSE
[pawelad]: https://pawelad.me/
