package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
	"gorm.io/gorm"
)

type SuperAdminService interface {
	FindByEmail(email string) (*model.SuperAdmin, error)
	Create(name, email string) (*model.SuperAdmin, error)
}

type superAdminService struct {
	repo repository.SuperAdminRepository
}

func NewSuperAdminService(repo repository.SuperAdminRepository) SuperAdminService {
	return &superAdminService{repo: repo}
}

func (s *superAdminService) FindByEmail(email string) (*model.SuperAdmin, error) {
	return s.repo.FindByEmail(email)
}

func (s *superAdminService) Create(name, email string) (*model.SuperAdmin, error) {
	if _, err := s.repo.FindByEmail(email); err == nil {
		return nil, ErrDuplicateSuperAdminEmail
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	admin := &model.SuperAdmin{
		ID:        uuid.New(),
		Key:       email,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
	}
	if err := s.repo.Create(admin); err != nil {
		return nil, err
	}
	return admin, nil
}
