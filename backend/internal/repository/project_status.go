package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type ProjectStatusRepository interface {
	Create(ps *model.ProjectStatus) error
	Update(ps *model.ProjectStatus) error
	FindByProjectID(projectID uuid.UUID) ([]model.ProjectStatus, error)
	FindByID(id uuid.UUID) (*model.ProjectStatus, error)
}

type projectStatusRepository struct {
	db *gorm.DB
}

func NewProjectStatusRepository(db *gorm.DB) ProjectStatusRepository {
	return &projectStatusRepository{db: db}
}

func (r *projectStatusRepository) Create(ps *model.ProjectStatus) error {
	return r.db.Create(ps).Error
}

func (r *projectStatusRepository) Update(ps *model.ProjectStatus) error {
	return r.db.Save(ps).Error
}

func (r *projectStatusRepository) FindByProjectID(projectID uuid.UUID) ([]model.ProjectStatus, error) {
	var rows []model.ProjectStatus
	err := r.db.Where("project_id = ?", projectID).Order(`"order" asc`).Find(&rows).Error
	return rows, err
}

func (r *projectStatusRepository) FindByID(id uuid.UUID) (*model.ProjectStatus, error) {
	var ps model.ProjectStatus
	err := r.db.First(&ps, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &ps, nil
}
