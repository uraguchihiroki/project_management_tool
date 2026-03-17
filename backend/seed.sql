-- ============================================================
-- seed.sql
-- マルチテナント初期化スクリプト（一度だけ手動で実行）
-- 実行方法:
--   docker exec -i pmt_db psql -U pmt_user -d pmt_db < backend/seed.sql
-- ============================================================

-- 1. roles テーブルの旧グローバルユニーク制約を削除
--    （GORM が (name, organization_id) の複合ユニークに付け替えるため）
DROP INDEX IF EXISTS uni_roles_name;

-- 2. "Ｆ．Ｒ．Ｓ．" 組織を固定UUIDで挿入
INSERT INTO organizations (id, name, created_at)
VALUES ('00000000-0000-0000-0000-000000000001', 'Ｆ．Ｒ．Ｓ．', NOW())
ON CONFLICT (id) DO NOTHING;

-- 3. 既存のプロジェクトを "Ｆ．Ｒ．Ｓ．" 組織に紐付け
UPDATE projects
SET organization_id = '00000000-0000-0000-0000-000000000001'
WHERE organization_id IS NULL;

-- 4. 既存の役職を "Ｆ．Ｒ．Ｓ．" 組織に紐付け
UPDATE roles
SET organization_id = '00000000-0000-0000-0000-000000000001'
WHERE organization_id IS NULL;

-- 5. organization_id に NOT NULL 制約を追加（データ移行完了後）
ALTER TABLE projects ALTER COLUMN organization_id SET NOT NULL;
ALTER TABLE roles    ALTER COLUMN organization_id SET NOT NULL;

-- 6. FK 制約を追加（参照整合性をDBレベルで保証）
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

-- 7. 既存ユーザーを全員 "Ｆ．Ｒ．Ｓ．" 組織のメンバーに追加
INSERT INTO organization_users (organization_id, user_id, is_org_admin, joined_at)
SELECT '00000000-0000-0000-0000-000000000001', id, is_admin, NOW()
FROM users
ON CONFLICT (organization_id, user_id) DO NOTHING;

-- 8. スーパーアドミンの初期レコード（メール変更可）
--    このメールアドレスで /super-admin/login からログインできます
INSERT INTO super_admins (id, name, email, created_at)
VALUES (
    gen_random_uuid(),
    'システム管理者',
    'superadmin@frs.example.com',
    NOW()
)
ON CONFLICT (email) DO NOTHING;

SELECT 'Seed completed successfully.' AS result;
