# テスト方針

## 概要

- **方式**: ブラックボックステスト（HTTP API 経由）
- **テスト DB**: インメモリ SQLite（PostgreSQL 不要）
- **実行**: `go test ./test/... -v`

---

## 実行方法

### バックエンド単体テスト

```powershell
cd backend
go test ./test/... -v
```

### E2E テスト（Playwright）

役職のドラッグ並び替えを検証する E2E テストがあります。

**前提条件**:
- バックエンド（`localhost:8080`）が起動していること
- フロントエンド（`localhost:3000`）が起動していること
- `seed.sql` 実行済みで組織（FRS）が存在すること

**実行**:
```powershell
cd frontend
npm run test:e2e
```

**注意**: ルート変更（例: `/roles/reorder` → `/roles/bulk/reorder`）を行った場合は、**バックエンドを再起動**してください。

---

## テストファイル一覧とカバー範囲

| ファイル | カバー範囲 |
|----------|------------|
| [setup_test.go](backend/test/setup_test.go) | テストサーバー起動、共通ヘルパー（req, createTestUser, createTestProject 等） |
| [user_test.go](backend/test/user_test.go) | ユーザー CRUD（Create, List, Get） |
| [project_test.go](backend/test/project_test.go) | プロジェクト CRUD、正常系フロー（一覧→作成→取得→更新→削除） |
| [issue_test.go](backend/test/issue_test.go) | Issue CRUD |
| [comment_test.go](backend/test/comment_test.go) | コメント CRUD |
| [role_test.go](backend/test/role_test.go) | 役職 CRUD、ユーザーへの役職割り当て、管理者一覧 |
| [workflow_test.go](backend/test/workflow_test.go) | ワークフロー CRUD、ステップ追加・更新・削除 |
| [department_test.go](backend/test/department_test.go) | 部署 CRUD、正常系フロー（一覧→作成→更新→削除）、ユーザー部署紐づけ |
| [template_test.go](backend/test/template_test.go) | テンプレート CRUD、テンプレートからの Issue 作成 |
| [approval_test.go](backend/test/approval_test.go) | 承認の自動作成、承認/却下、レベル・順序チェック |
| [organization_test.go](backend/test/organization_test.go) | 組織 CRUD、ユーザー追加、SuperAdmin ログイン、管理画面ユーザー一覧 |

---

## テストの特徴

- **認証なし**: 現状の API は認証ミドルウェアがないため、テストでは認証ヘッダー不要
- **組織スコープ**: テスト用固定組織 ID（`testOrgID`）を使用。プロジェクト・役職は組織に紐づけて作成
- **ステータス取得**: プロジェクト作成時にデフォルトステータスが自動作成される。Issue 作成時は `getFirstStatusID` で取得

---

## 新規 API 追加時のテスト追加手順

1. [setup_test.go](backend/test/setup_test.go) の `newTestServer` にルートを追加（既存の api グループに `api.GET/POST/...` を追加）
2. 新規テストファイルを作成（例: `xxx_test.go`）
3. `newTestServer(t)` でサーバーを取得
4. `ts.req(t, "METHOD", "/api/v1/path", body)` でリクエスト送信
5. `assertStatus`, `mustGetString`, `mustGetArray` 等でレスポンスを検証

### 例

```go
func TestXxx_Create(t *testing.T) {
	ts := newTestServer(t)
	userID := createTestUser(t, ts, "User", "user@example.com")
	projectID := createTestProject(t, ts, "PRJ", "Project", userID)

	status, resp := ts.req(t, "POST", "/api/v1/xxx", map[string]string{
		"field": "value",
	})
	assertStatus(t, status, http.StatusCreated, "create Xxx")
	id := mustGetString(t, resp, "data", "id")
	if id == "" {
		t.Error("id should not be empty")
	}
}
```
