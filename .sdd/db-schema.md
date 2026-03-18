# データベース設計

## ER図

```
organizations
├── id (PK)
├── name (UNIQUE)
├── admin_email
└── created_at

super_admins
├── id (PK)
├── name
├── email (UNIQUE)
└── created_at

users
├── id (PK)
├── name
├── email (UNIQUE)
├── avatar_url (nullable)
├── is_admin
└── created_at

organization_users (複合PK: organization_id, user_id)
├── organization_id (FK → organizations.id)
├── user_id (FK → users.id)
├── is_org_admin
└── joined_at

roles
├── id (PK, auto)
├── name
├── level
├── description
├── organization_id (FK → organizations.id, nullable)
└── created_at

user_roles (中間テーブル, many2many)
├── user_id (FK → users.id)
└── role_id (FK → roles.id)

projects
├── id (PK)
├── key (UNIQUE)
├── name
├── description (nullable)
├── owner_id (FK → users.id)
├── organization_id (FK → organizations.id, nullable)
└── created_at

statuses
├── id (PK)
├── project_id (FK → projects.id)
├── name
├── color (HEX)
└── order

issues
├── id (PK)
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
├── issue_id (FK → issues.id)
├── author_id (FK → users.id)
├── body
├── created_at
└── updated_at

workflows
├── id (PK, auto)
├── organization_id (FK → organizations.id)
├── name
├── description
└── created_at

workflow_steps
├── id (PK, auto)
├── workflow_id (FK → workflows.id)
├── order
├── name
├── required_level
├── status_id (FK → statuses.id, nullable)
├── approver_type (role / user / multiple)
├── approver_user_id (FK → users.id, nullable)
├── min_approvers (default 1)
├── exclude_reporter (default false)
└── exclude_assignee (default false)

issue_templates
├── id (PK, auto)
├── project_id (FK → projects.id)
├── name
├── description
├── body
├── default_priority
├── workflow_id (FK → workflows.id, nullable)
└── created_at

issue_approvals
├── id (PK)
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
| name | VARCHAR(200) | UNIQUE, NOT NULL | 組織名 |
| admin_email | VARCHAR(255) | nullable | 組織管理者のメールアドレス |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

### super_admins

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | スーパー管理者ID |
| name | VARCHAR(100) | NOT NULL | 表示名 |
| email | VARCHAR(255) | UNIQUE, NOT NULL | メールアドレス |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

### users

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | ユーザーID |
| name | VARCHAR(100) | NOT NULL | 表示名 |
| email | VARCHAR(255) | UNIQUE, NOT NULL | メールアドレス |
| avatar_url | TEXT | nullable | アバター画像URL |
| is_admin | BOOLEAN | DEFAULT false | システム管理者フラグ |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

### organization_users

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| organization_id | UUID | PK, FK | 組織ID |
| user_id | UUID | PK, FK | ユーザーID |
| is_org_admin | BOOLEAN | DEFAULT false | 組織管理者フラグ |
| joined_at | TIMESTAMP | NOT NULL | 参加日時 |

### roles

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | SERIAL | PK | 役職ID |
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
| key | VARCHAR(10) | UNIQUE, NOT NULL | 識別キー（大文字英数字） |
| name | VARCHAR(200) | NOT NULL | プロジェクト名 |
| description | TEXT | nullable | 説明 |
| owner_id | UUID | FK | オーナーユーザー |
| organization_id | UUID | FK, nullable | 所属組織 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

### statuses

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | ステータスID |
| project_id | UUID | FK | 所属プロジェクト |
| name | VARCHAR(50) | NOT NULL | ステータス名 |
| color | VARCHAR(7) | NOT NULL | HEXカラー (#RRGGBB) |
| order | INTEGER | NOT NULL | 表示順 |

### issues

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | UUID | PK | IssueID |
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
| issue_id | UUID | FK | 対象Issue |
| author_id | UUID | FK | 投稿者 |
| body | TEXT | NOT NULL | 本文 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

### workflows

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | SERIAL | PK | ワークフローID |
| organization_id | UUID | FK | 所属組織 |
| name | VARCHAR(200) | NOT NULL | ワークフロー名 |
| description | VARCHAR(500) | | 説明 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

### workflow_steps

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | SERIAL | PK | ステップID |
| workflow_id | INTEGER | FK | 所属ワークフロー |
| order | INTEGER | NOT NULL, DEFAULT 1 | 表示順 |
| name | VARCHAR(200) | NOT NULL | ステップ名 |
| required_level | INTEGER | NOT NULL, DEFAULT 1 | 承認に必要な役職レベル |
| status_id | UUID | FK, nullable | 紐づくステータス |
| approver_type | VARCHAR(20) | DEFAULT 'role' | role / user / multiple |
| approver_user_id | UUID | FK, nullable | 個人指定時のユーザーID |
| min_approvers | INTEGER | DEFAULT 1 | 最低承認者数 |
| exclude_reporter | BOOLEAN | DEFAULT false | 起票者を除外 |
| exclude_assignee | BOOLEAN | DEFAULT false | 担当者を除外 |

### issue_templates

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | SERIAL | PK | テンプレートID |
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
| issue_id | UUID | FK | 対象Issue |
| workflow_step_id | INTEGER | FK | ワークフローステップ |
| approver_id | UUID | FK, nullable | 承認者 |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'pending' | pending / approved / rejected |
| comment | TEXT | | コメント |
| acted_at | TIMESTAMP | nullable | 承認/却下日時 |
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
