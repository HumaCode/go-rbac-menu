package service

import (
	"errors"

	"github.com/HumaCode/go-rbac-menu/internal/domain"
	"github.com/HumaCode/go-rbac-menu/internal/repository"
)

type UserService interface {
	ListActive() ([]domain.User, error)
	GetByID(id string) (*domain.User, error)
	AssignRole(userID, roleID string) error
	RemoveRole(userID, roleID string) error
	ToggleActive(userID string, active int) error
}

type userService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

func NewUserService(userRepo repository.UserRepository, roleRepo repository.RoleRepository) UserService {
	return &userService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

func (s *userService) ListActive() ([]domain.User, error) {
	return s.userRepo.ListActive()
}

func (s *userService) GetByID(id string) (*domain.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *userService) AssignRole(userID, roleID string) error {
	// Verify user is active
	user, err := s.userRepo.FindActiveByID(userID)
	if err != nil || user.IsActive != 1 {
		return errors.New("user not found or inactive")
	}

	// Verify role is active
	role, err := s.roleRepo.FindByID(roleID)
	if err != nil || role.IsActive != 1 {
		return errors.New("role not found or inactive")
	}

	return s.userRepo.AssignRole(userID, roleID)
}

func (s *userService) RemoveRole(userID, roleID string) error {
	return s.userRepo.RemoveRole(userID, roleID)
}

func (s *userService) ToggleActive(userID string, active int) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	user.IsActive = active
	return s.userRepo.Update(user)
}

