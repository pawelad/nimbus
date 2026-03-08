"""
Uptime Kuma Sync Tool

This script automates the synchronization of Uptime Kuma monitors by scanning
docker-compose files for service definitions. It extracts service names and URLs
to ensure Uptime Kuma is always up-to-date with the deployed stack.

The script looks for `compose.yaml` files in subdirectories of a specified
stacks directory. It attempts to derive monitor names and URLs from:
 - `homepage.name` label
 - `homepage.href` label
 - `caddy` label (for URL)
 - Service name (fallback for name)
"""
# /// script
# dependencies = [
#   "uptime-kuma-api",
#   "pyyaml",
#   "cyclopts",
# ]
# ///

import os
import sys
import glob
import yaml
import logging
from pathlib import Path
from typing import Annotated, List, Dict, Any, Optional
from dataclasses import dataclass

from uptime_kuma_api import UptimeKumaApi, UptimeKumaException, Timeout, MonitorType, NotificationType
import cyclopts
import socketio.exceptions

logger = logging.getLogger(__name__)


app = cyclopts.App(help="Sync Uptime Kuma monitors from compose.yaml stacks.")


@dataclass
class MonitorDefinition:
    name: str
    url: str
    type: MonitorType
    interval: int = 60
    description: str = ""


def parse_stacks(stacks_dir: Path) -> List[MonitorDefinition]:
    """
    Scans a directory for docker-compose files and extracts monitor configurations.
    """
    monitors = []
    # Use glob to find compose.yaml files in subdirectories
    stack_files = list(stacks_dir.glob("*/compose.yaml"))

    logger.info(f"Found {len(stack_files)} stack files in {stacks_dir}")

    for file_path in stack_files:
        try:
            with file_path.open("r") as f:
                content = yaml.safe_load(f)
        except (OSError, yaml.YAMLError) as e:
            logger.error(f"Error parsing {file_path}: {e}")
            continue

        if not isinstance(content, dict) or "services" not in content:
            logger.warning(
                f"Skipping {file_path}: 'services' key not found or invalid format."
            )
            continue

        for service_name, service_config in content["services"].items():
            labels = service_config.get("labels", {})
            if not labels:
                continue

            # Determine Monitor Name
            name = labels.get("homepage.name")
            if not name:
                # Fallback to service name
                name = service_name.capitalize()

            # Determine URL
            url = labels.get("monitor.url")
            if not url:
                url = labels.get("homepage.href")
            if not url:
                # Try caddy label
                caddy_labels = labels.get("caddy", "")
                if caddy_labels:
                    # Take the first one. Assume https unless specified.
                    domains = caddy_labels.split()
                    if domains:
                        d = domains[0]
                        if not d.startswith("http"):
                            url = f"https://{d}"
                        else:
                            url = d

            if name and url:
                description = labels.get("homepage.description", "")
                monitors.append(
                    MonitorDefinition(
                        name=name,
                        url=url,
                        type=MonitorType.HTTP,
                        interval=60,
                        description=description,
                    )
                )
                logger.info(
                    f"  - Found: {name} -> {url} (from {file_path.parent.name})"
                )

    return monitors


def sync_monitors(
    api: Optional[UptimeKumaApi],
    monitors: List[MonitorDefinition],
    dry_run: bool = False,
    force_update: bool = False,
):
    """
    Synchronizes the list of desired monitors with the Uptime Kuma instance.
    """
    if api:
        logger.info("Fetching existing monitors...")
        try:
            existing_monitors = api.get_monitors()
            existing_map = {m["name"]: m for m in existing_monitors}
        except Exception as e:
            logger.error(f"Failed to fetch monitors: {e}")
            return
    else:
        logger.warning("Skipping fetch of existing monitors (No API connection).")
        existing_map = {}

    for target in monitors:
        name = target.name
        url = target.url
        description = target.description

        if name in existing_map:
            existing = existing_map[name]
            needs_update = False

            if existing["url"] != url:
                logger.info(f"  [UPDATE] {name}: URL '{existing['url']}' -> '{url}'")
                needs_update = True

            existing_desc = existing.get("description") or ""
            if existing_desc != description:
                logger.info(f"  [UPDATE] {name}: Description updated")
                needs_update = True

            if force_update:
                logger.info(f"  [FORCE] {name}: Force update enabled")
                needs_update = True

            if needs_update and not dry_run:
                try:
                    mid = existing["id"]
                    api.edit_monitor(mid, url=url, description=description, ignoreTls=True)
                    logger.info(f"    -> Updated {name}")
                except Exception as e:
                    logger.error(f"    -> Error updating {name}: {e}")
            elif needs_update and dry_run:
                logger.info(f"    -> (Dry Run) Would update {name}")
            else:
                logger.info(f"  [OK] {name} is up to date.")

        else:
            # Create
            logger.info(f"  [CREATE] New monitor: {name} ({url})")
            if not dry_run:
                try:
                    if api:
                        api.add_monitor(
                            type=target.type,
                            name=name,
                            url=url,
                            interval=target.interval,
                            ignoreTls=True,
                            description=target.description,
                        )
                        logger.info(f"    -> Created {name}")
                    else:
                        logger.error(
                            f"    -> Error: Cannot create {name}, no API connection."
                        )
                except Exception as e:
                    logger.error(f"    -> Error creating {name}: {e}")
            else:
                logger.info(f"    -> (Dry Run) Would create {name}")


NTFY_NOTIFICATION_NAME = "ntfy (Ansible Managed)"


def sync_notifications(
    api: Optional[UptimeKumaApi],
    ntfy_url: str,
    ntfy_topic: str,
    ntfy_token: str,
    dry_run: bool = False,
):
    """
    Ensures an ntfy notification provider exists in Uptime Kuma with the
    correct settings. Creates or updates as needed.
    """
    if not ntfy_url or not ntfy_topic:
        logger.info("ntfy not configured, skipping notification sync.")
        return None

    if not api:
        logger.warning("No API connection, skipping notification sync.")
        return None

    logger.info("Syncing ntfy notification provider...")

    try:
        existing_notifications = api.get_notifications()
    except Exception as e:
        logger.error(f"Failed to fetch notifications: {e}")
        return None

    # Look for our managed notification by name
    existing = None
    for n in existing_notifications:
        if n["name"] == NTFY_NOTIFICATION_NAME:
            existing = n
            break

    ntfy_params = dict(
        name=NTFY_NOTIFICATION_NAME,
        type=NotificationType.NTFY,
        ntfyserverurl=ntfy_url,
        ntfytopic=ntfy_topic,
        ntfyPriority=4,
        ntfyaccesstoken=ntfy_token,
        isDefault=True,
        applyExisting=True,
    )

    if existing:
        # Check if update needed
        needs_update = (
            existing.get("ntfyserverurl") != ntfy_url
            or existing.get("ntfytopic") != ntfy_topic
            or existing.get("ntfyaccesstoken") != ntfy_token
            or existing.get("ntfyPriority") != 4
        )
        if needs_update:
            logger.info(f"  [UPDATE] {NTFY_NOTIFICATION_NAME}")
            if not dry_run:
                try:
                    api.edit_notification(existing["id"], **ntfy_params)
                    logger.info(f"    -> Updated {NTFY_NOTIFICATION_NAME}")
                except Exception as e:
                    logger.error(f"    -> Error updating notification: {e}")
            else:
                logger.info(f"    -> (Dry Run) Would update {NTFY_NOTIFICATION_NAME}")
        else:
            logger.info(f"  [OK] {NTFY_NOTIFICATION_NAME} is up to date.")
        return existing["id"]
    else:
        logger.info(f"  [CREATE] {NTFY_NOTIFICATION_NAME}")
        if not dry_run:
            try:
                result = api.add_notification(**ntfy_params)
                notification_id = result.get("id")
                logger.info(f"    -> Created {NTFY_NOTIFICATION_NAME} (id={notification_id})")
                return notification_id
            except Exception as e:
                logger.error(f"    -> Error creating notification: {e}")
                return None
        else:
            logger.info(f"    -> (Dry Run) Would create {NTFY_NOTIFICATION_NAME}")
            return None


@app.default
def main(
    url: Annotated[
        str,
        cyclopts.Parameter(
            name=["--url"],
            env_var=["UPTIME_KUMA_URL"],
            help="Uptime Kuma URL",
        ),
    ] = "http://localhost:3001",
    user: Annotated[
        str,
        cyclopts.Parameter(
            name=["--user"],
            env_var=["UPTIME_KUMA_USER"],
            help="Uptime Kuma username",
        ),
    ] = "admin",
    password: Annotated[
        str,
        cyclopts.Parameter(
            name=["--password"],
            env_var=["UPTIME_KUMA_PASSWORD"],
            help="Uptime Kuma password",
        ),
    ] = "admin",
    dry_run: Annotated[
        bool,
        cyclopts.Parameter(
            name=["--check", "--dry-run"],
            help="Dry run mode (do not make changes)",
        ),
    ] = False,
    force: Annotated[
        bool,
        cyclopts.Parameter(
            name=["--force"],
            help="Force update all monitors",
        ),
    ] = False,
    stacks_dir: Annotated[
        Path, cyclopts.Parameter(help="Path to stacks directory")
    ] = Path("/data/stacks"),
    ntfy_url: Annotated[
        str,
        cyclopts.Parameter(
            name=["--ntfy-url"],
            env_var=["NTFY_URL"],
            help="ntfy server URL for notification provider",
        ),
    ] = "",
    ntfy_topic: Annotated[
        str,
        cyclopts.Parameter(
            name=["--ntfy-topic"],
            env_var=["NTFY_TOPIC"],
            help="ntfy topic for Uptime Kuma notifications",
        ),
    ] = "",
    ntfy_token: Annotated[
        str,
        cyclopts.Parameter(
            name=["--ntfy-token"],
            env_var=["NTFY_TOKEN"],
            help="ntfy access token",
        ),
    ] = "",
    timeout: Annotated[
        float,
        cyclopts.Parameter(
            name=["--timeout"],
            help="Connection timeout in seconds",
        ),
    ] = 30.0,
):
    """
    Sync specified Uptime Kuma monitors from compose.yaml files in the stacks directory.
    """
    logging.basicConfig(
        level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
    )

    api = None

    # In check mode, we can proceed without credentials if they fail, but we attempt basic connection if possible.
    # With defaults, we always have "some" credentials.

    logger.info(f"Connecting to Uptime Kuma at {url} (timeout={timeout}s)...")
    try:
        api = UptimeKumaApi(url, timeout=timeout)
        api.login(user, password)
    except (Timeout, socketio.exceptions.TimeoutError) as e:
        if dry_run:
            logger.warning(
                f"Warning: Connection timed out ({e}). Continuing in dry-run mode to check parsing only."
            )
        else:
            logger.exception(f"Connection timed out: {e}")
            logger.info("Check if Uptime Kuma is running and accessible.")
            sys.exit(1)
    except UptimeKumaException as e:
        if dry_run:
            logger.warning(
                f"Warning: Failed to connect ({e}). Continuing in dry-run mode to check parsing only."
            )
        else:
            logger.exception(f"Failed to connect: {e}")
            logger.info("Check your URL and credentials.")
            sys.exit(1)
    else:
        logger.info("Connected successfully.")

    logger.info(f"Scanning stacks in: {stacks_dir}")
    if not stacks_dir.exists():
        logger.error(f"Error: Stacks directory {stacks_dir} does not exist.")
        sys.exit(1)

    monitors_to_sync = parse_stacks(stacks_dir)

    sync_monitors(api, monitors_to_sync, dry_run=dry_run, force_update=force)

    # Sync ntfy notification provider
    sync_notifications(
        api,
        ntfy_url=ntfy_url,
        ntfy_topic=ntfy_topic,
        ntfy_token=ntfy_token,
        dry_run=dry_run,
    )

    if api:
        api.disconnect()


if __name__ == "__main__":
    app()
