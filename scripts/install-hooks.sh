#!/usr/bin/env bash
set -euo pipefail

if ! command -v cog &>/dev/null; then
  echo "Error: cocogitto (cog) is not installed."
  echo "Install: https://docs.cocogitto.io/guide/installation"
  exit 1
fi

cog install-hook commit-msg
echo "commit-msg hook installed."
