# Project Management Tool — AI Assistant Context

Jira / Redmine ライクなチケットベースのプロジェクト管理ツール。マルチテナント対応（組織ごとにデータ分離）。

## 技術スタック

| 役割 | 技術 |
|------|------|
| フロントエンド | Next.js 14 (React / TypeScript / Tailwind CSS) |
| バックエンド | Go 1.22 + Echo v4 + GORM |
| データベース | PostgreSQL 16 |
| 認証 | メールアドレスのみ（JWT 未実装） |

## ディレクトリ構成

```
project_management_tool/
├── .sdd/              # 設計ドキュメント（必読）
├── backend/            # Go API サーバー
│   ├── cmd/server/     # エントリポイント
│   ├── internal/      # handler, service, repository, model
│   ├── test/          # ブラックボックステスト
│   └── seed.sql       # 初期データ投入スクリプト
├── frontend/           # Next.js
│   └── src/app/       # App Router ページ
└── docker-compose.yml  # PostgreSQL
```

## 設計ドキュメント

**必ず [.sdd/README.md](.sdd/README.md) を参照すること。** ナビゲーションと各ドキュメントの役割が記載されている。

- [architecture.md](.sdd/architecture.md) — システム構成・マルチテナント
- [layer-responsibility.md](.sdd/layer-responsibility.md) — Handler / Service / Repository の責務
- [db-schema.md](.sdd/db-schema.md) — テーブル定義
- [api-spec.md](.sdd/api-spec.md) — REST API 仕様
- [key-flows.md](.sdd/key-flows.md) — 認証・組織・承認フロー
- [dev-guide.md](.sdd/dev-guide.md) — 新機能追加の手順

## 開発時の注意

1. **レイヤー責務を守る**: Handler は HTTP の入出力のみ。ビジネスロジックは Service、DB 操作は Repository。
2. **テストを実行**: `cd backend && go test ./test/... -v` でブラックボックステストを実行。
3. **API / DB 変更時**: `.sdd/api-spec.md` および `.sdd/db-schema.md` を更新する。
4. **バックエンド起動**: `go run` は使わない。毎回別の一時パスにビルドされるため、Windows ファイアウォールが毎回ブロックする。必ず `go build -o server.exe ./cmd/server` してから `.\server.exe` を実行する。
