package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	FindAll() ([]model.Project, error)
	FindByID(id uuid.UUID) (*model.Project, error)
	Create(project *model.Project) error
	Update(project *model.Project) error
	Delete(id uuid.UUID) error
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) FindAll() ([]model.Project, error) {
	var projects []model.Project
	err := r.db.Preload("Owner").Find(&projects).Error
	return projects, err
}

func (r *projectRepository) FindByID(id uuid.UUID) (*model.Project, error) {
	var project model.Project
	err := r.db.Preload("Owner").Preload("Statuses").First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) Create(project *model.Project) error {
	return r.db.Create(project).Error
}

func (r *projectRepository) Update(project *model.Project) error {
	return r.db.Save(project).Error
}

func (r *projectRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Project{}, "id = ?", id).Error
}
