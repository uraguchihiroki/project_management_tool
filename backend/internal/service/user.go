package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
	"gorm.io/gorm"
)

var ErrDuplicateEmailInOrg = errors.New("email already exists in organization")

type UserService interface {
	List() ([]model.User, error)
	ListWithRoles() ([]model.User, error)
	ListByOrg(orgID uuid.UUID) ([]model.User, error)
	Get(id uuid.UUID) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
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

func (s *userService) FindByEmail(email string) (*model.User, error) {
	return s.userRepo.FindByEmail(email)
}

func (s *userService) Create(name, email string) (*model.User, error) {
	// デフォルト組織（最初の組織）にユーザーを作成
	orgs, err := s.orgRepo.FindAll()
	if err != nil || len(orgs) == 0 {
		return nil, err
	}
	orgID := orgs[0].ID
	// 最初のユーザーを自動的に管理者にする
	count, err := s.userRepo.Count()
	if err != nil {
		return nil, err
	}
	userID := uuid.New()
	user := &model.User{
		ID:             userID,
		Key:            email,
		OrganizationID: orgID,
		Name:           name,
		Email:          email,
		IsAdmin:        count == 0,
		JoinedAt:       time.Now(),
		CreatedAt:      time.Now(),
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) CreateForOrg(orgID uuid.UUID, name, email string) (*model.User, error) {
	// 組織内で同一メールが既にいればエラー（1ユーザー＝1組織、組織内でemailユニーク）
	if existing, err := s.userRepo.FindByEmailAndOrg(orgID, email); err == nil && existing != nil {
		return nil, ErrDuplicateEmailInOrg
	}
	newUserID := uuid.New()
	user := &model.User{
		ID:             newUserID,
		Key:            email,
		OrganizationID: orgID,
		Name:           name,
		Email:          email,
		JoinedAt:       time.Now(),
		CreatedAt:      time.Now(),
	}
	if err := s.userRepo.Create(user); err != nil {
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
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if user.OrganizationID != orgID {
		return gorm.ErrRecordNotFound
	}
	return s.userRepo.Delete(userID)
}
