package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type RoleService interface {
	ListRoles() ([]model.Role, error)
	GetRole(id uint) (*model.Role, error)
	CreateRole(name string, level int, description string) (*model.Role, error)
	UpdateRole(id uint, name string, level int, description string) (*model.Role, error)
	DeleteRole(id uint) error
	AssignRolesToUser(userID uuid.UUID, roleIDs []uint) error
	GetUserRoles(userID uuid.UUID) ([]model.Role, error)
}

type roleService struct {
	roleRepo repository.RoleRepository
}

func NewRoleService(roleRepo repository.RoleRepository) RoleService {
	return &roleService{roleRepo: roleRepo}
}

func (s *roleService) ListRoles() ([]model.Role, error) {
	return s.roleRepo.FindAll()
}

func (s *roleService) GetRole(id uint) (*model.Role, error) {
	return s.roleRepo.FindByID(id)
}

func (s *roleService) CreateRole(name string, level int, description string) (*model.Role, error) {
	role := &model.Role{
		Name:        name,
		Level:       level,
		Description: description,
		CreatedAt:   time.Now(),
	}
	if err := s.roleRepo.Create(role); err != nil {
		return nil, err
	}
	return role, nil
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
