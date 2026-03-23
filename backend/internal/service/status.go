package service

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

var colorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type StatusService interface {
	Get(id uuid.UUID) (*model.Status, error)
	Create(orgID uuid.UUID, name, color, statusType string, order int) (*model.Status, error)
	ListByWorkflowID(workflowID uint) ([]model.Status, error)
	CreateForWorkflow(workflowID uint, name, color, statusType string, order int) (*model.Status, error)
	Update(id uuid.UUID, name, color string, order int) (*model.Status, error)
	Delete(id uuid.UUID) error
}

type statusService struct {
	statusRepo     repository.StatusRepository
	workflowRepo   repository.WorkflowRepository
	transitionRepo repository.WorkflowTransitionRepository
}

func NewStatusService(
	statusRepo repository.StatusRepository,
	workflowRepo repository.WorkflowRepository,
	transitionRepo repository.WorkflowTransitionRepository,
) StatusService {
	return &statusService{
		statusRepo:     statusRepo,
		workflowRepo:   workflowRepo,
		transitionRepo: transitionRepo,
	}
}

func (s *statusService) Get(id uuid.UUID) (*model.Status, error) {
	return s.statusRepo.FindByID(id)
}

func (s *statusService) Create(orgID uuid.UUID, name, color, statusType string, order int) (*model.Status, error) {
	if name == "" {
		return nil, fmt.Errorf("ステータス名は必須です")
	}
	if len(name) > 50 {
		return nil, fmt.Errorf("ステータス名は50文字以内で指定してください")
	}
	if color == "" || !colorRegex.MatchString(color) {
		return nil, fmt.Errorf("色は#RRGGBB形式で指定してください")
	}
	if statusType != "issue" && statusType != "project" {
		statusType = "issue"
	}
	wfName := "組織Issue"
	if statusType == "project" {
		wfName = "組織Project"
	}
	wf, err := s.workflowRepo.FindByOrgAndName(orgID, wfName)
	if err != nil {
		return nil, fmt.Errorf("ワークフロー %s が見つかりません（組織シードを実行してください）: %w", wfName, err)
	}
	statusID := uuid.New()
	key := "sts-" + statusID.String()
	status := &model.Status{
		ID:         statusID,
		Key:        key,
		WorkflowID: wf.ID,
		Name:       name,
		Color:      color,
		Order:      order,
		Type:       statusType,
	}
	if err := s.statusRepo.Create(status); err != nil {
		return nil, err
	}
	return s.statusRepo.FindByID(statusID)
}

func (s *statusService) ListByWorkflowID(workflowID uint) ([]model.Status, error) {
	return s.statusRepo.FindByWorkflowID(workflowID)
}

func (s *statusService) CreateForWorkflow(workflowID uint, name, color, statusType string, order int) (*model.Status, error) {
	if _, err := s.workflowRepo.FindByID(workflowID); err != nil {
		return nil, fmt.Errorf("ワークフローが見つかりません")
	}
	if name == "" {
		return nil, fmt.Errorf("ステータス名は必須です")
	}
	if len(name) > 50 {
		return nil, fmt.Errorf("ステータス名は50文字以内で指定してください")
	}
	if color == "" || !colorRegex.MatchString(color) {
		return nil, fmt.Errorf("色は#RRGGBB形式で指定してください")
	}
	if statusType != "issue" && statusType != "project" {
		statusType = "issue"
	}
	existing, err := s.statusRepo.FindByWorkflowID(workflowID)
	if err != nil {
		return nil, err
	}
	if order <= 0 {
		maxO := 0
		for _, st := range existing {
			if st.Order > maxO {
				maxO = st.Order
			}
		}
		order = maxO + 1
	}
	statusID := uuid.New()
	key := "sts-" + statusID.String()
	status := &model.Status{
		ID:         statusID,
		Key:        key,
		WorkflowID: workflowID,
		Name:       name,
		Color:      color,
		Order:      order,
		Type:       statusType,
	}
	if err := s.statusRepo.Create(status); err != nil {
		return nil, err
	}
	all, err := s.statusRepo.FindByWorkflowID(workflowID)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(all))
	for _, st := range all {
		ids = append(ids, st.ID)
	}
	if err := s.transitionRepo.SeedAllPairs(workflowID, ids); err != nil {
		return nil, err
	}
	return s.statusRepo.FindByID(statusID)
}

func (s *statusService) Update(id uuid.UUID, name, color string, order int) (*model.Status, error) {
	status, err := s.statusRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if status.StatusKey == "sts_start" || status.StatusKey == "sts_goal" {
		return nil, fmt.Errorf("システムステータスは変更できません")
	}
	if name != "" {
		if len(name) > 50 {
			return nil, fmt.Errorf("ステータス名は50文字以内で指定してください")
		}
		status.Name = name
	}
	if color != "" {
		if !colorRegex.MatchString(color) {
			return nil, fmt.Errorf("色は#RRGGBB形式で指定してください")
		}
		status.Color = color
	}
	status.Order = order
	if err := s.statusRepo.Update(status); err != nil {
		return nil, err
	}
	return s.statusRepo.FindByID(id)
}

func (s *statusService) Delete(id uuid.UUID) error {
	status, err := s.statusRepo.FindByID(id)
	if err != nil {
		return err
	}
	if status.StatusKey == "sts_start" || status.StatusKey == "sts_goal" {
		return fmt.Errorf("システムステータスは削除できません")
	}
	count, err := s.statusRepo.CountInUse(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("このステータスは使用中のため削除できません")
	}
	return s.statusRepo.Delete(id)
}
