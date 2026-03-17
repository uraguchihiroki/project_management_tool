package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type CreateIssueInput struct {
	Title       string
	Description *string
	StatusID    uuid.UUID
	Priority    string
	AssigneeID  *uuid.UUID
	ReporterID  uuid.UUID
}

type UpdateIssueInput struct {
	Title       *string
	Description *string
	StatusID    *uuid.UUID
	Priority    *string
	AssigneeID  *uuid.UUID
}

type IssueService interface {
	List(projectID uuid.UUID) ([]model.Issue, error)
	Get(projectID uuid.UUID, number int) (*model.Issue, error)
	Create(projectID uuid.UUID, input CreateIssueInput) (*model.Issue, error)
	Update(projectID uuid.UUID, number int, input UpdateIssueInput) (*model.Issue, error)
	Delete(projectID uuid.UUID, number int) error
}

type issueService struct {
	issueRepo   repository.IssueRepository
	projectRepo repository.ProjectRepository
}

func NewIssueService(issueRepo repository.IssueRepository, projectRepo repository.ProjectRepository) IssueService {
	return &issueService{issueRepo: issueRepo, projectRepo: projectRepo}
}

func (s *issueService) List(projectID uuid.UUID) ([]model.Issue, error) {
	return s.issueRepo.FindByProject(projectID)
}

func (s *issueService) Get(projectID uuid.UUID, number int) (*model.Issue, error) {
	return s.issueRepo.FindByNumber(projectID, number)
}

func (s *issueService) Create(projectID uuid.UUID, input CreateIssueInput) (*model.Issue, error) {
	// 採番
	nextNum, err := s.issueRepo.NextNumber(projectID)
	if err != nil {
		return nil, err
	}

	// デフォルト優先度
	priority := input.Priority
	if priority == "" {
		priority = "medium"
	}

	issue := &model.Issue{
		ID:          uuid.New(),
		Number:      nextNum,
		Title:       input.Title,
		Description: input.Description,
		StatusID:    input.StatusID,
		Priority:    priority,
		AssigneeID:  input.AssigneeID,
		ReporterID:  input.ReporterID,
		ProjectID:   projectID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.issueRepo.Create(issue); err != nil {
		return nil, err
	}
	// アソシエーションを含めて再取得
	return s.issueRepo.FindByNumber(projectID, issue.Number)
}

func (s *issueService) Update(projectID uuid.UUID, number int, input UpdateIssueInput) (*model.Issue, error) {
	issue, err := s.issueRepo.FindByNumber(projectID, number)
	if err != nil {
		return nil, err
	}
	if input.Title != nil {
		issue.Title = *input.Title
	}
	if input.Description != nil {
		issue.Description = input.Description
	}
	if input.StatusID != nil {
		issue.StatusID = *input.StatusID
	}
	if input.Priority != nil {
		issue.Priority = *input.Priority
	}
	if input.AssigneeID != nil {
		issue.AssigneeID = input.AssigneeID
	}
	issue.UpdatedAt = time.Now()
	if err := s.issueRepo.Update(issue); err != nil {
		return nil, err
	}
	return s.issueRepo.FindByNumber(projectID, number)
}

func (s *issueService) Delete(projectID uuid.UUID, number int) error {
	issue, err := s.issueRepo.FindByNumber(projectID, number)
	if err != nil {
		return err
	}
	return s.issueRepo.Delete(issue.ID)
}
