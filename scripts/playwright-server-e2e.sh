#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="/home/uraguchi/work/AI/project_management_tool"
FRONTEND_DIR="$ROOT_DIR/frontend"

# WSL から Windows 側 Playwright Server へ接続する既定値
# 優先: default gateway（WSL2 の Windows ホスト） / fallback: resolv.conf nameserver
detect_windows_host_ip() {
  local ip
  ip="$(ip route | awk '/^default/ {print $3; exit}')"
  if [[ -n "$ip" ]]; then
    printf '%s' "$ip"
    return
  fi
  grep -m 1 nameserver /etc/resolv.conf | awk '{print $2}'
}

WIN_HOST_IP="${WIN_HOST_IP:-$(detect_windows_host_ip)}"
PLAYWRIGHT_WS_ENDPOINT="${PLAYWRIGHT_WS_ENDPOINT:-ws://${WIN_HOST_IP}:9222/}"
export PLAYWRIGHT_WS_ENDPOINT

if [[ $# -eq 0 ]]; then
  echo "[info] Running all E2E tests via Playwright Server: ${PLAYWRIGHT_WS_ENDPOINT}"
  (cd "$FRONTEND_DIR" && npm run test:e2e)
else
  echo "[info] Running selected E2E tests via Playwright Server: ${PLAYWRIGHT_WS_ENDPOINT}"
  (cd "$FRONTEND_DIR" && npm run test:e2e -- "$@")
fi
