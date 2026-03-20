package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/pkg/keygen"
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
// 各ステップに1件の pending を作成（承認オブジェクトありのステップ用）
func (s *approvalService) InitializeForIssue(issueID uuid.UUID, workflowID uint) error {
	issue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		return fmt.Errorf("issue not found: %w", err)
	}
	workflow, err := s.workflowRepo.FindByID(workflowID)
	if err != nil {
		return fmt.Errorf("workflow not found: %w", err)
	}
	for _, step := range workflow.Steps {
		// sts_start, sts_goal は承認レコードを作成しない（常に通過）
		if step.Status != nil && (step.Status.StatusKey == "sts_start" || step.Status.StatusKey == "sts_goal") {
			continue
		}
		approvalID := uuid.New()
		approval := &model.IssueApproval{
			ID:             approvalID,
			Key:            keygen.UUIDKey(approvalID),
			OrganizationID: issue.OrganizationID,
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

	// 前のステップ（next_status_id == このステップの status_id）が承認済みかチェック
	workflow, err := s.workflowRepo.FindByID(step.WorkflowID)
	if err != nil {
		return nil, err
	}
	for i := range workflow.Steps {
		otherStep := &workflow.Steps[i]
		if otherStep.NextStatusID == nil || *otherStep.NextStatusID != step.StatusID {
			continue
		}
		// sts_start, sts_goal は常に通過済みとみなす
		if otherStep.Status != nil && (otherStep.Status.StatusKey == "sts_start" || otherStep.Status.StatusKey == "sts_goal") {
			continue
		}
		if !s.isStepComplete(approval.IssueID, otherStep) {
			prevName := otherStep.StatusID.String()
			if otherStep.Status != nil {
				prevName = otherStep.Status.Name
			}
			return nil, fmt.Errorf("前のステップ「%s」がまだ承認されていません", prevName)
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

	// 承認時: ステップ完了なら next_status_id へ遷移
	if action == "approved" && step.NextStatusID != nil {
		if s.isStepComplete(approval.IssueID, step) {
			_ = s.issueRepo.UpdateStatus(approval.IssueID, *step.NextStatusID)
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
	// approval_objects の role/user で承認者チェック
	if len(step.ApprovalObjects) == 0 {
		return nil // 承認オブジェクトなし＝誰でも可
	}
	eligible := false
	for _, ao := range step.ApprovalObjects {
		if ao.Type == "user" && ao.UserID != nil && *ao.UserID == approverID {
			eligible = true
			break
		}
		if ao.Type == "role" && ao.RoleID != nil {
			reqRole, err := s.roleRepo.FindByID(*ao.RoleID)
			if err != nil {
				continue
			}
			roles, err := s.roleRepo.FindRolesByUserID(approverID)
			if err != nil {
				continue
			}
			for _, r := range roles {
				if r.ID == *ao.RoleID {
					if ao.RoleOperator == "eq" && r.Level == reqRole.Level {
						eligible = true
						break
					}
					if (ao.RoleOperator == "gte" || ao.RoleOperator == "") && r.Level >= reqRole.Level {
						eligible = true
						break
					}
				}
			}
		}
	}
	if !eligible {
		return fmt.Errorf("このステップの承認者として登録されていません")
	}
	return nil
}

func (s *approvalService) isStepComplete(issueID uuid.UUID, step *model.WorkflowStep) bool {
	all, err := s.approvalRepo.FindByIssueID(issueID)
	if err != nil {
		return false
	}
	if len(step.ApprovalObjects) == 0 {
		// 承認オブジェクトなし＝1件でも承認あれば完了
		for _, a := range all {
			if a.WorkflowStepID == step.ID && a.Status == "approved" {
				return true
			}
		}
		return false
	}
	// 承認者ごとの最大 points を加算し、合計が threshold 以上で完了
	sum := 0
	for _, a := range all {
		if a.WorkflowStepID != step.ID || a.Status != "approved" || a.ApproverID == nil {
			continue
		}
		maxPoints := 0
		for _, ao := range step.ApprovalObjects {
			if ao.Type == "user" && ao.UserID != nil && *a.ApproverID == *ao.UserID && ao.Points > maxPoints {
				maxPoints = ao.Points
			}
			if ao.Type == "role" && ao.RoleID != nil {
				reqRole, err := s.roleRepo.FindByID(*ao.RoleID)
				if err != nil {
					continue
				}
				roles, _ := s.roleRepo.FindRolesByUserID(*a.ApproverID)
				for _, r := range roles {
					if r.ID == *ao.RoleID {
						ok := ao.RoleOperator == "gte" || ao.RoleOperator == ""
						if ao.RoleOperator == "eq" && r.Level == reqRole.Level {
							ok = true
						}
						if ok && ao.Points > maxPoints {
							maxPoints = ao.Points
						}
						break
					}
				}
			}
		}
		sum += maxPoints
	}
	return sum >= step.Threshold
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
	// 前ステップ（next_status_id == このステップの status_id）完了チェック
	workflow, err := s.workflowRepo.FindByID(step.WorkflowID)
	if err != nil {
		return nil, err
	}
	for i := range workflow.Steps {
		otherStep := &workflow.Steps[i]
		if otherStep.NextStatusID == nil || *otherStep.NextStatusID != step.StatusID {
			continue
		}
		if !s.isStepComplete(issueID, otherStep) {
			prevName := otherStep.StatusID.String()
			if otherStep.Status != nil {
				prevName = otherStep.Status.Name
			}
			return nil, fmt.Errorf("前のステップ「%s」がまだ承認されていません", prevName)
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
	issue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		return nil, fmt.Errorf("issue not found: %w", err)
	}
	now := time.Now()
	approvalID := uuid.New()
	approval := &model.IssueApproval{
		ID:             approvalID,
		Key:            keygen.UUIDKey(approvalID),
		OrganizationID: issue.OrganizationID,
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
	if step.NextStatusID != nil && s.isStepComplete(issueID, step) {
		_ = s.issueRepo.UpdateStatus(issueID, *step.NextStatusID)
	}
	return s.approvalRepo.FindByID(approval.ID)
}
