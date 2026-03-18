package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type StatusRepository interface {
	FindByProject(projectID uuid.UUID) ([]model.Status, error)
	FindByOrganizationID(orgID uuid.UUID) ([]model.Status, error)
	FindByID(id uuid.UUID) (*model.Status, error)
	Create(status *model.Status) error
	Update(status *model.Status) error
	Delete(id uuid.UUID) error
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
	var statuses []model.Status
	// プロジェクト所属 + 組織直下のステータス
	err := r.db.Where(
		"project_id IN (SELECT id FROM projects WHERE organization_id = ?) OR organization_id = ?",
		orgID, orgID,
	).Order(`"order" asc`).Find(&statuses).Error
	return statuses, err
}

func (r *statusRepository) FindByID(id uuid.UUID) (*model.Status, error) {
	var status model.Status
	err := r.db.First(&status, "id = ?", id).Error
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
