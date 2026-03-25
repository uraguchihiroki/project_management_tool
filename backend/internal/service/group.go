package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/pkg/keygen"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type GroupService interface {
	List(orgID uuid.UUID, kind *string) ([]model.Group, error)
	Create(orgID uuid.UUID, name string, kind *string, displayOrder int) (*model.Group, error)
	Get(id uuid.UUID) (*model.Group, error)
	Update(id uuid.UUID, name string, kind *string, displayOrder int) (*model.Group, error)
	Delete(id uuid.UUID) error
	ReplaceMembers(groupID uuid.UUID, userIDs []uuid.UUID) error
	ListMembers(groupID uuid.UUID) ([]uuid.UUID, error)
	ListGroupsByUser(userID uuid.UUID) ([]model.Group, error)
}

type groupService struct {
	groupRepo repository.GroupRepository
	ugRepo    repository.UserGroupRepository
	userRepo  repository.UserRepository
}

func NewGroupService(
	groupRepo repository.GroupRepository,
	ugRepo repository.UserGroupRepository,
	userRepo repository.UserRepository,
) GroupService {
	return &groupService{groupRepo: groupRepo, ugRepo: ugRepo, userRepo: userRepo}
}

func (s *groupService) List(orgID uuid.UUID, kind *string) ([]model.Group, error) {
	return s.groupRepo.ListByOrg(orgID, kind)
}

func (s *groupService) Create(orgID uuid.UUID, name string, kind *string, displayOrder int) (*model.Group, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	id := uuid.New()
	g := &model.Group{
		ID:             id,
		Key:            keygen.UUIDKey(id),
		OrganizationID: orgID,
		Name:           name,
		Kind:           kind,
		DisplayOrder:   displayOrder,
		CreatedAt:      time.Now(),
	}
	if err := s.groupRepo.Create(g); err != nil {
		return nil, err
	}
	return s.groupRepo.FindByID(g.ID)
}

func (s *groupService) Get(id uuid.UUID) (*model.Group, error) {
	return s.groupRepo.FindByID(id)
}

func (s *groupService) Update(id uuid.UUID, name string, kind *string, displayOrder int) (*model.Group, error) {
	g, err := s.groupRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	g.Name = name
	g.Kind = kind
	g.DisplayOrder = displayOrder
	if err := s.groupRepo.Update(g); err != nil {
		return nil, err
	}
	return s.groupRepo.FindByID(id)
}

func (s *groupService) Delete(id uuid.UUID) error {
	return s.groupRepo.Delete(id)
}

func (s *groupService) ReplaceMembers(groupID uuid.UUID, userIDs []uuid.UUID) error {
	g, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return err
	}
	for _, uid := range userIDs {
		u, err := s.userRepo.FindByID(uid)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}
		if u.OrganizationID != g.OrganizationID {
			return fmt.Errorf("user %s does not belong to group organization", uid)
		}
	}
	return s.ugRepo.ReplaceMembers(groupID, userIDs)
}

func (s *groupService) ListMembers(groupID uuid.UUID) ([]uuid.UUID, error) {
	return s.ugRepo.ListMemberIDs(groupID)
}

func (s *groupService) ListGroupsByUser(userID uuid.UUID) ([]model.Group, error) {
	return s.ugRepo.ListGroupsByUser(userID)
}
