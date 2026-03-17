# API仕様

## ベースURL

```
http://localhost:8080/api/v1
```

---

## エンドポイント一覧

### Users（ユーザー）

| Method | Path | 説明 |
|---|---|---|
| GET | /users | ユーザー一覧取得 |
| POST | /users | ユーザー作成 |
| GET | /users/:id | ユーザー詳細取得 |

### Projects（プロジェクト）

| Method | Path | 説明 |
|---|---|---|
| GET | /projects | プロジェクト一覧取得 |
| POST | /projects | プロジェクト作成 |
| GET | /projects/:id | プロジェクト詳細取得 |
| PUT | /projects/:id | プロジェクト更新 |
| DELETE | /projects/:id | プロジェクト削除 |

### Issues（チケット）

| Method | Path | 説明 |
|---|---|---|
| GET | /projects/:projectId/issues | Issue一覧取得 |
| POST | /projects/:projectId/issues | Issue作成 |
| GET | /projects/:projectId/issues/:number | Issue詳細取得 |
| PUT | /projects/:projectId/issues/:number | Issue更新 |
| DELETE | /projects/:projectId/issues/:number | Issue削除 |

### Statuses（ステータス）

| Method | Path | 説明 |
|---|---|---|
| GET | /projects/:projectId/statuses | ステータス一覧取得 |
| POST | /projects/:projectId/statuses | ステータス作成 |
| PUT | /projects/:projectId/statuses/:id | ステータス更新 |
| DELETE | /projects/:projectId/statuses/:id | ステータス削除 |

### Comments（コメント）

| Method | Path | 説明 |
|---|---|---|
| GET | /issues/:issueId/comments | コメント一覧取得 |
| POST | /issues/:issueId/comments | コメント投稿 |
| PUT | /issues/:issueId/comments/:id | コメント更新 |
| DELETE | /issues/:issueId/comments/:id | コメント削除 |

---

## レスポンス形式

### 成功

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

## Issue作成 リクエスト例

```json
POST /api/v1/projects/proj-uuid/issues

{
  "title": "ログイン画面のバリデーション修正",
  "description": "メールアドレス形式のバリデーションが動いていない",
  "status_id": "status-uuid",
  "priority": "high",
  "assignee_id": "user-uuid",
  "due_date": "2026-03-31"
}
```

---

## 優先度の値

| 値 | 表示名 |
|---|---|
| low | 低 |
| medium | 中 |
| high | 高 |
| critical | 緊急 |
