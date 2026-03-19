# Project Management Tool — AI Assistant Context

Jira / Redmine ライクなチケットベースのプロジェクト管理ツール。マルチテナント対応（組織ごとにデータ分離）。

## 設計原則（最優先）

**開発はローカル、運用はGCPなどのクラウド。** プランニング・設計において常に最上位で考慮すること。詳細は [.sdd/principles.md](.sdd/principles.md) を参照。

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

## WSL 開発環境（エージェント向け）

本プロジェクトは **Windows で Cursor を起動し、WSL にリモート接続**して開発する構成を推奨する。

- **Cursor**: Windows で起動 → Remote - WSL で接続 → プロジェクトフォルダを開く
- **ターミナル**: WSL Ubuntu の bash を使用（PowerShell は使用しない）
- **パス**: WSL 形式（例: `/home/uraguchi/work/AI/project_management_tool`）を使用する

エージェントがターミナルコマンドを実行する場合、WSL を明示的に呼び出す例:

```bash
wsl -d Ubuntu-22.04 -e bash -c "cd /home/uraguchi/work/AI/project_management_tool && 実行したいコマンド"
```

Cursor が WSL に接続済みの場合は、統合ターミナルは自動的に bash になるため、そのままコマンドを実行してよい。

## 開発時の注意

1. **レイヤー責務を守る**: Handler は HTTP の入出力のみ。ビジネスロジックは Service、DB 操作は Repository。
2. **テストを実行**: `cd backend && go test ./test/... -v` でブラックボックステストを実行。
3. **API / DB 変更時**: `.sdd/api-spec.md` および `.sdd/db-schema.md` を更新する。
4. **アプリ起動**（WSL 環境）: `bash scripts/start.sh` で DB・バックエンド・フロントエンドを一括起動。または `docker compose up -d db` の後、`go run ./cmd/server` と `npm run dev` を別ターミナルで実行。
