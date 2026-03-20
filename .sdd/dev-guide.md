# 開発ガイド

## 新機能追加の手順

### 1. モデル（必要に応じて）

- [backend/internal/model/model.go](backend/internal/model/model.go) に構造体を追加
- `main.go` の `db.AutoMigrate` に追加
- [db-schema.md](db-schema.md) を更新

### 2. Repository

- [backend/internal/repository/](backend/internal/repository/) に新規ファイルを作成
- CRUD や検索メソッドを実装
- ビジネスロジックは含めない（DB 操作のみ）

### 3. Service

- [backend/internal/service/](backend/internal/service/) に新規ファイルを作成
- ビジネスロジック、デフォルト値、採番、複数 Repository の調整を実装
- HTTP に依存しない形で書く（`echo.Context` を参照しない）

### 4. Handler

- [backend/internal/handler/](backend/internal/handler/) に新規ファイルを作成
- リクエストのパース、Service の呼び出し、レスポンスの返却のみ
- `main.go` にルートを追加
- [api-spec.md](api-spec.md) を更新

### 5. テスト

- [backend/test/](backend/test/) にテストを追加
- [setup_test.go](backend/test/setup_test.go) の `newTestServer` にルートを追加
- [testing.md](testing.md) を参照

---

## 命名規則・ディレクトリ配置

| 種別 | 規則 | 例 |
|------|------|-----|
| Handler | `XxxHandler`, `xxx.go` | `handler/issue.go` → `IssueHandler` |
| Service | `XxxService`, `xxx.go` | `service/issue.go` → `IssueService` |
| Repository | `XxxRepository`, `xxx.go` | `repository/issue.go` → `IssueRepository` |
| モデル | `model/` に集約 | `model/model.go` または `model/xxx.go` |

---

## seed.sql の扱い

- **実行**: `Get-Content backend/seed.sql | docker exec -i pmt_db psql -U pmt_user -d pmt_db`
- **べき等性**: `ON CONFLICT DO NOTHING` を使用しているため、複数回実行しても安全
- **変更時**: 既存環境への影響を考慮。`ALTER TABLE` や `UPDATE` は慎重に。新規カラムは `ADD COLUMN IF NOT EXISTS` を推奨

---

## 組織seed（CLI）

既存組織にステータス・役職・部署・サンプルプロジェクト等を投入する:

```bash
# 全組織に投入
./scripts/seed-org.sh --all

# 指定組織に投入
./scripts/seed-org.sh <org-id> [owner-id]
```

または:

```bash
cd backend
go run ./cmd/cli org seed --all
go run ./cmd/cli org seed --org-id=<uuid> [--owner-id=<uuid>]
```

- **--all**: 全組織にseed投入。各組織の管理者をオーナーとしてサンプルプロジェクトを作成
- **org-id**: 組織のUUID
- **owner-id**: 省略時は組織の管理者を使用。管理者がいない組織はサンプルプロジェクト・Issueは作成しない
- **冪等**: 既存レコードは更新、なければ作成。何度でも実行可能

---

## ドキュメント更新のタイミング

| 変更内容 | 更新するドキュメント |
|----------|----------------------|
| API の追加・変更 | [api-spec.md](api-spec.md) |
| テーブル・モデルの変更 | [db-schema.md](db-schema.md) |
| アーキテクチャ・構成の変更 | [architecture.md](architecture.md) |
| 新規フローの追加 | [key-flows.md](key-flows.md) |
| テスト方針の変更 | [testing.md](testing.md) |

---

## レイヤー責務の確認

詳細は [layer-responsibility.md](layer-responsibility.md) を参照。

- **Handler**: HTTP の入出力のみ。ビジネスロジックを持たない
- **Service**: ビジネスロジック。HTTP を知らない
- **Repository**: DB 操作のみ。ビジネスロジックを持たない
