package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type WorkflowRepository interface {
	FindAll() ([]model.Workflow, error)
	FindByOrganizationID(orgID uuid.UUID) ([]model.Workflow, error)
	FindByID(id uint) (*model.Workflow, error)
	Create(workflow *model.Workflow) error
	Update(workflow *model.Workflow) error
	Delete(id uint) error
	CreateStep(step *model.WorkflowStep) error
	UpdateStep(step *model.WorkflowStep) error
	DeleteStep(id uint) error
	FindStepByID(id uint) (*model.WorkflowStep, error)
	CountSteps(workflowID uint) (int64, error)
}

type workflowRepository struct {
	db *gorm.DB
}

func NewWorkflowRepository(db *gorm.DB) WorkflowRepository {
	return &workflowRepository{db: db}
}

func (r *workflowRepository) FindAll() ([]model.Workflow, error) {
	var workflows []model.Workflow
	err := r.db.Preload("Organization").Order("created_at DESC").Find(&workflows).Error
	return workflows, err
}

func (r *workflowRepository) FindByOrganizationID(orgID uuid.UUID) ([]model.Workflow, error) {
	var workflows []model.Workflow
	err := r.db.Where("organization_id = ?", orgID).Order("created_at DESC").Find(&workflows).Error
	return workflows, err
}

func (r *workflowRepository) FindByID(id uint) (*model.Workflow, error) {
	var workflow model.Workflow
	err := r.db.
		Preload("Organization").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC").Preload("Status")
		}).
		First(&workflow, id).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *workflowRepository) Create(workflow *model.Workflow) error {
	return r.db.Create(workflow).Error
}

func (r *workflowRepository) Update(workflow *model.Workflow) error {
	return r.db.Model(&model.Workflow{}).Where("id = ?", workflow.ID).Updates(map[string]interface{}{
		"name":        workflow.Name,
		"description": workflow.Description,
	}).Error
}

func (r *workflowRepository) Delete(id uint) error {
	// ステップを先に削除してからワークフローを削除
	if err := r.db.Where("workflow_id = ?", id).Delete(&model.WorkflowStep{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&model.Workflow{}, id).Error
}

func (r *workflowRepository) CreateStep(step *model.WorkflowStep) error {
	return r.db.Create(step).Error
}

func (r *workflowRepository) UpdateStep(step *model.WorkflowStep) error {
	return r.db.Model(&model.WorkflowStep{}).Where("id = ?", step.ID).Updates(map[string]interface{}{
		"name":             step.Name,
		"required_level":   step.RequiredLevel,
		"status_id":        step.StatusID,
		"order":            step.Order,
		"approver_type":    step.ApproverType,
		"approver_user_id": step.ApproverUserID,
		"min_approvers":    step.MinApprovers,
		"exclude_reporter": step.ExcludeReporter,
		"exclude_assignee": step.ExcludeAssignee,
	}).Error
}

func (r *workflowRepository) DeleteStep(id uint) error {
	return r.db.Delete(&model.WorkflowStep{}, id).Error
}

func (r *workflowRepository) FindStepByID(id uint) (*model.WorkflowStep, error) {
	var step model.WorkflowStep
	err := r.db.Preload("Status").First(&step, id).Error
	if err != nil {
		return nil, err
	}
	return &step, nil
}

func (r *workflowRepository) CountSteps(workflowID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.WorkflowStep{}).Where("workflow_id = ?", workflowID).Count(&count).Error
	return count, err
}
