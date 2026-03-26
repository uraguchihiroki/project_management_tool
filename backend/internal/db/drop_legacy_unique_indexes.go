package db

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// MigrateDropLegacyBusinessUniqueIndexes は、PK 以外の業務用 UNIQUE を落とす（既存 DB 向け・冪等）。
// モデルから uniqueIndex を外したあとも、PostgreSQL / SQLite に残った旧インデックスを除去する。
func MigrateDropLegacyBusinessUniqueIndexes(db *gorm.DB) error {
	names := []string{
		"idx_user_org_email",
		"idx_statuses_status_key",
		"idx_statuses_wf_name_order_active",
		"idx_role_name_org",
		"idx_project_org_key",
		"uni_organizations_name",
		"uni_super_admins_email",
		"idx_organizations_name",
		"idx_super_admins_email",
	}
	dialect := db.Dialector.Name()
	for _, name := range names {
		stmt := dropIndexSQL(dialect, name)
		if stmt == "" {
			continue
		}
		if err := db.Exec(stmt).Error; err != nil {
			if isDropIndexNotExistErr(dialect, err) {
				continue
			}
			return fmt.Errorf("drop legacy unique index %q: %w", name, err)
		}
	}
	return nil
}

func dropIndexSQL(dialect, indexName string) string {
	switch dialect {
	case "postgres":
		return fmt.Sprintf("DROP INDEX IF EXISTS %s", indexName)
	case "sqlite":
		return fmt.Sprintf("DROP INDEX IF EXISTS %s", indexName)
	default:
		return ""
	}
}

func isDropIndexNotExistErr(dialect string, err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	switch dialect {
	case "postgres":
		return strings.Contains(msg, "UndefinedObject") || strings.Contains(msg, "does not exist")
	case "sqlite":
		return strings.Contains(msg, "no such index")
	default:
		return false
	}
}
