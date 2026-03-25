# 仕様と実装の乖離（明示リスト）

`.sdd` を正としたとき、**実装がまだ追いついていない点**だけを列挙する。設計の正本は各ドキュメントに任せる。

ここに載っている項目は **ユーザーが内容を認識し、現時点で許容している乖離** とする。未合意のズレは書かない。新たな乖離を検知したら [AGENTS.md](../AGENTS.md) の **§7** に従い、ユーザーに確認してから追加する。

## スコープ合意（一文）

論理削除を `.sdd/db-schema.md` に沿って全テーブル・全 DELETE 経路へ適用する **対象の切り方・復元の要否・API の意味** は、別途詰めたうえで実装する。本ファイルは **現状の乖離の明示** に用いる。

## 合意済みの乖離

- **論理削除が実装されていない** — `.sdd/db-schema.md` の「論理削除（本番環境）」（全テーブル `deleted_at`、削除は `UPDATE` 相当）に対し、実装はテーブル・経路ごとに GORM ソフト削除・物理削除 SQL・モデル未対応（例: `workflow_transitions` に `deleted_at` なし）が混在している。

## 棚卸しメモ（db-schema との差分・実装タスク化の材料）

調査時点のコード参照。詳細は実装時に再確認すること。

| 観点 | メモ |
|------|------|
| `workflow_transitions` | モデルに `deleted_at` がなく、`repository/workflow_transition.go` の `Delete` は行を消す。 |
| `statuses` / `workflows` 等 | GORM の `DeletedAt` ありのモデルは `Delete` がソフト削除になりうる一方、一覧・JOIN で `deleted_at IS NULL` が漏れうる。 |
| 整合・マイグレーション | `internal/db/status_integrity.go` 等で `Unscoped().Delete` や `DELETE FROM workflow_transitions` の生 SQL あり。 |
| その他エンティティ | `internal/repository/*.go` の `Delete` が多数。db-schema の「原則」と一対一で未照合。 |
