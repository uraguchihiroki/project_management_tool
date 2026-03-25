package db

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const statusOneEntryPerWorkflowIndex = "idx_statuses_one_entry_per_workflow_active"

var legacySystemStatusIDs = []uuid.UUID{
	uuid.MustParse("30000000-0000-0000-0000-000000000001"),
	uuid.MustParse("30000000-0000-0000-0000-000000000002"),
}

// MigrateRemoveLegacyGlobalIssueStatuses は sts_start/sts_goal 相当の行・それを端点とする遷移を除去し、
// Issue の status_id が参照している場合は同一組織の任意の有効ステータスへ付け替える。冪等。
func MigrateRemoveLegacyGlobalIssueStatuses(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			DELETE FROM workflow_transitions
			WHERE from_status_id IN (?, ?) OR to_status_id IN (?, ?)
		`, legacySystemStatusIDs[0], legacySystemStatusIDs[1],
			legacySystemStatusIDs[0], legacySystemStatusIDs[1]).Error; err != nil {
			return fmt.Errorf("delete workflow_transitions legacy endpoints: %w", err)
		}

		dialect := tx.Dialector.Name()
		switch dialect {
		case "postgres":
			if err := tx.Exec(`
				UPDATE issues i
				SET status_id = sub.new_id
				FROM (
					SELECT i2.id AS issue_id,
						(SELECT s.id FROM statuses s
						 INNER JOIN workflows w ON w.id = s.workflow_id
						 WHERE w.organization_id = i2.organization_id
						   AND s.deleted_at IS NULL
						 ORDER BY s.display_order ASC, s.id ASC
						 LIMIT 1) AS new_id
					FROM issues i2
					WHERE i2.status_id IN (?, ?)
				) sub
				WHERE i.id = sub.issue_id AND sub.new_id IS NOT NULL
			`, legacySystemStatusIDs[0], legacySystemStatusIDs[1]).Error; err != nil {
				return fmt.Errorf("repoint issues from legacy status: %w", err)
			}
		case "sqlite":
			if err := tx.Exec(`
				UPDATE issues
				SET status_id = (
					SELECT s.id FROM statuses s
					INNER JOIN workflows w ON w.id = s.workflow_id
					WHERE w.organization_id = issues.organization_id
					  AND s.deleted_at IS NULL
					ORDER BY s.display_order ASC, s.id ASC
					LIMIT 1
				)
				WHERE status_id IN (?, ?)
				  AND EXISTS (
					SELECT 1 FROM statuses s
					INNER JOIN workflows w ON w.id = s.workflow_id
					WHERE w.organization_id = issues.organization_id AND s.deleted_at IS NULL
				  )
			`, legacySystemStatusIDs[0], legacySystemStatusIDs[1]).Error; err != nil {
				return fmt.Errorf("repoint issues from legacy status: %w", err)
			}
		default:
			return fmt.Errorf("repoint issues: unsupported dialect %q", dialect)
		}

		for _, col := range []string{"from_status_id", "to_status_id"} {
			q := fmt.Sprintf(`
				UPDATE issue_events SET %s = NULL WHERE %s IN (?, ?)
			`, col, col)
			if err := tx.Exec(q, legacySystemStatusIDs[0], legacySystemStatusIDs[1]).Error; err != nil {
				return fmt.Errorf("clear issue_events.%s legacy refs: %w", col, err)
			}
		}
		if err := tx.Exec(`
			DELETE FROM transition_alert_rules
			WHERE from_status_id IN (?, ?) OR to_status_id IN (?, ?)
		`, legacySystemStatusIDs[0], legacySystemStatusIDs[1],
			legacySystemStatusIDs[0], legacySystemStatusIDs[1]).Error; err != nil {
			return fmt.Errorf("delete alert rules referencing legacy statuses: %w", err)
		}

		if err := tx.Exec(`
			DELETE FROM statuses WHERE id IN (?, ?)
			   OR status_key IN ('sts_start', 'sts_goal')
		`, legacySystemStatusIDs[0], legacySystemStatusIDs[1]).Error; err != nil {
			return fmt.Errorf("delete legacy system statuses: %w", err)
		}
		return nil
	})
}

// MigrateStatusEntryUniqueIndex は is_entry 一意（論理削除除外）の部分インデックスを作成する。冪等。
func MigrateStatusEntryUniqueIndex(db *gorm.DB) error {
	dialect := db.Dialector.Name()
	var stmt string
	switch dialect {
	case "postgres":
		stmt = fmt.Sprintf(`
			CREATE UNIQUE INDEX IF NOT EXISTS %s
			ON statuses (workflow_id)
			WHERE is_entry = true AND deleted_at IS NULL
		`, statusOneEntryPerWorkflowIndex)
	case "sqlite":
		stmt = fmt.Sprintf(`
			CREATE UNIQUE INDEX IF NOT EXISTS %s
			ON statuses (workflow_id)
			WHERE is_entry IS 1 AND deleted_at IS NULL
		`, statusOneEntryPerWorkflowIndex)
	default:
		return fmt.Errorf("status entry unique index: unsupported dialect %q", dialect)
	}
	if err := db.Exec(stmt).Error; err != nil {
		return fmt.Errorf("create partial unique index on statuses.is_entry: %w", err)
	}
	return nil
}
