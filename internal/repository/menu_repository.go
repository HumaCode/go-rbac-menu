package repository

import (
	"errors"

	"go-rbac-menu/internal/domain"
	"gorm.io/gorm"
)

type MenuRepository interface {
	FindByID(id string) (*domain.Menu, error)
	ListActive() ([]domain.Menu, error)
	Create(menu *domain.Menu) error
	Update(menu *domain.Menu) error
	Delete(id string) error

	// Menu Permission linking
	AssignPermissionToMenu(menuID, permissionID string) (*domain.MenuPermission, error)
	GetMenuPermission(menuID, permissionID string) (*domain.MenuPermission, error)
	ListMenuPermissionsByMenuID(menuID string) ([]domain.MenuPermission, error)

	// Role Menu Permission linking
	AssignRoleMenuPermission(roleID, menuPermissionID string) error
	RemoveRoleMenuPermission(roleID, menuPermissionID string) error
	GetRoleMenuPermissions(roleIDs []string) ([]domain.RoleMenuPermission, error)

	// RBAC verification
	CheckUserPermission(userID, menuPath, permissionCode string) (bool, error)
}

type menuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &menuRepository{db: db}
}

func (r *menuRepository) FindByID(id string) (*domain.Menu, error) {
	var menu domain.Menu
	err := r.db.Where("id = ?", id).First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *menuRepository) ListActive() ([]domain.Menu, error) {
	var menus []domain.Menu
	err := r.db.
		Where("is_active = 1").
		Order("order_no asc, title asc").
		Find(&menus).Error
	return menus, err
}

func (r *menuRepository) Create(menu *domain.Menu) error {
	return r.db.Create(menu).Error
}

func (r *menuRepository) Update(menu *domain.Menu) error {
	return r.db.Save(menu).Error
}

func (r *menuRepository) Delete(id string) error {
	return r.db.Model(&domain.Menu{}).Where("id = ?", id).Update("is_active", 0).Error
}

func (r *menuRepository) AssignPermissionToMenu(menuID, permissionID string) (*domain.MenuPermission, error) {
	var existing domain.MenuPermission
	err := r.db.Where("menu_id = ? AND permission_id = ?", menuID, permissionID).First(&existing).Error
	if err == nil {
		if existing.IsActive != 1 {
			existing.IsActive = 1
			_ = r.db.Save(&existing)
		}
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	newMP := domain.MenuPermission{
		MenuID:       menuID,
		PermissionID: permissionID,
		IsActive:     1,
	}
	if err := r.db.Create(&newMP).Error; err != nil {
		return nil, err
	}
	return &newMP, nil
}

func (r *menuRepository) GetMenuPermission(menuID, permissionID string) (*domain.MenuPermission, error) {
	var mp domain.MenuPermission
	err := r.db.Where("menu_id = ? AND permission_id = ? AND is_active = 1", menuID, permissionID).First(&mp).Error
	if err != nil {
		return nil, err
	}
	return &mp, nil
}

func (r *menuRepository) ListMenuPermissionsByMenuID(menuID string) ([]domain.MenuPermission, error) {
	var list []domain.MenuPermission
	err := r.db.
		Preload("Permission", "is_active = 1").
		Where("menu_id = ? AND is_active = 1", menuID).
		Find(&list).Error
	return list, err
}

func (r *menuRepository) AssignRoleMenuPermission(roleID, menuPermissionID string) error {
	var existing domain.RoleMenuPermission
	err := r.db.Where("role_id = ? AND menu_permission_id = ?", roleID, menuPermissionID).First(&existing).Error
	if err == nil {
		if existing.IsActive != 1 {
			existing.IsActive = 1
			return r.db.Save(&existing).Error
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	rmp := domain.RoleMenuPermission{
		RoleID:           roleID,
		MenuPermissionID: menuPermissionID,
		IsActive:         1,
	}
	return r.db.Create(&rmp).Error
}

func (r *menuRepository) RemoveRoleMenuPermission(roleID, menuPermissionID string) error {
	return r.db.Where("role_id = ? AND menu_permission_id = ?", roleID, menuPermissionID).Delete(&domain.RoleMenuPermission{}).Error
}

func (r *menuRepository) GetRoleMenuPermissions(roleIDs []string) ([]domain.RoleMenuPermission, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}

	var results []domain.RoleMenuPermission
	err := r.db.
		Preload("MenuPermission", "is_active = 1").
		Preload("MenuPermission.Menu", "is_active = 1").
		Preload("MenuPermission.Permission", "is_active = 1").
		Where("role_id IN (?) AND is_active = 1", roleIDs).
		Find(&results).Error

	return results, err
}

func (r *menuRepository) CheckUserPermission(userID, menuPath, permissionCode string) (bool, error) {
	var count int64
	err := r.db.Table("users").
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Joins("JOIN roles ON roles.id = user_roles.role_id AND roles.is_active = 1").
		Joins("JOIN role_menu_permissions rmp ON rmp.role_id = roles.id AND rmp.is_active = 1").
		Joins("JOIN menu_permissions mp ON mp.id = rmp.menu_permission_id AND mp.is_active = 1").
		Joins("JOIN menus ON menus.id = mp.menu_id AND menus.is_active = 1").
		Joins("JOIN permissions p ON p.id = mp.permission_id AND p.is_active = 1").
		Where("users.id = ? AND users.is_active = 1", userID).
		Where("menus.path = ? AND p.code = ?", menuPath, permissionCode).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

