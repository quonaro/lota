#!/usr/bin/env bash
set -euo pipefail

VERSION=""
DRY_RUN=false

usage() {
  echo "Usage: $0 [--version X.Y.Z] [--dry-run]"
  echo ""
  echo "  --version X.Y.Z   force a specific version"
  echo "  --dry-run         print what would happen without making changes"
  echo ""
  echo "Without --version, the bump type is determined automatically"
  echo "from conventional commits (feat -> minor, fix -> patch, BREAKING -> major)."
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      VERSION="${2:-}"
      [[ -z "$VERSION" ]] && { echo "Error: --version requires a value"; usage; }
      shift 2
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo "Unknown option: $1"
      usage
      ;;
  esac
done

ARGS=()
if [[ -n "$VERSION" ]]; then
  ARGS+=("--version" "$VERSION")
else
  ARGS+=("--auto")
fi

[[ "$DRY_RUN" == true ]] && ARGS+=("--dry-run")

cog bump "${ARGS[@]}"
