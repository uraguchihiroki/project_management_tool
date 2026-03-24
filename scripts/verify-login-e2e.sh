#!/usr/bin/env bash
# ログイン〜管理画面到達を E2E で検証してから人に渡す用。
#
# 重要:
# - npx playwright run-server は「ブラウザをリモートで動かす」だけ。API ではない。
# - バックエンド :8080 とフロント :3000 が WSL 上で生きていることが必須。
# - Windows ブラウザ + WSL Next のとき、NEXT_PUBLIC_API_URL を localhost:8080 にすると
#   ブラウザは Windows 側 8080 を見に行きがちなので、通常は未設定で /api/v1 rewrite を使う。
#
# 使い方:
#   bash scripts/verify-login-e2e.sh
# Windows 側で run-server 済みなら:
#   PLAYWRIGHT_WS_ENDPOINT=ws://<WindowsのIP>:9222/ bash scripts/verify-login-e2e.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_LOG="${TMPDIR:-/tmp}/pmt-verify-backend.log"
FRONTEND_LOG="${TMPDIR:-/tmp}/pmt-verify-frontend.log"
export DATABASE_URL="${DATABASE_URL:-postgres://pmt_user:pmt_password@127.0.0.1:5432/pmt_db?sslmode=disable}"
export PORT="${PORT:-8080}"

cleanup() {
  trap - EXIT INT TERM
  if [[ -n "${FRONT_PID:-}" ]] && kill -0 "$FRONT_PID" 2>/dev/null; then
    kill -TERM "$FRONT_PID" 2>/dev/null || true
    wait "$FRONT_PID" 2>/dev/null || true
  fi
  if [[ -n "${BACK_PID:-}" ]] && kill -0 "$BACK_PID" 2>/dev/null; then
    kill -TERM -"$BACK_PID" 2>/dev/null || kill -TERM "$BACK_PID" 2>/dev/null || true
    wait "$BACK_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

echo "=== [1/4] Backend (go run) ==="
cd "$ROOT/backend"
if command -v setsid >/dev/null 2>&1; then
  setsid go run ./cmd/server >"$BACKEND_LOG" 2>&1 &
else
  go run ./cmd/server >"$BACKEND_LOG" 2>&1 &
fi
BACK_PID=$!
cd "$ROOT"

echo "  待機: GET /api/v1/health"
for _ in $(seq 1 45); do
  if curl -sf "http://127.0.0.1:${PORT}/api/v1/health" >/dev/null; then
    echo "  -> OK"
    break
  fi
  if ! kill -0 "$BACK_PID" 2>/dev/null; then
    echo "  !! バックエンドが終了しました。ログ:"
    cat "$BACKEND_LOG"
    exit 1
  fi
  sleep 1
done
if ! curl -sf "http://127.0.0.1:${PORT}/api/v1/health" >/dev/null; then
  echo "  !! health に到達できません"
  cat "$BACKEND_LOG"
  exit 1
fi

echo "=== [2/4] Frontend (npm run dev) ==="
# 同一オリジン /api/v1 を使わせる（rewrite で WSL の 8080 へ）
unset NEXT_PUBLIC_API_URL || true
cd "$ROOT/frontend"
if [[ ! -d node_modules ]]; then
  npm install
fi
npm run dev >"$FRONTEND_LOG" 2>&1 &
FRONT_PID=$!

echo "  待機: GET /login"
for _ in $(seq 1 90); do
  if curl -sf "http://127.0.0.1:3000/login" >/dev/null; then
    echo "  -> OK"
    break
  fi
  sleep 1
done
if ! curl -sf "http://127.0.0.1:3000/login" >/dev/null; then
  echo "  !! フロントに到達できません"
  cat "$FRONTEND_LOG" | tail -80
  exit 1
fi

echo "=== [3/4] Playwright（ログイン + 管理画面）==="
export PLAYWRIGHT_API_URL="${PLAYWRIGHT_API_URL:-http://127.0.0.1:${PORT}/api/v1}"
export PLAYWRIGHT_BASE_URL="${PLAYWRIGHT_BASE_URL:-http://127.0.0.1:3000}"
# 必ず frontend で実行（カレントがズレると spec が見つからない／真っ白のみ等に見える）
cd "$ROOT/frontend"
# WS 未設定ならローカル Chromium。Windows で run-server 済みなら PLAYWRIGHT_WS_ENDPOINT を渡す
npx playwright test e2e/login.spec.ts e2e/login-admin-dashboard.spec.ts --reporter=list

echo "=== [4/4] 完了 ==="
echo "ログイン〜管理画面の E2E を通過しました。"
