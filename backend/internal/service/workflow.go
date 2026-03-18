package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type AddStepInput struct {
	Name           string
	RequiredLevel  int
	StatusID       *uuid.UUID
	ApproverType   string
	ApproverUserID *uuid.UUID
	MinApprovers   int
	ExcludeReporter bool
	ExcludeAssignee bool
}

type UpdateStepInput struct {
	Name            string
	RequiredLevel   int
	StatusID        *uuid.UUID
	ApproverType    string
	ApproverUserID  *uuid.UUID
	MinApprovers    int
	ExcludeReporter bool
	ExcludeAssignee bool
}

type WorkflowService interface {
	ListAll() ([]model.Workflow, error)
	GetWorkflow(id uint) (*model.Workflow, error)
	CreateWorkflow(name, description string) (*model.Workflow, error)
	UpdateWorkflow(id uint, name, description string) (*model.Workflow, error)
	DeleteWorkflow(id uint) error
	Reorder(ids []uint) error
	ReorderSteps(workflowID uint, ids []uint) error
	AddStep(workflowID uint, input AddStepInput) (*model.WorkflowStep, error)
	UpdateStep(stepID uint, input UpdateStepInput) (*model.WorkflowStep, error)
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

func (s *workflowService) GetWorkflow(id uint) (*model.Workflow, error) {
	return s.workflowRepo.FindByID(id)
}

func (s *workflowService) CreateWorkflow(name, description string) (*model.Workflow, error) {
	maxOrder, err := s.workflowRepo.GetMaxOrder()
	if err != nil {
		return nil, err
	}
	workflow := &model.Workflow{
		Name:      name,
		Description: description,
		Order:     maxOrder + 1,
		CreatedAt: time.Now(),
	}
	if err := s.workflowRepo.Create(workflow); err != nil {
		return nil, err
	}
	return s.workflowRepo.FindByID(workflow.ID)
}

func (s *workflowService) Reorder(ids []uint) error {
	return s.workflowRepo.Reorder(ids)
}

func (s *workflowService) ReorderSteps(workflowID uint, ids []uint) error {
	return s.workflowRepo.ReorderSteps(workflowID, ids)
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

func (s *workflowService) AddStep(workflowID uint, input AddStepInput) (*model.WorkflowStep, error) {
	count, err := s.workflowRepo.CountSteps(workflowID)
	if err != nil {
		return nil, err
	}
	approverType := input.ApproverType
	if approverType == "" {
		approverType = "role"
	}
	minApprovers := input.MinApprovers
	if minApprovers < 1 {
		minApprovers = 1
	}
	step := &model.WorkflowStep{
		WorkflowID:      workflowID,
		Order:           int(count) + 1,
		Name:            input.Name,
		RequiredLevel:   input.RequiredLevel,
		StatusID:        input.StatusID,
		ApproverType:    approverType,
		ApproverUserID:  input.ApproverUserID,
		MinApprovers:    minApprovers,
		ExcludeReporter: input.ExcludeReporter,
		ExcludeAssignee: input.ExcludeAssignee,
	}
	if err := s.workflowRepo.CreateStep(step); err != nil {
		return nil, err
	}
	return s.workflowRepo.FindStepByID(step.ID)
}

func (s *workflowService) UpdateStep(stepID uint, input UpdateStepInput) (*model.WorkflowStep, error) {
	step, err := s.workflowRepo.FindStepByID(stepID)
	if err != nil {
		return nil, err
	}
	step.Name = input.Name
	step.RequiredLevel = input.RequiredLevel
	step.StatusID = input.StatusID
	if input.ApproverType != "" {
		step.ApproverType = input.ApproverType
	}
	step.ApproverUserID = input.ApproverUserID
	if input.MinApprovers >= 1 {
		step.MinApprovers = input.MinApprovers
	}
	step.ExcludeReporter = input.ExcludeReporter
	step.ExcludeAssignee = input.ExcludeAssignee
	if err := s.workflowRepo.UpdateStep(step); err != nil {
		return nil, err
	}
	return s.workflowRepo.FindStepByID(stepID)
}

func (s *workflowService) DeleteStep(stepID uint) error {
	return s.workflowRepo.DeleteStep(stepID)
}
