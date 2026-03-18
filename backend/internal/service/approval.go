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
	ApproveStep(issueID uuid.UUID, stepID uint, approverID uuid.UUID, comment string) (*model.IssueApproval, error)
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
// min_approvers=1 のステップのみ 1 件の pending を作成。min_approvers>1 は作成しない
func (s *approvalService) InitializeForIssue(issueID uuid.UUID, workflowID uint) error {
	workflow, err := s.workflowRepo.FindByID(workflowID)
	if err != nil {
		return fmt.Errorf("workflow not found: %w", err)
	}
	for _, step := range workflow.Steps {
		if step.MinApprovers < 1 {
			step.MinApprovers = 1
		}
		if step.MinApprovers == 1 {
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
	workflow, err := s.workflowRepo.FindByID(step.WorkflowID)
	if err != nil {
		return nil, err
	}
	for i := range workflow.Steps {
		otherStep := &workflow.Steps[i]
		if otherStep.Order >= step.Order {
			continue
		}
		if !s.isStepComplete(approval.IssueID, otherStep) {
			return nil, fmt.Errorf("前のステップ「%s」がまだ承認されていません", otherStep.Name)
		}
	}

	// 承認者チェック（却下は任意のユーザーが可、承認は条件チェックが必要）
	if action == "approved" {
		if err := s.checkApproverEligible(approval.IssueID, step, approverID); err != nil {
			return nil, err
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

	// 承認時: ステップ完了なら status_id を更新
	if action == "approved" && step.StatusID != nil {
		if s.isStepComplete(approval.IssueID, step) {
			_ = s.issueRepo.UpdateStatus(approval.IssueID, *step.StatusID)
		}
	}

	return s.approvalRepo.FindByID(approvalID)
}

func (s *approvalService) checkApproverEligible(issueID uuid.UUID, step *model.WorkflowStep, approverID uuid.UUID) error {
	issue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		return err
	}
	if step.ExcludeReporter && issue.ReporterID == approverID {
		return fmt.Errorf("起票者は承認者になれません")
	}
	if step.ExcludeAssignee && issue.AssigneeID != nil && *issue.AssigneeID == approverID {
		return fmt.Errorf("担当者は承認者になれません")
	}
	switch step.ApproverType {
	case "user":
		if step.ApproverUserID == nil || *step.ApproverUserID != approverID {
			return fmt.Errorf("このステップは指定ユーザーのみ承認可能です")
		}
	case "role", "":
		roles, err := s.roleRepo.FindRolesByUserID(approverID)
		if err != nil {
			return err
		}
		maxLevel := 0
		for _, r := range roles {
			if r.Level > maxLevel {
				maxLevel = r.Level
			}
		}
		if maxLevel < step.RequiredLevel {
			return fmt.Errorf("承認権限が不足しています（必要Level: %d、あなたのLevel: %d）", step.RequiredLevel, maxLevel)
		}
	case "multiple":
		// 複数人: 誰でも可（exclude チェックは上で済み）
		break
	default:
		return fmt.Errorf("不明な承認タイプ: %s", step.ApproverType)
	}
	return nil
}

func (s *approvalService) isStepComplete(issueID uuid.UUID, step *model.WorkflowStep) bool {
	all, err := s.approvalRepo.FindByIssueID(issueID)
	if err != nil {
		return false
	}
	count := 0
	for _, a := range all {
		if a.WorkflowStepID == step.ID && a.Status == "approved" {
			count++
		}
	}
	return count >= step.MinApprovers
}

// ApproveStep はステップに対する承認を行う。min_approvers>1 の場合は新規 IssueApproval を作成
func (s *approvalService) ApproveStep(issueID uuid.UUID, stepID uint, approverID uuid.UUID, comment string) (*model.IssueApproval, error) {
	step, err := s.workflowRepo.FindStepByID(stepID)
	if err != nil {
		return nil, fmt.Errorf("workflow step not found: %w", err)
	}
	if err := s.checkApproverEligible(issueID, step, approverID); err != nil {
		return nil, err
	}
	allApprovals, err := s.approvalRepo.FindByIssueID(issueID)
	if err != nil {
		return nil, err
	}
	// 前ステップ完了チェック
	for _, a := range allApprovals {
		otherStep, err := s.workflowRepo.FindStepByID(a.WorkflowStepID)
		if err != nil {
			continue
		}
		if otherStep.Order < step.Order && !s.isStepComplete(issueID, otherStep) {
			return nil, fmt.Errorf("前のステップ「%s」がまだ承認されていません", otherStep.Name)
		}
	}
	// 同一ユーザー重複チェック
	for _, a := range allApprovals {
		if a.WorkflowStepID == step.ID && a.ApproverID != nil && *a.ApproverID == approverID {
			return nil, fmt.Errorf("このステップはすでに承認済みです")
		}
	}
	// 却下チェック
	for _, a := range allApprovals {
		if a.WorkflowStepID == step.ID && a.Status == "rejected" {
			return nil, fmt.Errorf("このステップは却下されています")
		}
	}
	// 新規承認レコード作成
	now := time.Now()
	approval := &model.IssueApproval{
		ID:             uuid.New(),
		IssueID:        issueID,
		WorkflowStepID: stepID,
		ApproverID:     &approverID,
		Status:         "approved",
		Comment:        comment,
		ActedAt:        &now,
		CreatedAt:      now,
	}
	if err := s.approvalRepo.Create(approval); err != nil {
		return nil, err
	}
	if step.StatusID != nil && s.isStepComplete(issueID, step) {
		_ = s.issueRepo.UpdateStatus(issueID, *step.StatusID)
	}
	return s.approvalRepo.FindByID(approval.ID)
}
