package db

import (
	"fmt"

	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

// MigrateEnsureDefaultIssueEntryStatus は論理削除されていない statuses について、
// 同一 workflow_id で is_entry=true が0件のとき、display_order 最小の1行に is_entry=true を立てる（他行は false）。冪等。
func MigrateEnsureDefaultIssueEntryStatus(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var wfIDs []uint
		if err := tx.Model(&model.Status{}).
			Where("deleted_at IS NULL").
			Distinct("workflow_id").
			Pluck("workflow_id", &wfIDs).Error; err != nil {
			return fmt.Errorf("list workflow ids for entry backfill: %w", err)
		}
		for _, wfID := range wfIDs {
			var entryCount int64
			if err := tx.Model(&model.Status{}).
				Where("workflow_id = ? AND deleted_at IS NULL AND is_entry = ?", wfID, true).
				Count(&entryCount).Error; err != nil {
				return err
			}
			if entryCount > 0 {
				continue
			}
			var pick model.Status
			if err := tx.Where("workflow_id = ? AND deleted_at IS NULL", wfID).
				Order("display_order ASC, id ASC").
				First(&pick).Error; err != nil {
				return fmt.Errorf("pick default entry for workflow %d: %w", wfID, err)
			}
			if err := tx.Model(&model.Status{}).
				Where("workflow_id = ? AND deleted_at IS NULL AND id <> ?", wfID, pick.ID).
				Update("is_entry", false).Error; err != nil {
				return fmt.Errorf("clear is_entry peers workflow %d: %w", wfID, err)
			}
			if err := tx.Model(&model.Status{}).Where("id = ?", pick.ID).Update("is_entry", true).Error; err != nil {
				return fmt.Errorf("set default entry workflow %d: %w", wfID, err)
			}
		}
		return nil
	})
}
