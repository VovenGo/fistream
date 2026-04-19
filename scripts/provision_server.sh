#!/usr/bin/env bash
set -euo pipefail

# Run on server as root.

if ! command -v k3s >/dev/null 2>&1; then
  curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC='server --disable traefik --write-kubeconfig-mode 644' sh -
fi

systemctl enable --now k3s

# Required network ports: SSH, HTTP/S, Jitsi media, k3s API
if command -v ufw >/dev/null 2>&1; then
  ufw allow 22/tcp || true
  ufw allow 80/tcp || true
  ufw allow 443/tcp || true
  ufw allow 8443/tcp || true
  ufw allow 31000/udp || true
  ufw allow 6443/tcp || true
fi

install -d -m 0755 /etc/caddy
cp deploy/caddy/Caddyfile.fistream /etc/caddy/Caddyfile
systemctl reload caddy

echo 'server provision finished'
