package db

import "gorm.io/gorm"

// MigrateDropApprovalTables drops legacy approval-step tables.
// Safe to run repeatedly and skips missing tables.
func MigrateDropApprovalTables(db *gorm.DB) error {
	for _, table := range []string{
		"issue_approvals",
		"approval_objects",
		"workflow_steps",
	} {
		if db.Migrator().HasTable(table) {
			if err := db.Migrator().DropTable(table); err != nil {
				return err
			}
		}
	}
	return nil
}
