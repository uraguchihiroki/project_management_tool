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
INSERT INTO organizations (id, name, admin_email, created_at)
VALUES (
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

ALTER TABLE organization_users
    DROP CONSTRAINT IF EXISTS fk_org_users_organization;
ALTER TABLE organization_users
    ADD CONSTRAINT fk_org_users_organization
    FOREIGN KEY (organization_id) REFERENCES organizations(id);

ALTER TABLE organization_users
    DROP CONSTRAINT IF EXISTS fk_org_users_user;
ALTER TABLE organization_users
    ADD CONSTRAINT fk_org_users_user
    FOREIGN KEY (user_id) REFERENCES users(id);

-- 9. Add all existing users to F.R.S. organization
INSERT INTO organization_users (organization_id, user_id, is_org_admin, joined_at)
SELECT '00000000-0000-0000-0000-000000000001', id, is_admin, NOW()
FROM users
ON CONFLICT (organization_id, user_id) DO NOTHING;

-- 10. Set admin_email for FRS from first org admin user
UPDATE organizations o
SET admin_email = (
    SELECT u.email FROM users u
    JOIN organization_users ou ON ou.user_id = u.id
    WHERE ou.organization_id = o.id AND ou.is_org_admin = true
    LIMIT 1
)
WHERE o.id = '00000000-0000-0000-0000-000000000001'
  AND (o.admin_email IS NULL OR o.admin_email = '');

-- 11. Workflow: project_id → organization_id migration (Phase 3)
--    Add organization_id if column does not exist (GORM AutoMigrate may have added it)
ALTER TABLE workflows ADD COLUMN IF NOT EXISTS organization_id UUID;
-- Backfill from project's organization
UPDATE workflows w
SET organization_id = p.organization_id
FROM projects p
WHERE w.project_id = p.id AND w.organization_id IS NULL;
-- Set to FRS for any orphaned workflows
UPDATE workflows SET organization_id = '00000000-0000-0000-0000-000000000001'
WHERE organization_id IS NULL;
ALTER TABLE workflows ALTER COLUMN organization_id SET NOT NULL;
-- Drop project_id (ignore if already dropped)
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

-- 組織用デフォルトステータス（FRS用、project_id なし）
INSERT INTO statuses (id, project_id, organization_id, name, color, "order")
SELECT gen_random_uuid(), NULL, '00000000-0000-0000-0000-000000000001', n, c, o
FROM (VALUES ('未着手', '#6B7280', 1), ('進行中', '#3B82F6', 2), ('完了', '#10B981', 3)) AS t(n, c, o)
WHERE NOT EXISTS (SELECT 1 FROM statuses WHERE organization_id = '00000000-0000-0000-0000-000000000001' AND project_id IS NULL LIMIT 1);

-- 13a. Status: add type column (issue | project)
ALTER TABLE statuses ADD COLUMN IF NOT EXISTS type VARCHAR(20) NOT NULL DEFAULT 'issue';
UPDATE statuses SET type = 'issue' WHERE type IS NULL OR type = '';
UPDATE statuses SET type = 'issue' WHERE type NOT IN ('issue', 'project');
-- 組織用デフォルト Project ステータス（計画中, 進行中, 完了）
INSERT INTO statuses (id, project_id, organization_id, name, color, "order", type)
SELECT gen_random_uuid(), NULL, '00000000-0000-0000-0000-000000000001', n, c, o, 'project'
FROM (VALUES ('計画中', '#6B7280', 1), ('進行中', '#3B82F6', 2), ('完了', '#10B981', 3)) AS t(n, c, o)
WHERE NOT EXISTS (SELECT 1 FROM statuses WHERE organization_id = '00000000-0000-0000-0000-000000000001' AND project_id IS NULL AND type = 'project' LIMIT 1);

-- 13. WorkflowStep: Phase 5 承認対象拡張
ALTER TABLE workflow_steps ADD COLUMN IF NOT EXISTS approver_type VARCHAR(20) DEFAULT 'role';
ALTER TABLE workflow_steps ADD COLUMN IF NOT EXISTS approver_user_id UUID;
ALTER TABLE workflow_steps ADD COLUMN IF NOT EXISTS min_approvers INTEGER DEFAULT 1;
ALTER TABLE workflow_steps ADD COLUMN IF NOT EXISTS exclude_reporter BOOLEAN DEFAULT false;
ALTER TABLE workflow_steps ADD COLUMN IF NOT EXISTS exclude_assignee BOOLEAN DEFAULT false;

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
  SELECT id, ROW_NUMBER() OVER (PARTITION BY organization_id ORDER BY created_at) AS rn FROM workflows
) sub WHERE w.id = sub.id;

ALTER TABLE issue_templates ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 1;
UPDATE issue_templates t SET display_order = sub.rn FROM (
  SELECT id, ROW_NUMBER() OVER (PARTITION BY project_id ORDER BY name) AS rn FROM issue_templates
) sub WHERE t.id = sub.id;

-- 15. Create initial super admin account
--    Email: superadmin@frs.example.com  (change as needed)
INSERT INTO super_admins (id, name, email, created_at)
VALUES (
    gen_random_uuid(),
    E'\u30b7\u30b9\u30c6\u30e0\u7ba1\u7406\u8005',
    'superadmin@frs.example.com',
    NOW()
)
ON CONFLICT (email) DO NOTHING;

-- 16. Display order backfill complete
SELECT 'Seed completed successfully.' AS result;
