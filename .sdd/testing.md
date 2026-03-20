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
- バックエンド（既定 `http://localhost:8080/api/v1`、環境変数 `PLAYWRIGHT_API_URL` で変更可）が起動していること
- フロントエンド（既定 `http://localhost:3000`、`PLAYWRIGHT_BASE_URL`）が起動していること
- **組織が少なくとも1件**あること（推奨: `backend/seed.sql`）。組織0件の場合、ログイン E2E は `POST /super-admin/login` → 必要なら `POST /super-admin/organizations` で組織を1件作成する（`super_admins` にシード済みのメールが必要。既定 `E2E_SUPER_ADMIN_EMAIL=superadmin@frs.example.com`）

**WSL / Linux でブラウザが起動しない場合**（例: `libnspr4.so: cannot open shared object file`）:

Chromium 用の OS 依存ライブラリが未インストールです。いずれかを実行してください。

```bash
cd frontend
npx playwright install-deps chromium
# または（Ubuntu/Debian 例）
# sudo apt-get install -y libnss3 libnspr4 libatk1.0-0 libdrm2 libxkbcommon0 libxcomposite1 libxdamage1 libxfixes3 libxrandr2 libgbm1 libasound2
```

**実行**（WSL / bash）:

```bash
cd frontend
npm run test:e2e
```

**ログイン結線の最小スモーク**（バックエンド＋フロント起動が前提）:

```bash
cd frontend
npm run test:e2e:login
```

**安定してログイン E2E を通すには（推奨）**:
- `NEXT_PUBLIC_API_URL` を API と一致させたうえで **`npm run build` → `npm run start`** でフロントを起動する。E2E では本番相当が安定。`npm run dev:turbo`（Turbopack）単体では開発ツール表示やコンパイル競合でログイン E2E が不安定になりやすい（通常の `npm run dev` は Webpack）。

```bash
# 例: API が localhost:8080 のとき
cd frontend
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1 npm run build
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1 npm run start
# 別ターミナルで
cd frontend && npm run test:e2e:login
```

Windows PowerShell を使う場合は `cd` のみ PowerShell 相当で置き換えて同じ `npm run` を実行してください。

**注意**: ルート変更（例: `/roles/reorder` → `/roles/bulk/reorder`）を行った場合は、**バックエンドを再起動**してください。

---

### Playwright Server（Windows でブラウザ、WSL でテストを実行）

WSL 内だけだと Chromium の OS 依存ライブラリが不足することがある一方、**Windows 側で Playwright のブラウザを動かし、WSL から WebSocket で接続**する構成が使える。

**Windows（PowerShell）** — サーバを起動したままにする:

```powershell
npx playwright run-server --port 9222
# Listening on ws://localhost:9222/
```

**WSL（bash）** — バックエンド・フロントを起動したうえで、エンドポイントを指定して E2E を実行する。

- 既定では `scripts/playwright-server-e2e.sh` が **default gateway**（WSL2 の Windows ホスト想定）を使い `ws://<そのIP>:9222/` を組み立てる。
- うまく繋がらない場合は **手動で** Windows ホスト IP を指定する（例: `grep nameserver /etc/resolv.conf` の第2列、または `ip route` の default 先）。

```bash
# 例: 明示指定してワークフロー作成 E2E のみ
export PLAYWRIGHT_WS_ENDPOINT="ws://<Windows側IP>:9222/"
bash scripts/playwright-server-e2e.sh e2e/admin-workflows-create.spec.ts
```

同等の npm スクリプト（`frontend` 配下）:

```bash
cd frontend
npm run test:e2e:server -- e2e/login.spec.ts
```

`frontend/playwright.config.ts` は `PLAYWRIGHT_WS_ENDPOINT` が設定されているとき **`connectOptions.wsEndpoint` で上記サーバに接続**する。未設定なら従来どおりローカル Chromium を起動する。

**E2E の認証**: ログイン画面を通さないテストは `frontend/e2e/helpers.ts` の `setupAuth` が `POST /admin/login` と `POST /admin/switch-organization` で JWT を取得し、`sessionStorage` に `authToken` / `currentUser` / `currentOrg` を入れる。

---

## テストファイル一覧とカバー範囲

| ファイル | カバー範囲 |
|----------|------------|
| [setup_test.go](backend/test/setup_test.go) | テストサーバー起動、共通ヘルパー（req, reqNoAuth, createTestUser, createTestProject 等） |
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
| [login_test.go](backend/test/login_test.go) | **一般ユーザーログイン**（`POST /admin/login`）正常系・異常系、JWT で `GET /users/:id/organizations` |
| [cross_org_authorization_test.go](backend/test/cross_org_authorization_test.go) | クロス組織アクセス不可（多テナント境界） |
| [frontend/e2e/login.spec.ts](frontend/e2e/login.spec.ts) | **Playwright**: `/login` からのログイン成功・失敗（UI＋API 結線） |
| [frontend/e2e/helpers.ts](frontend/e2e/helpers.ts) | E2E 用 `ensureLoginableUser` / `setupAuth`（JWT・組織スコープ） |
| [scripts/playwright-server-e2e.sh](../scripts/playwright-server-e2e.sh) | WSL から `PLAYWRIGHT_WS_ENDPOINT` を付与して `npm run test:e2e` を実行 |

---

## テストの特徴

- **認証**: 多くの API は `Authorization: Bearer <JWT>` が必須。`newTestServer` はスーパーアドミン用トークンを `ts.req` に付与。公開エンドポイント（`POST /users`, `POST /admin/login` 等）は `ts.reqNoAuth` を使用
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
