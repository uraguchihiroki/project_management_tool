package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type UserService interface {
	List() ([]model.User, error)
	Get(id uuid.UUID) (*model.User, error)
	Create(name, email string) (*model.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) List() ([]model.User, error) {
	return s.userRepo.FindAll()
}

func (s *userService) Get(id uuid.UUID) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *userService) Create(name, email string) (*model.User, error) {
	user := &model.User{
		ID:        uuid.New(),
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}
