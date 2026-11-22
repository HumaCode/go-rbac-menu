package handler

import (
	"go-rbac-menu/internal/service"
	"go-rbac-menu/pkg/response"

	"github.com/nelthaarion/breeze/v2"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) ListUsers(ctx *breeze.Context) error {
	users, err := h.userService.ListActive()
	if err != nil {
		return response.Error(ctx, 500, "Failed to list users", err)
	}
	return response.Success(ctx, 200, "Users retrieved successfully", users)
}

func (h *UserHandler) GetUser(ctx *breeze.Context) error {
	id := ctx.Param("id")
	user, err := h.userService.GetByID(id)
	if err != nil {
		return response.Error(ctx, 404, "User not found", err)
	}
	return response.Success(ctx, 200, "User retrieved successfully", user)
}

type AssignRoleInput struct {
	RoleID string `json:"role_id"`
}

func (h *UserHandler) AssignRole(ctx *breeze.Context) error {
	userID := ctx.Param("id")
	var input AssignRoleInput
	if err := ctx.Bind(&input); err != nil {
		return response.Error(ctx, 400, "Invalid request payload", err)
	}

	if err := h.userService.AssignRole(userID, input.RoleID); err != nil {
		return response.Error(ctx, 400, err.Error(), err)
	}

	return response.Success(ctx, 200, "Role assigned successfully", nil)
}

func (h *UserHandler) RemoveRole(ctx *breeze.Context) error {
	userID := ctx.Param("id")
	var input AssignRoleInput
	if err := ctx.Bind(&input); err != nil {
		return response.Error(ctx, 400, "Invalid request payload", err)
	}

	if err := h.userService.RemoveRole(userID, input.RoleID); err != nil {
		return response.Error(ctx, 400, err.Error(), err)
	}

	return response.Success(ctx, 200, "Role removed successfully", nil)
}

type ToggleActiveInput struct {
	IsActive int `json:"is_active"`
}

func (h *UserHandler) ToggleActive(ctx *breeze.Context) error {
	userID := ctx.Param("id")
	var input ToggleActiveInput
	if err := ctx.Bind(&input); err != nil {
		return response.Error(ctx, 400, "Invalid request payload", err)
	}

	if err := h.userService.ToggleActive(userID, input.IsActive); err != nil {
		return response.Error(ctx, 400, err.Error(), err)
	}

	return response.Success(ctx, 200, "User status updated successfully", nil)
}

