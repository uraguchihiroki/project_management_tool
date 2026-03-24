# API仕様

**方針**: Issue 管理を主目的とする。**ワークフロー／承認**用エンドポイントは、設計上の正規 API からは **除外**（コードに残る場合はレガシー扱いで移行予定）。以下の一覧はその方針に合わせて記載する。

### 仕様の三本柱（.sdd README の **5 → 6 → 7**）

| 順 | ドキュメント | 内容 |
|----|--------------|------|
| **5** | [db-schema.md](db-schema.md) | **インプリント**（`issue_events` の追記行）、**イベントログ**、**Group** と **Issue↔Group / User↔Group** 多対多 |
| **6** | 本ドキュメント（api-spec） | 一覧・フィルタ・**イベント取得**など **クエリしやすい** API 契約 |
| **7** | [transition-permissions.md](transition-permissions.md) | **許可される遷移（形）**、**遷移アラート**、**監査の意味**（運用・通知・ログの分担） |

---

## ベースURL

```
http://localhost:8080/api/v1
```

## 認証・マルチテナント制御

- `POST /users`、`POST /admin/login`、`POST /super-admin/login` を除く API は `Authorization: Bearer <JWT>` が必要（`POST /admin/switch-organization` は要 JWT）。
- スーパーアドミン以外は、JWT の `organization_id` に一致するデータのみ返却する。
- 他組織の `org_id` / `project_id` / `issue_id` 等を指定した場合は、`403` または `404` を返す。

### テナント境界のパターン

- **組織の壁はサーバが張る。** フロントだけで一覧を会社 ID で絞って表示しても、テナント分離の代わりにはならない。
- **親子リソース**（例: ワークフロー配下のステータス）: パスの親 ID について「JWT の組織に属するか」を先に検証し、通ったあとは子を **親 ID**（例: `workflow_id`）で列挙する。詳細は [tenant-invariants.md](tenant-invariants.md) の (B)。
- **スーパーアドミン**で「選択中の1組織」だけを出す画面では、クエリに `org_id`（または仕様で定めた同等パラメータ）を付け、**サーバが**その組織の行だけ返す。全件取得してクライアントだけで絞るのはセキュリティの代わりにならない。

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

### Templates（Issueテンプレート）

| Method | Path | 説明 |
|--------|------|------|
| GET | /templates | テンプレート一覧取得 |
| POST | /templates | テンプレート作成 |
| GET | /templates/:id | テンプレート詳細取得 |
| PUT | /templates/:id | テンプレート更新 |
| DELETE | /templates/:id | テンプレート削除 |
| GET | /projects/:projectId/templates | プロジェクトのテンプレート一覧 |

### Organizations（組織）

| Method | Path | 説明 |
|--------|------|------|
| GET | /organizations | 組織一覧取得 |
| POST | /organizations | 組織作成 |
| GET | /users/:id/organizations | ユーザーの所属組織（1ユーザー＝1組織のため1件） |
| POST | /organizations/:orgId/users | 組織にユーザーを追加（既存ユーザーの name/email で新規ユーザーを作成） |
| GET | /organizations/:orgId/statuses | 組織のステータス一覧。`?type=issue` で Issue 用にフィルタ。`?exclude_system=1` で sts_start/sts_goal を除外 |
| POST | /organizations/:orgId/statuses | 組織の「組織Issue」「組織Project」固定ワークフローへステータス追加 |

### Workflows（組織スコープのワークフロー）

| Method | Path | 説明 |
|--------|------|------|
| GET | /workflows | ワークフロー一覧。**非スーパーアドミン**: JWT の `organization_id` のワークフローのみ。クエリ `org_id` を付ける場合は JWT と同一 UUID 必須（不一致は 403）。**スーパーアドミン**: `org_id` 省略時は全組織分、`org_id` 指定時はその組織のみ。無効な UUID は 400。 |
| POST | /workflows | 作成（body: `organization_id`, `name`, `description`） |
| PUT | /workflows/reorder | 表示順更新（body: `ids`） |
| GET | /workflows/:id | 詳細取得（他組織の ID は 404） |
| GET | /workflows/:id/statuses | **当該ワークフローに紐づく Status 一覧**（`order` 昇順）。`:id` のワークフローが JWT の組織に属さなければ **403/404**。列挙は `workflow_id = :id`。同一 `(name,type,order)` の重複は **DB 制約・マイグレーションで防ぐ**（[tenant-invariants.md](tenant-invariants.md)、[db-schema.md](db-schema.md)）。 |
| POST | /workflows/:id/statuses | **ステータス追加**。上記と同様に親ワークフローの組織を先に検証。body: `name`（必須）, `color`（省略時 `#6B7280`）, `type`（`issue` \| `project`、省略時 `issue`）, `order`（`0` または省略時は同一 WF 内の最大 `order` + 1）。作成後、当該 WF の **許可遷移を全ペア再シード**（Issue のステータス変更と整合） |
| PUT | /workflows/:id | 名前・説明の更新 |
| DELETE | /workflows/:id | 削除 |

### Statuses（個別更新・削除）

| Method | Path | 説明 |
|--------|------|------|
| PUT | /statuses/:id | 更新 |
| DELETE | /statuses/:id | 削除 |

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

### Groups（グループ）

組織内の **Group**（開示・共同文脈・通知の宛先・タグ的用途）。**Issue 文脈を主**とし、HR ディレクトリと完全一致させる必要はない（同期は `kind` 等で表現可能）。

| Method | Path | 説明 |
|--------|------|------|
| GET | /organizations/:orgId/groups | グループ一覧（`?kind=` でフィルタ可） |
| POST | /organizations/:orgId/groups | グループ作成 |
| GET | /groups/:id | グループ詳細 |
| PUT | /groups/:id | グループ更新 |
| DELETE | /groups/:id | グループ削除 |
| GET | /groups/:id/members | メンバー（User）一覧 |
| PUT | /groups/:id/members | メンバー一括置換または差分（実装で確定） |
| GET | /users/:id/groups | ユーザーが所属するグループ一覧 |

### Issues（チケット）

| Method | Path | 説明 |
|--------|------|------|
| GET | /projects/:projectId/issues | Issue一覧取得。**クエリ例**: `group_id`（Group に紐づく Issue のみ）、`status_id`、`assignee_id`、期間は **`updated_at` またはイベント API** と整合させる（一覧は軽量を優先） |
| POST | /projects/:projectId/issues | Issue作成（body に `group_ids` 任意） |
| GET | /projects/:projectId/issues/:number | Issue詳細取得（**groups** を含めてもよい） |
| PUT | /projects/:projectId/issues/:number | Issue更新（ステータス・担当変更時はサーバが **issue_events** に追記する想定） |
| DELETE | /projects/:projectId/issues/:number | Issue削除 |
| GET | /projects/:projectId/issues/:number/groups | Issue に付いた Group 一覧 |
| PUT | /projects/:projectId/issues/:number/groups | Issue ↔ Group の紐付け更新（多対多） |

### Issue events（インプリント・イベントログ・監査向け）

**追記のみ**の **インプリント**列を読む API（[db-schema.md](db-schema.md) の `issue_events` と対応。各行＝1 事実）。

| Method | Path | 説明 |
|--------|------|------|
| GET | /issues/:issueId/events | 当該 Issue の **インプリント**の時系列（`occurred_at` 昇順） |
| GET | /organizations/:orgId/issue-events | 組織横断。**クエリ例**: `event_type`、`from_occurred_at`、`to_occurred_at`（発生時刻の範囲。DB は `TIMESTAMPTZ`）、`actor_id`、`issue_id`。**監査**: 「完了遷移で `actor_id` = `assignee_id_at_occurred`」等はクライアントまたはバックエンドのレポートで集計 |

> **Note:** レスポンスの時刻は **ISO 8601（タイムゾーン付き）** とし、インプリントの `occurred_at`（DB `TIMESTAMPTZ`）と一致させる。

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

---

## 優先度の値

| 値 | 表示名 |
|----|--------|
| low | 低 |
| medium | 中 |
| high | 高 |
| critical | 緊急 |
