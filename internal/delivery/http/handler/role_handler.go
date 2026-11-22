package handler

import (
	"go-rbac-menu/internal/service"
	"go-rbac-menu/pkg/response"

	"github.com/nelthaarion/breeze/v2"
)

type RoleHandler struct {
	roleService service.RoleService
}

func NewRoleHandler(roleService service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

func (h *RoleHandler) ListRoles(ctx *breeze.Context) error {
	roles, err := h.roleService.ListActive()
	if err != nil {
		return response.Error(ctx, 500, "Failed to list roles", err)
	}
	return response.Success(ctx, 200, "Roles retrieved successfully", roles)
}

func (h *RoleHandler) GetRole(ctx *breeze.Context) error {
	id := ctx.Param("id")
	role, err := h.roleService.GetByID(id)
	if err != nil {
		return response.Error(ctx, 404, "Role not found", err)
	}
	return response.Success(ctx, 200, "Role retrieved successfully", role)
}

func (h *RoleHandler) CreateRole(ctx *breeze.Context) error {
	var input service.CreateRoleRequest
	if err := ctx.Bind(&input); err != nil {
		return response.Error(ctx, 400, "Invalid request payload", err)
	}

	if input.Code == "" || input.Name == "" {
		return response.Error(ctx, 400, "Code and name are required", nil)
	}

	role, err := h.roleService.Create(input)
	if err != nil {
		return response.Error(ctx, 500, "Failed to create role: "+err.Error(), err)
	}

	return response.Success(ctx, 201, "Role created successfully", role)
}

func (h *RoleHandler) UpdateRole(ctx *breeze.Context) error {
	id := ctx.Param("id")
	var input service.UpdateRoleRequest
	if err := ctx.Bind(&input); err != nil {
		return response.Error(ctx, 400, "Invalid request payload", err)
	}

	role, err := h.roleService.Update(id, input)
	if err != nil {
		return response.Error(ctx, 500, "Failed to update role: "+err.Error(), err)
	}

	return response.Success(ctx, 200, "Role updated successfully", role)
}

func (h *RoleHandler) DeleteRole(ctx *breeze.Context) error {
	id := ctx.Param("id")
	if err := h.roleService.Delete(id); err != nil {
		return response.Error(ctx, 500, "Failed to delete role", err)
	}
	return response.Success(ctx, 200, "Role deactivated successfully", nil)
}

