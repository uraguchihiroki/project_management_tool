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
	Create(orgID uuid.UUID, name, color string, order int) (*model.Status, error)
	ListByWorkflowID(workflowID uint) ([]model.Status, error)
	CreateForWorkflow(workflowID uint, name, color string, order int) (*model.Status, error)
	Update(id uuid.UUID, name, color string, order int) (*model.Status, error)
	Delete(id uuid.UUID) error
	ReorderForWorkflow(workflowID uint, orderedIDs []uuid.UUID) error
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

func (s *statusService) Create(orgID uuid.UUID, name, color string, order int) (*model.Status, error) {
	if name == "" {
		return nil, fmt.Errorf("ステータス名は必須です")
	}
	if len(name) > 50 {
		return nil, fmt.Errorf("ステータス名は50文字以内で指定してください")
	}
	if color == "" || !colorRegex.MatchString(color) {
		return nil, fmt.Errorf("色は#RRGGBB形式で指定してください")
	}
	wfName := "組織Issue"
	wf, err := s.workflowRepo.FindByOrgAndName(orgID, wfName)
	if err != nil {
		return nil, fmt.Errorf("ワークフロー %s が見つかりません（組織シードを実行してください）: %w", wfName, err)
	}
	existingOrg, err := s.statusRepo.FindByWorkflowID(wf.ID)
	if err != nil {
		return nil, err
	}
	for _, st := range existingOrg {
		if st.Name == name && st.DisplayOrder == order {
			return nil, fmt.Errorf("同一ワークフローに同じ表示順・名前のステータスが既にあります")
		}
	}
	statusID := uuid.New()
	key := "sts-" + statusID.String()
	status := &model.Status{
		ID:           statusID,
		Key:          key,
		WorkflowID:   wf.ID,
		Name:         name,
		Color:        color,
		DisplayOrder: order,
	}
	if err := s.statusRepo.Create(status); err != nil {
		return nil, err
	}
	return s.statusRepo.FindByID(statusID)
}

func (s *statusService) ListByWorkflowID(workflowID uint) ([]model.Status, error) {
	return s.statusRepo.FindByWorkflowID(workflowID)
}

func (s *statusService) CreateForWorkflow(workflowID uint, name, color string, order int) (*model.Status, error) {
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
	existing, err := s.statusRepo.FindByWorkflowID(workflowID)
	if err != nil {
		return nil, err
	}
	if order <= 0 {
		maxO := 0
		for _, st := range existing {
			if st.DisplayOrder > maxO {
				maxO = st.DisplayOrder
			}
		}
		order = maxO + 1
	}
	for _, st := range existing {
		if st.Name == name && st.DisplayOrder == order {
			return nil, fmt.Errorf("同一ワークフローに同じ表示順・名前のステータスが既にあります")
		}
	}
	statusID := uuid.New()
	key := "sts-" + statusID.String()
	status := &model.Status{
		ID:           statusID,
		Key:          key,
		WorkflowID:   workflowID,
		Name:         name,
		Color:        color,
		DisplayOrder: order,
	}
	if err := s.statusRepo.Create(status); err != nil {
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
	status.DisplayOrder = order
	peers, err := s.statusRepo.FindByWorkflowID(status.WorkflowID)
	if err != nil {
		return nil, err
	}
	for _, st := range peers {
		if st.ID == id {
			continue
		}
		if st.Name == status.Name && st.DisplayOrder == status.DisplayOrder {
			return nil, fmt.Errorf("同一ワークフローに同じ表示順・名前のステータスが既にあります")
		}
	}
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
	tref, err := s.transitionRepo.CountReferencingStatus(id)
	if err != nil {
		return err
	}
	if tref > 0 {
		return fmt.Errorf("このステータスは許可遷移で使用されているため削除できません")
	}
	count, err := s.statusRepo.CountInUse(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("このステータスは使用中のため削除できません")
	}
	wfCount, err := s.statusRepo.CountByWorkflowID(status.WorkflowID)
	if err != nil {
		return err
	}
	if wfCount <= 2 {
		return fmt.Errorf("ステータスはワークフロー内で最低2つ必要なため削除できません")
	}
	return s.statusRepo.Delete(id)
}

func (s *statusService) ReorderForWorkflow(workflowID uint, orderedIDs []uuid.UUID) error {
	existing, err := s.statusRepo.FindByWorkflowID(workflowID)
	if err != nil {
		return err
	}
	if len(orderedIDs) != len(existing) {
		return fmt.Errorf("ステータス ID の件数が一致しません")
	}
	want := make(map[uuid.UUID]struct{}, len(existing))
	for _, e := range existing {
		want[e.ID] = struct{}{}
	}
	for _, id := range orderedIDs {
		if _, ok := want[id]; !ok {
			return fmt.Errorf("無効なステータス ID が含まれます")
		}
		delete(want, id)
	}
	if len(want) != 0 {
		return fmt.Errorf("ステータス ID が不足しています")
	}
	return s.statusRepo.ReorderWorkflow(workflowID, orderedIDs)
}
