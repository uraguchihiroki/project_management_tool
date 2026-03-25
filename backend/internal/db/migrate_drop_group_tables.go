package db

import "gorm.io/gorm"

// MigrateDropGroupTables drops legacy Group-related tables.
// Safe to run repeatedly and skips missing tables.
func MigrateDropGroupTables(db *gorm.DB) error {
	for _, table := range []string{
		"issue_groups",
		"user_groups",
		"groups",
	} {
		if db.Migrator().HasTable(table) {
			if err := db.Migrator().DropTable(table); err != nil {
				return err
			}
		}
	}
	return nil
}
