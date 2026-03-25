package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	FindAll() ([]model.Project, error)
	FindByOrg(orgID uuid.UUID) ([]model.Project, error)
	FindByOrgAndName(orgID uuid.UUID, name string) (*model.Project, error)
	FindByID(id uuid.UUID) (*model.Project, error)
	Create(project *model.Project) error
	Update(project *model.Project) error
	Delete(id uuid.UUID) error
	Reorder(orgID *uuid.UUID, ids []uuid.UUID) error
	GetMaxOrder(orgID *uuid.UUID) (int, error)
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) FindAll() ([]model.Project, error) {
	var projects []model.Project
	err := r.db.Preload("Owner").Order("display_order ASC").Find(&projects).Error
	return projects, err
}

func (r *projectRepository) FindByOrg(orgID uuid.UUID) ([]model.Project, error) {
	var projects []model.Project
	err := r.db.Preload("Owner").Where("organization_id = ?", orgID).Order("display_order ASC").Find(&projects).Error
	return projects, err
}

func (r *projectRepository) FindByOrgAndName(orgID uuid.UUID, name string) (*model.Project, error) {
	var project model.Project
	err := r.db.Where("organization_id = ? AND name = ?", orgID, name).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) FindByID(id uuid.UUID) (*model.Project, error) {
	var project model.Project
	err := r.db.Preload("Owner").Preload("ProjectStatus").First(&project, "id = ?", id).Error
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

func (r *projectRepository) GetMaxOrder(orgID *uuid.UUID) (int, error) {
	var maxOrder int
	q := r.db.Model(&model.Project{})
	if orgID != nil {
		q = q.Where("organization_id = ?", orgID)
	}
	err := q.Select("COALESCE(MAX(display_order), 0)").Scan(&maxOrder).Error
	return maxOrder, err
}

func (r *projectRepository) Reorder(orgID *uuid.UUID, ids []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			q := tx.Model(&model.Project{}).Where("id = ?", id)
			if orgID != nil {
				q = q.Where("organization_id = ?", orgID)
			}
			if err := q.Update("display_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
