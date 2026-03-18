package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type AddStepInput struct {
	StatusID        uuid.UUID  // このステップのステータス
	NextStatusID    *uuid.UUID // 承認後ステータス（ゴールでは nil）
	Description     string
	Threshold       int
	ApprovalObjects []ApprovalObjectInput
	ExcludeReporter bool
	ExcludeAssignee bool
}

type UpdateStepInput struct {
	StatusID        *uuid.UUID
	NextStatusID    *uuid.UUID
	Description     string
	Threshold       int
	ApprovalObjects []ApprovalObjectInput
	ExcludeReporter bool
	ExcludeAssignee bool
}

type ApprovalObjectInput struct {
	Type            string
	RoleID          *uint
	RoleOperator    string
	UserID          *uuid.UUID
	Points          int
	ExcludeReporter bool
	ExcludeAssignee bool
}

type WorkflowService interface {
	ListAll() ([]model.Workflow, error)
	GetWorkflow(id uint) (*model.Workflow, error)
	GetStep(stepID uint) (*model.WorkflowStep, error)
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
	statusRepo   repository.StatusRepository
}

func NewWorkflowService(workflowRepo repository.WorkflowRepository, statusRepo repository.StatusRepository) WorkflowService {
	return &workflowService{workflowRepo: workflowRepo, statusRepo: statusRepo}
}

func (s *workflowService) ListAll() ([]model.Workflow, error) {
	return s.workflowRepo.FindAll()
}

func (s *workflowService) GetWorkflow(id uint) (*model.Workflow, error) {
	return s.workflowRepo.FindByID(id)
}

func (s *workflowService) GetStep(stepID uint) (*model.WorkflowStep, error) {
	return s.workflowRepo.FindStepByID(stepID)
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
	threshold := input.Threshold
	if threshold < 1 {
		threshold = 10
	}
	step := &model.WorkflowStep{
		WorkflowID:      workflowID,
		Order:           int(count) + 1,
		StatusID:        input.StatusID,
		NextStatusID:    input.NextStatusID,
		Description:     input.Description,
		Threshold:       threshold,
		ExcludeReporter: input.ExcludeReporter,
		ExcludeAssignee: input.ExcludeAssignee,
	}
	if err := s.workflowRepo.CreateStep(step); err != nil {
		return nil, err
	}
	for i, ao := range input.ApprovalObjects {
		obj := s.approvalObjectInputToModel(ao, step.ID, i+1)
		if obj != nil {
			_ = s.workflowRepo.CreateApprovalObject(obj)
		}
	}
	return s.workflowRepo.FindStepByID(step.ID)
}

func (s *workflowService) approvalObjectInputToModel(in ApprovalObjectInput, stepID uint, order int) *model.ApprovalObject {
	if in.Type != "role" && in.Type != "user" {
		return nil
	}
	obj := &model.ApprovalObject{
		WorkflowStepID:  stepID,
		Order:           order,
		Type:            in.Type,
		Points:          in.Points,
		ExcludeReporter: in.ExcludeReporter,
		ExcludeAssignee: in.ExcludeAssignee,
	}
	if in.Points < 1 {
		obj.Points = 1
	}
	if in.Type == "role" && in.RoleID != nil {
		obj.RoleID = in.RoleID
		obj.RoleOperator = in.RoleOperator
		if obj.RoleOperator == "" {
			obj.RoleOperator = "gte"
		}
	}
	if in.Type == "user" && in.UserID != nil {
		obj.UserID = in.UserID
	}
	return obj
}

func (s *workflowService) UpdateStep(stepID uint, input UpdateStepInput) (*model.WorkflowStep, error) {
	step, err := s.workflowRepo.FindStepByID(stepID)
	if err != nil {
		return nil, err
	}
	if input.StatusID != nil {
		step.StatusID = *input.StatusID
	}
	if input.NextStatusID != nil {
		step.NextStatusID = input.NextStatusID
	}
	step.Description = input.Description
	if input.Threshold >= 1 {
		step.Threshold = input.Threshold
	}
	step.ExcludeReporter = input.ExcludeReporter
	step.ExcludeAssignee = input.ExcludeAssignee
	if err := s.workflowRepo.UpdateStep(step); err != nil {
		return nil, err
	}
	if input.ApprovalObjects != nil {
		_ = s.workflowRepo.DeleteApprovalObjectsByStepID(stepID)
		for i, ao := range input.ApprovalObjects {
			obj := s.approvalObjectInputToModel(ao, stepID, i+1)
			if obj != nil {
				_ = s.workflowRepo.CreateApprovalObject(obj)
			}
		}
	}
	return s.workflowRepo.FindStepByID(stepID)
}

func (s *workflowService) DeleteStep(stepID uint) error {
	return s.workflowRepo.DeleteStep(stepID)
}
