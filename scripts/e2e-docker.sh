#!/bin/bash
# E2E テストを Docker 内で実行（libnspr4 等の依存関係なしで動作）
# db + backend + frontend を自動起動してから E2E を実行
# Usage: ./scripts/e2e-docker.sh

set -e
cd "$(dirname "$0")/.."

echo "Running E2E tests via docker compose (db + backend + frontend + e2e)..."
echo ""

docker compose -f docker-compose.e2e.yml run --rm e2e
