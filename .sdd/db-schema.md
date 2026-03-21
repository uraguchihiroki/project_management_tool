# データベース設計

本ドキュメントは **Issue 管理システム**を正とする設計へ寄せて更新する。稟議・承認ワークフロー向けのテーブルは **廃止方向**（実装に残る場合は [transition-permissions.md](transition-permissions.md) のレガシー扱い）。

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

## ステータス遷移の権限（TBD）

**採用スキーマは未確定。** 議論のたたき台は [transition-permissions.md](transition-permissions.md)。合意後に本ドキュメントへカラム・テーブルを追記する。

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
| level | INTEGER | NOT NULL, DEFAULT 1 | ヒエラルキー用の数値（ステータス遷移権限の門番に再利用するかは [transition-permissions.md](transition-permissions.md) で決定） |
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

---

## デフォルトステータス

新規プロジェクト作成時に以下のステータスを自動生成：

| order | name | color |
|-------|------|-------|
| 1 | 未着手 | #6B7280 |
| 2 | 進行中 | #3B82F6 |
| 3 | レビュー中 | #F59E0B |
| 4 | 完了 | #10B981 |
