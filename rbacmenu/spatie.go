package rbacmenu

import (
	"strings"

	"github.com/nelthaarion/breeze/v2"
)

// DefaultAbilities matches standard REST / HTTP actions
var DefaultAbilities = map[string]string{
	"GET":    "read",
	"POST":   "create",
	"PUT":    "update",
	"PATCH":  "update",
	"DELETE": "delete",
}

// Can checks if the given active roles have the required ability/permission.
// Format can be:
// - "{action} {menu_path}" (e.g. "read /admin/users", "delete /ppid/permohonan")
// - or simple permission code "{action}" (e.g. "read", "create")
func (m *Manager) Can(roleIDs []string, ability string) bool {
	if len(roleIDs) == 0 || ability == "" {
		return false
	}

	// Developer / Superadmin bypass (check if any role code is 'dev')
	var devCount int64
	m.db.Model(&Role{}).Where("id IN (?) AND code = ? AND is_active = 1", roleIDs, "dev").Count(&devCount)
	if devCount > 0 {
		return true
	}

	ability = strings.TrimSpace(ability)
	parts := strings.SplitN(ability, " ", 2)

	// If formatted as "{action} {menu_path}"
	if len(parts) == 2 {
		action := strings.TrimSpace(parts[0])
		menuPath := strings.TrimSpace(parts[1])
		hasAccess, err := m.CheckPermission(roleIDs, menuPath, action)
		return err == nil && hasAccess
	}

	// If only action code is provided
	var count int64
	err := m.db.Table("roles").
		Joins("JOIN role_menu_permissions rmp ON rmp.role_id = roles.id AND rmp.is_active = 1").
		Joins("JOIN menu_permissions mp ON mp.id = rmp.menu_permission_id AND mp.is_active = 1").
		Joins("JOIN permissions p ON p.id = mp.permission_id AND p.is_active = 1").
		Where("roles.id IN (?) AND roles.is_active = 1", roleIDs).
		Where("p.code = ?", parts[0]).
		Count(&count).Error

	return err == nil && count > 0
}

// CanUser checks ability directly for a user ID
func (m *Manager) CanUser(userID string, ability string) bool {
	var roleIDs []string
	err := m.db.Table("user_roles").
		Select("user_roles.role_id").
		Joins("JOIN roles ON roles.id = user_roles.role_id AND roles.is_active = 1").
		Where("user_roles.user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error

	if err != nil || len(roleIDs) == 0 {
		return false
	}

	return m.Can(roleIDs, ability)
}

// AttachMenuPermissions is the Go equivalent of Laravel's attachMenupermission trait:
// Automatically creates/finds permissions for the menu and assigns them to the menu and specified roles.
// Default actions if nil/empty: ["menu", "create", "read", "update", "delete"]
func (m *Manager) AttachMenuPermissions(menuID string, actions []string, roleIDsOrCodes []string) error {
	if len(actions) == 0 {
		actions = []string{"menu", "create", "read", "update", "delete"}
	}

	var menu Menu
	if err := m.db.Where("id = ? AND is_active = 1", menuID).First(&menu).Error; err != nil {
		return err
	}

	// 1. Resolve or create permissions for each action
	var menuPermissionIDs []string
	for _, action := range actions {
		action = strings.TrimSpace(action)
		if action == "" {
			continue
		}

		var perm Permission
		err := m.db.Where("code = ?", action).First(&perm).Error
		if err != nil {
			perm = Permission{
				Code:        action,
				Name:        strings.Title(action),
				Description: action + " permission",
				IsActive:    1,
			}
			if err := m.db.Create(&perm).Error; err != nil {
				return err
			}
		}

		// 2. Attach to menu via MenuPermission
		var mp MenuPermission
		err = m.db.Where("menu_id = ? AND permission_id = ?", menu.ID, perm.ID).First(&mp).Error
		if err != nil {
			mp = MenuPermission{
				MenuID:       menu.ID,
				PermissionID: perm.ID,
				IsActive:     1,
			}
			if err := m.db.Create(&mp).Error; err != nil {
				return err
			}
		} else if mp.IsActive != 1 {
			mp.IsActive = 1
			m.db.Save(&mp)
		}

		menuPermissionIDs = append(menuPermissionIDs, mp.ID)
	}

	// 3. Assign to roles if specified
	if len(roleIDsOrCodes) > 0 {
		var roles []Role
		m.db.Where("(id IN (?) OR code IN (?)) AND is_active = 1", roleIDsOrCodes, roleIDsOrCodes).Find(&roles)

		for _, role := range roles {
			for _, mpID := range menuPermissionIDs {
				_ = m.AssignRoleMenuPermission(role.ID, mpID)
			}
		}
	}

	return nil
}

// MatchMenuURL finds the best matching menu URL for the given request path
// Supports exact match and prefix match (e.g. /admin/users/123 matches /admin/users)
func MatchMenuURL(requestPath string, menuURLs []string) string {
	requestPath = strings.TrimRight(requestPath, "/")
	var bestMatch string

	for _, url := range menuURLs {
		cleanURL := strings.TrimRight(url, "/")
		if cleanURL == "" {
			continue
		}

		if requestPath == cleanURL || strings.HasPrefix(requestPath, cleanURL+"/") {
			// Choose the longest matching prefix (most specific match)
			if len(cleanURL) > len(bestMatch) {
				bestMatch = cleanURL
			}
		}
	}

	return bestMatch
}

// AutoAuthorizeMiddleware is the Go equivalent of Laravel's HasPermission callAction interceptor:
// Automatically intercepts requests, determines the action ('read', 'create', 'update', 'delete'),
// matches the closest registered menu path, and authorizes: "$action $matchedUrl"
func (m *Manager) AutoAuthorizeMiddleware() breeze.HandlerFunc {
	return func(ctx *breeze.Context) error {
		// 1. Get role IDs from context
		var roleIDs []string
		if val, exists := ctx.Get("user_roles"); exists {
			if codes, ok := val.([]string); ok {
				// Check for 'dev' superadmin bypass
				for _, c := range codes {
					if c == "dev" {
						return ctx.Next()
					}
				}
			}
		}

		if val, exists := ctx.Get("user_id"); exists {
			if uid, ok := val.(string); ok && uid != "" {
				_ = m.db.Table("user_roles").
					Select("user_roles.role_id").
					Joins("JOIN roles ON roles.id = user_roles.role_id AND roles.is_active = 1").
					Where("user_roles.user_id = ?", uid).
					Pluck("role_id", &roleIDs)
			}
		}

		if len(roleIDs) == 0 {
			ctx.Status(403)
			return ctx.JSON(map[string]interface{}{
				"success": false,
				"message": "Access denied: no active roles assigned",
			})
		}

		// 2. Map HTTP method to action ability
		httpMethod := string(ctx.Req.Method)
		action, ok := DefaultAbilities[httpMethod]
		if !ok {
			action = "read"
		}

		// 3. Fetch active menu paths (cached in slice)
		var menuURLs []string
		m.db.Model(&Menu{}).Where("is_active = 1").Order("LENGTH(path) DESC").Pluck("path", &menuURLs)

		// 4. Match URL
		reqPath := ctx.Req.Path
		matchedURL := MatchMenuURL(reqPath, menuURLs)

		if matchedURL != "" {
			requiredAbility := action + " " + matchedURL
			if !m.Can(roleIDs, requiredAbility) {
				ctx.Status(403)
				return ctx.JSON(map[string]interface{}{
					"success": false,
					"message": "Access denied: unauthorized for ability '" + requiredAbility + "'",
				})
			}
		}

		return ctx.Next()
	}
}
