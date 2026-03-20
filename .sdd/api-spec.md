# API仕様

## ベースURL

```
http://localhost:8080/api/v1
```

## 認証・マルチテナント制御

- `POST /users`、`POST /admin/login`、`POST /super-admin/login` を除く API は `Authorization: Bearer <JWT>` が必要（`POST /admin/switch-organization` は要 JWT）。
- スーパーアドミン以外は、JWT の `organization_id` に一致するデータのみ返却する。
- 他組織の `org_id` / `project_id` / `issue_id` 等を指定した場合は、`403` または `404` を返す。

---

## エンドポイント一覧

### Users（ユーザー）

| Method | Path | 説明 |
|--------|------|------|
| POST | /admin/login | 組織ユーザーログイン（JWT発行） |
| POST | /admin/switch-organization | 同一メールで別組織に切り替え（body: `organization_id`、該当組織のユーザー行に紐づく JWT を再発行） |
| GET | /users | ユーザー一覧取得 |
| POST | /users | ユーザー作成 |
| GET | /users/:id | ユーザー詳細取得 |
| PUT | /users/:id/admin | 管理者フラグ設定 |
| GET | /users/:id/roles | ユーザーの役職一覧取得 |
| PUT | /users/:id/roles | ユーザーに役職を割り当て |

### Roles（役職）

| Method | Path | 説明 |
|--------|------|------|
| GET | /roles | 役職一覧取得（org_id クエリで組織フィルタ可） |
| POST | /roles | 役職作成 |
| PUT | /roles/:id | 役職更新 |
| DELETE | /roles/:id | 役職削除 |

### Workflows（ワークフロー）

| Method | Path | 説明 |
|--------|------|------|
| GET | /workflows | ワークフロー一覧取得（組織スコープ。ステップ未追加の行も含む） |
| POST | /workflows | ワークフロー作成。スーパーアドミンは body に `organization_id` 必須。それ以外は JWT の組織スコープで作成（body の organization_id は無視可） |
| GET | /workflows/:id | ワークフロー詳細取得 |
| PUT | /workflows/:id | ワークフロー更新 |
| DELETE | /workflows/:id | ワークフロー削除 |
| POST | /workflows/:id/steps | ステップ追加（初回は sts_start + user + sts_goal を自動作成） |
| PUT | /workflows/:id/steps/:stepId | ステップ更新（next_status_id は無視。承認後ステータスは ReorderSteps でのみ更新） |
| PUT | /workflows/:id/steps/reorder | ステップ並び替え。ユーザーステップ ID の並び順のみ受け取り、承認後ステータスを確定 |
| DELETE | /workflows/:id/steps/:stepId | ステップ削除 |
| GET | /projects/:projectId/workflows | プロジェクトのワークフロー一覧 |

### Templates（Issueテンプレート）

| Method | Path | 説明 |
|--------|------|------|
| GET | /templates | テンプレート一覧取得 |
| POST | /templates | テンプレート作成 |
| GET | /templates/:id | テンプレート詳細取得 |
| PUT | /templates/:id | テンプレート更新 |
| DELETE | /templates/:id | テンプレート削除 |
| GET | /projects/:projectId/templates | プロジェクトのテンプレート一覧 |

### Approvals（承認）

| Method | Path | 説明 |
|--------|------|------|
| GET | /issues/:issueId/approvals | Issue の承認一覧取得 |
| POST | /approvals/:id/approve | 承認 |
| POST | /approvals/:id/reject | 却下 |

### Organizations（組織）

| Method | Path | 説明 |
|--------|------|------|
| GET | /organizations | 組織一覧取得 |
| POST | /organizations | 組織作成 |
| GET | /users/:id/organizations | ユーザーの所属組織（1ユーザー＝1組織のため1件） |
| POST | /organizations/:orgId/users | 組織にユーザーを追加（既存ユーザーの name/email で新規ユーザーを作成） |
| GET | /organizations/:orgId/statuses | 組織のステータス一覧。`?type=issue` で Issue 用にフィルタ。`?exclude_system=1` で sts_start/sts_goal を除外 |

### Super Admin

| Method | Path | 説明 |
|--------|------|------|
| POST | /super-admin/login | スーパー管理者ログイン（JWT発行） |
| GET | /super-admin/organizations | 組織一覧取得 |
| POST | /super-admin/organizations | 組織作成 |

### Admin（組織管理者向けユーザー管理）

| Method | Path | 説明 |
|--------|------|------|
| GET | /admin/users | ユーザー一覧取得（役職・組織付き、org_id クエリでフィルタ可） |
| POST | /admin/users | 組織にユーザーを作成（org_id 必須） |
| PUT | /admin/users/:id | ユーザー更新 |
| DELETE | /admin/users/:id | ユーザー削除（1ユーザー＝1組織のため、org_id で所属確認後に削除） |

### Projects（プロジェクト）

| Method | Path | 説明 |
|--------|------|------|
| GET | /projects | プロジェクト一覧取得（org_id クエリでフィルタ可） |
| POST | /projects | プロジェクト作成 |
| GET | /projects/:id | プロジェクト詳細取得 |
| PUT | /projects/:id | プロジェクト更新 |
| DELETE | /projects/:id | プロジェクト削除 |

> **Note:** プロジェクト詳細取得（GET /projects/:id）のレスポンスに `statuses` が含まれる。ステータスはプロジェクト作成時に自動生成され、専用の CRUD API はない。

### Issues（チケット）

| Method | Path | 説明 |
|--------|------|------|
| GET | /projects/:projectId/issues | Issue一覧取得 |
| POST | /projects/:projectId/issues | Issue作成 |
| GET | /projects/:projectId/issues/:number | Issue詳細取得 |
| PUT | /projects/:projectId/issues/:number | Issue更新 |
| DELETE | /projects/:projectId/issues/:number | Issue削除 |

### Comments（コメント）

| Method | Path | 説明 |
|--------|------|------|
| GET | /issues/:issueId/comments | コメント一覧取得 |
| POST | /issues/:issueId/comments | コメント投稿 |
| PUT | /issues/:issueId/comments/:id | コメント更新 |
| DELETE | /issues/:issueId/comments/:id | コメント削除 |

---

## レスポンス形式

### 成功（単体）

```json
{
  "data": { ... },
  "message": "success"
}
```

### 一覧

```json
{
  "data": [ ... ],
  "total": 100,
  "page": 1,
  "per_page": 20
}
```

### エラー

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "リソースが見つかりません"
  }
}
```

---

## リクエスト例

### Issue作成

```json
POST /api/v1/projects/proj-uuid/issues

{
  "title": "ログイン画面のバリデーション修正",
  "description": "メールアドレス形式のバリデーションが動いていない",
  "status_id": "status-uuid",
  "priority": "high",
  "assignee_id": "user-uuid",
  "reporter_id": "user-uuid",
  "due_date": "2026-03-31"
}
```

### プロジェクト作成

```json
POST /api/v1/projects

{
  "key": "PROJ",
  "name": "サンプルプロジェクト",
  "description": "説明",
  "owner_id": "user-uuid",
  "organization_id": "org-uuid"  // 必須
}
```

### ステップ並び替え（承認後ステータス確定）

```json
PUT /api/v1/workflows/:id/steps/reorder

{
  "ids": [3, 1, 2]  // ユーザーステップ ID の並び順のみ。sts_start/sts_goal は含まない
}
```

レスポンス: 204 No Content。並び順に応じて各ステップの `next_status_id` が自動更新される。

### ワークフロー作成

```json
POST /api/v1/workflows

{
  "organization_id": "org-uuid",  // 必須
  "name": "承認フロー",
  "description": "説明"
}
```

---

## 優先度の値

| 値 | 表示名 |
|----|--------|
| low | 低 |
| medium | 中 |
| high | 高 |
| critical | 緊急 |
