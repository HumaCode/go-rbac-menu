package handler

import (
	"github.com/HumaCode/go-rbac-menu/internal/service"
	"github.com/HumaCode/go-rbac-menu/pkg/response"

	"github.com/nelthaarion/breeze/v2"
)

type MenuHandler struct {
	menuService service.MenuService
}

func NewMenuHandler(menuService service.MenuService) *MenuHandler {
	return &MenuHandler{menuService: menuService}
}

// GetNavMenu returns the hierarchical navigation menu tree for the current authenticated user.
// Pass ?grouped=true to receive menus neatly grouped by category.
func (h *MenuHandler) GetNavMenu(ctx *breeze.Context) error {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		return response.Error(ctx, 401, "Unauthorized", nil)
	}
	userID := userIDVal.(string)

	grouped := ctx.Query("grouped")
	if grouped == "true" || grouped == "1" {
		categories, err := h.menuService.GetNavMenuCategorizedForUser(userID)
		if err != nil {
			return response.Error(ctx, 500, "Failed to load categorized navigation menu", err)
		}
		return response.Success(ctx, 200, "Categorized navigation menus retrieved successfully", categories)
	}

	tree, err := h.menuService.GetNavMenuTreeForUser(userID)
	if err != nil {
		return response.Error(ctx, 500, "Failed to load navigation menu", err)
	}

	return response.Success(ctx, 200, "Navigation menus retrieved successfully", tree)
}

// ListAllMenus returns a flat list of all active menus (for management)
func (h *MenuHandler) ListAllMenus(ctx *breeze.Context) error {
	menus, err := h.menuService.ListActiveMenus()
	if err != nil {
		return response.Error(ctx, 500, "Failed to list menus", err)
	}
	return response.Success(ctx, 200, "Menus retrieved successfully", menus)
}

func (h *MenuHandler) GetMenu(ctx *breeze.Context) error {
	id := ctx.Param("id")
	menu, err := h.menuService.GetMenuByID(id)
	if err != nil {
		return response.Error(ctx, 404, "Menu not found", err)
	}
	return response.Success(ctx, 200, "Menu retrieved successfully", menu)
}

func (h *MenuHandler) CreateMenu(ctx *breeze.Context) error {
	var input service.CreateMenuRequest
	if err := ctx.Bind(&input); err != nil {
		return response.Error(ctx, 400, "Invalid request payload", err)
	}

	if input.Title == "" || input.Path == "" {
		return response.Error(ctx, 400, "Title and path are required", nil)
	}

	menu, err := h.menuService.CreateMenu(input)
	if err != nil {
		return response.Error(ctx, 500, "Failed to create menu", err)
	}

	return response.Success(ctx, 201, "Menu created successfully", menu)
}

func (h *MenuHandler) UpdateMenu(ctx *breeze.Context) error {
	id := ctx.Param("id")
	var input service.UpdateMenuRequest
	if err := ctx.Bind(&input); err != nil {
		return response.Error(ctx, 400, "Invalid request payload", err)
	}

	menu, err := h.menuService.UpdateMenu(id, input)
	if err != nil {
		return response.Error(ctx, 500, "Failed to update menu: "+err.Error(), err)
	}

	return response.Success(ctx, 200, "Menu updated successfully", menu)
}

func (h *MenuHandler) DeleteMenu(ctx *breeze.Context) error {
	id := ctx.Param("id")
	if err := h.menuService.DeleteMenu(id); err != nil {
		return response.Error(ctx, 500, "Failed to delete menu", err)
	}
	return response.Success(ctx, 200, "Menu deleted successfully (deactivated)", nil)
}

type AssignMenuPermissionInput struct {
	PermissionID string `json:"permission_id"`
}

func (h *MenuHandler) AssignPermission(ctx *breeze.Context) error {
	menuID := ctx.Param("id")
	var input AssignMenuPermissionInput
	if err := ctx.Bind(&input); err != nil {
		return response.Error(ctx, 400, "Invalid request payload", err)
	}

	mp, err := h.menuService.AssignPermission(menuID, input.PermissionID)
	if err != nil {
		return response.Error(ctx, 500, "Failed to assign permission to menu", err)
	}

	return response.Success(ctx, 200, "Permission assigned to menu successfully", mp)
}

func (h *MenuHandler) ListMenuPermissions(ctx *breeze.Context) error {
	menuID := ctx.Param("id")
	mps, err := h.menuService.ListMenuPermissions(menuID)
	if err != nil {
		return response.Error(ctx, 500, "Failed to list menu permissions", err)
	}
	return response.Success(ctx, 200, "Menu permissions retrieved successfully", mps)
}

type RoleMenuPermissionInput struct {
	RoleID           string `json:"role_id"`
	MenuPermissionID string `json:"menu_permission_id"`
}

func (h *MenuHandler) AssignRoleMenuPermission(ctx *breeze.Context) error {
	var input RoleMenuPermissionInput
	if err := ctx.Bind(&input); err != nil {
		return response.Error(ctx, 400, "Invalid request payload", err)
	}

	if err := h.menuService.AssignRoleMenuPermission(input.RoleID, input.MenuPermissionID); err != nil {
		return response.Error(ctx, 500, "Failed to assign role menu permission", err)
	}

	return response.Success(ctx, 200, "Role menu permission granted successfully", nil)
}

func (h *MenuHandler) RemoveRoleMenuPermission(ctx *breeze.Context) error {
	var input RoleMenuPermissionInput
	if err := ctx.Bind(&input); err != nil {
		return response.Error(ctx, 400, "Invalid request payload", err)
	}

	if err := h.menuService.RemoveRoleMenuPermission(input.RoleID, input.MenuPermissionID); err != nil {
		return response.Error(ctx, 500, "Failed to remove role menu permission", err)
	}

	return response.Success(ctx, 200, "Role menu permission revoked successfully", nil)
}

