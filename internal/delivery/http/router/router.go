package router

import (
	"github.com/HumaCode/go-rbac-menu/config"
	"github.com/HumaCode/go-rbac-menu/internal/delivery/http/handler"
	"github.com/HumaCode/go-rbac-menu/internal/delivery/http/middleware"
	"github.com/HumaCode/go-rbac-menu/internal/repository"
	"github.com/HumaCode/go-rbac-menu/pkg/response"

	"github.com/nelthaarion/breeze/v2"
	breezeMw "github.com/nelthaarion/breeze/v2/middlewares"
)

type RouterDependencies struct {
	Config      *config.Config
	UserRepo    repository.UserRepository
	MenuRepo    repository.MenuRepository
	AuthHandler *handler.AuthHandler
	MenuHandler *handler.MenuHandler
	RoleHandler *handler.RoleHandler
	UserHandler *handler.UserHandler
}

func SetupRouter(deps *RouterDependencies) *breeze.Router {
	r := breeze.NewRouter()

	// Global Middlewares
	r.Use(breezeMw.RecoveryMiddleware())
	r.Use(breezeMw.LoggingMiddleware())
	r.Use(breezeMw.CORSMiddleware(breezeMw.CORSOptions{
		AllowOrigins:     "*",
		AllowMethods:     "GET, POST, PUT, DELETE, PATCH, OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: "true",
	}))

	// Auth Middleware
	authMw := middleware.AuthMiddleware(deps.Config, deps.UserRepo)

	// Health check
	r.Handle(breeze.GET, "/health", func(ctx *breeze.Context) error {
		return response.Success(ctx, 200, "Service is healthy", map[string]string{
			"status": "online",
		})
	})

	// Public Auth routes (HandleBlocking for DB I/O)
	r.HandleBlocking(breeze.POST, "/api/v1/auth/login", deps.AuthHandler.Login)
	r.HandleBlocking(breeze.POST, "/api/v1/auth/register", deps.AuthHandler.Register)

	// Protected Auth routes
	r.HandleBlocking(breeze.GET, "/api/v1/auth/profile", deps.AuthHandler.GetProfile, authMw)

	// Navigation Menu route (Dynamic Tree for logged in user)
	r.HandleBlocking(breeze.GET, "/api/v1/nav-menu", deps.MenuHandler.GetNavMenu, authMw)

	// Menu Management routes
	r.HandleBlocking(breeze.GET, "/api/v1/menus", deps.MenuHandler.ListAllMenus, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/menus", "view"))
	r.HandleBlocking(breeze.GET, "/api/v1/menus/:id", deps.MenuHandler.GetMenu, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/menus", "view"))
	r.HandleBlocking(breeze.POST, "/api/v1/menus", deps.MenuHandler.CreateMenu, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/menus", "create"))
	r.HandleBlocking(breeze.PUT, "/api/v1/menus/:id", deps.MenuHandler.UpdateMenu, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/menus", "edit"))
	r.HandleBlocking(breeze.DELETE, "/api/v1/menus/:id", deps.MenuHandler.DeleteMenu, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/menus", "delete"))

	// Menu Permissions link management
	r.HandleBlocking(breeze.POST, "/api/v1/menus/:id/permissions", deps.MenuHandler.AssignPermission, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/menus", "edit"))
	r.HandleBlocking(breeze.GET, "/api/v1/menus/:id/permissions", deps.MenuHandler.ListMenuPermissions, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/menus", "view"))

	// Role Menu Permissions assignment
	r.HandleBlocking(breeze.POST, "/api/v1/roles/menu-permissions", deps.MenuHandler.AssignRoleMenuPermission, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/roles", "edit"))
	r.HandleBlocking(breeze.DELETE, "/api/v1/roles/menu-permissions", deps.MenuHandler.RemoveRoleMenuPermission, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/roles", "edit"))

	// Role Management routes
	r.HandleBlocking(breeze.GET, "/api/v1/roles", deps.RoleHandler.ListRoles, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/roles", "view"))
	r.HandleBlocking(breeze.GET, "/api/v1/roles/:id", deps.RoleHandler.GetRole, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/roles", "view"))
	r.HandleBlocking(breeze.POST, "/api/v1/roles", deps.RoleHandler.CreateRole, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/roles", "create"))
	r.HandleBlocking(breeze.PUT, "/api/v1/roles/:id", deps.RoleHandler.UpdateRole, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/roles", "edit"))
	r.HandleBlocking(breeze.DELETE, "/api/v1/roles/:id", deps.RoleHandler.DeleteRole, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/roles", "delete"))

	// User Management routes
	r.HandleBlocking(breeze.GET, "/api/v1/users", deps.UserHandler.ListUsers, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/users", "view"))
	r.HandleBlocking(breeze.GET, "/api/v1/users/:id", deps.UserHandler.GetUser, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/users", "view"))
	r.HandleBlocking(breeze.POST, "/api/v1/users/:id/roles", deps.UserHandler.AssignRole, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/users", "edit"))
	r.HandleBlocking(breeze.DELETE, "/api/v1/users/:id/roles", deps.UserHandler.RemoveRole, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/users", "edit"))
	r.HandleBlocking(breeze.PATCH, "/api/v1/users/:id/status", deps.UserHandler.ToggleActive, authMw,
		middleware.RequirePermission(deps.MenuRepo, "/system/users", "edit"))

	return r
}

