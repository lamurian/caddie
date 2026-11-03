#!/usr/bin/env bash
# Starts, stops or shows the sandboxed browser stack (docker/compose.yml,
# ADR 0002). `up` first prepares .caddie/trace, the host side of the
# browser's writable /trace mount (ADR 0024), so the browser's uid 1000
# can write screenshots there. It never makes the directory world-writable.
# Usage: stack.sh up|down|status|logs
set -euo pipefail

cd "$(dirname "$0")/.."
COMPOSE=(docker compose -f docker/compose.yml)
TRACE=.caddie/trace
BROWSER_UID=1000

prepare_trace() {
  mkdir -p "$TRACE"
  if [[ "$(uname -s)" != Linux || "$(id -u)" == "$BROWSER_UID" ]]; then
    return
  fi
  if command -v setfacl >/dev/null 2>&1; then
    setfacl -m "u:$BROWSER_UID:rwx" "$TRACE"
    echo "granted uid $BROWSER_UID write access to $TRACE (setfacl)"
  else
    echo "warn: your uid is $(id -u), not $BROWSER_UID, and setfacl is missing;" >&2
    echo "      screenshots to /trace may fail. Install acl, then rerun, or see" >&2
    echo "      docker/README.md#trace-directory." >&2
  fi
}

case "${1:-}" in
  up)
    prepare_trace
    "${COMPOSE[@]}" up -d --build --wait
    echo "stack is up; check it with: caddie doctor"
    ;;
  down) "${COMPOSE[@]}" down ;;
  status) "${COMPOSE[@]}" ps ;;
  logs) "${COMPOSE[@]}" logs --no-color --tail 100 ;;
  *)
    echo "usage: $0 up|down|status|logs" >&2
    exit 2
    ;;
esac
