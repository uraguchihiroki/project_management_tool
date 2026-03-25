package db

import (
	"fmt"

	"gorm.io/gorm"
)

// MigrateJunctionOrganizationID は結合テーブルに organization_id を追加し、
// 既存データを backfill して NOT NULL 化する（PostgreSQL）。冪等。
func MigrateJunctionOrganizationID(db *gorm.DB) error {
	hasUserRoles := db.Migrator().HasTable("user_roles")
	hasUserGroups := db.Migrator().HasTable("user_groups")
	hasIssueGroups := db.Migrator().HasTable("issue_groups")

	if hasUserRoles {
		if err := ensureColumn(db, "user_roles", "organization_id", "uuid"); err != nil {
			return fmt.Errorf("user_roles add organization_id: %w", err)
		}
	}
	if hasUserGroups {
		if err := ensureColumn(db, "user_groups", "organization_id", "uuid"); err != nil {
			return fmt.Errorf("user_groups add organization_id: %w", err)
		}
	}
	if hasIssueGroups {
		if err := ensureColumn(db, "issue_groups", "organization_id", "uuid"); err != nil {
			return fmt.Errorf("issue_groups add organization_id: %w", err)
		}
	}

	switch db.Dialector.Name() {
	case "postgres":
		if err := db.Exec(`
			UPDATE user_roles ur
			SET organization_id = u.organization_id
			FROM users u
			WHERE ur.user_id = u.id
			  AND ur.organization_id IS NULL
		`).Error; err != nil {
			return fmt.Errorf("user_roles backfill org_id: %w", err)
		}
		if hasUserGroups {
			if err := db.Exec(`
				UPDATE user_groups ug
				SET organization_id = g.organization_id
				FROM groups g
				WHERE ug.group_id = g.id
				  AND ug.organization_id IS NULL
			`).Error; err != nil {
				return fmt.Errorf("user_groups backfill org_id: %w", err)
			}
			if err := db.Exec(`ALTER TABLE user_groups ALTER COLUMN organization_id SET NOT NULL`).Error; err != nil {
				return fmt.Errorf("user_groups organization_id not null: %w", err)
			}
		}
		if hasIssueGroups {
			if err := db.Exec(`
				UPDATE issue_groups ig
				SET organization_id = i.organization_id
				FROM issues i
				WHERE ig.issue_id = i.id
				  AND ig.organization_id IS NULL
			`).Error; err != nil {
				return fmt.Errorf("issue_groups backfill org_id: %w", err)
			}
			if err := db.Exec(`ALTER TABLE issue_groups ALTER COLUMN organization_id SET NOT NULL`).Error; err != nil {
				return fmt.Errorf("issue_groups organization_id not null: %w", err)
			}
		}
		if err := db.Exec(`ALTER TABLE user_roles ALTER COLUMN organization_id SET NOT NULL`).Error; err != nil {
			return fmt.Errorf("user_roles organization_id not null: %w", err)
		}
	case "sqlite":
		if err := db.Exec(`
			UPDATE user_roles
			SET organization_id = (
				SELECT u.organization_id FROM users u WHERE u.id = user_roles.user_id
			)
			WHERE organization_id IS NULL
		`).Error; err != nil {
			return fmt.Errorf("sqlite user_roles backfill org_id: %w", err)
		}
		if hasUserGroups {
			if err := db.Exec(`
				UPDATE user_groups
				SET organization_id = (
					SELECT g.organization_id FROM groups g WHERE g.id = user_groups.group_id
				)
				WHERE organization_id IS NULL
			`).Error; err != nil {
				return fmt.Errorf("sqlite user_groups backfill org_id: %w", err)
			}
		}
		if hasIssueGroups {
			if err := db.Exec(`
				UPDATE issue_groups
				SET organization_id = (
					SELECT i.organization_id FROM issues i WHERE i.id = issue_groups.issue_id
				)
				WHERE organization_id IS NULL
			`).Error; err != nil {
				return fmt.Errorf("sqlite issue_groups backfill org_id: %w", err)
			}
		}
	}

	return nil
}

func ensureColumn(db *gorm.DB, table, col, colType string) error {
	exists := false
	switch db.Dialector.Name() {
	case "postgres":
		var n int64
		if err := db.Raw(`
			SELECT COUNT(*) FROM information_schema.columns
			WHERE table_schema = current_schema()
			  AND table_name = ?
			  AND column_name = ?
		`, table, col).Scan(&n).Error; err != nil {
			return err
		}
		exists = n > 0
	case "sqlite":
		var n int64
		q := fmt.Sprintf(`SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name = ?`, table)
		if err := db.Raw(q, col).Scan(&n).Error; err != nil {
			return err
		}
		exists = n > 0
	default:
		return fmt.Errorf("unsupported dialect: %s", db.Dialector.Name())
	}
	if exists {
		return nil
	}
	return db.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, table, col, colType)).Error
}
