#!/bin/sh
set -e

# Create a directory for the certificate in a writable location
mkdir -p /tmp/certs

# Copy the certificate from the Caddy volume
cp /caddy-data/caddy/pki/authorities/local/root.crt /tmp/certs/root.crt

# Grant read permissions to everyone
chmod 644 /tmp/certs/root.crt
