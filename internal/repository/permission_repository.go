package repository

import (
	"go-rbac-menu/internal/domain"
	"gorm.io/gorm"
)

type PermissionRepository interface {
	ListActive() ([]domain.Permission, error)
	FindByID(id string) (*domain.Permission, error)
	FindByCode(code string) (*domain.Permission, error)
	Create(permission *domain.Permission) error
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) ListActive() ([]domain.Permission, error) {
	var permissions []domain.Permission
	err := r.db.Where("is_active = 1").Order("created_at asc").Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) FindByID(id string) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.Where("id = ?", id).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) FindByCode(code string) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.Where("code = ?", code).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) Create(permission *domain.Permission) error {
	return r.db.Create(permission).Error
}

