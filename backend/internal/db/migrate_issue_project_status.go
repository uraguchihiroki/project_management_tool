package db

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

// MigrateIssueProjectStatusSplitPre は AutoMigrate の前に実行する。
// 旧「組織Project」ワークフローと statuses.type 列を除去し、Issue 専用 statuses に揃える。
// 冪等。
func MigrateIssueProjectStatusSplitPre(db *gorm.DB) error {
	if !tableExists(db, "workflows") {
		return nil
	}
	dialect := db.Dialector.Name()

	// 旧ユニークインデックス（type 含む）を先に落とす（status_integrity の新インデックス名とは別）
	if err := dropStatusUniqueIndexIfExists(db, dialect, "idx_statuses_wf_name_type_order_active"); err != nil {
		return err
	}

	// 組織スコープのプロジェクト進行ワークフロー（Workflow は Issue のみにする）
	// 以下の DELETE はレガシー移行専用（本番データの論理削除 API ではない）
	if err := db.Exec(`
		DELETE FROM workflow_transitions
		WHERE workflow_id IN (SELECT id FROM workflows WHERE name = ?)
	`, "組織Project").Error; err != nil {
		return fmt.Errorf("migrate: delete org project workflow transitions: %w", err)
	}
	if err := db.Exec(`
		DELETE FROM statuses
		WHERE workflow_id IN (SELECT id FROM workflows WHERE name = ?)
	`, "組織Project").Error; err != nil {
		return fmt.Errorf("migrate: delete org project statuses: %w", err)
	}
	if err := db.Exec(`DELETE FROM workflows WHERE name = ?`, "組織Project").Error; err != nil {
		return fmt.Errorf("migrate: delete org project workflow: %w", err)
	}

	if columnExistsStatuses(db, "type") {
		if dialect == "postgres" {
			if err := db.Exec(`ALTER TABLE statuses DROP COLUMN IF EXISTS type`).Error; err != nil {
				return fmt.Errorf("migrate: drop statuses.type: %w", err)
			}
		} else if dialect == "sqlite" {
			// SQLite 3.35+ DROP COLUMN
			if err := db.Exec(`ALTER TABLE statuses DROP COLUMN type`).Error; err != nil {
				return fmt.Errorf("migrate: drop statuses.type (sqlite): %w", err)
			}
		} else {
			return fmt.Errorf("migrate: drop statuses.type: unsupported dialect %q", dialect)
		}
		log.Println("migrate: dropped statuses.type")
	}
	return nil
}

func tableExists(db *gorm.DB, table string) bool {
	dialect := db.Dialector.Name()
	if dialect == "sqlite" {
		var n int64
		_ = db.Raw(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&n)
		return n > 0
	}
	var n int64
	_ = db.Raw(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = ?
	`, table).Scan(&n)
	return n > 0
}

func columnExistsStatuses(db *gorm.DB, col string) bool {
	dialect := db.Dialector.Name()
	if dialect == "sqlite" {
		var n int64
		_ = db.Raw(`SELECT COUNT(*) FROM pragma_table_info('statuses') WHERE name = ?`, col).Scan(&n)
		return n > 0
	}
	var n int64
	_ = db.Raw(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'statuses' AND column_name = ?
	`, col).Scan(&n)
	return n > 0
}

func dropStatusUniqueIndexIfExists(db *gorm.DB, dialect, indexName string) error {
	if dialect == "postgres" || dialect == "sqlite" {
		return db.Exec(fmt.Sprintf(`DROP INDEX IF EXISTS %s`, indexName)).Error
	}
	return fmt.Errorf("drop index: unsupported dialect %q", dialect)
}

// MigrateProjectStatusSeed は AutoMigrate 後、project_status_id が無いプロジェクトに
// デフォルトの project_statuses のみ投入する（許可遷移は作らない）。冪等。
func MigrateProjectStatusSeed(db *gorm.DB) error {
	if !tableExists(db, "projects") || !tableExists(db, "project_statuses") {
		return nil
	}
	var projects []model.Project
	if err := db.Where("project_status_id IS NULL").Find(&projects).Error; err != nil {
		return fmt.Errorf("migrate seed: list projects: %w", err)
	}
	for i := range projects {
		p := &projects[i]
		if err := seedOneProjectStatuses(db, p.ID); err != nil {
			return err
		}
	}
	return nil
}

func seedOneProjectStatuses(db *gorm.DB, projectID uuid.UUID) error {
	defaults := []struct {
		Name  string
		Color string
		Order int
	}{
		{"計画中", "#6B7280", 1},
		{"進行中", "#3B82F6", 2},
		{"完了", "#10B981", 3},
	}
	var first uuid.UUID
	for i, d := range defaults {
		sid := uuid.New()
		ps := &model.ProjectStatus{
			ID:        sid,
			Key:       "pst-" + sid.String(),
			ProjectID: projectID,
			Name:      d.Name,
			Color:     d.Color,
			Order:     d.Order,
		}
		if err := db.Create(ps).Error; err != nil {
			return fmt.Errorf("migrate seed: create project_status: %w", err)
		}
		if i == 0 {
			first = sid
		}
	}
	if err := db.Model(&model.Project{}).Where("id = ?", projectID).Update("project_status_id", first).Error; err != nil {
		return fmt.Errorf("migrate seed: set project.project_status_id: %w", err)
	}
	return nil
}
