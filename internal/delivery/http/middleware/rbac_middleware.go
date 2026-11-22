package middleware

import (
	"go-rbac-menu/internal/repository"
	"go-rbac-menu/pkg/response"

	"github.com/nelthaarion/breeze/v2"
)

// RequirePermission ensures the authenticated user has an active permission on a specific menu path
func RequirePermission(menuRepo repository.MenuRepository, menuPath, permissionCode string) breeze.HandlerFunc {
	return func(ctx *breeze.Context) error {
		val, exists := ctx.Get("user_id")
		if !exists {
			return response.Error(ctx, 401, "Unauthorized", nil)
		}
		userID, ok := val.(string)
		if !ok || userID == "" {
			return response.Error(ctx, 401, "Unauthorized", nil)
		}

		// Check if user has 'dev' role: 'dev' role bypasses permission checks
		if rolesVal, hasRoles := ctx.Get("user_roles"); hasRoles {
			if roles, ok := rolesVal.([]string); ok {
				for _, r := range roles {
					if r == "dev" {
						return ctx.Next()
					}
				}
			}
		}

		hasAccess, err := menuRepo.CheckUserPermission(userID, menuPath, permissionCode)
		if err != nil {
			return response.Error(ctx, 500, "Failed to verify permissions", err)
		}

		if !hasAccess {
			return response.Error(ctx, 403, "Access denied: missing permission '"+permissionCode+"' for menu '"+menuPath+"'", nil)
		}

		return ctx.Next()
	}
}

