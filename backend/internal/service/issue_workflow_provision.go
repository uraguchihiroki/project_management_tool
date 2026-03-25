package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

// IssueWorkflowProvisioner はプロジェクトに紐づくデフォルト Issue 用ワークフロー（未着手・進行・完了＋許可遷移 4 本）の初回確保のみを担う。
// ProjectService とは別系統。project_status_bootstrap と同一ファイルに置かない。
type IssueWorkflowProvisioner struct {
	projectRepo    repository.ProjectRepository
	workflowRepo   repository.WorkflowRepository
	statusRepo     repository.StatusRepository
	transitionRepo repository.WorkflowTransitionRepository
}

func NewIssueWorkflowProvisioner(
	projectRepo repository.ProjectRepository,
	workflowRepo repository.WorkflowRepository,
	statusRepo repository.StatusRepository,
	transitionRepo repository.WorkflowTransitionRepository,
) *IssueWorkflowProvisioner {
	return &IssueWorkflowProvisioner{
		projectRepo:    projectRepo,
		workflowRepo:   workflowRepo,
		statusRepo:     statusRepo,
		transitionRepo: transitionRepo,
	}
}

// EnsureDefaultForProject は DefaultWorkflowID が無い場合のみ、ワークフロー名を「{プロジェクト名} - Issue」として作成し紐付ける。既にあれば no-op。
func (p *IssueWorkflowProvisioner) EnsureDefaultForProject(projectID uuid.UUID) error {
	proj, err := p.projectRepo.FindByID(projectID)
	if err != nil {
		return err
	}
	if proj.DefaultWorkflowID != nil {
		return nil
	}
	wfName := proj.Name + " - Issue"
	wfID, statusIDs, err := CreateOrgIssueWorkflowWithDefaultStatuses(
		p.workflowRepo, p.statusRepo, proj.OrganizationID, wfName,
	)
	if err != nil {
		return err
	}
	if err := SeedDefaultIssueWorkflowTransitions(p.transitionRepo, wfID, statusIDs); err != nil {
		return fmt.Errorf("seed transitions: %w", err)
	}
	proj.DefaultWorkflowID = &wfID
	return p.projectRepo.Update(proj)
}
