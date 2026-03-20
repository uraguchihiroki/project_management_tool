#!/bin/bash
# プロジェクト起動スクリプト（WSL / bash 用）
# 実行: bash scripts/start.sh
# または: ./scripts/start.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

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

# 2. バックエンド起動（バックグラウンド）
echo "[2/3] バックエンド起動..."
cd "$PROJECT_ROOT/backend"
go run ./cmd/server &
BACKEND_PID=$!
cd "$PROJECT_ROOT"

# Ctrl+C でバックエンドも終了させる
cleanup() {
  echo ""
  echo "終了中..."
  kill $BACKEND_PID 2>/dev/null || true
  exit 0
}
trap cleanup SIGINT SIGTERM

sleep 3

# 3. フロントエンド起動（フォアグラウンド = ログ表示）
echo "[3/3] フロントエンド起動..."
cd "$PROJECT_ROOT/frontend"
# nvm を使っている場合
if [ -s "$HOME/.nvm/nvm.sh" ]; then
  . "$HOME/.nvm/nvm.sh"
fi
# 初回は npm install
if [ ! -d "node_modules" ]; then
  echo "  -> npm install 実行中..."
  npm install
fi
echo ""
echo "ブラウザで http://localhost:3000 を開いてください"
echo "終了: Ctrl+C"
echo ""
npm run dev
