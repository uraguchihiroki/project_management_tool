#!/bin/bash
# プロジェクト起動スクリプト（WSL / bash 用）
# 実行: bash scripts/start.sh
# または: ./scripts/start.sh
#
# Ctrl+C またはシェル終了（EXIT）時に、バックエンド（go）をプロセスグループ単位で停止する。
# フロントはフォアグラウンドの npm が受け取るシグナルで止まる（通常は Ctrl+C で終了）。

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

BACKEND_PID=""
CLEANUP_DONE=0

cleanup() {
  if [ "$CLEANUP_DONE" -eq 1 ]; then
    return 0
  fi
  CLEANUP_DONE=1
  trap - EXIT INT TERM

  echo ""
  echo "終了中..."

  if [ -n "${BACKEND_PID:-}" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
    # setsid で起動した子が PG リーダーなので、負の PID でグループ全体に SIGTERM
    kill -TERM -"$BACKEND_PID" 2>/dev/null || kill -TERM "$BACKEND_PID" 2>/dev/null || true
    sleep 1
    kill -KILL -"$BACKEND_PID" 2>/dev/null || kill -KILL "$BACKEND_PID" 2>/dev/null || true
    wait "$BACKEND_PID" 2>/dev/null || true
  fi

  exit 0
}

trap cleanup INT TERM
trap cleanup EXIT

echo "=== Project Management Tool 起動 ==="

# 1. Docker (PostgreSQL) 起動
echo "[1/3] PostgreSQL 起動..."
docker compose up -d db

echo "  -> DB の起動を待機中..."
sleep 5
until docker exec pmt_db pg_isready -U pmt_user -d pmt_db > /dev/null 2>&1; do
  sleep 2
done
echo "  -> DB 起動完了"

# 2. バックエンド起動（別セッション = まとめて kill しやすい）
echo "[2/3] バックエンド起動..."
if command -v ss >/dev/null 2>&1 && ss -tlnp 2>/dev/null | grep -q ':8080 '; then
  echo "  !! 警告: ポート 8080 は既に使用中です。古い go run が残っていると bind に失敗します。"
  echo "     例: ss -tlnp | grep 8080 で PID を確認し kill する"
fi

cd "$PROJECT_ROOT/backend"
# setsid: 子が新セッションのリーダーになり、kill -TERM -$PID でコンパイル子プロセスごと終了しやすい
if command -v setsid >/dev/null 2>&1; then
  setsid go run ./cmd/server </dev/null &
else
  go run ./cmd/server </dev/null &
fi
BACKEND_PID=$!
cd "$PROJECT_ROOT"
echo "  -> バックエンド PID=$BACKEND_PID"
sleep 2

# 3. フロントエンド起動（フォアグラウンド = ログ表示）
echo "[3/3] フロントエンド起動..."
cd "$PROJECT_ROOT/frontend"
# nvm を使っている場合
if [ -s "$HOME/.nvm/nvm.sh" ]; then
  # shellcheck disable=SC1090
  . "$HOME/.nvm/nvm.sh"
fi
# 初回は npm install
if [ ! -d "node_modules" ]; then
  echo "  -> npm install 実行中..."
  npm install
fi
echo ""
echo "ブラウザで http://localhost:3000 を開いてください"
echo "終了: Ctrl+C（バックエンドも停止を試みます）"
echo ""

npm run dev
