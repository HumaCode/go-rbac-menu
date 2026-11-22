package service

import (
	"errors"

	"go-rbac-menu/internal/domain"
	"go-rbac-menu/internal/repository"
)

type RoleService interface {
	ListActive() ([]domain.Role, error)
	GetByID(id string) (*domain.Role, error)
	Create(req CreateRoleRequest) (*domain.Role, error)
	Update(id string, req UpdateRoleRequest) (*domain.Role, error)
	Delete(id string) error
}

type roleService struct {
	roleRepo repository.RoleRepository
}

type CreateRoleRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    *int   `json:"is_active"`
}

func NewRoleService(roleRepo repository.RoleRepository) RoleService {
	return &roleService{roleRepo: roleRepo}
}

func (s *roleService) ListActive() ([]domain.Role, error) {
	return s.roleRepo.ListActive()
}

func (s *roleService) GetByID(id string) (*domain.Role, error) {
	return s.roleRepo.FindByID(id)
}

func (s *roleService) Create(req CreateRoleRequest) (*domain.Role, error) {
	role := &domain.Role{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    1,
	}
	if err := s.roleRepo.Create(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) Update(id string, req UpdateRoleRequest) (*domain.Role, error) {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("role not found")
	}

	role.Name = req.Name
	role.Description = req.Description
	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}

	if err := s.roleRepo.Update(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) Delete(id string) error {
	return s.roleRepo.Delete(id)
}

