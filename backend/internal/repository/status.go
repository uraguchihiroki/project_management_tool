package repository

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type StatusRepository interface {
	FindByProject(projectID uuid.UUID) ([]model.Status, error)
	FindByWorkflowID(workflowID uint) ([]model.Status, error)
	FindByOrganizationID(orgID uuid.UUID) ([]model.Status, error)
	FindByOrganizationIDAndType(orgID uuid.UUID, statusType string) ([]model.Status, error)
	FindByOrganizationIDAndTypeExcludeSystem(orgID uuid.UUID, statusType string) ([]model.Status, error)
	FindByID(id uuid.UUID) (*model.Status, error)
	FindByStatusKeyInOrg(orgID uuid.UUID, key string) (*model.Status, error)
	Create(status *model.Status) error
	Update(status *model.Status) error
	// PersistWithEntryExclusive は st を保存し、st.IsEntry が true のとき同一 workflow の他行の is_entry を false にする（単一トランザクション）。
	PersistWithEntryExclusive(st *model.Status) error
	Delete(id uuid.UUID) error
	CountInUse(id uuid.UUID) (int64, error)
	CountByWorkflowID(workflowID uint) (int64, error)
	ReorderWorkflow(workflowID uint, statusIDs []uuid.UUID) error
}

type statusRepository struct {
	db *gorm.DB
}

func NewStatusRepository(db *gorm.DB) StatusRepository {
	return &statusRepository{db: db}
}

func (r *statusRepository) FindByProject(projectID uuid.UUID) ([]model.Status, error) {
	var p model.Project
	if err := r.db.First(&p, "id = ?", projectID).Error; err != nil {
		return nil, err
	}
	if p.DefaultWorkflowID == nil {
		return []model.Status{}, nil
	}
	return r.FindByWorkflowID(*p.DefaultWorkflowID)
}

func (r *statusRepository) FindByWorkflowID(workflowID uint) ([]model.Status, error) {
	var statuses []model.Status
	err := r.db.Where("workflow_id = ?", workflowID).Order("display_order ASC, id ASC").Find(&statuses).Error
	return statuses, err
}

func (r *statusRepository) FindByOrganizationID(orgID uuid.UUID) ([]model.Status, error) {
	return r.FindByOrganizationIDAndType(orgID, "")
}

// FindByOrganizationIDAndType は組織に属する全ワークフローの Issue 用 statuses を返す。statusType は互換のため残す（"project" は空配列を返す側で処理）。
func (r *statusRepository) FindByOrganizationIDAndType(orgID uuid.UUID, statusType string) ([]model.Status, error) {
	var statuses []model.Status
	q := r.db.Joins("JOIN workflows ON workflows.id = statuses.workflow_id").
		Where("workflows.organization_id = ?", orgID)
	err := q.Order("statuses.display_order ASC, statuses.id ASC").Find(&statuses).Error
	return statuses, err
}

func (r *statusRepository) FindByOrganizationIDAndTypeExcludeSystem(orgID uuid.UUID, statusType string) ([]model.Status, error) {
	return r.FindByOrganizationIDAndType(orgID, statusType)
}

func (r *statusRepository) FindByID(id uuid.UUID) (*model.Status, error) {
	var status model.Status
	err := r.db.Preload("Workflow").First(&status, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *statusRepository) FindByStatusKeyInOrg(orgID uuid.UUID, key string) (*model.Status, error) {
	var status model.Status
	err := r.db.Joins("JOIN workflows ON workflows.id = statuses.workflow_id").
		Where("workflows.organization_id = ? AND statuses.status_key = ?", orgID, key).
		First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *statusRepository) FindByStatusKey(key string) (*model.Status, error) {
	var status model.Status
	err := r.db.First(&status, "status_key = ?", key).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *statusRepository) Create(status *model.Status) error {
	return r.db.Create(status).Error
}

func (r *statusRepository) Update(status *model.Status) error {
	return r.db.Save(status).Error
}

func (r *statusRepository) PersistWithEntryExclusive(st *model.Status) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if st.IsEntry {
			if err := tx.Model(&model.Status{}).
				Where("workflow_id = ? AND id <> ? AND deleted_at IS NULL", st.WorkflowID, st.ID).
				Update("is_entry", false).Error; err != nil {
				return err
			}
		}
		return tx.Save(st).Error
	})
}

func (r *statusRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Status{}, "id = ?", id).Error
}

func (r *statusRepository) CountInUse(id uuid.UUID) (int64, error) {
	var issueCount int64
	if err := r.db.Model(&model.Issue{}).Where("status_id = ?", id).Count(&issueCount).Error; err != nil {
		return 0, err
	}
	return issueCount, nil
}

func (r *statusRepository) CountByWorkflowID(workflowID uint) (int64, error) {
	var n int64
	if err := r.db.Model(&model.Status{}).Where("workflow_id = ?", workflowID).Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

func (r *statusRepository) ReorderWorkflow(workflowID uint, statusIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, sid := range statusIDs {
			res := tx.Model(&model.Status{}).Where("id = ? AND workflow_id = ?", sid, workflowID).Update("display_order", i+1)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected != 1 {
				return fmt.Errorf("invalid status id for reorder")
			}
		}
		return nil
	})
}
