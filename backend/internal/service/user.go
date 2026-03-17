package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type UserService interface {
	List() ([]model.User, error)
	ListWithRoles() ([]model.User, error)
	ListByOrg(orgID uuid.UUID) ([]model.User, error)
	Get(id uuid.UUID) (*model.User, error)
	Create(name, email string) (*model.User, error)
	CreateForOrg(orgID uuid.UUID, name, email string) (*model.User, error)
	Update(id uuid.UUID, name string) error
	SetAdmin(id uuid.UUID, isAdmin bool) error
	RemoveFromOrg(orgID, userID uuid.UUID) error
}

type userService struct {
	userRepo repository.UserRepository
	orgRepo  repository.OrganizationRepository
}

func NewUserService(userRepo repository.UserRepository, orgRepo repository.OrganizationRepository) UserService {
	return &userService{userRepo: userRepo, orgRepo: orgRepo}
}

func (s *userService) List() ([]model.User, error) {
	return s.userRepo.FindAll()
}

func (s *userService) ListWithRoles() ([]model.User, error) {
	return s.userRepo.FindAllWithRoles()
}

func (s *userService) ListByOrg(orgID uuid.UUID) ([]model.User, error) {
	return s.userRepo.FindByOrg(orgID)
}

func (s *userService) Get(id uuid.UUID) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *userService) Create(name, email string) (*model.User, error) {
	// 最初のユーザーを自動的に管理者にする
	count, err := s.userRepo.Count()
	if err != nil {
		return nil, err
	}
	user := &model.User{
		ID:        uuid.New(),
		Name:      name,
		Email:     email,
		IsAdmin:   count == 0,
		CreatedAt: time.Now(),
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) CreateForOrg(orgID uuid.UUID, name, email string) (*model.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		user = &model.User{
			ID:        uuid.New(),
			Name:      name,
			Email:     email,
			CreatedAt: time.Now(),
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, err
		}
	}
	if err := s.orgRepo.AddUser(&model.OrganizationUser{
		OrganizationID: orgID,
		UserID:         user.ID,
		IsOrgAdmin:     false,
	}); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Update(id uuid.UUID, name string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	user.Name = name
	return s.userRepo.Update(user)
}

func (s *userService) SetAdmin(id uuid.UUID, isAdmin bool) error {
	return s.userRepo.UpdateAdmin(id, isAdmin)
}

func (s *userService) RemoveFromOrg(orgID, userID uuid.UUID) error {
	return s.orgRepo.RemoveUser(orgID, userID)
}
