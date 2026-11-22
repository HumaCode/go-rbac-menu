package repository

import (
	"github.com/HumaCode/go-rbac-menu/internal/domain"
	"gorm.io/gorm"
)

type RoleRepository interface {
	FindByID(id string) (*domain.Role, error)
	FindByCode(code string) (*domain.Role, error)
	ListActive() ([]domain.Role, error)
	Create(role *domain.Role) error
	Update(role *domain.Role) error
	Delete(id string) error
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) FindByID(id string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByCode(code string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.Where("code = ?", code).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) ListActive() ([]domain.Role, error) {
	var roles []domain.Role
	err := r.db.Where("is_active = 1").Order("created_at asc").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) Create(role *domain.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) Update(role *domain.Role) error {
	return r.db.Save(role).Error
}

func (r *roleRepository) Delete(id string) error {
	// Soft deactivate: set is_active = 0
	return r.db.Model(&domain.Role{}).Where("id = ?", id).Update("is_active", 0).Error
}

