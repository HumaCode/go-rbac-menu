package rbacmenu

import (
	"sort"
	"strings"
	"sync"
)

// In-memory cache for high-speed zero-query menu resolution
type menuCache struct {
	mu           sync.RWMutex
	urls         []string
	rawTree      []*Menu
	rawCategories []*NavCategory
	isLoaded     bool
}

var globalCache = &menuCache{}

// ClearCache flushes the in-memory menu caches.
// Call this whenever menus, roles, or permissions are created, updated, or deleted.
func (m *Manager) ClearCache() {
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()
	globalCache.urls = nil
	globalCache.rawTree = nil
	globalCache.rawCategories = nil
	globalCache.isLoaded = false
}

// URLMenu is the Go equivalent of Laravel's urlMenu():
// Retrieves all active menu URLs, caches them forever in memory,
// and sorts them from longest to shortest so that Submenus/Childmenus
// are prioritized over Parent Menus during route matching.
func (m *Manager) URLMenu() []string {
	globalCache.mu.RLock()
	if globalCache.isLoaded && globalCache.urls != nil {
		defer globalCache.mu.RUnlock()
		return globalCache.urls
	}
	globalCache.mu.RUnlock()

	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()

	// Double check after acquiring write lock
	if globalCache.isLoaded && globalCache.urls != nil {
		return globalCache.urls
	}

	var urls []string
	m.db.Model(&Menu{}).Where("is_active = 1").Pluck("path", &urls)

	// Sort from longest character length to shortest (prioritize submenus)
	sort.Slice(urls, func(i, j int) bool {
		return len(urls[i]) > len(urls[j])
	})

	globalCache.urls = urls
	globalCache.isLoaded = true
	return globalCache.urls
}

// getGlobalRawCategories loads the entire active menu tree once and caches it in memory.
// Zero database queries on subsequent calls!
func (m *Manager) getGlobalRawCategories() []*NavCategory {
	globalCache.mu.RLock()
	if globalCache.isLoaded && globalCache.rawCategories != nil {
		defer globalCache.mu.RUnlock()
		return globalCache.rawCategories
	}
	globalCache.mu.RUnlock()

	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()

	if globalCache.isLoaded && globalCache.rawCategories != nil {
		return globalCache.rawCategories
	}

	// 1 single query to fetch all active menus
	var allActiveMenus []Menu
	m.db.Where("is_active = 1").Order("order_no asc, title asc").Find(&allActiveMenus)

	allActiveMap := make(map[string]Menu)
	for _, menu := range allActiveMenus {
		allActiveMap[menu.ID] = menu
	}

	// Build in-memory full tree
	nodeMap := make(map[string]*Menu)
	for _, original := range allActiveMenus {
		node := &Menu{
			BaseModel: original.BaseModel,
			ParentID:  original.ParentID,
			Category:  original.Category,
			Type:      original.Type,
			Title:     original.Title,
			Path:      original.Path,
			Icon:      original.Icon,
			OrderNo:   original.OrderNo,
			IsActive:  original.IsActive,
			Children:  make([]*Menu, 0),
		}
		nodeMap[original.ID] = node
	}

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

	sortTree(rootMenus)
	categories := GroupMenusByCategory(rootMenus)

	globalCache.rawTree = rootMenus
	globalCache.rawCategories = categories
	globalCache.isLoaded = true

	return globalCache.rawCategories
}

// Menus is the Go equivalent of Laravel's menus() helper:
// 1. Reads raw menu hierarchy from in-memory cache (Zero database query for menus!).
// 2. Eager-loads all user permissions into an in-memory set in 1 single fast query.
// 3. Filters categories, menus, and submenus/childmenus purely in CPU memory.
// 4. Automatically removes empty categories.
func (m *Manager) Menus(roleIDs ...string) []*NavCategory {
	return m.MenusWithMaxDepth(0, roleIDs...)
}

// MenusWithMaxDepth allows restricting depth (e.g. maxDepth = 2 for Menu + Submenu only).
// Pass maxDepth = 0 to dynamically include all levels (childmenus).
func (m *Manager) MenusWithMaxDepth(maxDepth int, roleIDs ...string) []*NavCategory {
	if len(roleIDs) == 0 {
		return []*NavCategory{}
	}

	// 1. Check if user has 'dev' superadmin role
	var devCount int64
	m.db.Model(&Role{}).Where("id IN (?) AND code = ? AND is_active = 1", roleIDs, "dev").Count(&devCount)
	isDev := devCount > 0

	// 2. Eager-load all permitted menu paths for this user in 1 SINGLE query (NO N+1!)
	allowedPaths := make(map[string]bool)
	if !isDev {
		var paths []string
		m.db.Table("roles").
			Select("DISTINCT menus.path").
			Joins("JOIN role_menu_permissions rmp ON rmp.role_id = roles.id AND rmp.is_active = 1").
			Joins("JOIN menu_permissions mp ON mp.id = rmp.menu_permission_id AND mp.is_active = 1").
			Joins("JOIN menus ON menus.id = mp.menu_id AND menus.is_active = 1").
			Where("roles.id IN (?) AND roles.is_active = 1", roleIDs).
			Pluck("menus.path", &paths)

		for _, p := range paths {
			allowedPaths[strings.TrimSpace(p)] = true
		}
	}

	// 3. Get cached raw categories
	rawCategories := m.getGlobalRawCategories()

	// 4. Filter purely in CPU memory (clone nodes to preserve immutable cache)
	var filteredCategories []*NavCategory

	for _, cat := range rawCategories {
		var allowedMenus []*Menu

		for _, menu := range cat.Menus {
			cloned := m.filterMenuNode(menu, allowedPaths, isDev, 1, maxDepth)
			if cloned != nil {
				allowedMenus = append(allowedMenus, cloned)
			}
		}

		// Only include category if it has at least 1 allowed menu (like Laravel ->filter(fn($c) => $c->count() > 0))
		if len(allowedMenus) > 0 {
			filteredCategories = append(filteredCategories, &NavCategory{
				Category: cat.Category,
				OrderNo:  cat.OrderNo,
				Menus:    allowedMenus,
			})
		}
	}

	return filteredCategories
}

// filterMenuNode recursively filters a menu and its submenus in memory
func (m *Manager) filterMenuNode(node *Menu, allowedPaths map[string]bool, isDev bool, currentDepth, maxDepth int) *Menu {
	// Check if this menu or any of its submenus is permitted
	var allowedChildren []*Menu

	if maxDepth == 0 || currentDepth < maxDepth {
		for _, child := range node.Children {
			clonedChild := m.filterMenuNode(child, allowedPaths, isDev, currentDepth+1, maxDepth)
			if clonedChild != nil {
				allowedChildren = append(allowedChildren, clonedChild)
			}
		}
	}

	// This node is allowed if:
	// - Role is 'dev'
	// - OR its own path is explicitly permitted
	// - OR it has allowed children (acts as parent container)
	isSelfAllowed := isDev || allowedPaths[strings.TrimSpace(node.Path)]
	hasAllowedChildren := len(allowedChildren) > 0

	if !isSelfAllowed && !hasAllowedChildren {
		return nil
	}

	// Clone node to avoid mutating cache
	return &Menu{
		BaseModel:       node.BaseModel,
		ParentID:        node.ParentID,
		Category:        node.Category,
		Type:            node.Type,
		Title:           node.Title,
		Path:            node.Path,
		Icon:            node.Icon,
		OrderNo:         node.OrderNo,
		IsActive:        node.IsActive,
		Permissions:     node.Permissions,
		Children:        allowedChildren,
		MenuPermissions: node.MenuPermissions,
	}
}

// MenusForUser is a shortcut to get categorized filtered menus directly from a User ID
func (m *Manager) MenusForUser(userID string) []*NavCategory {
	roleIDs := m.getUserRoleIDs(userID)
	return m.Menus(roleIDs...)
}

