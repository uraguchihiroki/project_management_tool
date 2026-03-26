package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

// SeedDefaultProjectStatuses はプロジェクトにデフォルトの進行ステータス列を作成し、先頭ステータスの ID を返す（許可遷移は作らない）
func SeedDefaultProjectStatuses(
	psRepo repository.ProjectStatusRepository,
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
	var first uuid.UUID
	for i, d := range defaults {
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
		if i == 0 {
			first = sid
		}
	}
	return first, nil
}
