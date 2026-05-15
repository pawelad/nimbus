# Beryl AX (GL-MT3000)

The Beryl AX is a GL-iNet travel router running an OpenWRT-based firmware. In this project, it is integrated into the Ansible infrastructure to automate its connection to the self-hosted Headscale instance on **Zapp**.

## Ansible Configuration

The router is managed via a dedicated Ansible setup:

- **Inventory**: `src/ansible/inventories/beryl.ini`
- **Host Variables**: `src/ansible/inventories/host_vars/beryl.yml`
- **Playbook**: `src/ansible/playbooks/beryl_provision.yml`

## Automated Provisioning

You can configure the router by running:

> [!IMPORTANT]
> **Prerequisite**: You must manually install Python on the router before running this command:
> `ssh root@192.168.8.1 "opkg update && opkg install python3"`

```bash
make beryl-provision
```

## GL-iNet UI Configuration

To ensure that the Headscale **Split DNS** works correctly (allowing you to reach `*.pipusznicy.cloud` and `*.home` domains), the following settings should be verified in the router's web admin:

1. **DNS Settings**:
    - **Allow Custom DNS to Override VPN DNS**: Should be set to **OFF**. If this is ON, the router's local DNS logic will take precedence over Headscale's configuration.
2. **AdGuard Home**:
    - It is recommended to keep the router's internal AdGuard Home **OFF** to avoid conflicts with the primary AdGuard instance on **Kif**.

## Persistence

By modifying the system script `/usr/bin/gl_tailscale`, the Headscale integration becomes robust:
- It survives reboots.
- It survives toggling Tailscale ON/OFF in the GL-iNet Web UI.
- It survives connection loss or router restarts.

> [!WARNING]
> Since this setup modifies a system script (`/usr/bin/gl_tailscale`), these changes might be lost during a **firmware update**. If the connection is broken after an update, simply run `make beryl-provision` again.

## Why remove `--reset`?

The original GL-iNet script uses `--reset` in its `tailscale up` command, which explicitly wipes any custom state or login server settings. By removing it, Tailscale is allowed to cache its registration state, making the connection much more reliable.
