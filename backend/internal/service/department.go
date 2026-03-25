package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/pkg/keygen"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type GroupService interface {
	ListByOrganization(orgID uuid.UUID) ([]model.Group, error)
	Get(id uuid.UUID) (*model.Group, error)
	Create(orgID uuid.UUID, name string) (*model.Group, error)
	Update(id uuid.UUID, name string) (*model.Group, error)
	Delete(id uuid.UUID) error
	Reorder(orgID uuid.UUID, ids []uuid.UUID) error
	GetUserGroups(orgID, userID uuid.UUID) ([]model.Group, error)
	SetUserGroups(orgID, userID uuid.UUID, groupIDs []uuid.UUID) error
}

type groupService struct {
	groupRepo repository.GroupRepository
	orgRepo  repository.OrganizationRepository
}

func NewGroupService(groupRepo repository.GroupRepository, orgRepo repository.OrganizationRepository) GroupService {
	return &groupService{groupRepo: groupRepo, orgRepo: orgRepo}
}

func (s *groupService) ListByOrganization(orgID uuid.UUID) ([]model.Group, error) {
	return s.groupRepo.FindByOrganizationID(orgID)
}

func (s *groupService) Get(id uuid.UUID) (*model.Group, error) {
	return s.groupRepo.FindByID(id)
}

func (s *groupService) Create(orgID uuid.UUID, name string) (*model.Group, error) {
	maxOrder, err := s.groupRepo.GetMaxOrder(orgID)
	if err != nil {
		return nil, err
	}
	groupID := uuid.New()
	key := keygen.Slug(name)
	if key == "" {
		key = keygen.UUIDKey(groupID)
	}
	g := &model.Group{
		ID:             groupID,
		Key:            key,
		OrganizationID: orgID,
		Name:           name,
		Order:          maxOrder + 1,
		CreatedAt:      time.Now(),
	}
	if err := s.groupRepo.Create(g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *groupService) Update(id uuid.UUID, name string) (*model.Group, error) {
	g, err := s.groupRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		g.Name = name
	}
	if err := s.groupRepo.Update(g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *groupService) Reorder(orgID uuid.UUID, ids []uuid.UUID) error {
	return s.groupRepo.Reorder(orgID, ids)
}

func (s *groupService) Delete(id uuid.UUID) error {
	return s.groupRepo.Delete(id)
}

func (s *groupService) GetUserGroups(orgID, userID uuid.UUID) ([]model.Group, error) {
	return s.groupRepo.FindUserGroups(orgID, userID)
}

func (s *groupService) SetUserGroups(orgID, userID uuid.UUID, groupIDs []uuid.UUID) error {
	return s.groupRepo.SetUserGroups(orgID, userID, groupIDs)
}
