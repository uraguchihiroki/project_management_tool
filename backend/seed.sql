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

-- 11. Create initial super admin account
--    Email: superadmin@frs.example.com  (change as needed)
INSERT INTO super_admins (id, name, email, created_at)
VALUES (
    gen_random_uuid(),
    E'\u30b7\u30b9\u30c6\u30e0\u7ba1\u7406\u8005',
    'superadmin@frs.example.com',
    NOW()
)
ON CONFLICT (email) DO NOTHING;

SELECT 'Seed completed successfully.' AS result;
