package repository

import (
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
	FindByOrgNameType(orgID uuid.UUID, projectID *uuid.UUID, name, statusType string) (*model.Status, error)
	FindByID(id uuid.UUID) (*model.Status, error)
	FindByStatusKeyInOrg(orgID uuid.UUID, key string) (*model.Status, error)
	Create(status *model.Status) error
	Update(status *model.Status) error
	Delete(id uuid.UUID) error
	CountInUse(id uuid.UUID) (int64, error)
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
	err := r.db.Where("workflow_id = ?", workflowID).Order(`"order" asc`).Find(&statuses).Error
	return statuses, err
}

func (r *statusRepository) FindByOrganizationID(orgID uuid.UUID) ([]model.Status, error) {
	return r.FindByOrganizationIDAndType(orgID, "")
}

func (r *statusRepository) FindByOrganizationIDAndType(orgID uuid.UUID, statusType string) ([]model.Status, error) {
	var statuses []model.Status
	q := r.db.Joins("JOIN workflows ON workflows.id = statuses.workflow_id").
		Where("workflows.organization_id = ?", orgID)
	if statusType == "issue" || statusType == "project" {
		q = q.Where("statuses.type = ?", statusType)
	}
	err := q.Order(`statuses."order" asc`).Find(&statuses).Error
	return statuses, err
}

func (r *statusRepository) FindByOrganizationIDAndTypeExcludeSystem(orgID uuid.UUID, statusType string) ([]model.Status, error) {
	var statuses []model.Status
	q := r.db.Joins("JOIN workflows ON workflows.id = statuses.workflow_id").
		Where("workflows.organization_id = ?", orgID).
		Where("COALESCE(statuses.status_key, '') NOT IN ('sts_start','sts_goal')")
	if statusType == "issue" || statusType == "project" {
		q = q.Where("statuses.type = ?", statusType)
	}
	err := q.Order(`statuses."order" asc`).Find(&statuses).Error
	return statuses, err
}

func (r *statusRepository) FindByOrgNameType(orgID uuid.UUID, projectID *uuid.UUID, name, statusType string) (*model.Status, error) {
	var status model.Status
	q := r.db.Joins("JOIN workflows ON workflows.id = statuses.workflow_id").
		Where("workflows.organization_id = ? AND statuses.name = ? AND statuses.type = ?", orgID, name, statusType)
	if projectID != nil {
		q = q.Joins("JOIN projects ON projects.default_workflow_id = statuses.workflow_id").
			Where("projects.id = ?", *projectID)
	} else {
		// 組織直下の Issue / Project 用ステータスは専用ワークフロー名で区別
		if statusType == "issue" {
			q = q.Where("workflows.name = ?", "組織Issue")
		} else if statusType == "project" {
			q = q.Where("workflows.name = ?", "組織Project")
		}
	}
	err := q.First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
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
