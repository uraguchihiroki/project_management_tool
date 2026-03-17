# 設計ドキュメント (.sdd)

このフォルダには、プロジェクト管理ツールの設計・仕様ドキュメントが格納されています。

---

## ドキュメント一覧

| ドキュメント | 内容 |
|-------------|------|
| [architecture.md](architecture.md) | システムアーキテクチャ、技術スタック、ディレクトリ構成 |
| [layer-responsibility.md](layer-responsibility.md) | Handler / Service / Repository の責務分担と境界線 |
| [db-schema.md](db-schema.md) | データベース設計・テーブル定義 |
| [api-spec.md](api-spec.md) | REST API 仕様・エンドポイント一覧 |
| [key-flows.md](key-flows.md) | 認証・マルチテナント・承認の主要フロー |
| [testing.md](testing.md) | テスト方針・実行方法・カバー範囲 |
| [dev-guide.md](dev-guide.md) | 新機能追加の手順・規約・ドキュメント更新のタイミング |

---

## 初めて読む方へ

1. **全体像を把握する** → [architecture.md](architecture.md)
2. **コードの書き方・責務を理解する** → [layer-responsibility.md](layer-responsibility.md)
3. **データ構造を確認する** → [db-schema.md](db-schema.md)
4. **API の仕様を確認する** → [api-spec.md](api-spec.md)
5. **業務フローを理解する** → [key-flows.md](key-flows.md)
6. **開発・変更を行う** → [dev-guide.md](dev-guide.md)、[testing.md](testing.md)

---

## ドキュメント更新のルール

- **API を追加・変更したとき** → [api-spec.md](api-spec.md) を更新
- **テーブル・モデルを変更したとき** → [db-schema.md](db-schema.md) を更新
- **アーキテクチャ・構成が変わったとき** → [architecture.md](architecture.md) を更新
- **新規フローが追加されたとき** → [key-flows.md](key-flows.md) を更新
