package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type ApprovalService interface {
	GetApprovals(issueID uuid.UUID) ([]model.IssueApproval, error)
	InitializeForIssue(issueID uuid.UUID, workflowID uint) error
	Approve(approvalID uuid.UUID, approverID uuid.UUID, comment string) (*model.IssueApproval, error)
	Reject(approvalID uuid.UUID, approverID uuid.UUID, comment string) (*model.IssueApproval, error)
}

type approvalService struct {
	approvalRepo repository.ApprovalRepository
	workflowRepo repository.WorkflowRepository
	issueRepo    repository.IssueRepository
	roleRepo     repository.RoleRepository
}

func NewApprovalService(
	approvalRepo repository.ApprovalRepository,
	workflowRepo repository.WorkflowRepository,
	issueRepo repository.IssueRepository,
	roleRepo repository.RoleRepository,
) ApprovalService {
	return &approvalService{
		approvalRepo: approvalRepo,
		workflowRepo: workflowRepo,
		issueRepo:    issueRepo,
		roleRepo:     roleRepo,
	}
}

func (s *approvalService) GetApprovals(issueID uuid.UUID) ([]model.IssueApproval, error) {
	return s.approvalRepo.FindByIssueID(issueID)
}

// InitializeForIssue はWorkflowのステップに基づいてIssue承認レコードを生成する
func (s *approvalService) InitializeForIssue(issueID uuid.UUID, workflowID uint) error {
	workflow, err := s.workflowRepo.FindByID(workflowID)
	if err != nil {
		return fmt.Errorf("workflow not found: %w", err)
	}
	for _, step := range workflow.Steps {
		approval := &model.IssueApproval{
			ID:             uuid.New(),
			IssueID:        issueID,
			WorkflowStepID: step.ID,
			Status:         "pending",
			CreatedAt:      time.Now(),
		}
		if err := s.approvalRepo.Create(approval); err != nil {
			return err
		}
	}
	return nil
}

func (s *approvalService) Approve(approvalID uuid.UUID, approverID uuid.UUID, comment string) (*model.IssueApproval, error) {
	return s.act(approvalID, approverID, comment, "approved")
}

func (s *approvalService) Reject(approvalID uuid.UUID, approverID uuid.UUID, comment string) (*model.IssueApproval, error) {
	return s.act(approvalID, approverID, comment, "rejected")
}

func (s *approvalService) act(approvalID uuid.UUID, approverID uuid.UUID, comment, action string) (*model.IssueApproval, error) {
	approval, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return nil, fmt.Errorf("approval not found: %w", err)
	}
	if approval.Status != "pending" {
		return nil, fmt.Errorf("このステップはすでに%sです", approval.Status)
	}

	step, err := s.workflowRepo.FindStepByID(approval.WorkflowStepID)
	if err != nil {
		return nil, fmt.Errorf("workflow step not found: %w", err)
	}

	// 前のステップがすべて承認済みかチェック
	allApprovals, err := s.approvalRepo.FindByIssueID(approval.IssueID)
	if err != nil {
		return nil, err
	}
	for _, a := range allApprovals {
		if a.ID == approvalID {
			continue
		}
		otherStep, err := s.workflowRepo.FindStepByID(a.WorkflowStepID)
		if err != nil {
			continue
		}
		if otherStep.Order < step.Order && a.Status == "pending" {
			return nil, fmt.Errorf("前のステップ「%s」がまだ承認されていません", otherStep.Name)
		}
	}

	// 承認者のレベルチェック（却下は任意のユーザーが可、承認はレベルチェックが必要）
	if action == "approved" {
		roles, err := s.roleRepo.FindRolesByUserID(approverID)
		if err != nil {
			return nil, err
		}
		maxLevel := 0
		for _, r := range roles {
			if r.Level > maxLevel {
				maxLevel = r.Level
			}
		}
		if maxLevel < step.RequiredLevel {
			return nil, fmt.Errorf("承認権限が不足しています（必要Level: %d、あなたのLevel: %d）", step.RequiredLevel, maxLevel)
		}
	}

	// 操作を記録
	now := time.Now()
	approval.Status = action
	approval.ApproverID = &approverID
	approval.Comment = comment
	approval.ActedAt = &now
	if err := s.approvalRepo.Update(approval); err != nil {
		return nil, err
	}

	// 承認時: ステップにstatus_idが設定されていればIssueのステータスを更新
	if action == "approved" && step.StatusID != nil {
		_ = s.issueRepo.UpdateStatus(approval.IssueID, *step.StatusID)
	}

	return s.approvalRepo.FindByID(approvalID)
}
