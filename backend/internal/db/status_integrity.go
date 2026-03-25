package db

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

const statusWorkflowNameOrderLookupIndex = "idx_statuses_wf_name_order_active"

// MigrateStatusDedupe は (1) 同一 workflow 内の重複 statuses を参照先を付け替えて1行にまとめ、
// (2) 有効行のみ対象とする部分インデックス（非一意）を作成する。サーバ起動・テストの AutoMigrate 後に1回呼ぶ。冪等。
// 業務上の (workflow_id, name, display_order) の一意は Service で保証する。
func MigrateStatusDedupe(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := dedupeStatuses(tx); err != nil {
			return err
		}
		if err := dedupeWorkflowTransitions(tx); err != nil {
			return err
		}
		if err := ensureStatusPartialLookupIndex(tx); err != nil {
			return err
		}
		return nil
	})
}

func dedupeStatuses(tx *gorm.DB) error {
	var all []model.Status
	if err := tx.Find(&all).Error; err != nil {
		return fmt.Errorf("status dedupe: load statuses: %w", err)
	}

	groups := make(map[string][]model.Status)
	for _, s := range all {
		key := fmt.Sprintf("%d|%s|%d", s.WorkflowID, s.Name, s.DisplayOrder)
		groups[key] = append(groups[key], s)
	}

	remap := make(map[uuid.UUID]uuid.UUID)
	for key, sts := range groups {
		if len(sts) < 2 {
			continue
		}
		canonical := pickCanonicalStatus(sts)
		for _, s := range sts {
			if s.ID != canonical {
				remap[s.ID] = canonical
			}
		}
		log.Printf("status dedupe: merging %d rows for key %q -> canonical %s", len(sts), key, canonical)
	}

	if len(remap) == 0 {
		return nil
	}

	oldIDs := make([]uuid.UUID, 0, len(remap))
	for id := range remap {
		oldIDs = append(oldIDs, id)
	}
	sort.Slice(oldIDs, func(i, j int) bool {
		return oldIDs[i].String() < oldIDs[j].String()
	})

	for _, old := range oldIDs {
		newID := remap[old]
		if err := repointStatusFKs(tx, old, newID); err != nil {
			return err
		}
	}
	// 付け替えで (workflow, from, to) が重複しうるため、削除前に遷移行を畳む
	if err := dedupeWorkflowTransitions(tx); err != nil {
		return err
	}
	for _, old := range oldIDs {
		if err := tx.Unscoped().Delete(&model.Status{}, "id = ?", old).Error; err != nil {
			return fmt.Errorf("status dedupe: delete duplicate status %s: %w", old, err)
		}
	}

	return nil
}

func pickCanonicalStatus(sts []model.Status) uuid.UUID {
	if len(sts) == 1 {
		return sts[0].ID
	}
	var keyed []model.Status
	for _, s := range sts {
		if s.StatusKey == "sts_start" || s.StatusKey == "sts_goal" {
			keyed = append(keyed, s)
		}
	}
	if len(keyed) == 1 {
		return keyed[0].ID
	}
	sort.Slice(sts, func(i, j int) bool { return sts[i].ID.String() < sts[j].ID.String() })
	return sts[0].ID
}

func repointStatusFKs(tx *gorm.DB, old, newID uuid.UUID) error {
	if err := tx.Model(&model.Issue{}).Where("status_id = ?", old).Update("status_id", newID).Error; err != nil {
		return fmt.Errorf("repoint issues.status_id: %w", err)
	}
	if err := tx.Model(&model.WorkflowTransition{}).Where("from_status_id = ?", old).Update("from_status_id", newID).Error; err != nil {
		return fmt.Errorf("repoint workflow_transitions.from_status_id: %w", err)
	}
	if err := tx.Model(&model.WorkflowTransition{}).Where("to_status_id = ?", old).Update("to_status_id", newID).Error; err != nil {
		return fmt.Errorf("repoint workflow_transitions.to_status_id: %w", err)
	}
	if err := tx.Model(&model.IssueEvent{}).Where("from_status_id = ?", old).Update("from_status_id", newID).Error; err != nil {
		return fmt.Errorf("repoint issue_events.from_status_id: %w", err)
	}
	if err := tx.Model(&model.IssueEvent{}).Where("to_status_id = ?", old).Update("to_status_id", newID).Error; err != nil {
		return fmt.Errorf("repoint issue_events.to_status_id: %w", err)
	}
	if err := tx.Model(&model.TransitionAlertRule{}).Where("from_status_id = ?", old).Update("from_status_id", newID).Error; err != nil {
		return fmt.Errorf("repoint transition_alert_rules.from_status_id: %w", err)
	}
	if err := tx.Model(&model.TransitionAlertRule{}).Where("to_status_id = ?", old).Update("to_status_id", newID).Error; err != nil {
		return fmt.Errorf("repoint transition_alert_rules.to_status_id: %w", err)
	}
	return repointLegacyWorkflowStepsStatusID(tx, old, newID)
}

// レガシー DB に workflow_steps が残っている場合（GORM モデルからは外れているが FK が残る）
func repointLegacyWorkflowStepsStatusID(tx *gorm.DB, old, newID uuid.UUID) error {
	if err := tx.Exec(`UPDATE workflow_steps SET status_id = ? WHERE status_id = ?`, newID, old).Error; err != nil {
		if !isLegacyWorkflowStepsMissingErr(err) {
			return fmt.Errorf("repoint workflow_steps.status_id: %w", err)
		}
	}
	if err := tx.Exec(`UPDATE workflow_steps SET next_status_id = ? WHERE next_status_id = ?`, newID, old).Error; err != nil {
		if !isLegacyWorkflowStepsMissingErr(err) {
			return fmt.Errorf("repoint workflow_steps.next_status_id: %w", err)
		}
	}
	return nil
}

func isLegacyWorkflowStepsMissingErr(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "no such table") ||
		strings.Contains(msg, "UndefinedTable") ||
		strings.Contains(msg, "no such column") ||
		(strings.Contains(msg, "does not exist") && (strings.Contains(msg, "workflow_steps") || strings.Contains(msg, "column")))
}

// dedupeWorkflowTransitions は起動時マイグレーション専用。重複行の物理削除（業務 API の論理削除とは別）。
func dedupeWorkflowTransitions(tx *gorm.DB) error {
	dialect := tx.Dialector.Name()
	switch dialect {
	case "postgres":
		return tx.Exec(`
			DELETE FROM workflow_transitions wt
			USING workflow_transitions wt2
			WHERE wt.workflow_id = wt2.workflow_id
			  AND wt.from_status_id = wt2.from_status_id
			  AND wt.to_status_id = wt2.to_status_id
			  AND wt.id > wt2.id
		`).Error
	case "sqlite":
		return tx.Exec(`
			DELETE FROM workflow_transitions
			WHERE id NOT IN (
				SELECT MIN(id) FROM workflow_transitions GROUP BY workflow_id, from_status_id, to_status_id
			)
		`).Error
	default:
		return fmt.Errorf("dedupe workflow_transitions: unsupported dialect %q", dialect)
	}
}

func ensureStatusPartialLookupIndex(tx *gorm.DB) error {
	dialect := tx.Dialector.Name()
	var stmt string
	switch dialect {
	case "postgres", "sqlite":
		stmt = fmt.Sprintf(`
			CREATE INDEX IF NOT EXISTS %s
			ON statuses (workflow_id, name, display_order)
			WHERE deleted_at IS NULL
		`, statusWorkflowNameOrderLookupIndex)
	default:
		return fmt.Errorf("status lookup index: unsupported dialect %q", dialect)
	}
	if err := tx.Exec(stmt).Error; err != nil {
		return fmt.Errorf("create partial index on statuses: %w", err)
	}
	return nil
}
