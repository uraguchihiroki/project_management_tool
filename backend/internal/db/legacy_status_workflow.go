package db

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// PrepareStatusesWorkflowColumn は、レガシー DB（statuses に workflow_id が無い／NULL のまま）を
// AutoMigrate が NOT NULL 制約を付ける前に整合させる。
// PostgreSQL 想定（docker-compose の pmt_db）。
func PrepareStatusesWorkflowColumn(db *gorm.DB) error {
	var tableExists int64
	if err := db.Raw(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'statuses'
	`).Scan(&tableExists).Error; err != nil {
		return err
	}
	if tableExists == 0 {
		return nil
	}

	var hasWorkflowID int64
	if err := db.Raw(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'statuses' AND column_name = 'workflow_id'
	`).Scan(&hasWorkflowID).Error; err != nil {
		return err
	}
	if hasWorkflowID == 0 {
		if err := db.Exec(`ALTER TABLE statuses ADD COLUMN workflow_id bigint`).Error; err != nil {
			return fmt.Errorf("legacy migrate: add statuses.workflow_id: %w", err)
		}
		log.Println("legacy migrate: added nullable statuses.workflow_id")
	}

	// organization_id がある行 → 同一組織の先頭ワークフローへ
	if columnExists(db, "statuses", "organization_id") {
		if err := db.Exec(`
			UPDATE statuses s
			SET workflow_id = w.id
			FROM (
				SELECT DISTINCT ON (organization_id) id, organization_id
				FROM workflows
				WHERE deleted_at IS NULL
				ORDER BY organization_id, display_order ASC, id ASC
			) w
			WHERE s.workflow_id IS NULL
			  AND s.organization_id IS NOT NULL
			  AND s.organization_id = w.organization_id
		`).Error; err != nil {
			return fmt.Errorf("legacy migrate: backfill workflow_id from organization_id: %w", err)
		}
	}

	// project_id + projects.default_workflow_id
	if columnExists(db, "statuses", "project_id") {
		if err := db.Exec(`
			UPDATE statuses s
			SET workflow_id = p.default_workflow_id
			FROM projects p
			WHERE s.workflow_id IS NULL
			  AND s.project_id IS NOT NULL
			  AND s.project_id = p.id
			  AND p.default_workflow_id IS NOT NULL
		`).Error; err != nil {
			return fmt.Errorf("legacy migrate: backfill workflow_id from project: %w", err)
		}
		// default_workflow_id がまだ無いプロジェクト → 同一組織の先頭ワークフローへ
		if err := db.Exec(`
			UPDATE statuses s
			SET workflow_id = x.wfid
			FROM (
				SELECT DISTINCT ON (p.id) p.id AS pid, wf.id AS wfid
				FROM projects p
				JOIN workflows wf ON wf.organization_id = p.organization_id AND wf.deleted_at IS NULL
				ORDER BY p.id, wf.display_order ASC, wf.id ASC
			) x
			WHERE s.workflow_id IS NULL
			  AND s.project_id = x.pid
		`).Error; err != nil {
			return fmt.Errorf("legacy migrate: backfill workflow_id from project org workflow: %w", err)
		}
	}

	// Issue 経由でプロジェクト→デフォルトWF（1 Issue あたり先頭の project を採用）
	if columnExists(db, "issues", "status_id") && columnExists(db, "issues", "project_id") {
		if err := db.Exec(`
			UPDATE statuses s
			SET workflow_id = p.default_workflow_id
			FROM issues i
			JOIN projects p ON p.id = i.project_id
			WHERE s.workflow_id IS NULL
			  AND i.status_id = s.id
			  AND p.default_workflow_id IS NOT NULL
		`).Error; err != nil {
			return fmt.Errorf("legacy migrate: backfill workflow_id from issues: %w", err)
		}
		// organization_id が NULL の行（旧 sts_start 等）も Issue から辿れるなら紐づける
		if err := db.Exec(`
			UPDATE statuses s
			SET workflow_id = sub.wfid
			FROM (
				SELECT DISTINCT ON (i.status_id) i.status_id AS sid, p.default_workflow_id AS wfid
				FROM issues i
				JOIN projects p ON p.id = i.project_id
				WHERE p.default_workflow_id IS NOT NULL
				ORDER BY i.status_id, i.id
			) sub
			WHERE s.id = sub.sid AND s.workflow_id IS NULL
		`).Error; err != nil {
			return fmt.Errorf("legacy migrate: backfill workflow_id from issues (distinct): %w", err)
		}
	}

	// まだ NULL の組織ごとに最低1件のワークフローを作って紐づける
	if columnExists(db, "statuses", "organization_id") {
		if err := db.Exec(`
			INSERT INTO workflows (key, organization_id, name, description, display_order, created_at)
			SELECT 'wf-legacy-' || o.id::text, o.id, 'Legacy migration workflow', '', 1, NOW()
			FROM organizations o
			WHERE EXISTS (
				SELECT 1 FROM statuses s
				WHERE s.workflow_id IS NULL
				  AND s.organization_id = o.id
				  AND s.deleted_at IS NULL
			)
			AND NOT EXISTS (
				SELECT 1 FROM workflows w
				WHERE w.organization_id = o.id AND w.deleted_at IS NULL
			)
		`).Error; err != nil {
			return fmt.Errorf("legacy migrate: seed workflows for orphan statuses: %w", err)
		}
		if err := db.Exec(`
			UPDATE statuses s
			SET workflow_id = w.id
			FROM (
				SELECT DISTINCT ON (organization_id) id, organization_id
				FROM workflows
				WHERE deleted_at IS NULL
				ORDER BY organization_id, display_order ASC, id ASC
			) w
			WHERE s.workflow_id IS NULL
			  AND s.organization_id IS NOT NULL
			  AND s.organization_id = w.organization_id
		`).Error; err != nil {
			return fmt.Errorf("legacy migrate: second backfill from organization_id: %w", err)
		}
	}

	// NOTE: workflow_steps は承認ステップ系として廃止。承認テーブルは起動時に DROP されるため、ここでは参照しない。

	var remaining int64
	if err := db.Raw(`
		SELECT COUNT(*) FROM statuses
		WHERE workflow_id IS NULL AND deleted_at IS NULL
	`).Scan(&remaining).Error; err != nil {
		return err
	}
	if remaining == 0 {
		return nil
	}

	// まだ残るのは organization_id も Issue も無い孤立行が多い。開発用単一テナントでは先頭の WF に寄せて起動を優先する。
	var wfCount int64
	if err := db.Raw(`
		SELECT COUNT(DISTINCT organization_id) FROM workflows WHERE deleted_at IS NULL
	`).Scan(&wfCount).Error; err != nil {
		return err
	}
	var minWF uint
	if err := db.Raw(`
		SELECT id FROM workflows WHERE deleted_at IS NULL ORDER BY id ASC LIMIT 1
	`).Scan(&minWF).Error; err != nil || minWF == 0 {
		return fmt.Errorf("legacy migrate: %d statuses still have NULL workflow_id and no workflow row to attach", remaining)
	}
	if wfCount > 1 {
		return fmt.Errorf(
			"legacy migrate: %d statuses still have NULL workflow_id (multi-tenant DB); assign workflow_id manually or fix organization_id",
			remaining,
		)
	}
	log.Printf("legacy migrate: assigning %d orphan statuses to workflow_id=%d (single-tenant fallback)", remaining, minWF)
	if err := db.Exec(`UPDATE statuses SET workflow_id = ? WHERE workflow_id IS NULL AND deleted_at IS NULL`, minWF).Error; err != nil {
		return fmt.Errorf("legacy migrate: single-tenant fallback: %w", err)
	}
	return nil
}

func columnExists(db *gorm.DB, table, col string) bool {
	var n int64
	_ = db.Raw(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = ? AND column_name = ?
	`, table, col).Scan(&n)
	return n > 0
}
