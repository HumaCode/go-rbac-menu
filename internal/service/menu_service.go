package service

import (
	"errors"
	"sort"

	"go-rbac-menu/internal/domain"
	"go-rbac-menu/internal/repository"
)

type MenuService interface {
	GetNavMenuTreeForUser(userID string) ([]*domain.Menu, error)
	ListActiveMenus() ([]domain.Menu, error)
	GetMenuByID(id string) (*domain.Menu, error)
	CreateMenu(req CreateMenuRequest) (*domain.Menu, error)
	UpdateMenu(id string, req UpdateMenuRequest) (*domain.Menu, error)
	DeleteMenu(id string) error

	// Menu permissions
	AssignPermission(menuID, permissionID string) (*domain.MenuPermission, error)
	ListMenuPermissions(menuID string) ([]domain.MenuPermission, error)
	AssignRoleMenuPermission(roleID, menuPermissionID string) error
	RemoveRoleMenuPermission(roleID, menuPermissionID string) error
}

type menuService struct {
	menuRepo repository.MenuRepository
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

type CreateMenuRequest struct {
	ParentID *string `json:"parent_id"`
	Title    string  `json:"title"`
	Path     string  `json:"path"`
	Icon     string  `json:"icon"`
	OrderNo  int     `json:"order_no"`
}

type UpdateMenuRequest struct {
	ParentID *string `json:"parent_id"`
	Title    string  `json:"title"`
	Path     string  `json:"path"`
	Icon     string  `json:"icon"`
	OrderNo  int     `json:"order_no"`
	IsActive *int    `json:"is_active"`
}

func NewMenuService(menuRepo repository.MenuRepository, userRepo repository.UserRepository, roleRepo repository.RoleRepository) MenuService {
	return &menuService{
		menuRepo: menuRepo,
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

func (s *menuService) GetNavMenuTreeForUser(userID string) ([]*domain.Menu, error) {
	// 1. Get active roles of this active user
	roleIDs, err := s.userRepo.GetUserRoleIDs(userID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return []*domain.Menu{}, nil
	}

	// 2. Fetch role_menu_permissions where is_active = 1
	rmps, err := s.menuRepo.GetRoleMenuPermissions(roleIDs)
	if err != nil {
		return nil, err
	}

	// 3. Map permitted menus and their specific actions
	menuActionMap := make(map[string]map[string]bool) // menuID -> set of permission codes
	menuObjMap := make(map[string]domain.Menu)        // menuID -> Menu entity

	for _, rmp := range rmps {
		mp := rmp.MenuPermission
		if mp.IsActive != 1 || mp.Menu.IsActive != 1 || mp.Permission.IsActive != 1 {
			continue
		}

		mID := mp.MenuID
		if _, ok := menuActionMap[mID]; !ok {
			menuActionMap[mID] = make(map[string]bool)
		}
		menuActionMap[mID][mp.Permission.Code] = true
		menuObjMap[mID] = mp.Menu
	}

	// 4. Fetch all active menus to resolve parents
	allActiveMenus, err := s.menuRepo.ListActive()
	if err != nil {
		return nil, err
	}
	allActiveMenuMap := make(map[string]domain.Menu)
	for _, m := range allActiveMenus {
		allActiveMenuMap[m.ID] = m
	}

	// Ensure all ancestor/parent menus of permitted menus are also included
	includedMenuIDs := make(map[string]bool)
	for mID := range menuActionMap {
		currID := mID
		for {
			includedMenuIDs[currID] = true
			m, exists := allActiveMenuMap[currID]
			if !exists || m.ParentID == nil || *m.ParentID == "" {
				break
			}
			currID = *m.ParentID
		}
	}

	// 5. Construct Menu pointers for tree building
	nodeMap := make(map[string]*domain.Menu)
	for mID := range includedMenuIDs {
		original, ok := allActiveMenuMap[mID]
		if !ok {
			continue
		}

		// Convert actions set to slice
		var permissions []string
		if actionSet, hasActions := menuActionMap[mID]; hasActions {
			for code := range actionSet {
				permissions = append(permissions, code)
			}
			sort.Strings(permissions)
		}

		node := &domain.Menu{
			BaseModel:   original.BaseModel,
			ParentID:    original.ParentID,
			Title:       original.Title,
			Path:        original.Path,
			Icon:        original.Icon,
			OrderNo:     original.OrderNo,
			IsActive:    original.IsActive,
			Permissions: permissions,
			Children:    make([]*domain.Menu, 0),
		}
		nodeMap[mID] = node
	}

	// 6. Build hierarchy (Parent -> Children)
	var rootMenus []*domain.Menu
	for _, node := range nodeMap {
		if node.ParentID == nil || *node.ParentID == "" {
			rootMenus = append(rootMenus, node)
		} else if parentNode, exists := nodeMap[*node.ParentID]; exists {
			parentNode.Children = append(parentNode.Children, node)
		} else {
			// Parent is not included/not active, treat as root
			rootMenus = append(rootMenus, node)
		}
	}

	// 7. Sort tree nodes recursively by OrderNo asc
	sortMenuTree(rootMenus)

	return rootMenus, nil
}

func sortMenuTree(menus []*domain.Menu) {
	sort.Slice(menus, func(i, j int) bool {
		if menus[i].OrderNo == menus[j].OrderNo {
			return menus[i].Title < menus[j].Title
		}
		return menus[i].OrderNo < menus[j].OrderNo
	})

	for _, m := range menus {
		if len(m.Children) > 0 {
			sortMenuTree(m.Children)
		}
	}
}

func (s *menuService) ListActiveMenus() ([]domain.Menu, error) {
	return s.menuRepo.ListActive()
}

func (s *menuService) GetMenuByID(id string) (*domain.Menu, error) {
	return s.menuRepo.FindByID(id)
}

func (s *menuService) CreateMenu(req CreateMenuRequest) (*domain.Menu, error) {
	menu := &domain.Menu{
		ParentID: req.ParentID,
		Title:    req.Title,
		Path:     req.Path,
		Icon:     req.Icon,
		OrderNo:  req.OrderNo,
		IsActive: 1, // active by default
	}

	if err := s.menuRepo.Create(menu); err != nil {
		return nil, err
	}
	return menu, nil
}

func (s *menuService) UpdateMenu(id string, req UpdateMenuRequest) (*domain.Menu, error) {
	menu, err := s.menuRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("menu not found")
	}

	menu.ParentID = req.ParentID
	menu.Title = req.Title
	menu.Path = req.Path
	menu.Icon = req.Icon
	menu.OrderNo = req.OrderNo
	if req.IsActive != nil {
		menu.IsActive = *req.IsActive
	}

	if err := s.menuRepo.Update(menu); err != nil {
		return nil, err
	}
	return menu, nil
}

func (s *menuService) DeleteMenu(id string) error {
	return s.menuRepo.Delete(id)
}

func (s *menuService) AssignPermission(menuID, permissionID string) (*domain.MenuPermission, error) {
	return s.menuRepo.AssignPermissionToMenu(menuID, permissionID)
}

func (s *menuService) ListMenuPermissions(menuID string) ([]domain.MenuPermission, error) {
	return s.menuRepo.ListMenuPermissionsByMenuID(menuID)
}

func (s *menuService) AssignRoleMenuPermission(roleID, menuPermissionID string) error {
	return s.menuRepo.AssignRoleMenuPermission(roleID, menuPermissionID)
}

func (s *menuService) RemoveRoleMenuPermission(roleID, menuPermissionID string) error {
	return s.menuRepo.RemoveRoleMenuPermission(roleID, menuPermissionID)
}

