package repository

import (
	"database/sql"
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
	ExistsOtherThan(workflowID uint, fromID, toID uuid.UUID, exceptTransitionID uint) bool
	FindByWorkflowID(workflowID uint) ([]model.WorkflowTransition, error)
	Create(t *model.WorkflowTransition) error
	Update(fromStatusID, toStatusID uuid.UUID, id uint) error
	DeleteByID(id uint) error
	FindByID(id uint) (*model.WorkflowTransition, error)
	CountReferencingStatus(statusID uuid.UUID) (int64, error)
	ReorderWorkflow(workflowID uint, transitionIDs []uint) error
	MaxDisplayOrder(workflowID uint) (int, error)
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

// SeedAllPairs は同一ワークフロー内の各ステータス組み合わせに許可遷移を作成する（既存有効行はソフト削除のうえ再作成）
func (r *workflowTransitionRepository) SeedAllPairs(workflowID uint, statusIDs []uuid.UUID) error {
	if len(statusIDs) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workflow_id = ?", workflowID).Delete(&model.WorkflowTransition{}).Error; err != nil {
			return err
		}
		seq := 0
		for _, from := range statusIDs {
			for _, to := range statusIDs {
				seq++
				wt := &model.WorkflowTransition{
					Key:          keygen.UUIDKey(uuid.New()),
					WorkflowID:   workflowID,
					FromStatusID: from,
					ToStatusID:   to,
					DisplayOrder: seq,
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

func (r *workflowTransitionRepository) ExistsOtherThan(workflowID uint, fromID, toID uuid.UUID, exceptTransitionID uint) bool {
	var wt model.WorkflowTransition
	err := r.db.Where("workflow_id = ? AND from_status_id = ? AND to_status_id = ? AND id != ?", workflowID, fromID, toID, exceptTransitionID).
		First(&wt).Error
	return err == nil
}

func (r *workflowTransitionRepository) FindByWorkflowID(workflowID uint) ([]model.WorkflowTransition, error) {
	var rows []model.WorkflowTransition
	if err := r.db.Where("workflow_id = ?", workflowID).Order("display_order ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *workflowTransitionRepository) Create(t *model.WorkflowTransition) error {
	return r.db.Create(t).Error
}

func (r *workflowTransitionRepository) Update(fromStatusID, toStatusID uuid.UUID, id uint) error {
	return r.db.Model(&model.WorkflowTransition{}).Where("id = ?", id).Updates(map[string]interface{}{
		"from_status_id": fromStatusID,
		"to_status_id":   toStatusID,
	}).Error
}

func (r *workflowTransitionRepository) DeleteByID(id uint) error {
	return r.db.Delete(&model.WorkflowTransition{}, id).Error
}

func (r *workflowTransitionRepository) FindByID(id uint) (*model.WorkflowTransition, error) {
	var t model.WorkflowTransition
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *workflowTransitionRepository) CountReferencingStatus(statusID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Model(&model.WorkflowTransition{}).
		Where("from_status_id = ? OR to_status_id = ?", statusID, statusID).
		Count(&n).Error
	return n, err
}

// ReorderWorkflow は transition_ids の順に display_order を 1..n に振り直す（同一 WF 内のみ）
func (r *workflowTransitionRepository) ReorderWorkflow(workflowID uint, transitionIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, tid := range transitionIDs {
			if err := tx.Model(&model.WorkflowTransition{}).
				Where("id = ? AND workflow_id = ?", tid, workflowID).
				Update("display_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *workflowTransitionRepository) MaxDisplayOrder(workflowID uint) (int, error) {
	var max sql.NullInt64
	err := r.db.Model(&model.WorkflowTransition{}).Where("workflow_id = ?", workflowID).
		Select("COALESCE(MAX(display_order), 0)").Scan(&max).Error
	if err != nil {
		return 0, err
	}
	return int(max.Int64), nil
}
