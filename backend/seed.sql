-- ============================================================
-- seed.sql
-- Multi-tenant initialization script (run manually once)
-- Usage:
--   Get-Content backend/seed.sql | docker exec -i pmt_db psql -U pmt_user -d pmt_db
-- ============================================================

-- 1. Drop old global unique constraint on roles.name
--    (GORM will replace with composite unique index (name, organization_id))
DROP INDEX IF EXISTS uni_roles_name;

-- 2. Add admin_email column if not exists (GORM AutoMigrate may have added it)
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS admin_email VARCHAR(255);

-- 3. Insert "F.R.S." organization with fixed UUID
--    Name stored as Unicode escape to avoid encoding issues
INSERT INTO organizations (id, key, name, admin_email, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    E'\uff26\uff0e\uff32\uff0e\uff33\uff0e',
    '',
    NOW()
)
ON CONFLICT (id) DO NOTHING;


-- 5. Backfill existing projects to F.R.S. organization
UPDATE projects
SET organization_id = '00000000-0000-0000-0000-000000000001'
WHERE organization_id IS NULL;

-- 6. Backfill existing roles to F.R.S. organization
UPDATE roles
SET organization_id = '00000000-0000-0000-0000-000000000001'
WHERE organization_id IS NULL;

-- 7. Add NOT NULL constraint after data is populated
ALTER TABLE projects ALTER COLUMN organization_id SET NOT NULL;
ALTER TABLE roles    ALTER COLUMN organization_id SET NOT NULL;

-- 8. Add FK constraints for referential integrity at DB level
ALTER TABLE projects
    DROP CONSTRAINT IF EXISTS fk_projects_organization;
ALTER TABLE projects
    ADD CONSTRAINT fk_projects_organization
    FOREIGN KEY (organization_id) REFERENCES organizations(id);

ALTER TABLE roles
    DROP CONSTRAINT IF EXISTS fk_roles_organization;
ALTER TABLE roles
    ADD CONSTRAINT fk_roles_organization
    FOREIGN KEY (organization_id) REFERENCES organizations(id);

-- 9. users.organization_id 移行（organization_users 廃止対応）
ALTER TABLE users ADD COLUMN IF NOT EXISTS organization_id UUID;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_org_admin BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS joined_at TIMESTAMP;
-- organization_users からバックフィル（存在する場合）
UPDATE users u SET
  organization_id = ou.organization_id,
  is_org_admin = ou.is_org_admin,
  joined_at = ou.joined_at
FROM organization_users ou
WHERE ou.user_id = u.id AND u.organization_id IS NULL;
-- 未設定ユーザーは FRS に
UPDATE users SET
  organization_id = '00000000-0000-0000-0000-000000000001',
  is_org_admin = COALESCE(is_admin, false),
  joined_at = COALESCE(joined_at, NOW())
WHERE organization_id IS NULL;
ALTER TABLE users ALTER COLUMN organization_id SET NOT NULL;
DROP INDEX IF EXISTS idx_user_org_email;
CREATE INDEX IF NOT EXISTS idx_user_org_email ON users(organization_id, email);

-- 10. Set admin_email for FRS from first org admin user
UPDATE organizations o
SET admin_email = (
    SELECT u.email FROM users u
    WHERE u.organization_id = o.id AND u.is_org_admin = true
    LIMIT 1
)
WHERE o.id = '00000000-0000-0000-0000-000000000001'
  AND (o.admin_email IS NULL OR o.admin_email = '');

-- 11. Workflow: organization_id 追加（組織所属）
ALTER TABLE workflows ADD COLUMN IF NOT EXISTS organization_id UUID;
-- 未設定は FRS に
UPDATE workflows SET organization_id = '00000000-0000-0000-0000-000000000001'
WHERE organization_id IS NULL;
ALTER TABLE workflows ALTER COLUMN organization_id SET NOT NULL;
ALTER TABLE workflows DROP COLUMN IF EXISTS project_id;

-- 12. Issue & Status: Phase 4 migration
-- Issue: add organization_id, backfill, project_id nullable
ALTER TABLE issues ADD COLUMN IF NOT EXISTS organization_id UUID;
UPDATE issues i SET organization_id = p.organization_id
FROM projects p WHERE i.project_id = p.id AND i.organization_id IS NULL;
UPDATE issues SET organization_id = '00000000-0000-0000-0000-000000000001'
WHERE organization_id IS NULL;
ALTER TABLE issues ALTER COLUMN organization_id SET NOT NULL;
ALTER TABLE issues ALTER COLUMN project_id DROP NOT NULL;

-- Status: add organization_id, project_id nullable
ALTER TABLE statuses ADD COLUMN IF NOT EXISTS organization_id UUID;
ALTER TABLE statuses ALTER COLUMN project_id DROP NOT NULL;

-- 13a–13b. 廃止: statuses は Issue 専用（workflow_id）。プロジェクト進行は project_statuses。
-- 旧 type 列はサーバ起動時 MigrateIssueProjectStatusSplitPre で DROP。
-- 組織横断の Issue 列・project_statuses の初期値は API 経由のシード（組織作成・プロジェクト作成）を正とする。

-- 13. 承認ステップ系テーブル（workflow_steps/approval_objects/issue_approvals）は廃止（Issue トラッキングを正とする）

-- 13c. projects.status カラム削除（ステータステーブル参照に移行したため未使用）
ALTER TABLE projects DROP COLUMN IF EXISTS status;

-- 13d. workflows.organization_id は組織所属のため保持（削除しない）

-- 14. Display order columns for drag-and-drop reordering
ALTER TABLE roles ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 1;
UPDATE roles r SET display_order = sub.rn FROM (
  SELECT id, ROW_NUMBER() OVER (PARTITION BY COALESCE(organization_id::text, '') ORDER BY level DESC, name) AS rn FROM roles
) sub WHERE r.id = sub.id;

ALTER TABLE projects ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 1;
UPDATE projects p SET display_order = sub.rn FROM (
  SELECT id, ROW_NUMBER() OVER (PARTITION BY COALESCE(organization_id::text, '') ORDER BY created_at) AS rn FROM projects
) sub WHERE p.id = sub.id;

ALTER TABLE workflows ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 1;
UPDATE workflows w SET display_order = sub.rn FROM (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at) AS rn FROM workflows
) sub WHERE w.id = sub.id;

ALTER TABLE issue_templates ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 1;
UPDATE issue_templates t SET display_order = sub.rn FROM (
  SELECT id, ROW_NUMBER() OVER (PARTITION BY project_id ORDER BY name) AS rn FROM issue_templates
) sub WHERE t.id = sub.id;

-- 14e. 承認ステップ系（workflow_steps/approval_objects）は廃止

-- 14g. ステータスベースワークフローステップ
ALTER TABLE statuses ADD COLUMN IF NOT EXISTS status_key VARCHAR(50);
CREATE INDEX IF NOT EXISTS idx_statuses_status_key ON statuses(status_key) WHERE status_key IS NOT NULL AND status_key != '';

-- sts_start, sts_goal システムステータス（全会社共通・グローバル）
INSERT INTO statuses (id, project_id, organization_id, name, color, "order", status_key)
VALUES
  ('30000000-0000-0000-0000-000000000001', NULL, NULL, 'sts_start', '#9CA3AF', 0, 'sts_start'),
  ('30000000-0000-0000-0000-000000000002', NULL, NULL, 'sts_goal', '#9CA3AF', 99, 'sts_goal')
ON CONFLICT (id) DO UPDATE SET status_key = EXCLUDED.status_key, name = EXCLUDED.name, organization_id = NULL;

UPDATE statuses SET organization_id = NULL WHERE status_key IN ('sts_start','sts_goal');

-- workflow_steps: next_status_id は廃止

-- 14f. comments, issue_templates に organization_id 追加
ALTER TABLE comments ADD COLUMN IF NOT EXISTS organization_id UUID;
UPDATE comments c SET organization_id = i.organization_id FROM issues i WHERE c.issue_id = i.id AND c.organization_id IS NULL;
UPDATE comments SET organization_id = '00000000-0000-0000-0000-000000000001' WHERE organization_id IS NULL;
ALTER TABLE comments ALTER COLUMN organization_id SET NOT NULL;

-- workflow_steps/approval_objects/issue_approvals は廃止のため organization_id 付与も不要

ALTER TABLE issue_templates ADD COLUMN IF NOT EXISTS organization_id UUID;
UPDATE issue_templates it SET organization_id = p.organization_id FROM projects p WHERE it.project_id = p.id AND it.organization_id IS NULL;
UPDATE issue_templates SET organization_id = '00000000-0000-0000-0000-000000000001' WHERE organization_id IS NULL;
ALTER TABLE issue_templates ALTER COLUMN organization_id SET NOT NULL;

-- issue_approvals は廃止

-- 14f2. organization_users 廃止（users に organization_id を移行済み）
DROP TABLE IF EXISTS organization_users CASCADE;

-- 14f3. 論理削除: 全テーブルに deleted_at 追加
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE super_admins ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE organization_user_groups ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE roles ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE statuses ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE issues ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE comments ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE workflows ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE issue_templates ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_organizations_deleted_at ON organizations(deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_groups_deleted_at ON groups(deleted_at);
CREATE INDEX IF NOT EXISTS idx_organization_user_groups_deleted_at ON organization_user_groups(deleted_at);
CREATE INDEX IF NOT EXISTS idx_roles_deleted_at ON roles(deleted_at);
CREATE INDEX IF NOT EXISTS idx_projects_deleted_at ON projects(deleted_at);
CREATE INDEX IF NOT EXISTS idx_statuses_deleted_at ON statuses(deleted_at);
CREATE INDEX IF NOT EXISTS idx_issues_deleted_at ON issues(deleted_at);
CREATE INDEX IF NOT EXISTS idx_comments_deleted_at ON comments(deleted_at);
CREATE INDEX IF NOT EXISTS idx_workflows_deleted_at ON workflows(deleted_at);
CREATE INDEX IF NOT EXISTS idx_issue_templates_deleted_at ON issue_templates(deleted_at);
-- 承認ステップ系テーブルの deleted_at / index は廃止

-- 14h. 全テーブルに key カラム追加（VARCHAR(255), NOT NULL）
-- projects は既に key あり（組織内識別キー）のためスキップ
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS key VARCHAR(255);
UPDATE organizations SET key = id::text WHERE key IS NULL OR key = '';
ALTER TABLE organizations ALTER COLUMN key SET NOT NULL;

ALTER TABLE super_admins ADD COLUMN IF NOT EXISTS key VARCHAR(255);
UPDATE super_admins SET key = id::text WHERE key IS NULL OR key = '';
ALTER TABLE super_admins ALTER COLUMN key SET NOT NULL;

ALTER TABLE users ADD COLUMN IF NOT EXISTS key VARCHAR(255);
UPDATE users SET key = id::text WHERE key IS NULL OR key = '';
ALTER TABLE users ALTER COLUMN key SET NOT NULL;

ALTER TABLE roles ADD COLUMN IF NOT EXISTS key VARCHAR(255);
UPDATE roles SET key = 'role-' || id::text WHERE key IS NULL OR key = '';
ALTER TABLE roles ALTER COLUMN key SET NOT NULL;

ALTER TABLE groups ADD COLUMN IF NOT EXISTS key VARCHAR(255);
UPDATE groups SET key = id::text WHERE key IS NULL OR key = '';
ALTER TABLE groups ALTER COLUMN key SET NOT NULL;

ALTER TABLE organization_user_groups ADD COLUMN IF NOT EXISTS key VARCHAR(255);
UPDATE organization_user_groups SET key = organization_id::text || '-' || user_id::text || '-' || group_id::text WHERE key IS NULL OR key = '';
ALTER TABLE organization_user_groups ALTER COLUMN key SET NOT NULL;

ALTER TABLE statuses ADD COLUMN IF NOT EXISTS key VARCHAR(255);
UPDATE statuses SET key = COALESCE(NULLIF(TRIM(status_key), ''), 'sts-' || id::text) WHERE key IS NULL OR key = '';
UPDATE statuses SET key = id::text WHERE key IS NULL OR key = '';  -- fallback for any remaining
ALTER TABLE statuses ALTER COLUMN key SET NOT NULL;

ALTER TABLE issues ADD COLUMN IF NOT EXISTS key VARCHAR(255);
UPDATE issues SET key = id::text WHERE key IS NULL OR key = '';
ALTER TABLE issues ALTER COLUMN key SET NOT NULL;

ALTER TABLE comments ADD COLUMN IF NOT EXISTS key VARCHAR(255);
UPDATE comments SET key = id::text WHERE key IS NULL OR key = '';
ALTER TABLE comments ALTER COLUMN key SET NOT NULL;

ALTER TABLE workflows ADD COLUMN IF NOT EXISTS key VARCHAR(255);
UPDATE workflows SET key = 'wf-' || id::text WHERE key IS NULL OR key = '';
ALTER TABLE workflows ALTER COLUMN key SET NOT NULL;


ALTER TABLE issue_templates ADD COLUMN IF NOT EXISTS key VARCHAR(255);
UPDATE issue_templates SET key = 'tmpl-' || id::text WHERE key IS NULL OR key = '';
ALTER TABLE issue_templates ALTER COLUMN key SET NOT NULL;

-- 承認ステップ系テーブルの key 追加は廃止

-- user_roles（GORM many2many 中間テーブル）
ALTER TABLE user_roles ADD COLUMN IF NOT EXISTS key VARCHAR(255);
UPDATE user_roles SET key = user_id::text || '-' || role_id::text WHERE key IS NULL OR key = '';
ALTER TABLE user_roles ALTER COLUMN key SET NOT NULL;

-- 15. Create initial super admin account
--    Email: superadmin@frs.example.com  (change as needed)
INSERT INTO super_admins (id, key, name, email, created_at)
SELECT gen_random_uuid(), 'superadmin@frs.example.com', E'\u30b7\u30b9\u30c6\u30e0\u7ba1\u7406\u8005', 'superadmin@frs.example.com', NOW()
WHERE NOT EXISTS (SELECT 1 FROM super_admins WHERE email = 'superadmin@frs.example.com');
UPDATE super_admins SET key = COALESCE(NULLIF(key, ''), id::text) WHERE key IS NULL OR key = '';

-- 16. Display order backfill complete
SELECT 'Seed completed successfully.' AS result;
