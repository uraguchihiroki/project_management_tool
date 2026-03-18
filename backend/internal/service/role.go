package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type RoleService interface {
	ListRoles(orgID *uuid.UUID) ([]model.Role, error)
	GetRole(id uint) (*model.Role, error)
	CreateRole(name string, level int, description string, orgID *uuid.UUID) (*model.Role, error)
	UpdateRole(id uint, name string, level int, description string) (*model.Role, error)
	DeleteRole(id uint) error
	Reorder(orgID *uuid.UUID, ids []uint) error
	AssignRolesToUser(userID uuid.UUID, roleIDs []uint) error
	GetUserRoles(userID uuid.UUID) ([]model.Role, error)
}

type roleService struct {
	roleRepo repository.RoleRepository
}

func NewRoleService(roleRepo repository.RoleRepository) RoleService {
	return &roleService{roleRepo: roleRepo}
}

func (s *roleService) ListRoles(orgID *uuid.UUID) ([]model.Role, error) {
	if orgID != nil {
		return s.roleRepo.FindByOrg(*orgID)
	}
	return s.roleRepo.FindAll()
}

func (s *roleService) GetRole(id uint) (*model.Role, error) {
	return s.roleRepo.FindByID(id)
}

func (s *roleService) CreateRole(name string, level int, description string, orgID *uuid.UUID) (*model.Role, error) {
	maxOrder, err := s.roleRepo.GetMaxOrder(orgID)
	if err != nil {
		return nil, err
	}
	role := &model.Role{
		Name:           name,
		Level:          level,
		Order:          maxOrder + 1,
		Description:    description,
		OrganizationID: orgID,
		CreatedAt:      time.Now(),
	}
	if err := s.roleRepo.Create(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) Reorder(orgID *uuid.UUID, ids []uint) error {
	return s.roleRepo.Reorder(orgID, ids)
}

func (s *roleService) UpdateRole(id uint, name string, level int, description string) (*model.Role, error) {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	role.Name = name
	role.Level = level
	role.Description = description
	if err := s.roleRepo.Update(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) DeleteRole(id uint) error {
	return s.roleRepo.Delete(id)
}

func (s *roleService) AssignRolesToUser(userID uuid.UUID, roleIDs []uint) error {
	return s.roleRepo.AssignRolesToUser(userID, roleIDs)
}

func (s *roleService) GetUserRoles(userID uuid.UUID) ([]model.Role, error) {
	return s.roleRepo.FindRolesByUserID(userID)
}
