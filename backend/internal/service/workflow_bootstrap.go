package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/pkg/keygen"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

// CreateOrgIssueWorkflowWithDefaultStatuses は組織に紐づく Issue 用ワークフロー行と、未着手・進行・完了の 3 ステータス行のみを作成する。
// projects テーブルや default_workflow_id は更新しない。許可遷移は呼び出し側で SeedDefaultIssueWorkflowTransitions を使う。
func CreateOrgIssueWorkflowWithDefaultStatuses(
	workflowRepo repository.WorkflowRepository,
	statusRepo repository.StatusRepository,
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
		{"進行", "#3B82F6", 2},
		{"完了", "#10B981", 3},
	}
	ids := make([]uuid.UUID, 0, len(defaultStatuses))
	for _, ds := range defaultStatuses {
		sid := uuid.New()
		st := &model.Status{
			ID:           sid,
			Key:          "sts-" + sid.String(),
			WorkflowID:   wf.ID,
			Name:         ds.Name,
			Color:        ds.Color,
			DisplayOrder: ds.Order,
		}
		if err := statusRepo.Create(st); err != nil {
			return 0, nil, err
		}
		ids = append(ids, sid)
	}
	return wf.ID, ids, nil
}

// SeedDefaultIssueWorkflowTransitions は「未着手・進行・完了」の 3 ステータス（CreateOrgIssueWorkflowWithDefaultStatuses と同一順）向けに許可遷移 4 本を追加する。既に存在する辺はスキップ（冪等寄り）。
// 辺: 未着手↔進行、進行↔完了（未着手→完了の直送は含めない）。
func SeedDefaultIssueWorkflowTransitions(
	transitionRepo repository.WorkflowTransitionRepository,
	workflowID uint,
	statusIDs []uuid.UUID,
) error {
	if len(statusIDs) != 3 {
		return fmt.Errorf("SeedDefaultIssueWorkflowTransitions: want 3 status ids, got %d", len(statusIDs))
	}
	a, b, c := statusIDs[0], statusIDs[1], statusIDs[2]
	pairs := []struct{ from, to uuid.UUID }{
		{a, b}, {b, a}, {b, c}, {c, b},
	}
	for _, p := range pairs {
		if transitionRepo.Exists(workflowID, p.from, p.to) {
			continue
		}
		nextDO, err := transitionRepo.MaxDisplayOrder(workflowID)
		if err != nil {
			return err
		}
		row := &model.WorkflowTransition{
			Key:          keygen.UUIDKey(uuid.New()),
			WorkflowID:   workflowID,
			FromStatusID: p.from,
			ToStatusID:   p.to,
			DisplayOrder: nextDO + 1,
			CreatedAt:    time.Now(),
		}
		if err := transitionRepo.Create(row); err != nil {
			return err
		}
	}
	return nil
}
