package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type IssueRepository interface {
	FindByProject(projectID uuid.UUID) ([]model.Issue, error)
	FindByNumber(projectID uuid.UUID, number int) (*model.Issue, error)
	FindByID(id uuid.UUID) (*model.Issue, error)
	Create(issue *model.Issue) error
	Update(issue *model.Issue) error
	UpdateStatus(id uuid.UUID, statusID uuid.UUID) error
	Delete(id uuid.UUID) error
	NextNumber(projectID uuid.UUID) (int, error)
}

type issueRepository struct {
	db *gorm.DB
}

func NewIssueRepository(db *gorm.DB) IssueRepository {
	return &issueRepository{db: db}
}

func (r *issueRepository) FindByProject(projectID uuid.UUID) ([]model.Issue, error) {
	var issues []model.Issue
	err := r.db.
		Preload("Status").
		Preload("Assignee").
		Preload("Reporter").
		Where("project_id = ?", projectID).
		Order("number desc").
		Find(&issues).Error
	return issues, err
}

func (r *issueRepository) FindByNumber(projectID uuid.UUID, number int) (*model.Issue, error) {
	var issue model.Issue
	err := r.db.
		Preload("Status").
		Preload("Assignee").
		Preload("Reporter").
		Preload("Comments.Author").
		Where("project_id = ? AND number = ?", projectID, number).
		First(&issue).Error
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

func (r *issueRepository) FindByID(id uuid.UUID) (*model.Issue, error) {
	var issue model.Issue
	err := r.db.First(&issue, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

func (r *issueRepository) Create(issue *model.Issue) error {
	return r.db.Create(issue).Error
}

func (r *issueRepository) UpdateStatus(id uuid.UUID, statusID uuid.UUID) error {
	return r.db.Model(&model.Issue{}).Where("id = ?", id).Update("status_id", statusID).Error
}

func (r *issueRepository) Update(issue *model.Issue) error {
	return r.db.Model(&model.Issue{}).Where("id = ?", issue.ID).Updates(map[string]interface{}{
		"title":       issue.Title,
		"description": issue.Description,
		"status_id":   issue.StatusID,
		"priority":    issue.Priority,
		"assignee_id": issue.AssigneeID,
		"due_date":    issue.DueDate,
		"updated_at":  issue.UpdatedAt,
	}).Error
}

func (r *issueRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Issue{}, "id = ?", id).Error
}

func (r *issueRepository) NextNumber(projectID uuid.UUID) (int, error) {
	var maxNumber int
	r.db.Model(&model.Issue{}).Where("project_id = ?", projectID).Select("COALESCE(MAX(number), 0)").Scan(&maxNumber)
	return maxNumber + 1, nil
}
