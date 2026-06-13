#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

export XUI_DB_FOLDER="${XUI_DB_FOLDER:-$ROOT_DIR/db}"
export XUI_LOG_FOLDER="${XUI_LOG_FOLDER:-$ROOT_DIR/log}"
export XUI_BIN_FOLDER="${XUI_BIN_FOLDER:-$ROOT_DIR/bin}"

mkdir -p "$XUI_DB_FOLDER" "$XUI_LOG_FOLDER" "$XUI_BIN_FOLDER"

usage() {
  cat <<'EOF'
Usage: ./run-local.sh [run|restart|stop|status] [go-run-args...]

Commands:
  run      Stop any existing Nova_x-ui process, provision binaries, then start.
  restart  Same as run.
  stop     Stop any existing Nova_x-ui process and exit.
  status   Show whether Nova_x-ui is currently running.

Without a command, the script behaves like `run`.
EOF
}

list_existing_pids() {
  pgrep -x Nova_x-ui 2>/dev/null || true
}

status_existing_process() {
  local pids

  pids="$(list_existing_pids)"
  if [[ -z "$pids" ]]; then
    echo "Nova_x-ui is not running"
    return 0
  fi

  echo "Nova_x-ui is running"
  while IFS= read -r pid; do
    [[ -z "$pid" ]] && continue
    echo "  PID $pid: $(ps -p "$pid" -o command= 2>/dev/null || echo Nova_x-ui)"
  done <<< "$pids"

  local listen_ports
  listen_ports="$(lsof -nP -a -c Nova_x-ui -iTCP -sTCP:LISTEN 2>/dev/null | awk 'NR>1 {print $9}' | sort -u | tr '\n' ' ' | sed 's/[[:space:]]*$//')"
  if [[ -n "$listen_ports" ]]; then
    echo "  Listening: $listen_ports"
  fi
}

stop_existing_process() {
  local pids

  pids="$(list_existing_pids)"
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
  pids="$(list_existing_pids)"
  if [[ -n "$pids" ]]; then
    kill -9 $pids 2>/dev/null || true
  fi
}

start_local() {
  if ! command -v python3 >/dev/null 2>&1; then
    echo "python3 is required to provision Xray binaries for local runs" >&2
    exit 1
  fi

  python3 "$ROOT_DIR/scripts/provision_xray.py"

  if [[ $# -gt 0 ]]; then
    exec go run . "$@"
  fi

  exec go run . run
}

ACTION="run"
if [[ $# -gt 0 ]]; then
  case "$1" in
    run|restart|stop|status)
      ACTION="$1"
      shift
      ;;
    -h|--help|help)
      usage
      exit 0
      ;;
  esac
fi

case "$ACTION" in
  status)
    status_existing_process
    ;;
  stop)
    stop_existing_process
    ;;
  restart)
    stop_existing_process
    start_local "$@"
    ;;
  run)
    stop_existing_process
    start_local "$@"
    ;;
  *)
    usage >&2
    exit 1
    ;;
esac
