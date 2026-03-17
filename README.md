# Project Management Tool

Jira / Redmine ライクなチケットベースのプロジェクト管理ツール。

## 技術スタック

- **フロントエンド**: Next.js 14 (React / TypeScript / Tailwind CSS)
- **バックエンド**: Go 1.22 + Echo v4
- **データベース**: PostgreSQL 16
- **コンテナ**: Docker Compose

## ローカル起動手順

### 前提条件

- Docker Desktop がインストールされていること
- Go 1.22+ がインストールされていること
- Node.js 20+ がインストールされていること

### 起動

```bash
# リポジトリのクローン
git clone git@github.com:uraguchihiroki/project_management_tool.git
cd project_management_tool

# Docker で PostgreSQL 起動
docker-compose up -d db

# バックエンド起動
cd backend
go run cmd/server/main.go

# フロントエンド起動（別ターミナル）
cd frontend
npm install
npm run dev
```

### アクセス

| サービス | URL |
|---|---|
| フロントエンド | http://localhost:3000 |
| バックエンド API | http://localhost:8080/api/v1 |
| PostgreSQL | localhost:5432 |

## ドキュメント

設計資料は [.sdd/](.sdd/) フォルダを参照。

| ドキュメント | 内容 |
|---|---|
| [architecture.md](.sdd/architecture.md) | システムアーキテクチャ |
| [db-schema.md](.sdd/db-schema.md) | DB設計・テーブル定義 |
| [api-spec.md](.sdd/api-spec.md) | REST API仕様 |
