package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/pkg/keygen"
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
	CreateWorkflow(orgID uuid.UUID, name, description string) (*model.Workflow, error)
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

func (s *workflowService) ReorderSteps(workflowID uint, ids []uint) error {
	return s.workflowRepo.ReorderStepsWithNextStatus(workflowID, ids)
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
	workflow, err := s.workflowRepo.FindByID(workflowID)
	if err != nil {
		return nil, err
	}
	count, err := s.workflowRepo.CountSteps(workflowID)
	if err != nil {
		return nil, err
	}
	threshold := input.Threshold
	if threshold < 1 {
		threshold = 10
	}

	if count == 0 {
		return s.addFirstStep(workflow, input, threshold)
	}
	return s.addSubsequentStep(workflow, input, threshold)
}

func (s *workflowService) addFirstStep(workflow *model.Workflow, input AddStepInput, threshold int) (*model.WorkflowStep, error) {
	stsStart, err := s.statusRepo.FindByStatusKey("sts_start")
	if err != nil {
		return nil, err
	}
	stsGoal, err := s.statusRepo.FindByStatusKey("sts_goal")
	if err != nil {
		return nil, err
	}

	// 1. sts_start (order=0, next=user)
	stepStart := &model.WorkflowStep{
		OrganizationID: workflow.OrganizationID,
		WorkflowID:     workflow.ID,
		Order:          0,
		StatusID:       stsStart.ID,
		NextStatusID:   &input.StatusID,
		Description:    "",
		Threshold:      10,
	}
	if err := s.workflowRepo.CreateStep(stepStart); err != nil {
		return nil, err
	}
	stepStart.Key = keygen.PrefixedID("ws", stepStart.ID)
	_ = s.workflowRepo.UpdateStep(stepStart)

	// 2. user step (order=1, next=sts_goal)
	nextGoal := stsGoal.ID
	step := &model.WorkflowStep{
		OrganizationID:  workflow.OrganizationID,
		WorkflowID:      workflow.ID,
		Order:           1,
		StatusID:        input.StatusID,
		NextStatusID:    &nextGoal,
		Description:     input.Description,
		Threshold:       threshold,
		ExcludeReporter: input.ExcludeReporter,
		ExcludeAssignee: input.ExcludeAssignee,
	}
	if err := s.workflowRepo.CreateStep(step); err != nil {
		return nil, err
	}
	step.Key = keygen.PrefixedID("ws", step.ID)
	_ = s.workflowRepo.UpdateStep(step)

	// 3. sts_goal (order=2, next=nil)
	stepGoal := &model.WorkflowStep{
		OrganizationID: workflow.OrganizationID,
		WorkflowID:     workflow.ID,
		Order:          2,
		StatusID:       stsGoal.ID,
		NextStatusID:   nil,
		Description:    "",
		Threshold:      10,
	}
	if err := s.workflowRepo.CreateStep(stepGoal); err != nil {
		return nil, err
	}
	stepGoal.Key = keygen.PrefixedID("ws", stepGoal.ID)
	_ = s.workflowRepo.UpdateStep(stepGoal)

	for i, ao := range input.ApprovalObjects {
		obj := s.approvalObjectInputToModel(ao, step.ID, workflow.OrganizationID, i+1)
		if obj != nil {
			_ = s.workflowRepo.CreateApprovalObject(obj)
			obj.Key = keygen.PrefixedID("ao", obj.ID)
			_ = s.workflowRepo.UpdateApprovalObject(obj)
		}
	}
	return s.workflowRepo.FindStepByID(step.ID)
}

func (s *workflowService) addSubsequentStep(workflow *model.Workflow, input AddStepInput, threshold int) (*model.WorkflowStep, error) {
	steps, err := s.workflowRepo.FindStepsByWorkflowID(workflow.ID)
	if err != nil {
		return nil, err
	}
	var stsGoalStep *model.WorkflowStep
	var prevUserStep *model.WorkflowStep
	for i := range steps {
		if steps[i].Status != nil && steps[i].Status.StatusKey == "sts_goal" {
			stsGoalStep = &steps[i]
			if i > 0 {
				prevUserStep = &steps[i-1]
			}
			break
		}
	}
	if stsGoalStep == nil {
		return nil, errors.New("sts_goal step not found")
	}

	insertOrder := stsGoalStep.Order
	stsGoalStatusID := stsGoalStep.StatusID

	// 新ステップを sts_goal の直前に挿入
	step := &model.WorkflowStep{
		OrganizationID:  workflow.OrganizationID,
		WorkflowID:      workflow.ID,
		Order:           insertOrder,
		StatusID:        input.StatusID,
		NextStatusID:    &stsGoalStatusID,
		Description:     input.Description,
		Threshold:       threshold,
		ExcludeReporter: input.ExcludeReporter,
		ExcludeAssignee: input.ExcludeAssignee,
	}
	if err := s.workflowRepo.CreateStep(step); err != nil {
		return nil, err
	}
	step.Key = keygen.PrefixedID("ws", step.ID)
	_ = s.workflowRepo.UpdateStep(step)

	// sts_goal の order を +1
	stsGoalStep.Order = insertOrder + 1
	_ = s.workflowRepo.UpdateStep(stsGoalStep)

	// 直前ユーザーステップの next_status_id を新ステップに
	if prevUserStep != nil {
		prevUserStep.NextStatusID = &input.StatusID
		_ = s.workflowRepo.UpdateStep(prevUserStep)
	}

	for i, ao := range input.ApprovalObjects {
		obj := s.approvalObjectInputToModel(ao, step.ID, workflow.OrganizationID, i+1)
		if obj != nil {
			_ = s.workflowRepo.CreateApprovalObject(obj)
			obj.Key = keygen.PrefixedID("ao", obj.ID)
			_ = s.workflowRepo.UpdateApprovalObject(obj)
		}
	}
	return s.workflowRepo.FindStepByID(step.ID)
}

func (s *workflowService) approvalObjectInputToModel(in ApprovalObjectInput, stepID uint, orgID uuid.UUID, order int) *model.ApprovalObject {
	if in.Type != "role" && in.Type != "user" {
		return nil
	}
	obj := &model.ApprovalObject{
		OrganizationID:  orgID,
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
	// next_status_id は ReorderSteps でのみ更新。ここでは無視する
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
		workflow, _ := s.workflowRepo.FindByID(step.WorkflowID)
		orgID := uuid.Nil
		if workflow != nil {
			orgID = workflow.OrganizationID
		}
		for i, ao := range input.ApprovalObjects {
			obj := s.approvalObjectInputToModel(ao, stepID, orgID, i+1)
			if obj != nil {
				_ = s.workflowRepo.CreateApprovalObject(obj)
				obj.Key = keygen.PrefixedID("ao", obj.ID)
				_ = s.workflowRepo.UpdateApprovalObject(obj)
			}
		}
	}
	return s.workflowRepo.FindStepByID(stepID)
}

func (s *workflowService) DeleteStep(stepID uint) error {
	return s.workflowRepo.DeleteStep(stepID)
}
