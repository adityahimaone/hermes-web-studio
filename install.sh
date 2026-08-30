#!/bin/sh
set -eu

install_dir="${HERMES_WEBUI_INSTALL_DIR:-$HOME/.local/bin}"
mkdir -p "$install_dir"
if command -v pnpm >/dev/null 2>&1; then
  make build
else
  echo "pnpm is required to build Hermes Web Studio" >&2
  exit 1
fi
cp hermes-web-studio "$install_dir/hermes-web-studio"
chmod 0755 "$install_dir/hermes-web-studio"
echo "Installed $install_dir/hermes-web-studio"
