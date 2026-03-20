# システムアーキテクチャ

## 概要

> **設計原則**: 開発はローカル、運用はGCP等のクラウド。詳細は [principles.md](principles.md) を参照。

Jira / Redmine ライクなチケットベースのプロジェクト管理ツール。  
Go 製 REST API + Next.js フロントエンド + PostgreSQL の3層構成。  
マルチテナント対応（組織ごとにデータを分離）。

---

## システム構成図

```mermaid
graph TD
    Browser["Browser / Next.js (React/TS)<br/>localhost:3000"]

    subgraph Backend["Go API Server (Echo) — localhost:8080"]
        Handler["Handler<br/>(ルーティング・リクエスト処理)"]
        Service["Service<br/>(ビジネスロジック)"]
        Repository["Repository<br/>(DB操作)"]
        Handler --> Service
        Service --> Repository
    end

    DB[("PostgreSQL<br/>localhost:5432")]

    Browser -->|"REST API (JSON)"| Handler
    Repository -->|"SQL (GORM)"| DB
```

---

## マルチテナント構成

| 種別 | 説明 |
|------|------|
| **SuperAdmin** | システム全体の管理者。組織の作成のみ可能。メールアドレスのみでログイン。 |
| **Organization** | 会社・組織。プロジェクト・役職・ユーザーは組織に紐づく。 |
| **User** | 1ユーザー＝1組織。`organization_id` で所属組織、`is_org_admin` で組織管理者を識別。 |
| **組織管理者** | 所属組織内のユーザー作成・更新・削除、管理画面へのアクセスが可能。 |

---

## 技術スタック

| レイヤー | 技術 | バージョン |
|---------|------|------------|
| フロントエンド | Next.js (App Router) | 14.x |
| UI | Tailwind CSS | 最新 |
| バックエンド | Go + Echo | Go 1.22 / Echo v4 |
| ORM | GORM | v2 |
| データベース | PostgreSQL | 16 |
| コンテナ | Docker Compose | - |
| 認証 | メールアドレスのみ | JWT 未実装 |

---

## ディレクトリ構成

```
project_management_tool/
├── .sdd/                        # 設計ドキュメント
│   ├── README.md                # ナビゲーション
│   ├── architecture.md         # このファイル
│   ├── layer-responsibility.md # レイヤー責務定義
│   ├── db-schema.md            # DB設計
│   ├── api-spec.md             # API仕様
│   ├── key-flows.md            # 主要フロー
│   ├── testing.md              # テスト方針
│   └── dev-guide.md            # 開発ガイド
├── backend/                     # Go APIサーバー
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── handler/             # HTTPハンドラー
│   │   ├── service/             # ビジネスロジック
│   │   ├── repository/         # DB操作
│   │   ├── model/               # データモデル
│   │   └── middleware/         # ミドルウェア
│   ├── test/                    # ブラックボックステスト（インメモリ SQLite）
│   ├── seed.sql                 # 初期データ投入（手動実行）
│   ├── go.mod
│   └── Dockerfile
├── frontend/                    # Next.js
│   ├── src/
│   │   ├── app/                 # App Router
│   │   ├── components/          # UIコンポーネント
│   │   ├── lib/                 # APIクライアント等
│   │   └── types/               # TypeScript型定義
│   ├── package.json
│   └── Dockerfile
├── docker-compose.yml
├── README.md
└── AGENTS.md                    # Cursor 用コンテキスト
```

> **Note:** DB マイグレーションは GORM の AutoMigrate を使用。`migrations/` フォルダは使用していない。

---

## 将来の拡張方針（GCP / AWS 対応）

| 項目 | ローカル | クラウド |
|------|----------|----------|
| DB | Docker PostgreSQL | Cloud SQL / RDS |
| バックエンド | ローカル実行 | Cloud Run / ECS |
| フロントエンド | ローカル実行 | Cloud Run / Amplify |
| 認証 | メールのみ | Firebase Auth / Cognito |
| ストレージ | ローカル | GCS / S3 |
