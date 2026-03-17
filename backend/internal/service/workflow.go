package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type WorkflowService interface {
	ListAll() ([]model.Workflow, error)
	ListByProject(projectID uuid.UUID) ([]model.Workflow, error)
	GetWorkflow(id uint) (*model.Workflow, error)
	CreateWorkflow(projectID uuid.UUID, name, description string) (*model.Workflow, error)
	UpdateWorkflow(id uint, name, description string) (*model.Workflow, error)
	DeleteWorkflow(id uint) error
	AddStep(workflowID uint, name string, requiredLevel int, statusID *uuid.UUID) (*model.WorkflowStep, error)
	UpdateStep(stepID uint, name string, requiredLevel int, statusID *uuid.UUID, order int) (*model.WorkflowStep, error)
	DeleteStep(stepID uint) error
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

func (s *workflowService) ListByProject(projectID uuid.UUID) ([]model.Workflow, error) {
	return s.workflowRepo.FindByProjectID(projectID)
}

func (s *workflowService) GetWorkflow(id uint) (*model.Workflow, error) {
	return s.workflowRepo.FindByID(id)
}

func (s *workflowService) CreateWorkflow(projectID uuid.UUID, name, description string) (*model.Workflow, error) {
	workflow := &model.Workflow{
		ProjectID:   projectID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}
	if err := s.workflowRepo.Create(workflow); err != nil {
		return nil, err
	}
	return s.workflowRepo.FindByID(workflow.ID)
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

func (s *workflowService) AddStep(workflowID uint, name string, requiredLevel int, statusID *uuid.UUID) (*model.WorkflowStep, error) {
	// 既存ステップ数に基づいてorder自動採番
	count, err := s.workflowRepo.CountSteps(workflowID)
	if err != nil {
		return nil, err
	}
	step := &model.WorkflowStep{
		WorkflowID:    workflowID,
		Order:         int(count) + 1,
		Name:          name,
		RequiredLevel: requiredLevel,
		StatusID:      statusID,
	}
	if err := s.workflowRepo.CreateStep(step); err != nil {
		return nil, err
	}
	return s.workflowRepo.FindStepByID(step.ID)
}

func (s *workflowService) UpdateStep(stepID uint, name string, requiredLevel int, statusID *uuid.UUID, order int) (*model.WorkflowStep, error) {
	step, err := s.workflowRepo.FindStepByID(stepID)
	if err != nil {
		return nil, err
	}
	step.Name = name
	step.RequiredLevel = requiredLevel
	step.StatusID = statusID
	step.Order = order
	if err := s.workflowRepo.UpdateStep(step); err != nil {
		return nil, err
	}
	return s.workflowRepo.FindStepByID(stepID)
}

func (s *workflowService) DeleteStep(stepID uint) error {
	return s.workflowRepo.DeleteStep(stepID)
}
