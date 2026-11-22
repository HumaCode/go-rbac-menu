package service

import (
	"testing"

	"go-rbac-menu/internal/domain"
)

func TestSortMenuTree(t *testing.T) {
	parentID := "01HQ0000000000000000000001"
	child1 := &domain.Menu{
		BaseModel: domain.BaseModel{ID: "01HQ0000000000000000000002"},
		ParentID:  &parentID,
		Title:     "User Management",
		OrderNo:   2,
	}
	child2 := &domain.Menu{
		BaseModel: domain.BaseModel{ID: "01HQ0000000000000000000003"},
		ParentID:  &parentID,
		Title:     "Analytics",
		OrderNo:   1,
	}

	root1 := &domain.Menu{
		BaseModel: domain.BaseModel{ID: parentID},
		Title:     "Settings",
		OrderNo:   2,
		Children:  []*domain.Menu{child1, child2},
	}
	root2 := &domain.Menu{
		BaseModel: domain.BaseModel{ID: "01HQ0000000000000000000004"},
		Title:     "Dashboard",
		OrderNo:   1,
	}

	tree := []*domain.Menu{root1, root2}
	sortMenuTree(tree)

	// Root level should be Dashboard (OrderNo 1) then Settings (OrderNo 2)
	if tree[0].Title != "Dashboard" {
		t.Errorf("expected tree[0] to be Dashboard, got %s", tree[0].Title)
	}
	if tree[1].Title != "Settings" {
		t.Errorf("expected tree[1] to be Settings, got %s", tree[1].Title)
	}

	// Children of Settings should be Analytics (OrderNo 1) then User Management (OrderNo 2)
	if tree[1].Children[0].Title != "Analytics" {
		t.Errorf("expected child[0] to be Analytics, got %s", tree[1].Children[0].Title)
	}
	if tree[1].Children[1].Title != "User Management" {
		t.Errorf("expected child[1] to be User Management, got %s", tree[1].Children[1].Title)
	}
}

