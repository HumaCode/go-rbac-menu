package rbacmenu

import (
	"errors"
	"sort"

	"gorm.io/gorm"
)

// AutoMigrate migrates RBAC and Menu tables to any target database
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Role{},
		&UserRole{},
		&Permission{},
		&Menu{},
		&MenuPermission{},
		&RoleMenuPermission{},
	)
}

// Manager provides high-level RBAC & Dynamic Menu operations
type Manager struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

// GetNavMenuTree generates the hierarchical menu tree for given active role IDs.
// Strictly includes only records with is_active = 1.
func (m *Manager) GetNavMenuTree(roleIDs []string) ([]*Menu, error) {
	if len(roleIDs) == 0 {
		return []*Menu{}, nil
	}

	// 1. Fetch active role_menu_permissions
	var rmps []RoleMenuPermission
	err := m.db.
		Preload("MenuPermission", "is_active = 1").
		Preload("MenuPermission.Menu", "is_active = 1").
		Preload("MenuPermission.Permission", "is_active = 1").
		Where("role_id IN (?) AND is_active = 1", roleIDs).
		Find(&rmps).Error
	if err != nil {
		return nil, err
	}

	// 2. Map actions to menus
	menuActionMap := make(map[string]map[string]bool)
	for _, rmp := range rmps {
		mp := rmp.MenuPermission
		if mp.IsActive != 1 || mp.Menu.IsActive != 1 || mp.Permission.IsActive != 1 {
			continue
		}
		if _, ok := menuActionMap[mp.MenuID]; !ok {
			menuActionMap[mp.MenuID] = make(map[string]bool)
		}
		menuActionMap[mp.MenuID][mp.Permission.Code] = true
	}

	// 3. Fetch all active menus to resolve parents
	var allActiveMenus []Menu
	err = m.db.Where("is_active = 1").Order("order_no asc, title asc").Find(&allActiveMenus).Error
	if err != nil {
		return nil, err
	}
	allActiveMap := make(map[string]Menu)
	for _, menu := range allActiveMenus {
		allActiveMap[menu.ID] = menu
	}

	// 4. Ensure ancestor parents are included
	includedMenuIDs := make(map[string]bool)
	for mID := range menuActionMap {
		currID := mID
		for {
			includedMenuIDs[currID] = true
			parent, exists := allActiveMap[currID]
			if !exists || parent.ParentID == nil || *parent.ParentID == "" {
				break
			}
			currID = *parent.ParentID
		}
	}

	// 5. Build nodes
	nodeMap := make(map[string]*Menu)
	for mID := range includedMenuIDs {
		original, ok := allActiveMap[mID]
		if !ok {
			continue
		}

		var perms []string
		if actionSet, ok := menuActionMap[mID]; ok {
			for code := range actionSet {
				perms = append(perms, code)
			}
			sort.Strings(perms)
		}

		node := &Menu{
			BaseModel:   original.BaseModel,
			ParentID:    original.ParentID,
			Category:    original.Category,
			Type:        original.Type,
			Title:       original.Title,
			Path:        original.Path,
			Icon:        original.Icon,
			OrderNo:     original.OrderNo,
			IsActive:    original.IsActive,
			Permissions: perms,
			Children:    make([]*Menu, 0),
		}
		nodeMap[mID] = node
	}

	// 6. Assemble tree
	var rootMenus []*Menu
	for _, node := range nodeMap {
		if node.ParentID == nil || *node.ParentID == "" {
			rootMenus = append(rootMenus, node)
		} else if parentNode, exists := nodeMap[*node.ParentID]; exists {
			parentNode.Children = append(parentNode.Children, node)
		} else {
			rootMenus = append(rootMenus, node)
		}
	}

	// 7. Sort recursively by OrderNo
	sortTree(rootMenus)
	return rootMenus, nil
}

// GetNavMenuCategorized returns root menus grouped cleanly by Category.
// Each root menu contains its full nested hierarchy (Submenus, Childmenus, etc.).
func (m *Manager) GetNavMenuCategorized(roleIDs []string) ([]*NavCategory, error) {
	rootTree, err := m.GetNavMenuTree(roleIDs)
	if err != nil {
		return nil, err
	}

	return GroupMenusByCategory(rootTree), nil
}

// GroupMenusByCategory groups an existing tree of root menus by their Category attribute
func GroupMenusByCategory(rootMenus []*Menu) []*NavCategory {
	categoryMap := make(map[string]*NavCategory)
	var categoryOrder []string

	for _, menu := range rootMenus {
		catName := menu.Category
		if catName == "" {
			catName = "UTAMA"
		}

		if _, exists := categoryMap[catName]; !exists {
			categoryMap[catName] = &NavCategory{
				Category: catName,
				OrderNo:  menu.OrderNo,
				Menus:    make([]*Menu, 0),
			}
			categoryOrder = append(categoryOrder, catName)
		}

		// Update category min order_no
		if menu.OrderNo < categoryMap[catName].OrderNo {
			categoryMap[catName].OrderNo = menu.OrderNo
		}

		categoryMap[catName].Menus = append(categoryMap[catName].Menus, menu)
	}

	var result []*NavCategory
	for _, catName := range categoryOrder {
		result = append(result, categoryMap[catName])
	}

	// Sort categories by OrderNo asc
	sort.Slice(result, func(i, j int) bool {
		if result[i].OrderNo == result[j].OrderNo {
			return result[i].Category < result[j].Category
		}
		return result[i].OrderNo < result[j].OrderNo
	})

	return result
}

func sortTree(menus []*Menu) {
	sort.Slice(menus, func(i, j int) bool {
		if menus[i].OrderNo == menus[j].OrderNo {
			return menus[i].Title < menus[j].Title
		}
		return menus[i].OrderNo < menus[j].OrderNo
	})

	for _, m := range menus {
		if len(m.Children) > 0 {
			sortTree(m.Children)
		}
	}
}

// CheckPermission checks whether any of the given roleIDs have the active permission for a menu path
func (m *Manager) CheckPermission(roleIDs []string, menuPath, permissionCode string) (bool, error) {
	if len(roleIDs) == 0 {
		return false, nil
	}

	var count int64
	err := m.db.Table("roles").
		Joins("JOIN role_menu_permissions rmp ON rmp.role_id = roles.id AND rmp.is_active = 1").
		Joins("JOIN menu_permissions mp ON mp.id = rmp.menu_permission_id AND mp.is_active = 1").
		Joins("JOIN menus ON menus.id = mp.menu_id AND menus.is_active = 1").
		Joins("JOIN permissions p ON p.id = mp.permission_id AND p.is_active = 1").
		Where("roles.id IN (?) AND roles.is_active = 1", roleIDs).
		Where("menus.path = ? AND p.code = ?", menuPath, permissionCode).
		Count(&count).Error

	return count > 0, err
}

// CheckUserPermission checks permission directly for a userID using user_roles table
func (m *Manager) CheckUserPermission(userID, menuPath, permissionCode string) (bool, error) {
	var count int64
	err := m.db.Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id AND roles.is_active = 1").
		Joins("JOIN role_menu_permissions rmp ON rmp.role_id = roles.id AND rmp.is_active = 1").
		Joins("JOIN menu_permissions mp ON mp.id = rmp.menu_permission_id AND mp.is_active = 1").
		Joins("JOIN menus ON menus.id = mp.menu_id AND menus.is_active = 1").
		Joins("JOIN permissions p ON p.id = mp.permission_id AND p.is_active = 1").
		Where("user_roles.user_id = ?", userID).
		Where("menus.path = ? AND p.code = ?", menuPath, permissionCode).
		Count(&count).Error

	return count > 0, err
}

// AssignRoleMenuPermission grants an active menu permission to a role
func (m *Manager) AssignRoleMenuPermission(roleID, menuPermissionID string) error {
	var existing RoleMenuPermission
	err := m.db.Where("role_id = ? AND menu_permission_id = ?", roleID, menuPermissionID).First(&existing).Error
	if err == nil {
		if existing.IsActive != 1 {
			existing.IsActive = 1
			return m.db.Save(&existing).Error
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	rmp := RoleMenuPermission{
		RoleID:           roleID,
		MenuPermissionID: menuPermissionID,
		IsActive:         1,
	}
	return m.db.Create(&rmp).Error
}

// RemoveRoleMenuPermission revokes a menu permission from a role
func (m *Manager) RemoveRoleMenuPermission(roleID, menuPermissionID string) error {
	return m.db.Where("role_id = ? AND menu_permission_id = ?", roleID, menuPermissionID).Delete(&RoleMenuPermission{}).Error
}

