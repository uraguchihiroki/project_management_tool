package repository

import (
	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindAll() ([]model.User, error)
	FindAllWithRoles() ([]model.User, error)
	FindByID(id uuid.UUID) (*model.User, error)
	Create(user *model.User) error
	UpdateAdmin(id uuid.UUID, isAdmin bool) error
	Count() (int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindAll() ([]model.User, error) {
	var users []model.User
	err := r.db.Find(&users).Error
	return users, err
}

func (r *userRepository) FindByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAllWithRoles() ([]model.User, error) {
	var users []model.User
	err := r.db.Preload("Roles").Find(&users).Error
	return users, err
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) UpdateAdmin(id uuid.UUID, isAdmin bool) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("is_admin", isAdmin).Error
}

func (r *userRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Count(&count).Error
	return count, err
}
