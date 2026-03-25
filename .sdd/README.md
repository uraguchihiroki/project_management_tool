# 設計ドキュメント (.sdd)

このフォルダには、**Issue 管理ツール**の設計・仕様ドキュメントが格納されています。

---

## 最重要事項

→ [AGENTS.md](../AGENTS.md#最重要事項人ai-ともに必ず認識すること) を参照（正本）

---

## ドキュメント一覧

| ドキュメント | 内容 |
|-------------|------|
| [principles.md](principles.md) | 設計原則（開発はローカル、運用はクラウド）・Issue 管理の方針 |
| [dev-environment.md](dev-environment.md) | 開発環境の構成・リソース配置・使用ツール |
| [architecture.md](architecture.md) | システムアーキテクチャ、技術スタック、ディレクトリ構成 |
| [tenant-invariants.md](tenant-invariants.md) | テナント（組織）不変条件の正本（JWT・親子 API・スーパーアドミン・禁止事項） |
| [layer-responsibility.md](layer-responsibility.md) | Handler / Service / Repository の責務分担と境界線 |
| [domain-model.md](domain-model.md) | エンティティ関係・ドメインモデル |
| [transition-permissions.md](transition-permissions.md) | ステータス遷移・**遷移アラート**・監査（**§5–§7** と [db-schema](db-schema.md) / [api-spec](api-spec.md) の役割分担） |
| [db-schema.md](db-schema.md) | データベース設計・テーブル定義（**インプリント**・イベントログ） |
| [spec-implementation-divergences.md](spec-implementation-divergences.md) | **仕様と実装の乖離**（ユーザー合意済みのみ列挙。正本は各 `.sdd`） |
| [api-spec.md](api-spec.md) | REST API 仕様・エンドポイント一覧 |
| [key-flows.md](key-flows.md) | 認証・マルチテナント・ステータス遷移の権限の主要フロー |
| [visual-flow/](visual-flow/) | 画面設計・遷移（Visual Flow 方式） |
| [testing.md](testing.md) | テスト方針・実行方法・カバー範囲・**BB の AAA / 日本語コメント / ログ / アンチパターン** |
| [dev-guide.md](dev-guide.md) | 新機能追加の手順・規約・ドキュメント更新のタイミング |

---

## 初めて読む方へ

1. **設計原則を確認する** → [principles.md](principles.md)
2. **全体像を把握する** → [architecture.md](architecture.md)
3. **コードの書き方・責務を理解する** → [layer-responsibility.md](layer-responsibility.md)
4. **エンティティ関係を把握する** → [domain-model.md](domain-model.md)
5. **データ構造を確認する** → [db-schema.md](db-schema.md)
6. **API の仕様を確認する** → [api-spec.md](api-spec.md)
7. **ステータス遷移の権限の論点** → [transition-permissions.md](transition-permissions.md)
8. **業務フローを理解する** → [key-flows.md](key-flows.md)
9. **画面設計・遷移を確認する** → [visual-flow/transition-flow.md](visual-flow/transition-flow.md)、[visual-flow/conventions.md](visual-flow/conventions.md)
10. **開発・変更を行う** → [dev-guide.md](dev-guide.md)、[testing.md](testing.md)

---

## ドキュメント更新のルール

- **API を追加・変更したとき** → [api-spec.md](api-spec.md) を更新
- **エンティティ関係・ドメインが変わったとき** → [domain-model.md](domain-model.md) を更新
- **テーブル・モデルを変更したとき** → [db-schema.md](db-schema.md) を更新
- **アーキテクチャ・構成が変わったとき** → [architecture.md](architecture.md) を更新
- **新規フローが追加されたとき** → [key-flows.md](key-flows.md) を更新
- **ステータス遷移の権限の決定・変更** → [transition-permissions.md](transition-permissions.md)、必要に応じて [domain-model.md](domain-model.md) / [db-schema.md](db-schema.md) を更新
- **画面追加・変更時** → [visual-flow/](visual-flow/) を更新（[transition-flow.md](visual-flow/transition-flow.md)、該当 v_xxx.md）
- **テナント不変条件（誰が組織境界を張るか・親子の責務）を変えたとき** → [tenant-invariants.md](tenant-invariants.md) を更新し、必要なら [api-spec.md](api-spec.md) / [architecture.md](architecture.md) と整合させる
