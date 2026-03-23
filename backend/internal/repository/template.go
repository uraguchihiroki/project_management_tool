package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type TemplateRepository interface {
	FindAll() ([]model.IssueTemplate, error)
	FindByProjectID(projectID uuid.UUID) ([]model.IssueTemplate, error)
	FindByID(id uint) (*model.IssueTemplate, error)
	Create(template *model.IssueTemplate) error
	Update(template *model.IssueTemplate) error
	Delete(id uint) error
	Reorder(projectID uuid.UUID, ids []uint) error
	GetMaxOrder(projectID uuid.UUID) (int, error)
}

type templateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) TemplateRepository {
	return &templateRepository{db: db}
}

func (r *templateRepository) FindAll() ([]model.IssueTemplate, error) {
	var templates []model.IssueTemplate
	err := r.db.Preload("Project").Order("created_at DESC").Find(&templates).Error
	return templates, err
}

func (r *templateRepository) FindByProjectID(projectID uuid.UUID) ([]model.IssueTemplate, error) {
	var templates []model.IssueTemplate
	err := r.db.Where("project_id = ?", projectID).Order("display_order ASC").Find(&templates).Error
	return templates, err
}

func (r *templateRepository) FindByID(id uint) (*model.IssueTemplate, error) {
	var template model.IssueTemplate
	err := r.db.Preload("Project").First(&template, id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *templateRepository) Create(template *model.IssueTemplate) error {
	return r.db.Create(template).Error
}

func (r *templateRepository) Update(template *model.IssueTemplate) error {
	updates := map[string]interface{}{
		"name":             template.Name,
		"description":      template.Description,
		"body":             template.Body,
		"default_priority": template.DefaultPriority,
	}
	if template.Key != "" {
		updates["key"] = template.Key
	}
	return r.db.Model(&model.IssueTemplate{}).Where("id = ?", template.ID).Updates(updates).Error
}

func (r *templateRepository) Delete(id uint) error {
	return r.db.Delete(&model.IssueTemplate{}, id).Error
}

func (r *templateRepository) GetMaxOrder(projectID uuid.UUID) (int, error) {
	var maxOrder int
	err := r.db.Model(&model.IssueTemplate{}).Where("project_id = ?", projectID).
		Select("COALESCE(MAX(display_order), 0)").Scan(&maxOrder).Error
	return maxOrder, err
}

func (r *templateRepository) Reorder(projectID uuid.UUID, ids []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.IssueTemplate{}).
				Where("id = ? AND project_id = ?", id, projectID).
				Update("display_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
