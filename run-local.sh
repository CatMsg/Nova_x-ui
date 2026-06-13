#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

export XUI_DB_FOLDER="${XUI_DB_FOLDER:-$ROOT_DIR/db}"
export XUI_LOG_FOLDER="${XUI_LOG_FOLDER:-$ROOT_DIR/log}"
export XUI_BIN_FOLDER="${XUI_BIN_FOLDER:-$ROOT_DIR/bin}"
RUN_LOCAL_PID_FILE="${RUN_LOCAL_PID_FILE:-$ROOT_DIR/.run-local.pid}"
RUN_LOCAL_LOG_FILE="${RUN_LOCAL_LOG_FILE:-$XUI_LOG_FOLDER/run-local.log}"
RUN_LOCAL_BINARY="${RUN_LOCAL_BINARY:-$XUI_BIN_FOLDER/Nova_x-ui}"
RUN_LOCAL_START_TIMEOUT="${RUN_LOCAL_START_TIMEOUT:-60}"

mkdir -p "$XUI_DB_FOLDER" "$XUI_LOG_FOLDER" "$XUI_BIN_FOLDER"

usage() {
  cat <<'EOF'
Usage: ./run-local.sh [run|restart|stop|status|logs] [go-run-args...]

Commands:
  run      Stop any existing Nova_x-ui process, provision binaries, then start.
  restart  Same as run.
  stop     Stop any existing Nova_x-ui process and exit.
  status   Show whether Nova_x-ui is currently running.
  logs     Print the last startup log, or follow it with -f.

Without a command, the script behaves like `run`.
EOF
}

read_pid_file() {
  if [[ -f "$RUN_LOCAL_PID_FILE" ]]; then
    tr -d '[:space:]' < "$RUN_LOCAL_PID_FILE"
  fi
}

is_nova_process() {
  local pid="$1"
  local command_line

  if [[ -z "$pid" ]] || ! kill -0 "$pid" 2>/dev/null; then
    return 1
  fi

  command_line="$(ps -p "$pid" -o command= 2>/dev/null | sed 's/^ *//')"
  [[ "$command_line" == *"Nova_x-ui"* ]]
}

collect_running_pids() {
  {
    read_pid_file
    pgrep -x Nova_x-ui 2>/dev/null || true
    lsof -tiTCP:2026 -sTCP:LISTEN 2>/dev/null || true
    lsof -tiTCP:2027 -sTCP:LISTEN 2>/dev/null || true
  } | sed '/^$/d' | sort -u | while IFS= read -r pid; do
    if is_nova_process "$pid"; then
      echo "$pid"
    fi
  done
}

process_listening_ports() {
  local pid="$1"

  lsof -nP -a -p "$pid" -iTCP -sTCP:LISTEN 2>/dev/null | awk 'NR>1 {print $9}' | sort -u | tr '\n' ' ' | sed 's/[[:space:]]*$//'
}

resolve_running_pid() {
  collect_running_pids | head -n 1
}

print_process_details() {
  local pid="$1"
  local started_at
  local command_line
  local listen_ports

  started_at="$(ps -p "$pid" -o lstart= 2>/dev/null | sed 's/^ *//')"
  command_line="$(ps -p "$pid" -o command= 2>/dev/null | sed 's/^ *//')"
  listen_ports="$(process_listening_ports "$pid")"

  echo "  PID: $pid"
  echo "  Started: ${started_at:-unknown}"
  echo "  Command: ${command_line:-unknown}"
  echo "  PID file: $RUN_LOCAL_PID_FILE"
  echo "  Log file: $RUN_LOCAL_LOG_FILE"
  if [[ -n "$listen_ports" ]]; then
    echo "  Listening: $listen_ports"
  fi
}

status_existing_process() {
  local pids

  pids="$(collect_running_pids)"
  if [[ -z "$pids" ]]; then
    echo "Nova_x-ui is not running"
    if [[ -f "$RUN_LOCAL_PID_FILE" ]]; then
      echo "  Stale pid file: $(cat "$RUN_LOCAL_PID_FILE" 2>/dev/null || true)"
    fi
    return 0
  fi

  echo "Nova_x-ui is running"
  while IFS= read -r pid; do
    [[ -z "$pid" ]] && continue
    print_process_details "$pid"
  done <<< "$pids"
}

tail_logs() {
  local lines="${1:-120}"

  if [[ ! -f "$RUN_LOCAL_LOG_FILE" ]]; then
    echo "Log file not found: $RUN_LOCAL_LOG_FILE" >&2
    return 1
  fi

  tail -n "$lines" "$RUN_LOCAL_LOG_FILE"
}

stop_existing_process() {
  local pids
  local remaining

  pids="$(collect_running_pids)"

  if [[ -z "$pids" ]]; then
    rm -f "$RUN_LOCAL_PID_FILE"
    echo "Nova_x-ui is not running"
    return 0
  fi

  echo "Stopping Nova_x-ui process(es):"
  while IFS= read -r pid; do
    [[ -z "$pid" ]] && continue
    print_process_details "$pid"
    kill -TERM "$pid" 2>/dev/null || true
  done <<< "$pids"

  for _ in {1..20}; do
    remaining="$(collect_running_pids)"
    if [[ -z "$remaining" ]]; then
      rm -f "$RUN_LOCAL_PID_FILE"
      return 0
    fi
    sleep 0.25
  done

  echo "Force stopping remaining Nova_x-ui process(es)"
  while IFS= read -r pid; do
    [[ -z "$pid" ]] && continue
    kill -KILL "$pid" 2>/dev/null || true
  done <<< "$remaining"

  for _ in {1..20}; do
    remaining="$(collect_running_pids)"
    if [[ -z "$remaining" ]]; then
      rm -f "$RUN_LOCAL_PID_FILE"
      return 0
    fi
    sleep 0.25
  done

  echo "Warning: some Nova_x-ui process(es) may still be running" >&2
  return 1
}

wait_for_startup() {
  local launcher_pid="$1"
  local timeout="$2"
  local pids
  local pid
  local listen_ports

  for _ in $(seq 1 "$timeout"); do
    pids="$(collect_running_pids)"
    if [[ -n "$pids" ]]; then
      while IFS= read -r pid; do
        [[ -z "$pid" ]] && continue
        listen_ports="$(process_listening_ports "$pid")"
        if [[ -n "$listen_ports" ]]; then
          echo "$pid|$listen_ports"
          return 0
        fi
      done <<< "$pids"
    fi

    if ! kill -0 "$launcher_pid" 2>/dev/null && [[ -z "$pids" ]]; then
      return 1
    fi

    sleep 1
  done

  return 2
}

start_local() {
  local launch_command

  if ! command -v python3 >/dev/null 2>&1; then
    echo "python3 is required to provision Xray binaries for local runs" >&2
    exit 1
  fi

  if ! command -v go >/dev/null 2>&1; then
    echo "go is required to start Nova_x-ui locally" >&2
    exit 1
  fi

  stop_existing_process >/dev/null 2>&1 || true

  : > "$RUN_LOCAL_LOG_FILE"

  if ! python3 "$ROOT_DIR/scripts/provision_xray.py" 2>&1 | tee -a "$RUN_LOCAL_LOG_FILE"; then
    echo "Nova_x-ui failed during Xray provisioning" >&2
    echo "--- last log lines ---" >&2
    tail_logs 120 >&2 || true
    return 1
  fi

  echo "Building local Nova_x-ui binary..."
  if ! go build -o "$RUN_LOCAL_BINARY" . 2>&1 | tee -a "$RUN_LOCAL_LOG_FILE"; then
    echo "Nova_x-ui build failed" >&2
    echo "--- last log lines ---" >&2
    tail_logs 120 >&2 || true
    return 1
  fi

  if [[ $# -gt 0 ]]; then
    launch_command="$RUN_LOCAL_BINARY $*"
  else
    launch_command="$RUN_LOCAL_BINARY run"
  fi

  echo "Launching Nova_x-ui..."
  echo "  PID file: $RUN_LOCAL_PID_FILE"
  echo "  Log file: $RUN_LOCAL_LOG_FILE"
  echo "  Command: $launch_command"

  printf '%s\n' "$$" > "$RUN_LOCAL_PID_FILE"
  exec > >(tee -a "$RUN_LOCAL_LOG_FILE") 2>&1

  if [[ $# -gt 0 ]]; then
    exec "$RUN_LOCAL_BINARY" "$@"
  fi

  exec "$RUN_LOCAL_BINARY" run
}

print_logs() {
  local follow="${1:-false}"
  local lines="${2:-120}"

  if [[ "$follow" == "true" ]]; then
    tail -n "$lines" -f "$RUN_LOCAL_LOG_FILE"
    return 0
  fi

  tail_logs "$lines"
}

ACTION="run"
if [[ $# -gt 0 ]]; then
  case "$1" in
    run|restart|stop|status|logs)
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
    start_local "$@"
    ;;
  run)
    start_local "$@"
    ;;
  logs)
    if [[ "${1:-}" == "-f" || "${1:-}" == "--follow" ]]; then
      shift || true
      print_logs true "${1:-120}"
    else
      print_logs false "${1:-120}"
    fi
    ;;
  *)
    usage >&2
    exit 1
    ;;
esac
