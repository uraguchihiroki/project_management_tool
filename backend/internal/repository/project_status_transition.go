package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type ProjectStatusTransitionRepository interface {
	Exists(projectID uuid.UUID, fromID, toID uuid.UUID) bool
}

type projectStatusTransitionRepository struct {
	db *gorm.DB
}

func NewProjectStatusTransitionRepository(db *gorm.DB) ProjectStatusTransitionRepository {
	return &projectStatusTransitionRepository{db: db}
}

func (r *projectStatusTransitionRepository) Exists(projectID uuid.UUID, fromID, toID uuid.UUID) bool {
	var pt model.ProjectStatusTransition
	err := r.db.Where("project_id = ? AND from_project_status_id = ? AND to_project_status_id = ?", projectID, fromID, toID).
		First(&pt).Error
	return err == nil
}
