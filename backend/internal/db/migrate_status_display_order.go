package db

import (
	"fmt"

	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

// MigrateStatusOrderToDisplayOrder はレガシー列名 `order` を `display_order` に寄せる（新規DBでは no-op）。
// MigrateStatusDedupeAndUniqueIndex より前に呼ぶこと。
func MigrateStatusOrderToDisplayOrder(db *gorm.DB) error {
	if err := db.Exec("DROP INDEX IF EXISTS " + statusUniqueIndexName).Error; err != nil {
		return fmt.Errorf("drop status unique index: %w", err)
	}

	dialect := db.Dialector.Name()
	switch dialect {
	case "sqlite":
		type colInfo struct {
			Name string
		}
		var cols []colInfo
		if err := db.Raw("PRAGMA table_info(statuses)").Scan(&cols).Error; err != nil {
			return fmt.Errorf("pragma table_info statuses: %w", err)
		}
		hasOrder := false
		hasDisplay := false
		for _, c := range cols {
			if c.Name == "order" {
				hasOrder = true
			}
			if c.Name == "display_order" {
				hasDisplay = true
			}
		}
		if hasOrder && !hasDisplay {
			if err := db.Exec(`ALTER TABLE statuses RENAME COLUMN "order" TO display_order`).Error; err != nil {
				return fmt.Errorf("sqlite rename order to display_order: %w", err)
			}
		} else if hasOrder && hasDisplay {
			if err := db.Exec(`UPDATE statuses SET display_order = "order" WHERE "order" IS NOT NULL`).Error; err != nil {
				return fmt.Errorf("sqlite copy order to display_order: %w", err)
			}
			if err := db.Exec(`ALTER TABLE statuses DROP COLUMN "order"`).Error; err != nil {
				return fmt.Errorf("sqlite drop order column: %w", err)
			}
		}
	case "postgres":
		var n int64
		if err := db.Raw(`
			SELECT COUNT(*) FROM information_schema.columns
			WHERE table_schema = current_schema() AND table_name = 'statuses' AND column_name = 'order'
		`).Scan(&n).Error; err != nil {
			return fmt.Errorf("postgres check order column: %w", err)
		}
		if n > 0 {
			var nd int64
			if err := db.Raw(`
				SELECT COUNT(*) FROM information_schema.columns
				WHERE table_schema = current_schema() AND table_name = 'statuses' AND column_name = 'display_order'
			`).Scan(&nd).Error; err != nil {
				return fmt.Errorf("postgres check display_order column: %w", err)
			}
			if nd == 0 {
				if err := db.Exec(`ALTER TABLE statuses RENAME COLUMN "order" TO display_order`).Error; err != nil {
					return fmt.Errorf("postgres rename order to display_order: %w", err)
				}
			} else {
				if err := db.Exec(`UPDATE statuses SET display_order = "order"`).Error; err != nil {
					return fmt.Errorf("postgres copy order to display_order: %w", err)
				}
				if err := db.Exec(`ALTER TABLE statuses DROP COLUMN "order"`).Error; err != nil {
					return fmt.Errorf("postgres drop order column: %w", err)
				}
			}
		}
	default:
		return fmt.Errorf("MigrateStatusOrderToDisplayOrder: unsupported dialect %q", dialect)
	}
	return nil
}

// MigrateWorkflowTransitionDisplayOrder は workflow_transitions.display_order を埋める（既存行は workflow 内 id 順で 1..n）。
func MigrateWorkflowTransitionDisplayOrder(db *gorm.DB) error {
	var wfIDs []uint
	if err := db.Model(&model.WorkflowTransition{}).Distinct().Pluck("workflow_id", &wfIDs).Error; err != nil {
		return fmt.Errorf("list workflow_ids: %w", err)
	}
	for _, wf := range wfIDs {
		var rows []model.WorkflowTransition
		if err := db.Where("workflow_id = ?", wf).Order("id ASC").Find(&rows).Error; err != nil {
			return err
		}
		for i := range rows {
			if err := db.Model(&model.WorkflowTransition{}).Where("id = ?", rows[i].ID).
				Update("display_order", i+1).Error; err != nil {
				return fmt.Errorf("set transition display_order: %w", err)
			}
		}
	}
	return nil
}
