package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/pkg/keygen"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

// WorkflowService は組織スコープのワークフロー CRUD（承認ステップは廃止）
type WorkflowService interface {
	ListAll() ([]model.Workflow, error)
	GetWorkflow(id uint) (*model.Workflow, error)
	CreateWorkflow(orgID uuid.UUID, name, description string) (*model.Workflow, error)
	UpdateWorkflow(id uint, name, description string) (*model.Workflow, error)
	DeleteWorkflow(id uint) error
	Reorder(ids []uint) error
}

type workflowService struct {
	workflowRepo repository.WorkflowRepository
}

func NewWorkflowService(workflowRepo repository.WorkflowRepository) WorkflowService {
	return &workflowService{workflowRepo: workflowRepo}
}

func (s *workflowService) ListAll() ([]model.Workflow, error) {
	return s.workflowRepo.FindAll()
}

func (s *workflowService) GetWorkflow(id uint) (*model.Workflow, error) {
	return s.workflowRepo.FindByID(id)
}

func (s *workflowService) CreateWorkflow(orgID uuid.UUID, name, description string) (*model.Workflow, error) {
	maxOrder, err := s.workflowRepo.GetMaxOrder()
	if err != nil {
		return nil, err
	}
	workflow := &model.Workflow{
		OrganizationID: orgID,
		Name:           name,
		Description:    description,
		Order:          maxOrder + 1,
		CreatedAt:      time.Now(),
	}
	if err := s.workflowRepo.Create(workflow); err != nil {
		return nil, err
	}
	key := keygen.Slug(name)
	if key == "" {
		key = keygen.PrefixedID("wf", workflow.ID)
	}
	workflow.Key = key
	_ = s.workflowRepo.Update(workflow)
	return s.workflowRepo.FindByID(workflow.ID)
}

func (s *workflowService) Reorder(ids []uint) error {
	return s.workflowRepo.Reorder(ids)
}

func (s *workflowService) UpdateWorkflow(id uint, name, description string) (*model.Workflow, error) {
	workflow, err := s.workflowRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	workflow.Name = name
	workflow.Description = description
	if err := s.workflowRepo.Update(workflow); err != nil {
		return nil, err
	}
	return s.workflowRepo.FindByID(id)
}

func (s *workflowService) DeleteWorkflow(id uint) error {
	return s.workflowRepo.Delete(id)
}
