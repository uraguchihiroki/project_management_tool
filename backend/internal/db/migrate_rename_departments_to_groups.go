package db

import (
	"fmt"

	"gorm.io/gorm"
)

// MigrateRenameDepartmentsToGroups renames legacy department tables to group tables (PostgreSQL only).
// It is idempotent and safe to run multiple times.
func MigrateRenameDepartmentsToGroups(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}

	// departments -> groups
	hasDepartments := db.Migrator().HasTable("departments")
	hasGroups := db.Migrator().HasTable("groups")
	if hasDepartments && !hasGroups {
		if err := db.Exec(`ALTER TABLE "departments" RENAME TO "groups"`).Error; err != nil {
			return fmt.Errorf("rename departments->groups: %w", err)
		}
	}

	// organization_user_departments -> organization_user_groups
	hasOUD := db.Migrator().HasTable("organization_user_departments")
	hasOUG := db.Migrator().HasTable("organization_user_groups")
	if hasOUD && !hasOUG {
		if err := db.Exec(`ALTER TABLE "organization_user_departments" RENAME TO "organization_user_groups"`).Error; err != nil {
			return fmt.Errorf("rename organization_user_departments->organization_user_groups: %w", err)
		}
	}

	// department_id -> group_id
	// Do this after table rename so we always target the new table name when present.
	if db.Migrator().HasTable("organization_user_groups") {
		if db.Migrator().HasColumn("organization_user_groups", "department_id") && !db.Migrator().HasColumn("organization_user_groups", "group_id") {
			if err := db.Exec(`ALTER TABLE "organization_user_groups" RENAME COLUMN "department_id" TO "group_id"`).Error; err != nil {
				return fmt.Errorf("rename organization_user_groups.department_id->group_id: %w", err)
			}
		}
	}

	return nil
}

