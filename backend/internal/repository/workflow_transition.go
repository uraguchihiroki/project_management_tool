package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/pkg/keygen"
	"gorm.io/gorm"
)

type WorkflowTransitionRepository interface {
	DeleteByWorkflowID(workflowID uint) error
	SeedAllPairs(workflowID uint, statusIDs []uuid.UUID) error
	Exists(workflowID uint, fromID, toID uuid.UUID) bool
}

type workflowTransitionRepository struct {
	db *gorm.DB
}

func NewWorkflowTransitionRepository(db *gorm.DB) WorkflowTransitionRepository {
	return &workflowTransitionRepository{db: db}
}

func (r *workflowTransitionRepository) DeleteByWorkflowID(workflowID uint) error {
	return r.db.Where("workflow_id = ?", workflowID).Delete(&model.WorkflowTransition{}).Error
}

// SeedAllPairs は同一ワークフロー内の任意遷移を許可（全ペア）
func (r *workflowTransitionRepository) SeedAllPairs(workflowID uint, statusIDs []uuid.UUID) error {
	if len(statusIDs) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workflow_id = ?", workflowID).Delete(&model.WorkflowTransition{}).Error; err != nil {
			return err
		}
		for _, from := range statusIDs {
			for _, to := range statusIDs {
				wt := &model.WorkflowTransition{
					Key:          keygen.UUIDKey(uuid.New()),
					WorkflowID:   workflowID,
					FromStatusID: from,
					ToStatusID:   to,
					CreatedAt:    time.Now(),
				}
				if err := tx.Create(wt).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *workflowTransitionRepository) Exists(workflowID uint, fromID, toID uuid.UUID) bool {
	var wt model.WorkflowTransition
	err := r.db.Where("workflow_id = ? AND from_status_id = ? AND to_status_id = ?", workflowID, fromID, toID).
		First(&wt).Error
	return err == nil
}
