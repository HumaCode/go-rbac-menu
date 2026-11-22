package seeder

import (
	"log"

	"github.com/HumaCode/go-rbac-menu/internal/domain"
	"github.com/HumaCode/go-rbac-menu/pkg/utils"
	"gorm.io/gorm"
)

func SeedDatabase(db *gorm.DB) error {
	log.Println("[INFO] Checking and seeding initial data...")

	// 1. Seed Roles: dev, admin, user
	roles := []domain.Role{
		{Code: "dev", Name: "Developer / Superadmin", Description: "Full developer access to all systems", IsActive: 1},
		{Code: "admin", Name: "System Administrator", Description: "Administrative access to business operations", IsActive: 1},
		{Code: "user", Name: "Standard User", Description: "Regular user with basic access", IsActive: 1},
	}

	roleMap := make(map[string]*domain.Role)
	for _, r := range roles {
		var existing domain.Role
		if err := db.Where("code = ?", r.Code).First(&existing).Error; err != nil {
			newRole := r
			if err := db.Create(&newRole).Error; err != nil {
				return err
			}
			roleMap[r.Code] = &newRole
		} else {
			roleMap[r.Code] = &existing
		}
	}

	// 2. Seed Permissions: view, create, edit, delete
	perms := []domain.Permission{
		{Code: "view", Name: "View", Description: "Can view and read data", IsActive: 1},
		{Code: "create", Name: "Create", Description: "Can create new records", IsActive: 1},
		{Code: "edit", Name: "Edit", Description: "Can update existing records", IsActive: 1},
		{Code: "delete", Name: "Delete", Description: "Can remove or deactivate records", IsActive: 1},
	}

	permMap := make(map[string]*domain.Permission)
	for _, p := range perms {
		var existing domain.Permission
		if err := db.Where("code = ?", p.Code).First(&existing).Error; err != nil {
			newPerm := p
			if err := db.Create(&newPerm).Error; err != nil {
				return err
			}
			permMap[p.Code] = &newPerm
		} else {
			permMap[p.Code] = &existing
		}
	}

	// 3. Seed Menus (Tree structure)
	// Root: Dashboard
	dashboardMenu := getOrCreateMenu(db, nil, "Dashboard", "/dashboard", "layout-dashboard", 1)

	// Root: System Management (Parent)
	systemMenu := getOrCreateMenu(db, nil, "System Management", "/system", "settings", 2)

	// Submenus under System Management
	userMenu := getOrCreateMenu(db, &systemMenu.ID, "User Management", "/system/users", "users", 1)
	roleMenu := getOrCreateMenu(db, &systemMenu.ID, "Role Management", "/system/roles", "shield", 2)
	menuMenu := getOrCreateMenu(db, &systemMenu.ID, "Menu Management", "/system/menus", "menu", 3)

	// 4. Seed Menu Permissions
	// Dashboard has 'view'
	dashViewMP := getOrCreateMenuPermission(db, dashboardMenu.ID, permMap["view"].ID)

	// User menu has: view, create, edit, delete
	userViewMP := getOrCreateMenuPermission(db, userMenu.ID, permMap["view"].ID)
	userCreateMP := getOrCreateMenuPermission(db, userMenu.ID, permMap["create"].ID)
	userEditMP := getOrCreateMenuPermission(db, userMenu.ID, permMap["edit"].ID)
	userDeleteMP := getOrCreateMenuPermission(db, userMenu.ID, permMap["delete"].ID)

	// Role menu has: view, create, edit, delete
	roleViewMP := getOrCreateMenuPermission(db, roleMenu.ID, permMap["view"].ID)
	roleCreateMP := getOrCreateMenuPermission(db, roleMenu.ID, permMap["create"].ID)
	roleEditMP := getOrCreateMenuPermission(db, roleMenu.ID, permMap["edit"].ID)
	roleDeleteMP := getOrCreateMenuPermission(db, roleMenu.ID, permMap["delete"].ID)

	// Menu menu has: view, create, edit, delete
	menuViewMP := getOrCreateMenuPermission(db, menuMenu.ID, permMap["view"].ID)
	menuCreateMP := getOrCreateMenuPermission(db, menuMenu.ID, permMap["create"].ID)
	menuEditMP := getOrCreateMenuPermission(db, menuMenu.ID, permMap["edit"].ID)
	menuDeleteMP := getOrCreateMenuPermission(db, menuMenu.ID, permMap["delete"].ID)

	allMPs := []*domain.MenuPermission{
		dashViewMP,
		userViewMP, userCreateMP, userEditMP, userDeleteMP,
		roleViewMP, roleCreateMP, roleEditMP, roleDeleteMP,
		menuViewMP, menuCreateMP, menuEditMP, menuDeleteMP,
	}

	// 5. Assign Role Menu Permissions
	// 'dev' gets all menu permissions
	devRole := roleMap["dev"]
	for _, mp := range allMPs {
		assignRoleMenuPermission(db, devRole.ID, mp.ID)
	}

	// 'admin' gets Dashboard (view), Users (view, create, edit), Roles (view), Menus (view)
	adminRole := roleMap["admin"]
	adminMPs := []*domain.MenuPermission{
		dashViewMP,
		userViewMP, userCreateMP, userEditMP,
		roleViewMP,
		menuViewMP,
	}
	for _, mp := range adminMPs {
		assignRoleMenuPermission(db, adminRole.ID, mp.ID)
	}

	// 'user' gets Dashboard (view)
	userRole := roleMap["user"]
	assignRoleMenuPermission(db, userRole.ID, dashViewMP.ID)

	// 6. Seed Default Users
	seedUser(db, "dev", "dev@example.com", "Developer Admin", "dev123", devRole.ID)
	seedUser(db, "admin", "admin@example.com", "System Admin", "admin123", adminRole.ID)
	seedUser(db, "user", "user@example.com", "Standard User", "user123", userRole.ID)

	log.Println("[INFO] Seeding completed successfully.")
	return nil
}

func getOrCreateMenu(db *gorm.DB, parentID *string, title, path, icon string, orderNo int) *domain.Menu {
	var m domain.Menu
	if err := db.Where("path = ?", path).First(&m).Error; err != nil {
		m = domain.Menu{
			ParentID: parentID,
			Title:    title,
			Path:     path,
			Icon:     icon,
			OrderNo:  orderNo,
			IsActive: 1,
		}
		_ = db.Create(&m)
	}
	return &m
}

func getOrCreateMenuPermission(db *gorm.DB, menuID, permID string) *domain.MenuPermission {
	var mp domain.MenuPermission
	if err := db.Where("menu_id = ? AND permission_id = ?", menuID, permID).First(&mp).Error; err != nil {
		mp = domain.MenuPermission{
			MenuID:       menuID,
			PermissionID: permID,
			IsActive:     1,
		}
		_ = db.Create(&mp)
	}
	return &mp
}

func assignRoleMenuPermission(db *gorm.DB, roleID, menuPermID string) {
	var rmp domain.RoleMenuPermission
	if err := db.Where("role_id = ? AND menu_permission_id = ?", roleID, menuPermID).First(&rmp).Error; err != nil {
		rmp = domain.RoleMenuPermission{
			RoleID:           roleID,
			MenuPermissionID: menuPermID,
			IsActive:         1,
		}
		_ = db.Create(&rmp)
	}
}

func seedUser(db *gorm.DB, username, email, name, password, roleID string) {
	var existing domain.User
	if err := db.Where("username = ?", username).First(&existing).Error; err != nil {
		hash, _ := utils.HashPassword(password)
		user := domain.User{
			Username:     username,
			Email:        email,
			Name:         name,
			PasswordHash: hash,
			IsActive:     1,
		}
		if err := db.Create(&user).Error; err == nil {
			userRole := domain.UserRole{
				UserID: user.ID,
				RoleID: roleID,
			}
			_ = db.Create(&userRole)
			log.Printf("[SEED] Created user '%s' with role", username)
		}
	}
}

