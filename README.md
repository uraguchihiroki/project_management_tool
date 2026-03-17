# Project Management Tool

Jira / Redmine ライクなチケットベースのプロジェクト管理ツール。  
マルチテナント対応（会社・組織ごとにデータを分離）。

## 技術スタック

| 役割 | 技術 |
|---|---|
| フロントエンド | Next.js 14 (React / TypeScript / Tailwind CSS) |
| バックエンド | Go 1.22 + Echo v4 + GORM |
| データベース | PostgreSQL 16 |
| コンテナ | Docker Compose |

---

## ログイン URL まとめ

| ユーザー種別 | URL | 備考 |
|---|---|---|
| 一般ユーザー / 組織管理者 | http://localhost:3000/login | メールアドレスのみで登録・ログイン |
| スーパー管理者 | http://localhost:3000/super-admin/login | 会社・組織の作成のみ可能 |

### 初期スーパー管理者アカウント

| 項目 | 値 |
|---|---|
| メールアドレス | `superadmin@frs.example.com` |
| パスワード | なし（メールアドレスのみ） |

> **Note:** seed.sql を実行済みの環境でのみ有効。メールアドレスは `backend/seed.sql` 内で変更可能。

### ログイン後の動作

- 所属組織が **1 件** → 自動選択されてプロジェクト一覧へ
- 所属組織が **複数** → `/select-org` で組織を選択してからプロジェクト一覧へ
- ログイン画面の「管理者としてログイン」チェックを付けると管理画面 (`/admin`) にアクセス可能

---

## ローカル起動手順

### 前提条件

- Docker Desktop がインストールされていること
- Go 1.22+ がインストールされていること
- Node.js 20+ がインストールされていること

---

### 1. Docker（PostgreSQL）の起動

```powershell
# プロジェクトルートで実行
docker-compose up -d db
```

起動確認:

```powershell
docker ps
# pmt_db が "Up" かつ "(healthy)" になっていればOK
```

---

### 2. バックエンドの起動

```powershell
cd backend

# 初回またはコード変更後はビルドが必要
go build -o server.exe ./cmd/server/

# サーバー起動
.\server.exe
```

起動確認:

```powershell
# 別ターミナルで
Invoke-RestMethod http://localhost:8080/api/v1/organizations
# {"data":[...]} が返ればOK
```

> **go run ではなく go build を使う理由:**  
> `go run` は毎回一時ファイルを生成するため、Windows ファイアウォールが毎回ネットワーク許可を求めてくる。  
> `go build` で `server.exe` を作れば最初の一度だけ許可すれば済む。

---

### 3. フロントエンドの起動

```powershell
cd frontend

# 初回のみ（依存パッケージのインストール）
npm install

# 開発サーバー起動
npm run dev
```

起動確認:  
ブラウザで http://localhost:3000 を開く

---

### 4. 初期データの投入（初回のみ）

初回起動時はデータベースが空の状態です。以下を実行して初期データを投入してください。

```powershell
# プロジェクトルートで実行
Get-Content backend/seed.sql | docker exec -i pmt_db psql -U pmt_user -d pmt_db
```

実行されること:

| 処理 | 内容 |
|---|---|
| 組織作成 | 「Ｆ．Ｒ．Ｓ．」を固定 UUID で挿入 |
| データ紐付け | 既存のプロジェクト・役職を Ｆ．Ｒ．Ｓ．組織に紐付け |
| FK 制約追加 | `projects.organization_id`, `roles.organization_id` を NOT NULL + FK 化 |
| ユーザー追加 | 既存ユーザーを全員 Ｆ．Ｒ．Ｓ．組織のメンバーに追加 |
| SA 作成 | スーパー管理者 `superadmin@frs.example.com` を挿入 |

> **べき等:** `ON CONFLICT DO NOTHING` を使っているため、2 回以上実行しても安全。

---

### 5. 起動順序のまとめ

```
[Terminal 1]  docker-compose up -d db
[Terminal 2]  cd backend  →  go build -o server.exe ./cmd/server/  →  .\server.exe
[Terminal 3]  cd frontend  →  npm run dev
[初回のみ]   Get-Content backend/seed.sql | docker exec -i pmt_db psql -U pmt_user -d pmt_db
```

---

### 6. ローカル環境の停止手順

```powershell
# [Terminal 2] バックエンド → Ctrl+C で停止

# [Terminal 3] フロントエンド → Ctrl+C で停止

# [Terminal 1] Docker（PostgreSQL）を停止
docker-compose stop db
```

> **データを消さずに止める場合** は `stop`。  
> **コンテナごと削除する場合**（DB データも消える）は `docker-compose down`。

---

## アクセス先一覧

| サービス | URL |
|---|---|
| フロントエンド（一般） | http://localhost:3000 |
| フロントエンド（スーパー管理者） | http://localhost:3000/super-admin/login |
| バックエンド API | http://localhost:8080/api/v1 |
| PostgreSQL | localhost:5432 |

---

## ドキュメント

設計資料は [.sdd/](.sdd/) フォルダを参照。

| ドキュメント | 内容 |
|---|---|
| [architecture.md](.sdd/architecture.md) | システムアーキテクチャ |
| [layer-responsibility.md](.sdd/layer-responsibility.md) | Handler / Service / Repository の責務分担 |
| [db-schema.md](.sdd/db-schema.md) | DB 設計・テーブル定義 |
| [api-spec.md](.sdd/api-spec.md) | REST API 仕様 |

---

## テスト実行

```powershell
cd backend
go test ./test/... -v
```

- テスト DB: インメモリ SQLite（PostgreSQL 不要）
- テスト件数: 150+
