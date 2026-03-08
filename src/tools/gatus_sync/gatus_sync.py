"""
Gatus Sync Tool

This script automates the synchronization of Gatus endpoints by scanning
docker-compose files for service definitions. It extracts service names,
groups, and URLs from Homepage/Caddy labels to ensure Gatus is properly populated.
"""
# /// script
# dependencies = [
#   "pyyaml",
#   "cyclopts",
# ]
# ///

import os
import sys
import yaml
import logging
from pathlib import Path
from typing import Annotated, List, Dict, Any

import cyclopts

logger = logging.getLogger(__name__)

app = cyclopts.App(help="Sync Gatus config from compose.yaml stacks.")


def parse_stacks(stacks_dir: Path) -> List[Dict[str, Any]]:
    """
    Scans a directory for docker-compose files and extracts endpoint configurations.
    """
    endpoints = []
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

        if not content.get("services"):
            logger.warning(
                f"Skipping {file_path}: 'services' key not found or invalid format."
            )
            continue

        for service_name, service_config in content["services"].items():
            labels = service_config.get("labels", {})
            if not labels:
                continue

            # Determine Target Name
            name = labels.get("homepage.name")
            group = labels.get("homepage.group", "Apps")
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
                    # e.g caddy: "aiostreams.home aiostreams.pipusznicy"
                    # e.g caddy_0: "aiostreams.home aiostreams.pipusznicy"
                    domains = caddy_labels.split()
                    if domains:
                        d = domains[0]
                        if not d.startswith("http"):
                            url = f"https://{d}"
                        else:
                            url = d

                # Check for caddy_0, caddy_1, etc. if 'caddy' not present
                if not url:
                    for key, val in labels.items():
                        if (
                            key.startswith("caddy_")
                            and not key.endswith(".tls")
                            and not key.endswith(".reverse_proxy")
                        ):
                            domains = str(val).split()
                            if domains:
                                d = domains[0]
                                if not d.startswith("http"):
                                    url = f"https://{d}"
                                else:
                                    url = d
                                break

            if name and url:
                endpoint = {
                    "name": name,
                    "group": group,
                    "url": url,
                    "interval": "60s",
                    "conditions": ["[STATUS] == 200"],
                    "client": {"insecure": True},
                    "alerts": [{"type": "ntfy"}],
                }
                endpoints.append(endpoint)
                logger.info(
                    f"  - Found endpoint: {group} / {name} -> {url} (from {file_path.parent.name})"
                )

    return endpoints


@app.default
def main(
    output_config: Annotated[
        Path,
        cyclopts.Parameter(
            name=["--output", "-o"],
            help="Path to gatus config.yaml output file",
        ),
    ] = Path("/config/config.yaml"),
    stacks_dir: Annotated[
        Path, cyclopts.Parameter(help="Path to stacks directory")
    ] = Path("/data/stacks"),
):
    """
    Generate Gatus configuration from compose.yaml files in the stacks directory.
    """
    logging.basicConfig(
        level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
    )

    logger.info(f"Scanning stacks in: {stacks_dir}")
    if not stacks_dir.exists():
        logger.error(f"Error: Stacks directory {stacks_dir} does not exist.")
        sys.exit(1)

    endpoints = parse_stacks(stacks_dir)
    
    # Gatus v5.35+ requires at least one endpoint to start. 
    # Add a self-health check to guarantee it never crashes with an empty config.
    endpoints.append({
        "name": "Health",
        "group": "Internal",
        "url": "http://localhost:8080/health",
        "interval": "60s",
        "conditions": ["[STATUS] == 200"],
    })

    # Sort endpoints by group and name for consistent output
    import operator

    endpoints.sort(key=operator.itemgetter("group", "name"))

    config = {
        "metrics": True,
        "endpoints": endpoints,
    }

    import json
    
    config_yaml = yaml.dump(config, sort_keys=False)
    
    if output_config.exists():
        try:
            with output_config.open("r") as f:
                existing_config = yaml.safe_load(f)
            if existing_config == config:
                logger.info("Config is up to date. No changes required.")
                return
            else:
                try:
                    with output_config.open("w") as f:
                        f.write(config_yaml)
                    logger.info("[UPDATE] Config updated.")
                except OSError as e:
                    logger.error(f"Failed to write config file: {e}")
                    sys.exit(1)
        except (OSError, yaml.YAMLError) as e:
            logger.warning(f"Error reading existing config: {e}. Overwriting.")
            try:
                with output_config.open("w") as f:
                    f.write(config_yaml)
                logger.info("[UPDATE] Config overwritten.")
            except OSError as e:
                logger.error(f"Failed to write config file: {e}")
                sys.exit(1)
    else:
        try:
            output_config.parent.mkdir(parents=True, exist_ok=True)
            with output_config.open("w") as f:
                f.write(config_yaml)
            logger.info("[CREATE] Config created.")
        except OSError as e:
            logger.error(f"Failed to write config file: {e}")
            sys.exit(1)

    logger.info("Done.")


if __name__ == "__main__":
    app()
