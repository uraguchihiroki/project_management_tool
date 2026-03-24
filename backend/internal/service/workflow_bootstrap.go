package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/pkg/keygen"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

// CreateWorkflowWithIssueStatuses は組織スコープのワークフローと Issue 用デフォルトステータス列・全ペア遷移を作成する
func CreateWorkflowWithIssueStatuses(
	workflowRepo repository.WorkflowRepository,
	statusRepo repository.StatusRepository,
	transitionRepo repository.WorkflowTransitionRepository,
	orgID uuid.UUID,
	workflowName string,
) (uint, []uuid.UUID, error) {
	maxOrder, err := workflowRepo.GetMaxOrder()
	if err != nil {
		return 0, nil, err
	}
	wf := &model.Workflow{
		Key:            keygen.UUIDKey(uuid.New()),
		OrganizationID: orgID,
		Name:           workflowName,
		Description:    "",
		Order:          maxOrder + 1,
		CreatedAt:      time.Now(),
	}
	if err := workflowRepo.Create(wf); err != nil {
		return 0, nil, err
	}

	defaultStatuses := []struct {
		Name  string
		Color string
		Order int
	}{
		{"未着手", "#6B7280", 1},
		{"進行中", "#3B82F6", 2},
		{"レビュー中", "#F59E0B", 3},
		{"完了", "#10B981", 4},
	}
	ids := make([]uuid.UUID, 0, len(defaultStatuses))
	for _, ds := range defaultStatuses {
		sid := uuid.New()
		st := &model.Status{
			ID:         sid,
			Key:        "sts-" + sid.String(),
			WorkflowID: wf.ID,
			Name:       ds.Name,
			Color:      ds.Color,
			Order:      ds.Order,
		}
		if err := statusRepo.Create(st); err != nil {
			return 0, nil, err
		}
		ids = append(ids, sid)
	}
	if err := transitionRepo.SeedAllPairs(wf.ID, ids); err != nil {
		return 0, nil, err
	}
	return wf.ID, ids, nil
}
