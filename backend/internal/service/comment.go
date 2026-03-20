package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type CommentService interface {
	List(issueID uuid.UUID) ([]model.Comment, error)
	Create(issueID, authorID uuid.UUID, body string) (*model.Comment, error)
	Update(id uuid.UUID, body string) (*model.Comment, error)
	Delete(id uuid.UUID) error
}

type commentService struct {
	commentRepo repository.CommentRepository
	issueRepo   repository.IssueRepository
}

func NewCommentService(commentRepo repository.CommentRepository, issueRepo repository.IssueRepository) CommentService {
	return &commentService{commentRepo: commentRepo, issueRepo: issueRepo}
}

func (s *commentService) List(issueID uuid.UUID) ([]model.Comment, error) {
	return s.commentRepo.FindByIssue(issueID)
}

func (s *commentService) Create(issueID, authorID uuid.UUID, body string) (*model.Comment, error) {
	issue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		return nil, err
	}
	comment := &model.Comment{
		ID:             uuid.New(),
		OrganizationID: issue.OrganizationID,
		IssueID:        issueID,
		AuthorID:       authorID,
		Body:           body,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}
	// Author情報を含めて再取得
	if err := s.commentRepo.FindByID(comment.ID, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *commentService) Update(id uuid.UUID, body string) (*model.Comment, error) {
	comment := &model.Comment{}
	if err := s.commentRepo.FindByID(id, comment); err != nil {
		return nil, err
	}
	comment.Body = body
	comment.UpdatedAt = time.Now()
	if err := s.commentRepo.Update(comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *commentService) Delete(id uuid.UUID) error {
	return s.commentRepo.Delete(id)
}
