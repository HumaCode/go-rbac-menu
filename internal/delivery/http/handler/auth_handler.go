package handler

import (
	"github.com/HumaCode/go-rbac-menu/internal/service"
	"github.com/HumaCode/go-rbac-menu/pkg/response"

	"github.com/nelthaarion/breeze/v2"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(ctx *breeze.Context) error {
	var input LoginInput
	if err := ctx.Bind(&input); err != nil {
		return response.Error(ctx, 400, "Invalid request payload", err)
	}

	if input.Username == "" || input.Password == "" {
		return response.Error(ctx, 400, "Username and password are required", nil)
	}

	res, err := h.authService.Login(input.Username, input.Password)
	if err != nil {
		return response.Error(ctx, 401, err.Error(), nil)
	}

	return response.Success(ctx, 200, "Login successful", res)
}

func (h *AuthHandler) Register(ctx *breeze.Context) error {
	var input service.RegisterRequest
	if err := ctx.Bind(&input); err != nil {
		return response.Error(ctx, 400, "Invalid request payload", err)
	}

	if input.Username == "" || input.Email == "" || input.Password == "" || input.Name == "" {
		return response.Error(ctx, 400, "All fields (username, email, password, name) are required", nil)
	}

	user, err := h.authService.Register(input)
	if err != nil {
		return response.Error(ctx, 400, "Registration failed: "+err.Error(), err)
	}

	return response.Success(ctx, 201, "User registered successfully", user)
}

func (h *AuthHandler) GetProfile(ctx *breeze.Context) error {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		return response.Error(ctx, 401, "Unauthorized", nil)
	}
	userID := userIDVal.(string)

	user, roles, err := h.authService.GetProfile(userID)
	if err != nil {
		return response.Error(ctx, 404, err.Error(), nil)
	}

	return response.Success(ctx, 200, "Profile fetched successfully", map[string]interface{}{
		"user":  user,
		"roles": roles,
	})
}

