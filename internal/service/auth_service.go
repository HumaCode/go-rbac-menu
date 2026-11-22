package service

import (
	"errors"

	"go-rbac-menu/config"
	"go-rbac-menu/internal/domain"
	"go-rbac-menu/internal/repository"
	"go-rbac-menu/pkg/utils"
)

type AuthService interface {
	Login(username, password string) (*LoginResponse, error)
	Register(req RegisterRequest) (*domain.User, error)
	GetProfile(userID string) (*domain.User, []string, error)
}

type authService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
	cfg      *config.Config
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
	Roles []string     `json:"roles"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func NewAuthService(userRepo repository.UserRepository, roleRepo repository.RoleRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		roleRepo: roleRepo,
		cfg:      cfg,
	}
}

func (s *authService) Login(username, password string) (*LoginResponse, error) {
	// Only users with is_active = 1 can login
	user, err := s.userRepo.FindActiveByUsername(username)
	if err != nil {
		return nil, errors.New("invalid username or account is inactive")
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	// Fetch active roles
	roles, err := s.userRepo.GetUserRoleCodes(user.ID)
	if err != nil {
		return nil, err
	}

	token, err := utils.GenerateToken(
		s.cfg.JWTSecret,
		s.cfg.JWTExpireHours,
		user.ID,
		user.Username,
		user.Email,
		roles,
	)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token: token,
		User:  user,
		Roles: roles,
	}, nil
}

func (s *authService) Register(req RegisterRequest) (*domain.User, error) {
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Name:         req.Name,
		IsActive:     1, // active by default
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Assign default role 'user' (active)
	defaultRole, err := s.roleRepo.FindByCode("user")
	if err == nil && defaultRole.IsActive == 1 {
		_ = s.userRepo.AssignRole(user.ID, defaultRole.ID)
	}

	return user, nil
}

func (s *authService) GetProfile(userID string) (*domain.User, []string, error) {
	user, err := s.userRepo.FindActiveByID(userID)
	if err != nil {
		return nil, nil, errors.New("user not found or inactive")
	}

	roles, err := s.userRepo.GetUserRoleCodes(user.ID)
	if err != nil {
		return nil, nil, err
	}

	return user, roles, nil
}

