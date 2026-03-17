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
	Create(name string) (*model.Organization, error)
	GetUserOrganizations(userID uuid.UUID) ([]model.Organization, error)
	AddUser(orgID, userID uuid.UUID, isOrgAdmin bool) error
}

type organizationService struct {
	orgRepo repository.OrganizationRepository
}

func NewOrganizationService(orgRepo repository.OrganizationRepository) OrganizationService {
	return &organizationService{orgRepo: orgRepo}
}

func (s *organizationService) List() ([]model.Organization, error) {
	return s.orgRepo.FindAll()
}

func (s *organizationService) Get(id uuid.UUID) (*model.Organization, error) {
	return s.orgRepo.FindByID(id)
}

func (s *organizationService) Create(name string) (*model.Organization, error) {
	org := &model.Organization{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
	}
	if err := s.orgRepo.Create(org); err != nil {
		return nil, err
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
