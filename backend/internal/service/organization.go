package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type OrganizationService interface {
	List() ([]model.Organization, error)
	Get(id uuid.UUID) (*model.Organization, error)
	Create(name, adminEmail, adminName string) (*model.Organization, error)
	GetUserOrganizations(userID uuid.UUID) ([]model.Organization, error)
	AddUser(orgID, userID uuid.UUID, isOrgAdmin bool) error
}

type organizationService struct {
	orgRepo  repository.OrganizationRepository
	userRepo repository.UserRepository
}

func NewOrganizationService(orgRepo repository.OrganizationRepository, userRepo repository.UserRepository) OrganizationService {
	return &organizationService{orgRepo: orgRepo, userRepo: userRepo}
}

func (s *organizationService) List() ([]model.Organization, error) {
	return s.orgRepo.FindAll()
}

func (s *organizationService) Get(id uuid.UUID) (*model.Organization, error) {
	return s.orgRepo.FindByID(id)
}

func (s *organizationService) Create(name, adminEmail, adminName string) (*model.Organization, error) {
	org := &model.Organization{
		ID:         uuid.New(),
		Name:       name,
		AdminEmail: adminEmail,
		CreatedAt:  time.Now(),
	}
	if err := s.orgRepo.Create(org); err != nil {
		return nil, err
	}
	if adminEmail != "" {
		user, err := s.userRepo.FindByEmail(adminEmail)
		if err != nil {
			user = &model.User{
				ID:        uuid.New(),
				Name:      adminName,
				Email:     adminEmail,
				IsAdmin:   true,
				CreatedAt: time.Now(),
			}
			if adminName == "" {
				user.Name = adminEmail
			}
			if err := s.userRepo.Create(user); err != nil {
				return org, err
			}
		} else {
			_ = s.userRepo.UpdateAdmin(user.ID, true)
		}
		_ = s.orgRepo.AddUser(&model.OrganizationUser{
			OrganizationID: org.ID,
			UserID:         user.ID,
			IsOrgAdmin:     true,
		})
	}
	return org, nil
}

func (s *organizationService) GetUserOrganizations(userID uuid.UUID) ([]model.Organization, error) {
	return s.orgRepo.FindByUserID(userID)
}

func (s *organizationService) AddUser(orgID, userID uuid.UUID, isOrgAdmin bool) error {
	orgUser := &model.OrganizationUser{
		OrganizationID: orgID,
		UserID:         userID,
		IsOrgAdmin:     isOrgAdmin,
	}
	return s.orgRepo.AddUser(orgUser)
}
