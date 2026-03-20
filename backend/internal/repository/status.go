package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type StatusRepository interface {
	FindByProject(projectID uuid.UUID) ([]model.Status, error)
	FindByOrganizationID(orgID uuid.UUID) ([]model.Status, error)
	FindByOrganizationIDAndType(orgID uuid.UUID, statusType string) ([]model.Status, error)
	FindByOrgNameType(orgID uuid.UUID, projectID *uuid.UUID, name, statusType string) (*model.Status, error)
	FindByID(id uuid.UUID) (*model.Status, error)
	FindByStatusKey(key string) (*model.Status, error)
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
	var statuses []model.Status
	err := r.db.Where("project_id = ?", projectID).Order(`"order" asc`).Find(&statuses).Error
	return statuses, err
}

func (r *statusRepository) FindByOrganizationID(orgID uuid.UUID) ([]model.Status, error) {
	return r.FindByOrganizationIDAndType(orgID, "")
}

func (r *statusRepository) FindByOrganizationIDAndType(orgID uuid.UUID, statusType string) ([]model.Status, error) {
	var statuses []model.Status
	q := r.db.Where(
		"project_id IN (SELECT id FROM projects WHERE organization_id = ?) OR organization_id = ? OR status_key IN ('sts_start','sts_goal')",
		orgID, orgID,
	)
	if statusType == "issue" || statusType == "project" {
		q = q.Where("type = ?", statusType)
	}
	err := q.Order(`"order" asc`).Find(&statuses).Error
	return statuses, err
}

func (r *statusRepository) FindByOrgNameType(orgID uuid.UUID, projectID *uuid.UUID, name, statusType string) (*model.Status, error) {
	var status model.Status
	q := r.db.Where("name = ? AND type = ?", name, statusType)
	if projectID != nil {
		q = q.Where("project_id = ?", projectID)
	} else {
		q = q.Where("organization_id = ? AND project_id IS NULL", orgID)
	}
	err := q.First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *statusRepository) FindByID(id uuid.UUID) (*model.Status, error) {
	var status model.Status
	err := r.db.First(&status, "id = ?", id).Error
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
	var issueCount, stepCount int64
	if err := r.db.Model(&model.Issue{}).Where("status_id = ?", id).Count(&issueCount).Error; err != nil {
		return 0, err
	}
	if err := r.db.Model(&model.WorkflowStep{}).Where("status_id = ? OR next_status_id = ?", id, id).Count(&stepCount).Error; err != nil {
		return 0, err
	}
	return issueCount + stepCount, nil
}
