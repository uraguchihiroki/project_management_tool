# 開発ガイド

## 新機能追加の手順

### 1. モデル（必要に応じて）

- [backend/internal/model/model.go](backend/internal/model/model.go) に構造体を追加
- `main.go` の `db.AutoMigrate` に追加
- [db-schema.md](db-schema.md) を更新

### 2. Repository

- [backend/internal/repository/](backend/internal/repository/) に新規ファイルを作成
- CRUD や検索メソッドを実装
- ビジネスロジックは含めない（DB 操作のみ）

### 3. Service

- [backend/internal/service/](backend/internal/service/) に新規ファイルを作成
- ビジネスロジック、デフォルト値、採番、複数 Repository の調整を実装
- HTTP に依存しない形で書く（`echo.Context` を参照しない）

### 4. Handler

- [backend/internal/handler/](backend/internal/handler/) に新規ファイルを作成
- リクエストのパース、Service の呼び出し、レスポンスの返却のみ
- `main.go` にルートを追加
- [api-spec.md](api-spec.md) を更新
- テナント（組織境界）に触れる変更では [tenant-invariants.md](tenant-invariants.md)（不変条件の変更時）、[backend/test/TENANT_TEST_MATRIX.md](../backend/test/TENANT_TEST_MATRIX.md)（行の追加・`テナントBB` 列）も更新する

### 5. テスト

- [backend/test/](backend/test/) にテストを追加
- [setup_test.go](backend/test/setup_test.go) の `newTestServer` にルートを追加
- [testing.md](testing.md) を参照（**ブラックボックステストの AAA・日本語コメント・ログ・アンチパターン**は同ドキュメントの「記述仕様」節）

---

## 命名規則・ディレクトリ配置

| 種別 | 規則 | 例 |
|------|------|-----|
| Handler | `XxxHandler`, `xxx.go` | `handler/issue.go` → `IssueHandler` |
| Service | `XxxService`, `xxx.go` | `service/issue.go` → `IssueService` |
| Repository | `XxxRepository`, `xxx.go` | `repository/issue.go` → `IssueRepository` |
| モデル | `model/` に集約 | `model/model.go` または `model/xxx.go` |

---

## Bootstrap・初期投入（`*_bootstrap.go`）の規約

- **`workflow_bootstrap.go`** … 組織スコープの **Issue 用** `workflows` / `statuses`（および呼び出し側で付ける遷移シード）のみ。`projects` 行や `default_workflow_id` は更新しない。
- **`project_status_bootstrap.go`** … **プロジェクト進行**用の `project_statuses` のみ。
- **同一の `*_bootstrap.go` に Issue 用とプロジェクト進行用を混在させない**（別ファイルに分ける）。
- **`ProjectService`** は Issue ワークフロー行の作成・`default_workflow_id` の設定を行わない。プロジェクトにデフォルト Issue ワークフローを紐付けるのは **`IssueWorkflowProvisioner`**（`issue_workflow_provision.go`）、**`POST /projects/:id/default-issue-workflow`**、または **`IssueService.Create` 前段の lazy 確保**（実装参照）。

---

## seed.sql の扱い

- **実行**: `Get-Content backend/seed.sql | docker exec -i pmt_db psql -U pmt_user -d pmt_db`
- **べき等性**: `ON CONFLICT DO NOTHING` を使用しているため、複数回実行しても安全
- **変更時**: 既存環境への影響を考慮。`ALTER TABLE` や `UPDATE` は慎重に。新規カラムは `ADD COLUMN IF NOT EXISTS` を推奨

---

## 組織seed（CLI）

既存組織にステータス・役職・グループ・サンプルプロジェクト等を投入する:

```bash
# 全組織に投入
./scripts/seed-org.sh --all

# 指定組織に投入
./scripts/seed-org.sh <org-id> [owner-id]
```

または:

```bash
cd backend
go run ./cmd/cli org seed --all
go run ./cmd/cli org seed --org-id=<uuid> [--owner-id=<uuid>]
```

- **--all**: 全組織にseed投入。各組織の管理者をオーナーとしてサンプルプロジェクトを作成
- **org-id**: 組織のUUID
- **owner-id**: 省略時は組織の管理者を使用。管理者がいない組織はサンプルプロジェクト・Issueは作成しない
- **冪等**: 既存レコードは更新、なければ作成。何度でも実行可能

---

## ドキュメント更新のタイミング

| 変更内容 | 更新するドキュメント |
|----------|----------------------|
| API の追加・変更 | [api-spec.md](api-spec.md) |
| テーブル・モデルの変更 | [db-schema.md](db-schema.md) |
| アーキテクチャ・構成の変更 | [architecture.md](architecture.md) |
| 新規フローの追加 | [key-flows.md](key-flows.md) |
| ステータス遷移の権限の決定・変更 | [transition-permissions.md](transition-permissions.md)、[domain-model.md](domain-model.md)、[db-schema.md](db-schema.md) |
| テスト方針の変更 | [testing.md](testing.md) |

---

## いつバックエンド／フロントを再起動するか（Django 経験者向け）

**Django の `runserver`** はコード変更を検知して **自動リロード**しやすいので、「動かしっぱなしでだいたい追従する」感覚になりがちです。

このリポジトリは **Go + Next.js** なので、次のように考えると迷いにくいです。

| 対象 | ホットに近い挙動 | **再起動した方がよいタイミング** |
|------|------------------|-----------------------------------|
| **バックエンド**（`go run ./cmd/server`） | 基本なし（常に **プロセス＝ビルド結果**） | `.go` を直した／`git pull` した／依存や `go.mod` が変わった／**挙動がおかしい**ときは **必ず再起動** |
| **フロント**（`npm run dev`） | React や多くの TS は **Fast Refresh** | `next.config.js` 変更、**環境変数**（`NEXT_PUBLIC_*` や `.env*`）、ルート構成の大きな変更、**キャッシュのせいか分からない不具合**のときは再起動 |
| **両方** | — | **スキーマ・マイグレーション・API 契約を変えたあと**、ログインや API が古いままに見えるとき |

**覚え方**: 「Django の自動リロード相当」を **手でやるのが Go**。**フロントはだいたい HMR だが設定と env は再起動**。

`bash scripts/start.sh` 利用時は **Ctrl+C** でフロントとバックエンド停止を試みる。コードを大量に変えたあとは、**止めてからもう一度 `start.sh`** が確実です。

---

## トラブルシュート（ログイン・API が立たない）

- **IDE で `frontend/.../routes.d.ts` が無い／`next-env.d.ts` が赤い（clone 直後）**  
  Next の生成物 `.next` がまだ無い状態。[README.md](../README.md) の **「Cursor で TypeScript エラー（`routes.d.ts` が無い等）が出るとき」** を参照。`cd frontend` のうえ `npm install` の後、**`npm run build` または `npm run dev`** を一度実行する。
- **AI 向け**: 画面・挙動の不具合報告では、**いきなり実装を疑わず**、[AGENTS.md](../AGENTS.md) の **「### 6」** と [`.cursor/rules/browser-first-investigation.mdc`](../.cursor/rules/browser-first-investigation.mdc) に従い、**まずブラウザで再現**し、続けて **疎通・再起動**を確認する。
- **`bash scripts/start.sh` を使っているのにログインできない**  
  - スクリプトは **フロント（npm）をフォアグラウンド**にするため、ターミナルにはフロントのログばかり出ることがある。  
  - **`go run ./cmd/server` がマイグレーションで即終了している**と、**フロントだけ生きて 8080 に API がいない**状態になり、ログインが失敗する。  
  - 起動直後に `curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:8080/api/v1/health` で **200** になるか確認する。
- **レガシー DB で `statuses.workflow_id` まわり**  
  - 起動時に `PrepareStatusesWorkflowColumn`（`internal/db/legacy_status_workflow.go`）が **AutoMigrate より前**に走り、NULL の `workflow_id` を埋めてから NOT NULL 化する。  
  - 複数テナントで孤立行が残る場合はログに従い手動修正が必要なことがある。
- **ステータス重複の除去・再発防止**  
  - `AutoMigrate` の前に `MigrateIssueProjectStatusSplitPre`（`internal/db/migrate_issue_project_status.go`）が旧 `statuses.type` と「組織Project」ワークフローを整理する。続けて **`MigrateJunctionTablesSurrogatePK`**（`internal/db/migrate_junction_surrogate_pk.go`）が PostgreSQL 上の結合テーブル（`user_roles` 等）を **複合 PK → UUID 単独 PK** に移行する（SQLite はスキップ）。`AutoMigrate` の後に `MigrateDropLegacyBusinessUniqueIndexes` で旧業務 UNIQUE を除去し、`MigrateProjectStatusSeed` で `project_statuses` を欠けるプロジェクトへ投入し、その後に `MigrateStatusDedupe`（`internal/db/status_integrity.go`）が走る。既存の同一 `(workflow_id, name, display_order)` 重複を参照付け替えのうえ削除し、**非一意**の部分インデックス（任意）を付与する。業務一意は Service で保証する（[principles.md](principles.md)）。

---

## レイヤー責務の確認

詳細は [layer-responsibility.md](layer-responsibility.md) を参照。

- **Handler**: HTTP の入出力のみ。ビジネスロジックを持たない
- **Service**: ビジネスロジック。HTTP を知らない
- **Repository**: DB 操作のみ。ビジネスロジックを持たない
