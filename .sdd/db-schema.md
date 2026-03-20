# データベース設計

## Key カラム（全テーブル共通）

全テーブルに `key` カラム（VARCHAR(255), NOT NULL）を設け、API/URL 用の識別子とする。

- 意味のある値がある場合: スラッグや識別子を格納（例: projects.key, statuses.status_key）
- 書き込む内容がない場合: PK の UID を格納（UUID の文字列、または `prefix-{id}` 形式）

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

---

## ER図

```
organizations（グローバル）
├── id (PK)
├── key (VARCHAR(255), NOT NULL)
├── name (UNIQUE)
├── admin_email
└── created_at

super_admins（グローバル）
├── id (PK)
├── key (VARCHAR(255), NOT NULL)
├── name
├── email (UNIQUE)
└── created_at

users（1ユーザー＝1組織）
├── id (PK)
├── key (VARCHAR(255), NOT NULL)
├── organization_id (FK → organizations.id, NOT NULL)
├── name
├── email（組織内UNIQUE: (organization_id, email)）
├── avatar_url (nullable)
├── is_admin
├── is_org_admin
├── joined_at
└── created_at

roles
├── id (PK, auto)
├── key (VARCHAR(255), NOT NULL)
├── name
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
├── key（組織内UNIQUE: (organization_id, key)）
├── name
├── description (nullable)
├── owner_id (FK → users.id)
├── organization_id (FK → organizations.id, NOT NULL)
└── created_at

statuses
├── id (PK)
├── key (VARCHAR(255), NOT NULL)
├── project_id (FK → projects.id)
├── name
├── color (HEX)
└── order

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
├── workflow_id (FK → workflows.id, nullable)
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

workflows
├── id (PK, auto)
├── key (VARCHAR(255), NOT NULL)
├── organization_id (FK → organizations.id, NOT NULL)
├── name
├── description
└── created_at

workflow_steps（ステータス参照の双方向リスト）
├── id (PK, auto)
├── key (VARCHAR(255), NOT NULL)
├── organization_id (FK → organizations.id, NOT NULL)
├── workflow_id (FK → workflows.id)
├── status_id (FK → statuses.id, NOT NULL)   # このステップのステータス。表示名は status.name を使用
├── next_status_id (FK → statuses.id, nullable) # 承認後ステータス。ゴールでは NULL
├── description (nullable)           # ステップの説明
├── threshold (default 10)           # 閾値（点数合計>=で遷移）。ゴールでは無効
approval_objects (承認オブジェクト, 1ステップ:N。goal ステップには紐づかない)
├── id (PK, auto)
├── key (VARCHAR(255), NOT NULL)
├── organization_id (FK → organizations.id, NOT NULL)
├── workflow_step_id (FK → workflow_steps.id)
├── order
├── type (role / user)
├── role_id (FK → roles.id, nullable)      # type=role のとき
├── role_operator (eq / gte, nullable)     # イコール / 以上
├── user_id (FK → users.id, nullable)      # type=user のとき
├── points (default 1)                      # 承認時に加算する点数。同一人物は1回のみ、複数該当時は最高点で加算
├── exclude_reporter (default false)        # 起票者をこの承認オブジェクトの承認者から除外
└── exclude_assignee (default false)        # 担当者をこの承認オブジェクトの承認者から除外

issue_templates
├── id (PK, auto)
├── key (VARCHAR(255), NOT NULL)
├── organization_id (FK → organizations.id, NOT NULL)
├── project_id (FK → projects.id)
├── name
├── description
├── body
├── default_priority
├── workflow_id (FK → workflows.id, nullable)
└── created_at

issue_approvals
├── id (PK)
├── key (VARCHAR(255), NOT NULL)
├── organization_id (FK → organizations.id, NOT NULL)
├── issue_id (FK → issues.id)
├── workflow_step_id (FK → workflow_steps.id)
├── approver_id (FK → users.id, nullable)
├── status (pending / approved / rejected)
├── comment
├── acted_at (nullable)
└── created_at
```

---

## テーブル定義

### organizations

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | 組織ID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子 |
| name | VARCHAR(200) | UNIQUE, NOT NULL | 組織名 |
| admin_email | VARCHAR(255) | nullable | 組織管理者のメールアドレス |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

### super_admins

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | スーパー管理者ID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子 |
| name | VARCHAR(100) | NOT NULL | 表示名 |
| email | VARCHAR(255) | UNIQUE, NOT NULL | メールアドレス |
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

> **Note:** (organization_id, email) でユニークインデックス。

### roles

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | SERIAL | PK | 役職ID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子 |
| name | VARCHAR(100) | NOT NULL | 役職名 |
| level | INTEGER | NOT NULL, DEFAULT 1 | 承認に必要なレベル |
| description | VARCHAR(500) | | 説明 |
| organization_id | UUID | FK, nullable | 所属組織（NULL はグローバル） |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

> **Note:** (name, organization_id) でユニークインデックス。

### projects

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | プロジェクトID |
| key | VARCHAR(10) | NOT NULL | 識別キー（組織内でユニーク）。API/URL 用にも使用 |
| name | VARCHAR(200) | NOT NULL | プロジェクト名 |
| description | TEXT | nullable | 説明 |
| owner_id | UUID | FK | オーナーユーザー |
| organization_id | UUID | FK, NOT NULL | 所属組織 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

> **Note:** (organization_id, key) でユニークインデックス。

### statuses

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | ステータスID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子（status_key があれば流用、なければ id） |
| project_id | UUID | FK, nullable | 所属プロジェクト（組織用は NULL） |
| organization_id | UUID | FK, nullable | 所属組織 |
| name | VARCHAR(50) | NOT NULL | ステータス名 |
| color | VARCHAR(7) | NOT NULL | HEXカラー (#RRGGBB) |
| order | INTEGER | NOT NULL | 表示順 |
| type | VARCHAR(20) | NOT NULL | issue / project |
| status_key | VARCHAR(50) | nullable, UNIQUE | システム用: sts_start, sts_goal。NULL=ユーザー定義 |

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
| workflow_id | INTEGER | FK, nullable | ワークフロー |
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

### workflows

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | SERIAL | PK | ワークフローID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子 |
| organization_id | UUID | FK, NOT NULL | 所属組織 |
| name | VARCHAR(200) | NOT NULL | ワークフロー名 |
| description | VARCHAR(500) | | 説明 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

### workflow_steps（ステータス参照の双方向リスト）

表示順は意味を持たない。status_id → next_status_id のリンクで辿る。

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | SERIAL | PK | ステップID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子 |
| organization_id | UUID | FK, NOT NULL | 所属組織 |
| workflow_id | INTEGER | FK | 所属ワークフロー |
| status_id | UUID | FK, NOT NULL | このステップのステータス。表示名は status.name |
| next_status_id | UUID | FK, nullable | 承認後ステータス。ゴールでは NULL |
| description | TEXT | nullable | ステップの説明 |
| threshold | INTEGER | NOT NULL, DEFAULT 10 | 閾値（点数合計>=で遷移）。ゴールでは無効 |

**システムステータス（ユーザー変更不可）:** sts_start（最初のステップ）, sts_goal（最後のステップ）

### approval_objects（承認オブジェクト）

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | SERIAL | PK | 承認オブジェクトID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子 |
| organization_id | UUID | FK, NOT NULL | 所属組織 |
| workflow_step_id | INTEGER | FK | 所属ステップ |
| order | INTEGER | NOT NULL, DEFAULT 1 | 表示順 |
| type | VARCHAR(20) | NOT NULL | role / user |
| role_id | INTEGER | FK, nullable | type=role のとき対象役職 |
| role_operator | VARCHAR(10) | nullable | type=role のとき: eq（イコール）/ gte（以上） |
| user_id | UUID | FK, nullable | type=user のとき対象ユーザー |
| points | INTEGER | NOT NULL, DEFAULT 1 | 承認時に加算する点数 |
| exclude_reporter | BOOLEAN | DEFAULT false | 起票者をこの承認オブジェクトの承認者から除外 |
| exclude_assignee | BOOLEAN | DEFAULT false | 担当者をこの承認オブジェクトの承認者から除外 |

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
| workflow_id | INTEGER | FK, nullable | 紐づくワークフロー |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

### issue_approvals

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | 承認ID |
| key | VARCHAR(255) | NOT NULL | API/URL 用識別子（id を格納） |
| organization_id | UUID | FK, NOT NULL | 所属組織 |
| issue_id | UUID | FK | 対象Issue |
| workflow_step_id | INTEGER | FK | ワークフローステップ |
| approver_id | UUID | FK, nullable | 承認者 |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'pending' | pending / approved / rejected |
| comment | TEXT | | コメント |
| acted_at | TIMESTAMP | nullable | 承認/却下日時 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

---

## ワークフローステップ（ステータス参照の双方向リスト）

表示順は意味を持たない。`status_id` → `next_status_id` のリンクで辿る。

| IDX | status | 説明 | 閾値 | 承認後ステータス |
|:----|:-------|:-----|:-----|:----------------|
| 1 | sts_start | 最初のステップ。ユーザー変更不可 | 10 | 未着手 |
| 2 | 未着手 | ユーザーで変更可能 | 10 | 進行中 |
| 3 | 進行中 | ユーザーで変更可能 | 10 | sts_goal |
| 4 | sts_goal | 最後のステップ。ユーザー変更不可 | - | - |

**システムステータス:** `sts_start`, `sts_goal` は `status_key` で識別し、編集・削除不可。

---

## デフォルトステータス

新規プロジェクト作成時に以下のステータスを自動生成：

| order | name | color |
|-------|------|-------|
| 1 | 未着手 | #6B7280 |
| 2 | 進行中 | #3B82F6 |
| 3 | レビュー中 | #F59E0B |
| 4 | 完了 | #10B981 |
