package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/pkg/keygen"
	"gorm.io/gorm"
)

type ProjectStatusTransitionRepository interface {
	DeleteByProjectID(projectID uuid.UUID) error
	SeedAllPairs(projectID uuid.UUID, statusIDs []uuid.UUID) error
	Exists(projectID uuid.UUID, fromID, toID uuid.UUID) bool
}

type projectStatusTransitionRepository struct {
	db *gorm.DB
}

func NewProjectStatusTransitionRepository(db *gorm.DB) ProjectStatusTransitionRepository {
	return &projectStatusTransitionRepository{db: db}
}

func (r *projectStatusTransitionRepository) DeleteByProjectID(projectID uuid.UUID) error {
	return r.db.Unscoped().Where("project_id = ?", projectID).Delete(&model.ProjectStatusTransition{}).Error
}

// SeedAllPairs は同一プロジェクト内の任意遷移を許可（全ペア）
func (r *projectStatusTransitionRepository) SeedAllPairs(projectID uuid.UUID, statusIDs []uuid.UUID) error {
	if len(statusIDs) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 全ペア再生成: 論理削除のみだと同一 (project,from,to) の再挿入で重複しうるため Unscoped
		if err := tx.Unscoped().Where("project_id = ?", projectID).Delete(&model.ProjectStatusTransition{}).Error; err != nil {
			return err
		}
		for _, from := range statusIDs {
			for _, to := range statusIDs {
				pt := &model.ProjectStatusTransition{
					Key:                 keygen.UUIDKey(uuid.New()),
					ProjectID:           projectID,
					FromProjectStatusID: from,
					ToProjectStatusID:   to,
					CreatedAt:           time.Now(),
				}
				if err := tx.Create(pt).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *projectStatusTransitionRepository) Exists(projectID uuid.UUID, fromID, toID uuid.UUID) bool {
	var pt model.ProjectStatusTransition
	err := r.db.Where("project_id = ? AND from_project_status_id = ? AND to_project_status_id = ?", projectID, fromID, toID).
		First(&pt).Error
	return err == nil
}
