package repository

import (
	"errors"

	"github.com/HumaCode/go-rbac-menu/internal/domain"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindActiveByUsername(username string) (*domain.User, error)
	FindActiveByID(id string) (*domain.User, error)
	FindByID(id string) (*domain.User, error)
	ListActive() ([]domain.User, error)
	Create(user *domain.User) error
	Update(user *domain.User) error
	AssignRole(userID, roleID string) error
	RemoveRole(userID, roleID string) error
	GetUserRoleIDs(userID string) ([]string, error)
	GetUserRoleCodes(userID string) ([]string, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindActiveByUsername(username string) (*domain.User, error) {
	var user domain.User
	err := r.db.
		Preload("Roles", "is_active = 1").
		Where("username = ? AND is_active = 1", username).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindActiveByID(id string) (*domain.User, error) {
	var user domain.User
	err := r.db.
		Preload("Roles", "is_active = 1").
		Where("id = ? AND is_active = 1", id).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id string) (*domain.User, error) {
	var user domain.User
	err := r.db.Preload("Roles").Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) ListActive() ([]domain.User, error) {
	var users []domain.User
	err := r.db.
		Preload("Roles", "is_active = 1").
		Where("is_active = 1").
		Order("created_at desc").
		Find(&users).Error
	return users, err
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) AssignRole(userID, roleID string) error {
	var existing domain.UserRole
	err := r.db.Where("user_id = ? AND role_id = ?", userID, roleID).First(&existing).Error
	if err == nil {
		return nil // already assigned
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	userRole := domain.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.Create(&userRole).Error
}

func (r *userRepository) RemoveRole(userID, roleID string) error {
	return r.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&domain.UserRole{}).Error
}

func (r *userRepository) GetUserRoleIDs(userID string) ([]string, error) {
	var roleIDs []string
	err := r.db.Table("user_roles").
		Select("user_roles.role_id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.is_active = 1", userID).
		Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

func (r *userRepository) GetUserRoleCodes(userID string) ([]string, error) {
	var codes []string
	err := r.db.Table("roles").
		Select("roles.code").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.is_active = 1", userID).
		Pluck("code", &codes).Error
	return codes, err
}

