package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type CommentRepository interface {
	FindByIssue(issueID uuid.UUID) ([]model.Comment, error)
	FindByID(id uuid.UUID, comment *model.Comment) error
	Create(comment *model.Comment) error
	Update(comment *model.Comment) error
	Delete(id uuid.UUID) error
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) FindByID(id uuid.UUID, comment *model.Comment) error {
	return r.db.Preload("Author").First(comment, "id = ?", id).Error
}

func (r *commentRepository) FindByIssue(issueID uuid.UUID) ([]model.Comment, error) {
	var comments []model.Comment
	err := r.db.Preload("Author").Where("issue_id = ?", issueID).Order("created_at asc").Find(&comments).Error
	return comments, err
}

func (r *commentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

func (r *commentRepository) Update(comment *model.Comment) error {
	return r.db.Save(comment).Error
}

func (r *commentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Comment{}, "id = ?", id).Error
}
