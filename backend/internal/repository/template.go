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
}

type templateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) TemplateRepository {
	return &templateRepository{db: db}
}

func (r *templateRepository) FindAll() ([]model.IssueTemplate, error) {
	var templates []model.IssueTemplate
	err := r.db.Preload("Project").Preload("Workflow").Order("created_at DESC").Find(&templates).Error
	return templates, err
}

func (r *templateRepository) FindByProjectID(projectID uuid.UUID) ([]model.IssueTemplate, error) {
	var templates []model.IssueTemplate
	err := r.db.Preload("Workflow").Where("project_id = ?", projectID).Order("name ASC").Find(&templates).Error
	return templates, err
}

func (r *templateRepository) FindByID(id uint) (*model.IssueTemplate, error) {
	var template model.IssueTemplate
	err := r.db.Preload("Project").Preload("Workflow").First(&template, id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *templateRepository) Create(template *model.IssueTemplate) error {
	return r.db.Create(template).Error
}

func (r *templateRepository) Update(template *model.IssueTemplate) error {
	return r.db.Model(&model.IssueTemplate{}).Where("id = ?", template.ID).Updates(map[string]interface{}{
		"name":             template.Name,
		"description":      template.Description,
		"body":             template.Body,
		"default_priority": template.DefaultPriority,
		"workflow_id":      template.WorkflowID,
	}).Error
}

func (r *templateRepository) Delete(id uint) error {
	return r.db.Delete(&model.IssueTemplate{}, id).Error
}
