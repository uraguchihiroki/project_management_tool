package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

// SeedDefaultProjectStatuses はプロジェクトにデフォルトの進行ステータス列と全ペア遷移を作成し、先頭ステータスの ID を返す。
func SeedDefaultProjectStatuses(
	psRepo repository.ProjectStatusRepository,
	pstRepo repository.ProjectStatusTransitionRepository,
	projectID uuid.UUID,
) (firstID uuid.UUID, err error) {
	defaults := []struct {
		Name  string
		Color string
		Order int
	}{
		{"計画中", "#6B7280", 1},
		{"進行中", "#3B82F6", 2},
		{"完了", "#10B981", 3},
	}
	ids := make([]uuid.UUID, 0, len(defaults))
	for _, d := range defaults {
		sid := uuid.New()
		ps := &model.ProjectStatus{
			ID:        sid,
			Key:       "pst-" + sid.String(),
			ProjectID: projectID,
			Name:      d.Name,
			Color:     d.Color,
			Order:     d.Order,
		}
		if err := psRepo.Create(ps); err != nil {
			return uuid.Nil, fmt.Errorf("create project status: %w", err)
		}
		ids = append(ids, sid)
	}
	if err := pstRepo.SeedAllPairs(projectID, ids); err != nil {
		return uuid.Nil, err
	}
	return ids[0], nil
}
