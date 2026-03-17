package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type ApprovalRepository interface {
	FindByIssueID(issueID uuid.UUID) ([]model.IssueApproval, error)
	FindByID(id uuid.UUID) (*model.IssueApproval, error)
	Create(approval *model.IssueApproval) error
	Update(approval *model.IssueApproval) error
}

type approvalRepository struct {
	db *gorm.DB
}

func NewApprovalRepository(db *gorm.DB) ApprovalRepository {
	return &approvalRepository{db: db}
}

func (r *approvalRepository) FindByIssueID(issueID uuid.UUID) ([]model.IssueApproval, error) {
	var approvals []model.IssueApproval
	err := r.db.
		Preload("WorkflowStep.Status").
		Preload("Approver").
		Where("issue_id = ?", issueID).
		Find(&approvals).Error
	return approvals, err
}

func (r *approvalRepository) FindByID(id uuid.UUID) (*model.IssueApproval, error) {
	var approval model.IssueApproval
	err := r.db.
		Preload("WorkflowStep.Status").
		Preload("Approver").
		First(&approval, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &approval, nil
}

func (r *approvalRepository) Create(approval *model.IssueApproval) error {
	return r.db.Create(approval).Error
}

func (r *approvalRepository) Update(approval *model.IssueApproval) error {
	return r.db.Model(&model.IssueApproval{}).Where("id = ?", approval.ID).Updates(map[string]interface{}{
		"status":      approval.Status,
		"approver_id": approval.ApproverID,
		"comment":     approval.Comment,
		"acted_at":    approval.ActedAt,
	}).Error
}
