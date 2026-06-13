#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

export XUI_DB_FOLDER="${XUI_DB_FOLDER:-$ROOT_DIR/db}"
export XUI_LOG_FOLDER="${XUI_LOG_FOLDER:-$ROOT_DIR/log}"
export XUI_BIN_FOLDER="${XUI_BIN_FOLDER:-$ROOT_DIR/bin}"

mkdir -p "$XUI_DB_FOLDER" "$XUI_LOG_FOLDER" "$XUI_BIN_FOLDER"

stop_existing_process() {
  local pids

  pids="$(pgrep -x Nova_x-ui 2>/dev/null || true)"
  if [[ -z "$pids" ]]; then
    return 0
  fi

  echo "Stopping existing Nova_x-ui process(es): $pids"
  kill $pids 2>/dev/null || true

  for _ in {1..20}; do
    if ! pgrep -x Nova_x-ui >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.25
  done

  echo "Force stopping remaining Nova_x-ui process(es)"
  kill -9 $(pgrep -x Nova_x-ui 2>/dev/null || true) 2>/dev/null || true
}

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required to provision Xray binaries for local runs" >&2
  exit 1
fi

stop_existing_process

python3 "$ROOT_DIR/scripts/provision_xray.py"

if [[ $# -gt 0 ]]; then
  exec go run . "$@"
else
  exec go run . run
fi
