# テナント（組織）境界 — ブラックボックステスト マトリクス

**目的**: スーパーアドミン以外の JWT で呼べる API について、**他組織データが返らない／他組織 ID 操作で 403/404** であることをブラックボックステストで追跡する。

**凡例**

| 列 | 意味 |
|----|------|
| テナントBB | `done` = 意図したテナント用テストあり / `partial` = 間接のみ / `todo` = 未着手 |
| 備考 | スーパーアドミン専用は `—`（本マトリクス対象外） |

`main.go` の `api` グループ（`RequireJWT`）を棚卸しした一覧。更新時はルート追加・削除に合わせて本表を直すこと。

## 一覧

| Method | Path | テナントBB | 主なテスト・備考 |
|--------|------|------------|-------------------|
| GET | /users | **done** | `user_tenant_test.go`（一覧は JWT org のみ） |
| GET | /users/:id | **done** | `user_tenant_test.go`（他 org ユーザー 404） |
| POST | /admin/switch-organization | — | 組織切替（別観点） |
| PUT | /users/:id/admin | todo | |
| GET | /users/:id/roles | todo | |
| PUT | /users/:id/roles | todo | |
| GET | /roles | todo | org クエリと JWT |
| POST | /roles | todo | |
| PUT | /roles/bulk/reorder | todo | |
| PUT | /roles/:id | todo | |
| DELETE | /roles/:id | todo | |
| GET | /workflows | **done** | `workflow_tenant_test.go`（一覧 org のみ、org_id クエリ、SA+org_id） |
| POST | /workflows | **done** | `workflow_tenant_test.go`（body の他 org ID は無視して JWT org に作成）ほか |
| PUT | /workflows/reorder | **done** | `workflow_tenant_test.go`（他 org ID 混在は 403） |
| GET | /workflows/:id | **done** | `workflow_tenant_test.go`（他 org 404） |
| GET | /workflows/:id/statuses | **done** | `workflow_status_test.go`（別組織 404、重複は DB ユニークで防止） |
| POST | /workflows/:id/statuses | **done** | 同上 |
| PUT | /workflows/:id | **done** | `workflow_tenant_test.go`（他 org 404） |
| DELETE | /workflows/:id | **done** | 同上 |
| GET | /templates | **done** | `template_tenant_test.go`（他 org テンプレは一覧に出ない） |
| POST | /templates | **done** | `template_tenant_test.go`（他 org project_id は 403） |
| GET | /templates/:id | **done** | `template_tenant_test.go`（他 org 404） |
| PUT | /templates/:id | todo | |
| DELETE | /templates/:id | todo | |
| GET | /projects/:projectId/templates | todo | |
| PUT | /projects/:projectId/templates/reorder | todo | |
| GET | /organizations | todo | グローバル一覧の仕様確認 |
| POST | /organizations | todo | |
| GET | /users/:id/organizations | todo | |
| POST | /organizations/:orgId/users | todo | |
| GET | /organizations/:orgId/departments | todo | |
| POST | /organizations/:orgId/departments | todo | |
| PUT | /organizations/:orgId/departments/reorder | todo | |
| PUT | /organizations/:orgId/departments/:id | todo | |
| DELETE | /organizations/:orgId/departments/:id | todo | |
| GET | /users/:id/departments | todo | |
| PUT | /users/:id/departments | todo | |
| GET | /super-admin/organizations | — | SA 専用 |
| POST | /super-admin/organizations | — | SA 専用 |
| GET | /admin/users | todo | |
| POST | /admin/users | todo | |
| PUT | /admin/users/:id | todo | |
| DELETE | /admin/users/:id | todo | |
| GET | /projects | **done** | `project_tenant_test.go`（一覧は JWT org のみ） |
| GET | /organizations/:orgId/statuses | todo | |
| POST | /organizations/:orgId/statuses | todo | |
| PUT | /statuses/:id | todo | |
| DELETE | /statuses/:id | todo | |
| POST | /projects | **done** | `project_tenant_test.go`（他 org_id は 403） |
| POST | /projects/:id/default-issue-workflow | todo | |
| PUT | /projects/reorder | todo | |
| GET | /projects/:id | **done** | `project_tenant_test.go`（他 org 404） |
| PUT | /projects/:id | **done** | 同上 |
| DELETE | /projects/:id | **done** | 同上 |
| GET | /projects/:projectId/issues | todo | |
| POST | /projects/:projectId/issues | todo | |
| GET | /organizations/:orgId/issues | todo | |
| POST | /organizations/:orgId/issues | todo | |
| GET | /organizations/:orgId/issues/:number | todo | |
| PUT | /organizations/:orgId/issues/:number | todo | |
| DELETE | /organizations/:orgId/issues/:number | todo | |
| GET | /projects/:projectId/issues/:number | todo | |
| PUT | /projects/:projectId/issues/:number | todo | |
| DELETE | /projects/:projectId/issues/:number | todo | |
| GET | /organizations/:orgId/issue-events | todo | |
| GET | /issues/:issueId/events | todo | |
| GET | /issues/:issueId/comments | todo | |
| POST | /issues/:issueId/comments | todo | |
| PUT | /issues/:issueId/comments/:id | todo | |
| DELETE | /issues/:issueId/comments/:id | todo | |

## 第1波（完了）

- `GET /workflows` のテナント境界: [workflow_tenant_test.go](workflow_tenant_test.go)

## 今後

- 上表の `todo` をカテゴリ（Issues / Projects / …）ごとに潰し、`partial` を `done` に上げる。
- 既存の `*_test.go` にテナント観点のアサーションがある場合は「主なテスト」にファイル名を追記する。

## 過去バグ・デグレ防止レジストリ

一度誤解・不具合になった点を**短く固定**し、同じ解釈ミスで実装しないための参照用。詳しい不変条件は [.sdd/tenant-invariants.md](../../.sdd/tenant-invariants.md)。

| 論点 | 正しい理解 | テスト・根拠 |
|------|------------|--------------|
| `GET /workflows`（スーパーアドミン） | `org_id` なしでは全組織分を返し得る。管理画面で「選択中の1社」だけ見せるときは **`org_id` をクエリで付け、サーバがその組織だけ返す**（フロントだけで絞らない）。非スーパーアドミンは JWT の org のみ。 | [workflow_tenant_test.go](workflow_tenant_test.go) |
| `GET/POST /workflows/:id/statuses` | テナント境界は **親ワークフローが JWT の組織に属するか**の検証。通過後の列挙は **`workflow_id = :id`**。同一 `(name, display_order)` の重複は **Service** で拒否（起動時 `MigrateStatusDedupe` はレガシー重複の整理のみ）。 | [workflow_status_test.go](workflow_status_test.go)（別組織 404、重複 POST は 400） |
