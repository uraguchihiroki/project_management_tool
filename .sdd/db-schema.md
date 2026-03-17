# データベース設計

## ER図

```
users
├── id (PK)
├── name
├── email (UNIQUE)
├── avatar_url
└── created_at

projects
├── id (PK)
├── key         (例: "PROJ", プロジェクト識別子)
├── name
├── description
├── owner_id (FK → users.id)
└── created_at

issues
├── id (PK)
├── number      (プロジェクト内の連番 例: PROJ-1)
├── title
├── description
├── status_id (FK → statuses.id)
├── priority    (low / medium / high / critical)
├── assignee_id (FK → users.id, nullable)
├── reporter_id (FK → users.id)
├── project_id  (FK → projects.id)
├── due_date    (nullable)
├── created_at
└── updated_at

statuses
├── id (PK)
├── project_id (FK → projects.id)
├── name        (例: "未着手", "進行中", "レビュー中", "完了")
├── color       (HEXカラーコード)
└── order       (表示順)

comments
├── id (PK)
├── issue_id (FK → issues.id)
├── author_id (FK → users.id)
├── body
├── created_at
└── updated_at

labels
├── id (PK)
├── project_id (FK → projects.id)
├── name
└── color

issue_labels (中間テーブル)
├── issue_id (FK → issues.id)
└── label_id (FK → labels.id)
```

---

## テーブル定義

### users

| カラム | 型 | 制約 | 説明 |
|---|---|---|---|
| id | UUID | PK | ユーザーID |
| name | VARCHAR(100) | NOT NULL | 表示名 |
| email | VARCHAR(255) | UNIQUE, NOT NULL | メールアドレス |
| avatar_url | TEXT | nullable | アバター画像URL |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

### projects

| カラム | 型 | 制約 | 説明 |
|---|---|---|---|
| id | UUID | PK | プロジェクトID |
| key | VARCHAR(10) | UNIQUE, NOT NULL | 識別キー（大文字英数字） |
| name | VARCHAR(200) | NOT NULL | プロジェクト名 |
| description | TEXT | nullable | 説明 |
| owner_id | UUID | FK | オーナーユーザー |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

### issues

| カラム | 型 | 制約 | 説明 |
|---|---|---|---|
| id | UUID | PK | IssueID |
| number | INTEGER | NOT NULL | プロジェクト内連番 |
| title | VARCHAR(500) | NOT NULL | タイトル |
| description | TEXT | nullable | 詳細説明 |
| status_id | UUID | FK | ステータス |
| priority | VARCHAR(20) | NOT NULL | 優先度 |
| assignee_id | UUID | FK, nullable | 担当者 |
| reporter_id | UUID | FK | 起票者 |
| project_id | UUID | FK | 所属プロジェクト |
| due_date | DATE | nullable | 期日 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

### statuses

| カラム | 型 | 制約 | 説明 |
|---|---|---|---|
| id | UUID | PK | ステータスID |
| project_id | UUID | FK | 所属プロジェクト |
| name | VARCHAR(50) | NOT NULL | ステータス名 |
| color | VARCHAR(7) | NOT NULL | HEXカラー (#RRGGBB) |
| order | INTEGER | NOT NULL | 表示順 |

### comments

| カラム | 型 | 制約 | 説明 |
|---|---|---|---|
| id | UUID | PK | コメントID |
| issue_id | UUID | FK | 対象Issue |
| author_id | UUID | FK | 投稿者 |
| body | TEXT | NOT NULL | 本文 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

---

## デフォルトステータス

新規プロジェクト作成時に以下のステータスを自動生成：

| order | name | color |
|---|---|---|
| 1 | 未着手 | #6B7280 |
| 2 | 進行中 | #3B82F6 |
| 3 | レビュー中 | #F59E0B |
| 4 | 完了 | #10B981 |
