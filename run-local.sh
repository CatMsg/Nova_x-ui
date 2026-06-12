#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

export XUI_DB_FOLDER="${XUI_DB_FOLDER:-$ROOT_DIR/db}"
export XUI_LOG_FOLDER="${XUI_LOG_FOLDER:-$ROOT_DIR/log}"
export XUI_BIN_FOLDER="${XUI_BIN_FOLDER:-$ROOT_DIR/bin}"

mkdir -p "$XUI_DB_FOLDER" "$XUI_LOG_FOLDER" "$XUI_BIN_FOLDER"

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required to provision Xray binaries for local runs" >&2
  exit 1
fi

python3 "$ROOT_DIR/scripts/provision_xray.py"

if [[ $# -gt 0 ]]; then
  exec go run . "$@"
else
  exec go run . run
fi
