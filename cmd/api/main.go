package main

import (
	"log"
	"runtime"

	"github.com/HumaCode/go-rbac-menu/config"
	"github.com/HumaCode/go-rbac-menu/internal/delivery/http/handler"
	"github.com/HumaCode/go-rbac-menu/internal/delivery/http/router"
	"github.com/HumaCode/go-rbac-menu/internal/domain"
	"github.com/HumaCode/go-rbac-menu/internal/repository"
	"github.com/HumaCode/go-rbac-menu/internal/seeder"
	"github.com/HumaCode/go-rbac-menu/internal/service"

	"github.com/nelthaarion/breeze/v2"
)

func main() {
	log.Println("[INFO] Starting Go RBAC Menu Service...")

	// 1. Load Configurations
	cfg := config.LoadConfig()

	// 2. Initialize Database Connection
	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("[FATAL] Database initialization failed: %v", err)
	}

	// 3. Auto Migration (ULID primary keys & relational schemas)
	log.Println("[INFO] Running database migrations...")
	err = db.AutoMigrate(
		&domain.User{},
		&domain.Role{},
		&domain.UserRole{},
		&domain.Permission{},
		&domain.Menu{},
		&domain.MenuPermission{},
		&domain.RoleMenuPermission{},
	)
	if err != nil {
		log.Fatalf("[FATAL] Database migration failed: %v", err)
	}
	log.Println("[INFO] Database migrations finished successfully.")

	// 4. Seed Initial Data (roles: dev/admin/user, permissions, menus, menu_permissions, default users)
	if err := seeder.SeedDatabase(db); err != nil {
		log.Printf("[WARN] Seeder warning: %v", err)
	}

	// 5. Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	_ = permRepo
	menuRepo := repository.NewMenuRepository(db)

	// 6. Initialize Services
	authService := service.NewAuthService(userRepo, roleRepo, cfg)
	menuService := service.NewMenuService(menuRepo, userRepo, roleRepo)
	roleService := service.NewRoleService(roleRepo)
	userService := service.NewUserService(userRepo, roleRepo)

	// 7. Initialize Handlers
	authHandler := handler.NewAuthHandler(authService)
	menuHandler := handler.NewMenuHandler(menuService)
	roleHandler := handler.NewRoleHandler(roleService)
	userHandler := handler.NewUserHandler(userService)

	// 8. Setup Breeze Router
	r := router.SetupRouter(&router.RouterDependencies{
		Config:      cfg,
		UserRepo:    userRepo,
		MenuRepo:    menuRepo,
		AuthHandler: authHandler,
		MenuHandler: menuHandler,
		RoleHandler: roleHandler,
		UserHandler: userHandler,
	})

	// 9. Initialize Worker Pool
	workerConcurrency := runtime.NumCPU() * 4
	if workerConcurrency < 8 {
		workerConcurrency = 8
	}
	pool := breeze.NewEventLoopWorkerPool(workerConcurrency)

	// 10. Start Breeze Server
	app := breeze.New(r, pool)
	log.Printf("[INFO] Server listening on port :%d (Workers: %d)", cfg.AppPort, workerConcurrency)
	if err := app.Run(cfg.AppPort, true); err != nil {
		log.Fatalf("[FATAL] Breeze server terminated with error: %v", err)
	}
}

