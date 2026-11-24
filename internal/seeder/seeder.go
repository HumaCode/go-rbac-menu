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

	// 3. Seed Menus (Categorized & 3-Level Hierarchy)
	// --- KATEGORI 1: UTAMA ---
	dashboardMenu := getOrCreateMenu(db, nil, "UTAMA", "menu", "Dashboard", "/dashboard", "layout-dashboard", 1)

	// --- KATEGORI 2: LAYANAN PPID (Contoh 3 Level: Menu -> Submenu -> Childmenu) ---
	ppidRootMenu := getOrCreateMenu(db, nil, "LAYANAN PPID", "menu", "Layanan Permohonan", "/ppid", "folder", 2)
	// Level 2 (Submenu):
	ppidPermohonan := getOrCreateMenu(db, &ppidRootMenu.ID, "LAYANAN PPID", "menu", "Permohonan Informasi", "/ppid/permohonan", "file-text", 1)
	ppidKeberatan := getOrCreateMenu(db, &ppidRootMenu.ID, "LAYANAN PPID", "menu", "Pengajuan Keberatan", "/ppid/keberatan", "alert-circle", 2)
	// Level 3 (Childmenu):
	ppidMasuk := getOrCreateMenu(db, &ppidPermohonan.ID, "LAYANAN PPID", "menu", "Permohonan Baru", "/ppid/permohonan/masuk", "inbox", 1)
	ppidProses := getOrCreateMenu(db, &ppidPermohonan.ID, "LAYANAN PPID", "menu", "Dalam Proses", "/ppid/permohonan/proses", "clock", 2)
	ppidSelesai := getOrCreateMenu(db, &ppidPermohonan.ID, "LAYANAN PPID", "menu", "Selesai", "/ppid/permohonan/selesai", "check-circle", 3)

	// --- KATEGORI 3: PENGATURAN SISTEM ---
	systemMenu := getOrCreateMenu(db, nil, "PENGATURAN SISTEM", "menu", "System Management", "/system", "settings", 3)
	// Level 2 (Submenu):
	userMenu := getOrCreateMenu(db, &systemMenu.ID, "PENGATURAN SISTEM", "menu", "User Management", "/system/users", "users", 1)
	roleMenu := getOrCreateMenu(db, &systemMenu.ID, "PENGATURAN SISTEM", "menu", "Role Management", "/system/roles", "shield", 2)
	menuMenu := getOrCreateMenu(db, &systemMenu.ID, "PENGATURAN SISTEM", "menu", "Menu Management", "/system/menus", "menu", 3)

	// 4. Seed Menu Permissions
	dashViewMP := getOrCreateMenuPermission(db, dashboardMenu.ID, permMap["view"].ID)

	ppidRootViewMP := getOrCreateMenuPermission(db, ppidRootMenu.ID, permMap["view"].ID)
	ppidPermohonanViewMP := getOrCreateMenuPermission(db, ppidPermohonan.ID, permMap["view"].ID)
	ppidKeberatanViewMP := getOrCreateMenuPermission(db, ppidKeberatan.ID, permMap["view"].ID)
	ppidMasukViewMP := getOrCreateMenuPermission(db, ppidMasuk.ID, permMap["view"].ID)
	ppidProsesViewMP := getOrCreateMenuPermission(db, ppidProses.ID, permMap["view"].ID)
	ppidSelesaiViewMP := getOrCreateMenuPermission(db, ppidSelesai.ID, permMap["view"].ID)

	userViewMP := getOrCreateMenuPermission(db, userMenu.ID, permMap["view"].ID)
	userCreateMP := getOrCreateMenuPermission(db, userMenu.ID, permMap["create"].ID)
	userEditMP := getOrCreateMenuPermission(db, userMenu.ID, permMap["edit"].ID)
	userDeleteMP := getOrCreateMenuPermission(db, userMenu.ID, permMap["delete"].ID)

	roleViewMP := getOrCreateMenuPermission(db, roleMenu.ID, permMap["view"].ID)
	roleCreateMP := getOrCreateMenuPermission(db, roleMenu.ID, permMap["create"].ID)
	roleEditMP := getOrCreateMenuPermission(db, roleMenu.ID, permMap["edit"].ID)
	roleDeleteMP := getOrCreateMenuPermission(db, roleMenu.ID, permMap["delete"].ID)

	menuViewMP := getOrCreateMenuPermission(db, menuMenu.ID, permMap["view"].ID)
	menuCreateMP := getOrCreateMenuPermission(db, menuMenu.ID, permMap["create"].ID)
	menuEditMP := getOrCreateMenuPermission(db, menuMenu.ID, permMap["edit"].ID)
	menuDeleteMP := getOrCreateMenuPermission(db, menuMenu.ID, permMap["delete"].ID)

	allMPs := []*domain.MenuPermission{
		dashViewMP,
		ppidRootViewMP, ppidPermohonanViewMP, ppidKeberatanViewMP,
		ppidMasukViewMP, ppidProsesViewMP, ppidSelesaiViewMP,
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

	// 'admin' gets Dashboard, PPID, and Management
	adminRole := roleMap["admin"]
	adminMPs := []*domain.MenuPermission{
		dashViewMP,
		ppidRootViewMP, ppidPermohonanViewMP, ppidKeberatanViewMP,
		ppidMasukViewMP, ppidProsesViewMP, ppidSelesaiViewMP,
		userViewMP, userCreateMP, userEditMP,
		roleViewMP,
		menuViewMP,
	}
	for _, mp := range adminMPs {
		assignRoleMenuPermission(db, adminRole.ID, mp.ID)
	}

	// 'user' gets Dashboard & PPID Layanan View
	userRole := roleMap["user"]
	userRoleMPs := []*domain.MenuPermission{
		dashViewMP,
		ppidRootViewMP, ppidPermohonanViewMP, ppidMasukViewMP,
	}
	for _, mp := range userRoleMPs {
		assignRoleMenuPermission(db, userRole.ID, mp.ID)
	}

	// 6. Seed Default Users
	seedUser(db, "dev", "dev@example.com", "Developer Admin", "dev123", devRole.ID)
	seedUser(db, "admin", "admin@example.com", "System Admin", "admin123", adminRole.ID)
	seedUser(db, "user", "user@example.com", "Standard User", "user123", userRole.ID)

	log.Println("[INFO] Seeding completed successfully.")
	return nil
}

func getOrCreateMenu(db *gorm.DB, parentID *string, category, menuType, title, path, icon string, orderNo int) *domain.Menu {
	var m domain.Menu
	if err := db.Where("path = ?", path).First(&m).Error; err != nil {
		m = domain.Menu{
			ParentID: parentID,
			Category: category,
			Type:     menuType,
			Title:    title,
			Path:     path,
			Icon:     icon,
			OrderNo:  orderNo,
			IsActive: 1,
		}
		_ = db.Create(&m)
	} else {
		// Update category and type if empty
		if m.Category == "" || m.Type == "" {
			m.Category = category
			m.Type = menuType
			_ = db.Save(&m)
		}
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
