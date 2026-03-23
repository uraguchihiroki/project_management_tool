package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type WorkflowRepository interface {
	FindAll() ([]model.Workflow, error)
	FindByID(id uint) (*model.Workflow, error)
	FindByOrgAndName(orgID uuid.UUID, name string) (*model.Workflow, error)
	Create(workflow *model.Workflow) error
	Update(workflow *model.Workflow) error
	Delete(id uint) error
	Reorder(ids []uint) error
	GetMaxOrder() (int, error)
}

type workflowRepository struct {
	db *gorm.DB
}

func NewWorkflowRepository(db *gorm.DB) WorkflowRepository {
	return &workflowRepository{db: db}
}

func (r *workflowRepository) FindAll() ([]model.Workflow, error) {
	var workflows []model.Workflow
	err := r.db.Order("workflows.display_order ASC").Find(&workflows).Error
	return workflows, err
}

func (r *workflowRepository) FindByID(id uint) (*model.Workflow, error) {
	var workflow model.Workflow
	err := r.db.First(&workflow, id).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *workflowRepository) FindByOrgAndName(orgID uuid.UUID, name string) (*model.Workflow, error) {
	var workflow model.Workflow
	err := r.db.Where("organization_id = ? AND name = ?", orgID, name).First(&workflow).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *workflowRepository) Create(workflow *model.Workflow) error {
	return r.db.Create(workflow).Error
}

func (r *workflowRepository) Update(workflow *model.Workflow) error {
	updates := map[string]interface{}{
		"name":        workflow.Name,
		"description": workflow.Description,
	}
	if workflow.Key != "" {
		updates["key"] = workflow.Key
	}
	return r.db.Model(&model.Workflow{}).Where("id = ?", workflow.ID).Updates(updates).Error
}

func (r *workflowRepository) Delete(id uint) error {
	var sids []uuid.UUID
	_ = r.db.Model(&model.Status{}).Where("workflow_id = ?", id).Pluck("id", &sids)
	if len(sids) > 0 {
		_ = r.db.Where("workflow_id = ?", id).Delete(&model.WorkflowTransition{}).Error
	}
	if err := r.db.Where("workflow_id = ?", id).Delete(&model.Status{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&model.Workflow{}, id).Error
}

func (r *workflowRepository) Reorder(ids []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.Workflow{}).Where("id = ?", id).Update("display_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *workflowRepository) GetMaxOrder() (int, error) {
	var maxOrder int
	err := r.db.Model(&model.Workflow{}).Select("COALESCE(MAX(display_order), 0)").Scan(&maxOrder).Error
	return maxOrder, err
}
