# データベース設計

本ドキュメントは **Issue 管理システム**を正とする設計へ寄せて更新する。稟議・承認ワークフロー向けのテーブルは **廃止方向**（実装に残る場合は [transition-permissions.md](transition-permissions.md) のレガシー扱い）。

## Key カラム（全テーブル共通）

全テーブルに `key` カラム（VARCHAR(255), NOT NULL）を設け、API/URL 用の識別子とする。

- 意味のある値がある場合: スラッグや識別子を格納（例: projects.key, statuses.status_key）
- 書き込む内容がない場合: PK の UID を格納（UUID の文字列、または `prefix-{id}` 形式）

---

## DB における一意性（PK のみ）

**スキーマ上の一意制約は主キーに限定する。** 組織名・メール・(organization_id, email)・(organization_id, key)・(name, organization_id) on roles・同一ワークフロー内の (name, display_order)（未削除行）・空でない `status_key` の重複禁止などは、**Service 層で保証する**。検索用には必要に応じて **非一意** のインデックスを張ってよい。

方針の根拠は [principles.md](principles.md)「データベースの一意制約（PK のみ）」。

---

## 論理削除（本番環境）

**本番では基本的にすべてのデータは論理削除とする。**

- 全テーブルで共通のカラム名を使用する: **`deleted_at`**
- このカラムに日時が入っていたら削除されたレコードとみなす
- `deleted_at` が NULL のレコードのみ有効（未削除）
- クエリ時は原則 `WHERE deleted_at IS NULL` を付与する

| カラム | 型 | 説明 |
|--------|-----|------|
| deleted_at | TIMESTAMP | NULL = 有効、日時が入っている = 削除済み |

> **Note:** 実装時は各テーブルに `deleted_at` を追加し、Repository 層で削除時は物理削除ではなく `UPDATE ... SET deleted_at = NOW()` とする。一覧取得・検索時は `deleted_at IS NULL` を条件に含める。

**実装（Go / GORM）:** モデルに `gorm.DeletedAt` を付与し、業務経路の `Delete` は原則ソフト削除とする。**起動時マイグレーション**（`internal/db` の重複除去・レガシー除去など）では意図的に生 `DELETE` や `Unscoped().Delete` が残る。多対多の**結合テーブル**で「全メンバー差し替え」等を行う箇所は、**複合主キー (user_id, role_id) 等での再登録**のためソフト削除済み行が残ると衝突するので `Unscoped` による物理削除を使う（DB に業務用 UNIQUE を張らない方針でも、**主キー**は変わらず衝突しうる）。

---

## インプリント（Imprint）

本システムにおける **インプリント**とは、Issue 等に対して **起きた事実を 1 件として追記する** 不変レコード（**`issue_events` の 1 行**）を指す。

| 観点 | 定義 |
|------|------|
| **メタファー** | 「印を押した」**跡**（押し跡）。文字が流れる **ログ** より、**事実が残った印** に近い。 |
| **実体** | 追記専用。業務行の `updated_at` だけでは表せない **誰が・いつ・何が・当時の担当** を保持する。 |
| **クエリ** | **テナント・Issue・種別（`event_type`）・期間（`occurred_at`）・操作者（`actor_id`）** などで **後から絞り込める** 列を持つ（実質は監査・集計用の **クエリキー**の束）。 |
| **時刻** | **主語を「タイムスタンプ」にしない**。発生時刻はインプリントの属性として **`occurred_at`（`TIMESTAMPTZ`）** を持つ。 |
| **コメントのスタンプ** | Issue コメントに付ける **リアクション（見ました・いいね等）** とは **別機能・別名**。**インプリント**はシステムが記録する **事実のインプリント** と区別する。 |

### 実装方針（インプリント）

| 方針 | 内容 |
|------|------|
| **時刻（属性）** | 各インプリントに **`occurred_at`**。PostgreSQL では **`TIMESTAMPTZ`**（タイムゾーン付き）を推奨。UTC 格納・表示時変換。 |
| **エンティティの時刻** | 通常レコードの作成・更新は `created_at` / `updated_at`（**インプリント**とは別物。`updated_at` だけでは監査に代えない）。 |
| **インデックス** | よく使う絞り込み軸に **`(organization_id, occurred_at)`**、**`(issue_id, occurred_at)`** など複合インデックスを検討する。 |
| **deleted_at** | インプリントは追記のみを業務ルールとするが、**テナント／契約終了時**に全データを一括論理削除し、システムに異常がないか確認してから物理削除する運用に備え、行に `deleted_at` を持てるようにする（通常運用では削除 API を使わない想定）。 |

---

## イベントログ（インプリントの列）

ステータス変更・担当変更・開示範囲変更など、**1 操作ごとにインプリントを 1 行追記**する。**事後の監査・レポート・「作業者と変更者が同一だったか」**の検査に使う。**追記のみ（append-only）**を原則とし、業務データの `UPDATE` だけでは代用しない。

### 設計の要点

| 要点 | 内容 |
|------|------|
| **列の意味が固定** | `event_type`（例: `issue.status_changed`）は **アプリ定数または DB 制約**で列挙し、自由文字だけにしない（集計・JOIN が楽）。 |
| **スナップショット** | 監査に必要な値は **イベント時点のコピー**を持つ（例: `assignee_id_at_occurred`）。後から Issue を更新されても当時が分かる。 |
| **外部キー** | `organization_id` / `issue_id` / `actor_id` を持ち、**テナント・Issue・誰が**で絞り込みやすくする。 |
| **拡張** | 追加属性は **`payload`（JSONB）** に逃がす余地を残す（インデックスは GIN または生成列で必要分のみ）。 |

### 例（概念スキーマ）

```
issue_events（名称は実装で確定）
├── id (PK)
├── key (VARCHAR(255), NOT NULL)
├── organization_id (FK → organizations.id, NOT NULL)
├── issue_id (FK → issues.id, NOT NULL)
├── actor_id (FK → users.id, NOT NULL)     # 操作したユーザー
├── event_type (VARCHAR(80), NOT NULL)    # 例: issue.status_changed
├── occurred_at (TIMESTAMPTZ, NOT NULL)   # インプリントの発生時刻（期間クエリの軸）
├── from_status_id (nullable)
├── to_status_id (nullable)
├── assignee_id_at_occurred (nullable)    # 発生時点の担当スナップショット
└── payload (JSONB, nullable)             # 補足（任意）
```

**推奨インデックス例**: `(organization_id, occurred_at)`、`(issue_id, occurred_at)`、`(event_type, occurred_at)`（監査パターンに応じて選ぶ）。

詳細なルール（どの操作で行を残すか）は [transition-permissions.md](transition-permissions.md) と [key-flows.md](key-flows.md) に合わせて確定する。

---

## ER図

```
organizations（グローバル）
├── id (PK)
├── key (VARCHAR(255), NOT NULL)
├── name（アプリで一意）
├── admin_email
└── created_at

super_admins（グローバル）
├── id (PK)
├── key (VARCHAR(255), NOT NULL)
├── name
├── email（アプリで一意）
└── created_at

users（1ユーザー＝1組織）
├── id (PK)
├── key (VARCHAR(255), NOT NULL)
├── organization_id (FK → organizations.id, NOT NULL)
├── name
├── email（組織内はアプリで一意: (organization_id, email)）
├── avatar_url (nullable)
├── is_admin
├── is_org_admin
├── joined_at
└── created_at

roles
├── id (PK, auto)
├── key (VARCHAR(255), NOT NULL)
├── name（同一 organization_id 内はアプリで一意）
├── level
├── description
├── organization_id (FK → organizations.id, nullable)
└── created_at

user_roles (中間テーブル, many2many)
├── user_id (FK → users.id)
├── role_id (FK → roles.id)
└── key (VARCHAR(255), NOT NULL)

projects
├── id (PK)
├── key（組織内はアプリで一意: (organization_id, key)）
├── name
├── description (nullable)
├── owner_id (FK → users.id)
├── organization_id (FK → organizations.id, NOT NULL)
├── default_workflow_id (FK → workflows.id, nullable) — Issue 用ワークフロー
├── project_status_id (FK → project_statuses.id, nullable) — 現在のプロジェクト進行
└── created_at

statuses（Issue 専用。常に workflow_id 必須。Workflow / workflow_transitions と組み合わせて Issue のカンバン列と遷移を表す）

├── id (PK)
├── key (VARCHAR(255), NOT NULL)
├── workflow_id (FK → workflows.id, NOT NULL)
├── name
├── color (HEX)
├── order
├── status_key (nullable) — 例: sts_start, sts_goal
└── deleted_at

project_statuses（プロジェクト進行。Workflow は使用しない）

├── id (PK)
├── key (VARCHAR(255), NOT NULL)
├── project_id (FK → projects.id, NOT NULL)
├── name, color, order
├── status_key (nullable)
└── deleted_at

project_status_transitions（同一プロジェクト内の許可遷移 from → to）

├── id (PK, auto)
├── project_id (FK → projects.id, NOT NULL)
├── from_project_status_id, to_project_status_id
└── created_at

issues
├── id (PK)
├── key (VARCHAR(255), NOT NULL)
├── number (プロジェクト内連番)
├── title
├── description (nullable)
├── status_id (FK → statuses.id)
├── priority
├── assignee_id (FK → users.id, nullable)
├── reporter_id (FK → users.id)
├── project_id (FK → projects.id)
├── due_date (nullable)
├── template_id (FK → issue_templates.id, nullable)
├── created_at
└── updated_at

comments
├── id (PK)
├── key (VARCHAR(255), NOT NULL)
├── organization_id (FK → organizations.id, NOT NULL)
├── issue_id (FK → issues.id)
├── author_id (FK → users.id)
├── body
├── created_at
└── updated_at

groups（組織スコープ。部署コピー・タグ・通知用など用途を `kind` 等で区別）
├── id (PK)
├── key (VARCHAR(255), NOT NULL)
├── organization_id (FK → organizations.id, NOT NULL)
├── name
├── kind (nullable)                     # 例: team / tag / notification / sync_from_hr 等・実装で確定
└── created_at

user_groups（ユーザー ↔ Group 多対多）
├── user_id (FK → users.id)
├── group_id (FK → groups.id)
└── key (VARCHAR(255), NOT NULL)

issue_groups（Issue ↔ Group 多対多・開示・共同文脈）
├── issue_id (FK → issues.id)
├── group_id (FK → groups.id)
├── role (nullable)                     # 例: primary / collaborator / tag・実装で確定
└── key (VARCHAR(255), NOT NULL)

issue_events（インプリントの列・追記のみ。上記「イベントログ」節と同一概念）
├── （上記セクションの列定義に準拠）

issue_templates
├── id (PK, auto)
├── key (VARCHAR(255), NOT NULL)
├── organization_id (FK → organizations.id, NOT NULL)
├── project_id (FK → projects.id)
├── name
├── description
├── body
├── default_priority
└── created_at
```

> **レガシー（移行予定）**: 実装 DB に `workflows` / `workflow_steps` / `approval_objects` / `issue_approvals` および `issues.workflow_id` / `issue_templates.workflow_id` が残っている場合がある。Issue 管理を正とする設計では **これらは廃止方向**。[transition-permissions.md](transition-permissions.md) で合意したあと、スキーマから除去する。

---

## ステータス遷移・Group（仕様の柱の一部）

**許可遷移・遷移アラート・監査**の意味論は [transition-permissions.md](transition-permissions.md)（**7**）。本ドキュメントでは **テーブル**として `groups` / `issue_groups` / `user_groups` / `issue_events` を置く（上記 ER・各節）。

- **役職（roles）** は稟議・ディレクトリ型のマスタとして必須ではない。**Issue 文脈の Group** を主とし、`roles` / `user_roles` は既存互換・補助として扱う（廃止は別議論）。
- **Position** 専用テーブルは必須としない。表示順が必要なら `groups.display_order` 等で足りる想定（詳細は transition-permissions）。

---

## テーブル定義

### organizations

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | 組織ID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子 |
| name | VARCHAR(200) | NOT NULL | 組織名（**アプリ**で一意） |
| admin_email | VARCHAR(255) | nullable | 組織管理者のメールアドレス |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

### super_admins

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | スーパー管理者ID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子 |
| name | VARCHAR(100) | NOT NULL | 表示名 |
| email | VARCHAR(255) | NOT NULL | メールアドレス（**アプリ**で一意） |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

### users

1 ユーザー＝1 組織。同一メールでも組織が違えば別レコード。

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | ユーザーID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子 |
| organization_id | UUID | FK, NOT NULL | 所属組織 |
| name | VARCHAR(100) | NOT NULL | 表示名 |
| email | VARCHAR(255) | NOT NULL | メールアドレス（組織内でユニーク） |
| avatar_url | TEXT | nullable | アバター画像URL |
| is_admin | BOOLEAN | DEFAULT false | システム管理者フラグ |
| is_org_admin | BOOLEAN | DEFAULT false | 組織管理者フラグ |
| joined_at | TIMESTAMP | NOT NULL | 参加日時 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

> **Note:** (organization_id, email) の一意は **Service** で保証。DB には非一意インデックスのみ（任意）。

### roles

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | SERIAL | PK | 役職ID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子 |
| name | VARCHAR(100) | NOT NULL | 役職名 |
| level | INTEGER | NOT NULL, DEFAULT 1 | ヒエラルキー用の数値（遷移アラートの想定アクター表現等に再利用するかは [transition-permissions.md](transition-permissions.md) で決定） |
| description | VARCHAR(500) | | 説明 |
| organization_id | UUID | FK, nullable | 所属組織（NULL はグローバル） |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

> **Note:** (name, organization_id) の一意は **Service** で保証。

### projects

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | プロジェクトID |
| key | VARCHAR(10) | NOT NULL | 識別キー（組織内は**アプリ**で一意）。API/URL 用にも使用 |
| name | VARCHAR(200) | NOT NULL | プロジェクト名 |
| description | TEXT | nullable | 説明 |
| owner_id | UUID | FK | オーナーユーザー |
| organization_id | UUID | FK, NOT NULL | 所属組織 |
| default_workflow_id | BIGINT | FK → workflows.id, nullable | Issue 用デフォルトワークフロー |
| project_status_id | UUID | FK → project_statuses.id, nullable | 現在のプロジェクト進行 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

> **Note:** (organization_id, key) の一意は **Service** で保証。

### statuses（Issue 専用）

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | ステータスID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子 |
| workflow_id | BIGINT | FK → workflows.id, NOT NULL | 所属ワークフロー |
| name | VARCHAR(50) | NOT NULL | ステータス名 |
| color | VARCHAR(7) | NOT NULL | HEXカラー (#RRGGBB) |
| display_order | INTEGER | NOT NULL | 表示順 |
| status_key | VARCHAR(50) | nullable | システム用: sts_start, sts_goal。NULL=ユーザー定義 |

> **重複防止**: 同一 `workflow_id` で、論理削除されていない行について `(name, display_order)` の一意は **Service** で保証する。起動時 `MigrateStatusDedupe`（`internal/db/status_integrity.go`）でレガシー重複行を畳む。旧 `type` 列は `MigrateIssueProjectStatusSplitPre`（`internal/db/migrate_issue_project_status.go`）で除去する。レガシー列名 `order` は `MigrateStatusOrderToDisplayOrder` で `display_order` に寄せる。必要に応じ **非一意** の複合インデックスで検索を補助してよい。
>
> **件数下限（Issue 用）**: 同一 `workflow_id` に紐づく有効な `statuses` は **少なくとも 2 行**を保つ（削除で 1 行以下になる操作は API で拒否）。運用上の意味・用語の正本は [domain-model.md](domain-model.md)。

### project_statuses

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | |
| key | VARCHAR(255) | NOT NULL | |
| project_id | UUID | FK → projects.id, NOT NULL | |
| name | VARCHAR(50) | NOT NULL | |
| color | VARCHAR(7) | NOT NULL | |
| order | INTEGER | NOT NULL | |
| status_key | VARCHAR(50) | nullable | |
| deleted_at | TIMESTAMP | nullable | 論理削除 |

### project_status_transitions

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | SERIAL | PK | |
| key | VARCHAR(255) | NOT NULL | |
| project_id | UUID | FK → projects.id, NOT NULL | |
| from_project_status_id | UUID | NOT NULL | |
| to_project_status_id | UUID | NOT NULL | |
| created_at | TIMESTAMP | NOT NULL | |

### issues

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | IssueID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子（{project_key}-{number} or id） |
| number | INTEGER | NOT NULL | プロジェクト内連番 |
| title | VARCHAR(500) | NOT NULL | タイトル |
| description | TEXT | nullable | 詳細説明 |
| status_id | UUID | FK | ステータス |
| priority | VARCHAR(20) | NOT NULL, DEFAULT 'medium' | 優先度 |
| assignee_id | UUID | FK, nullable | 担当者 |
| reporter_id | UUID | FK | 起票者 |
| project_id | UUID | FK | 所属プロジェクト |
| due_date | DATE | nullable | 期日 |
| template_id | INTEGER | FK, nullable | テンプレート |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

### comments

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | コメントID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子（id を格納） |
| organization_id | UUID | FK, NOT NULL | 所属組織 |
| issue_id | UUID | FK | 対象Issue |
| author_id | UUID | FK | 投稿者 |
| body | TEXT | NOT NULL | 本文 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

### issue_templates

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | SERIAL | PK | テンプレートID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子 |
| organization_id | UUID | FK, NOT NULL | 所属組織 |
| project_id | UUID | FK | 所属プロジェクト |
| name | VARCHAR(200) | NOT NULL | テンプレート名 |
| description | VARCHAR(500) | | 説明 |
| body | TEXT | | 本文テンプレート |
| default_priority | VARCHAR(20) | NOT NULL, DEFAULT 'medium' | デフォルト優先度 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

### groups

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | グループID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子 |
| organization_id | UUID | FK, NOT NULL | 所属組織 |
| name | VARCHAR(200) | NOT NULL | 表示名 |
| kind | VARCHAR(50) | nullable | 用途区分（例: team / tag / notification）。集計・フィルタ用 |
| display_order | INTEGER | nullable | 一覧の並び（任意） |
| created_at | TIMESTAMPTZ | NOT NULL | 作成日時 |

### user_groups

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| user_id | UUID | FK, NOT NULL | ユーザー |
| group_id | UUID | FK, NOT NULL | グループ |
| key | VARCHAR(255) | NOT NULL | API/URL 用 |

> **Note:** 複合 **PK** (user_id, group_id)。1 ユーザーが複数グループに入るときは行が複数（兼務）。

### issue_groups

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| issue_id | UUID | FK, NOT NULL | Issue |
| group_id | UUID | FK, NOT NULL | グループ |
| role | VARCHAR(50) | nullable | 例: primary / collaborator / tag |
| key | VARCHAR(255) | NOT NULL | API/URL 用 |

> **Note:** 複合 **PK** (issue_id, group_id)。

### issue_events

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | イベントID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子 |
| organization_id | UUID | FK, NOT NULL | テナント絞り込み用 |
| issue_id | UUID | FK, NOT NULL | 対象 Issue |
| actor_id | UUID | FK, NOT NULL | 操作ユーザー |
| event_type | VARCHAR(80) | NOT NULL | 列挙値（例: issue.status_changed） |
| occurred_at | TIMESTAMPTZ | NOT NULL | 発生時刻（**クエリの主軸**） |
| from_status_id | UUID | FK, nullable | 遷移前 |
| to_status_id | UUID | FK, nullable | 遷移後 |
| assignee_id_at_occurred | UUID | FK, nullable | 当時の担当スナップショット |
| payload | JSONB | nullable | 拡張属性 |

---

## デフォルトステータス

新規プロジェクト作成時に以下のステータスを自動生成：

| order | name | color |
|-------|------|-------|
| 1 | 未着手 | #6B7280 |
| 2 | 進行中 | #3B82F6 |
| 3 | レビュー中 | #F59E0B |
| 4 | 完了 | #10B981 |
