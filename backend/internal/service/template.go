package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type TemplateService interface {
	ListAll() ([]model.IssueTemplate, error)
	ListByProject(projectID uuid.UUID) ([]model.IssueTemplate, error)
	GetTemplate(id uint) (*model.IssueTemplate, error)
	CreateTemplate(projectID uuid.UUID, name, description, body, defaultPriority string, workflowID *uint) (*model.IssueTemplate, error)
	UpdateTemplate(id uint, name, description, body, defaultPriority string, workflowID *uint) (*model.IssueTemplate, error)
	DeleteTemplate(id uint) error
	Reorder(projectID uuid.UUID, ids []uint) error
}

type templateService struct {
	templateRepo repository.TemplateRepository
}

func NewTemplateService(templateRepo repository.TemplateRepository) TemplateService {
	return &templateService{templateRepo: templateRepo}
}

func (s *templateService) ListAll() ([]model.IssueTemplate, error) {
	return s.templateRepo.FindAll()
}

func (s *templateService) ListByProject(projectID uuid.UUID) ([]model.IssueTemplate, error) {
	return s.templateRepo.FindByProjectID(projectID)
}

func (s *templateService) GetTemplate(id uint) (*model.IssueTemplate, error) {
	return s.templateRepo.FindByID(id)
}

func (s *templateService) CreateTemplate(projectID uuid.UUID, name, description, body, defaultPriority string, workflowID *uint) (*model.IssueTemplate, error) {
	if defaultPriority == "" {
		defaultPriority = "medium"
	}
	maxOrder, err := s.templateRepo.GetMaxOrder(projectID)
	if err != nil {
		return nil, err
	}
	template := &model.IssueTemplate{
		ProjectID:       projectID,
		Name:            name,
		Description:     description,
		Body:            body,
		DefaultPriority: defaultPriority,
		WorkflowID:      workflowID,
		Order:           maxOrder + 1,
		CreatedAt:       time.Now(),
	}
	if err := s.templateRepo.Create(template); err != nil {
		return nil, err
	}
	return s.templateRepo.FindByID(template.ID)
}

func (s *templateService) UpdateTemplate(id uint, name, description, body, defaultPriority string, workflowID *uint) (*model.IssueTemplate, error) {
	if defaultPriority == "" {
		defaultPriority = "medium"
	}
	tmpl, err := s.templateRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	tmpl.Name = name
	tmpl.Description = description
	tmpl.Body = body
	tmpl.DefaultPriority = defaultPriority
	tmpl.WorkflowID = workflowID
	if err := s.templateRepo.Update(tmpl); err != nil {
		return nil, err
	}
	return s.templateRepo.FindByID(id)
}

func (s *templateService) DeleteTemplate(id uint) error {
	return s.templateRepo.Delete(id)
}

func (s *templateService) Reorder(projectID uuid.UUID, ids []uint) error {
	return s.templateRepo.Reorder(projectID, ids)
}
