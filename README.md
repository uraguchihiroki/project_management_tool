# Project Management Tool

Jira / Redmine ライクなチケットベースのプロジェクト管理ツール。  
マルチテナント対応（会社・組織ごとにデータを分離）。

## 技術スタック

| 役割 | 技術 |
|---|---|
| フロントエンド | Next.js 14 (React / TypeScript / Tailwind CSS) |
| バックエンド | Go 1.22+ + Echo v4 + GORM |
| データベース | PostgreSQL 16 |
| コンテナ | Docker Compose |

---

## 開発環境（WSL Ubuntu 推奨）

本プロジェクトは **WSL2 上の Ubuntu 22.04.5** での開発を推奨します。

- **Docker Desktop は使用しない**（WSL 内の Docker Engine を使用）
- **PowerShell は使用しない**（bash を使用）
- Windows 版 Cursor から WSL にリモート接続して開発
- ブラウザ（Chrome 等）は Windows のものを使用可能（`localhost:3000` でアクセス）

### 初回セットアップ

1. WSL Ubuntu を用意し、GitHub から clone:
   ```bash
   mkdir -p ~/work/AI
   cd ~/work/AI
   git clone git@github.com:uraguchihiroki/project_management_tool.git
   cd project_management_tool
   ```
   > HTTPS 利用時: `https://github.com/uraguchihiroki/project_management_tool.git`  
   > clone 前に SSH 鍵または `gh auth login` の設定が必要です。

2. セットアップスクリプトを実行（Docker, Go, Node.js, GitHub CLI, GCP CLI をインストール）:
   ```bash
   bash scripts/setup-wsl.sh
   ```

3. 新しいターミナルを開き、Docker を起動:
   ```bash
   sudo service docker start   # または sudo systemctl start docker
   ```

4. Cursor で WSL に接続: `Ctrl+Shift+P` → 「WSL: Connect to WSL」→ `~/work/AI/project_management_tool` を開く

### Cursor の起動方法（重要）

**Cursor は Windows で起動し、WSL にリモート接続する**構成を推奨します。

| 推奨 | 非推奨 |
|------|--------|
| Windows で Cursor を起動 → WSL に接続 → プロジェクトを開く | WSL から `cursor .` で起動 |

WSL のターミナルから `cursor .` で起動すると、各種トラブルが発生することがあります。正しい手順:

1. **Windows** で Cursor を起動（スタートメニューやショートカットから）
2. `Ctrl+Shift+P` → 「**Remote: Connect to WSL**」または「**Connect to WSL**」を実行
3. WSL に接続後、`File` → `Open Folder` で `~/work/AI/project_management_tool` を開く

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

## ローカル起動手順（WSL / bash）

### 毎回の起動手順（クイックリファレンス）

初回セットアップ後は、次の 3 ステップでアプリを起動できます。詳細は [.sdd/dev-environment.md](.sdd/dev-environment.md) を参照。

1. **Docker を起動**（WSL のターミナルで）:
   ```bash
   sudo service docker start
   ```

2. **Cursor で WSL に接続** → プロジェクトフォルダ `~/work/AI/project_management_tool` を開く

3. **アプリを起動**（Cursor のターミナルで）:
   ```bash
   bash scripts/start.sh
   ```

ブラウザで `http://localhost:3000` にアクセス。

---

### 前提条件

- WSL Ubuntu 22.04.5 上に以下がインストールされていること:
  - Docker Engine（Docker Desktop は使用しない）
  - Go 1.22+
  - Node.js 20+
- `scripts/setup-wsl.sh` を実行済みであること

---

### 方法 A: 一括起動（推奨）

```bash
bash scripts/start.sh
```

DB → バックエンド → フロントエンドの順で起動します。終了は `Ctrl+C`。

---

### 方法 B: ターミナルを分けて起動

#### 1. Docker（PostgreSQL）の起動

```bash
# プロジェクトルートで実行
docker compose up -d db
```

起動確認:

```bash
docker ps
# pmt_db が "Up" かつ "(healthy)" になっていればOK
```

---

#### 2. バックエンドの起動

```bash
cd backend
go run ./cmd/server
```

起動確認:

```bash
# 別ターミナルで
curl http://localhost:8080/api/v1/organizations
# {"data":[...]} が返ればOK
```

---

#### 3. フロントエンドの起動

```bash
cd frontend

# 初回のみ（依存パッケージのインストール）
npm install

# 開発サーバー起動
npm run dev
```

起動確認:  
ブラウザで http://localhost:3000 を開く

---

#### 4. 初期データの投入（初回のみ）

初回起動時はデータベースが空の状態です。以下を実行して初期データを投入してください。

```bash
# プロジェクトルートで実行
cat backend/seed.sql | docker exec -i pmt_db psql -U pmt_user -d pmt_db
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

#### 5. 起動順序のまとめ

```
[Terminal 1]  docker compose up -d db
[Terminal 2]  cd backend  →  go run ./cmd/server
[Terminal 3]  cd frontend  →  npm run dev
[初回のみ]   cat backend/seed.sql | docker exec -i pmt_db psql -U pmt_user -d pmt_db
```

---

#### 6. ローカル環境の停止手順

```bash
# バックエンド → 起動したターミナルで Ctrl+C

# フロントエンド → 起動したターミナルで Ctrl+C

# Docker（PostgreSQL）を停止
docker compose stop db
```

> **データを消さずに止める場合** は `stop`。  
> **コンテナごと削除する場合**（DB データも消える）は `docker compose down`。

#### ポートが埋まっていて起動できない場合

「Port 3000 is in use」や「Unable to acquire lock」が出た場合は、前のプロセスが残っています。

```bash
# 3000番ポートを使っているプロセスを確認
lsof -i :3000
# または
ss -tlnp | grep 3000

# そのPIDを強制終了（例: PID が 12345 の場合）
kill -9 12345

# ロックファイルを削除
rm -f frontend/.next/dev/lock
```

#### スクリプト実行で `$'\r': command not found` が出る場合

Windows からコピーしたファイルは CRLF 改行のため、bash でエラーになります。以下で修正してから再実行してください:

```bash
sed -i 's/\r$//' scripts/*.sh
bash scripts/setup-wsl.sh   # または bash scripts/start.sh
```

---

## アクセス先一覧

| サービス | URL |
|---|---|
| フロントエンド（一般） | http://localhost:3000 |
| フロントエンド（スーパー管理者） | http://localhost:3000/super-admin/login |
| バックエンド API | http://localhost:8080/api/v1 |
| PostgreSQL | localhost:5432 |

> **Note:** WSL2 は `localhost` を Windows に自動フォワードするため、Windows の Chrome 等からそのままアクセスできます。

---

## ドキュメント

設計資料は [.sdd/](.sdd/) フォルダを参照。[.sdd/README.md](.sdd/README.md) にナビゲーションあり。

| ドキュメント | 内容 |
|---|---|
| [.sdd/README.md](.sdd/README.md) | ドキュメント一覧・ナビゲーション |
| [architecture.md](.sdd/architecture.md) | システムアーキテクチャ |
| [layer-responsibility.md](.sdd/layer-responsibility.md) | Handler / Service / Repository の責務分担 |
| [domain-model.md](.sdd/domain-model.md) | エンティティ関係・ドメインモデル |
| [db-schema.md](.sdd/db-schema.md) | DB 設計・テーブル定義 |
| [api-spec.md](.sdd/api-spec.md) | REST API 仕様 |
| [key-flows.md](.sdd/key-flows.md) | 認証・マルチテナント・承認の主要フロー |
| [testing.md](.sdd/testing.md) | テスト方針・実行方法 |
| [dev-guide.md](.sdd/dev-guide.md) | 新機能追加の手順・規約 |

**AI アシスタント用**: [AGENTS.md](AGENTS.md) にプロジェクト概要と設計ドキュメントへのリンクを記載。

---

## テスト実行

```bash
cd backend
go test ./test/... -v
```

- テスト DB: インメモリ SQLite（PostgreSQL 不要）
- テスト件数: 150+
