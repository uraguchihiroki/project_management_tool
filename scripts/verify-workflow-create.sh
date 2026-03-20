#!/usr/bin/env bash
# ワークフロー追加（組織ユーザーの JWT で POST /workflows）をブラックボックステストで検証する。
#
# - backend/test は SQLite メモリ上で httptest サーバを起動するため、PostgreSQL・Docker・sudo は不要。
# - 実ブラウザでの確認は別途 scripts/start.sh（DB は docker compose）が必要。
#
# 実行: bash scripts/verify-workflow-create.sh
set -e
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT/backend"
echo "=== verify: TestWorkflow_Create_AsOrgUserJWT (org JWT, not super-admin) ==="
go test ./test/... -v -count=1 -run '^TestWorkflow_Create_AsOrgUserJWT$'
echo "=== OK ==="
