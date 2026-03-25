# 仕様と実装の乖離（明示リスト）

`.sdd` を正としたとき、**実装がまだ追いついていない点**だけを列挙する。設計の正本は各ドキュメントに任せる。

ここに載っている項目は **ユーザーが内容を認識し、現時点で許容している乖離** とする。未合意のズレは書かない。新たな乖離を検知したら [AGENTS.md](../AGENTS.md) の **§7** に従い、ユーザーに確認してから追加する。

## 運用ルール（状態列）

表の **`状態`** 列は次のとおり。

| 値 | 意味 |
|----|------|
| `open` | 追跡中（乖離の解消・方針確定をまだ終えていない、または許容として監視中）。 |
| `done` | 解消済み、または許容として確定し **このリストでの追跡を終えた**（履歴として行は残してよい）。 |

- 新規に行を足すときは原則 **`open`**。
- **`done`** に更新した事実は、コミット・PR・チャットなどでいつ誰が変えたか分かればよい（本ファイルに日付列は設けない）。

## スコープ合意（一文）

論理削除を `.sdd/db-schema.md` に沿って全テーブル・全 DELETE 経路へ適用する **対象の切り方・復元の要否・API の意味** は、別途詰めたうえで実装する。本ファイルは **現状の乖離の明示** に用いる。

## 読み方（このリストで言う「乖離」）

- **「テーブルに `deleted_at` が無い」ことだけ**を指しているわけではない。業務モデルには GORM の `DeletedAt` を付け、通常の `Delete` はソフト削除に寄せてある。
- ここで問題にしているのは、`.sdd/db-schema.md` の **「削除は UPDATE 相当（論理削除）」という理想**に対して、実装上 **どうしても別の削除の仕方になる経路**が残っていること、および **起動時の移行処理**が業務 API とは別のルールで `DELETE` を使うこと。

## 合意済みの乖離

| 項目 | 内容 | 状態 |
|------|------|------|
| 論理削除まわりの残差 | **起動時マイグレーションのみ** … サーバ起動時に [`internal/db`](../backend/internal/db) が実行する処理（レガシー列の除去・重複 statuses のマージ・旧業務 UNIQUE の DROP・任意の非一意インデックス作成・結合テーブルの代理 PK 移行・`workflow_transitions` の重複行畳み込みなど）は、**業務 API ではなく DB 移行・整合用**である。ここでは意図的に **生 `DELETE`** や **`Unscoped().Delete`** が使われる（[`cmd/server/main.go`](../backend/cmd/server/main.go) の起動順で呼ばれる）。**業務 Repository**（結合テーブルの付け替え、`workflow_transitions` / `project_status_transitions` の許可遷移の更新含む）は原則 **ソフト削除＋新行 `Create`** であり、**`Unscoped` による物理削除**は業務経路では使わない。業務上の一意は DB の UNIQUE ではなく Service で保証する（[principles.md](principles.md)）。詳細は [db-schema.md](db-schema.md) の論理削除ノート参照。 | open |

## 棚卸しメモ（参考）

| 観点 | メモ | 状態 |
|------|------|------|
| マイグレーション | `status_integrity` / `migrate_issue_project_status` の物理 DELETE は **移行専用**（業務の論理削除と別物）。 | done |
| 結合テーブル | `AssignRolesToUser` / `ReplaceMembers` / `ReplaceForIssue` / 部署の `SetUserDepartments` は **論理削除＋新行 Create**（代理 PK により `Unscoped` 不要）。重複 ID は入力側で除去。 | done |
| organization_id 絶対化 | 結合テーブル `user_roles` / `user_groups` / `issue_groups` に `organization_id` を保持し、作成時に埋める。既存データは起動時マイグレーション `MigrateJunctionOrganizationID` で backfill。 | done |
| 許可遷移 | **デフォルト Issue ワークフロー**（未着手・進行・完了の 3 列）を作る経路（`CreateOrgIssueWorkflowWithDefaultStatuses` の呼び出し元・`IssueWorkflowProvisioner`・組織シードなど）では、`workflow_transitions` に **4 本**を自動投入（未着手↔進行・進行↔完了）。**`project_status_transitions`** および **上記以外の Issue ワークフロー**では作成時に自動投入しない（`POST .../transitions` 等）。 | done |
| Project と Issue WF | **`ProjectService.Create` は Issue ワークフローを作らない**。`default_workflow_id` は `IssueWorkflowProvisioner` / `POST /projects/:id/default-issue-workflow` / `IssueService.Create` 前段の lazy 確保で紐付ける。 | done |
