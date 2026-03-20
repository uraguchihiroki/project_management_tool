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
	FindByEmail(email string) (*model.User, error)
	FindAllByEmail(email string) ([]model.User, error)
	FindByEmailAndOrg(orgID uuid.UUID, email string) (*model.User, error)
	FindByOrg(orgID uuid.UUID) ([]model.User, error)
	Create(user *model.User) error
	Update(user *model.User) error
	UpdateAdmin(id uuid.UUID, isAdmin bool) error
	Delete(id uuid.UUID) error
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

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAllByEmail(email string) ([]model.User, error) {
	var users []model.User
	err := r.db.Where("email = ?", email).Preload("Organization").Find(&users).Error
	return users, err
}

func (r *userRepository) FindByEmailAndOrg(orgID uuid.UUID, email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("organization_id = ? AND email = ?", orgID, email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByOrg(orgID uuid.UUID) ([]model.User, error) {
	var users []model.User
	err := r.db.Where("organization_id = ?", orgID).
		Preload("Roles", "organization_id = ?", orgID).
		Find(&users).Error
	return users, err
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.User{}, "id = ?", id).Error
}

func (r *userRepository) FindAllWithRoles() ([]model.User, error) {
	var users []model.User
	err := r.db.Preload("Roles").Find(&users).Error
	return users, err
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *model.User) error {
	return r.db.Model(user).Updates(map[string]interface{}{
		"name":  user.Name,
		"email": user.Email,
	}).Error
}

func (r *userRepository) UpdateAdmin(id uuid.UUID, isAdmin bool) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("is_admin", isAdmin).Error
}

func (r *userRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Count(&count).Error
	return count, err
}
