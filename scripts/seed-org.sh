#!/bin/bash
# 組織にseedデータを投入する
# 使用例:
#   ./scripts/seed-org.sh --all              # 全組織に投入
#   ./scripts/seed-org.sh <org-id> [owner-id] # 指定組織に投入

set -e

if [ -z "$1" ]; then
  echo "Usage: $0 --all"
  echo "       $0 <org-id> [owner-id]"
  echo "  --all: 全組織にseed投入"
  echo "  org-id: 組織のUUID"
  echo "  owner-id: サンプルプロジェクトのオーナーUUID（省略時は組織の管理者を使用）"
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

if [ "$1" = "--all" ]; then
  ARGS="org seed --all"
else
  ARGS="org seed --org-id=$1"
  if [ -n "$2" ]; then
    ARGS="$ARGS --owner-id=$2"
  fi
fi

cd backend
go run ./cmd/cli $ARGS
