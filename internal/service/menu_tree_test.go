package service

import (
	"testing"

	"github.com/HumaCode/go-rbac-menu/internal/domain"
)

func TestSortMenuTreeAnd3LevelHierarchy(t *testing.T) {
	parentID := "01HQ0000000000000000000001"
	childID := "01HQ0000000000000000000002"
	grandchildID := "01HQ0000000000000000000003"

	grandchild := &domain.Menu{
		BaseModel: domain.BaseModel{ID: grandchildID},
		ParentID:  &childID,
		Title:     "Permohonan Baru (Childmenu Level 3)",
		OrderNo:   1,
	}

	child := &domain.Menu{
		BaseModel: domain.BaseModel{ID: childID},
		ParentID:  &parentID,
		Title:     "Permohonan Informasi (Submenu Level 2)",
		OrderNo:   1,
		Children:  []*domain.Menu{grandchild},
	}

	root := &domain.Menu{
		BaseModel: domain.BaseModel{ID: parentID},
		Category:  "LAYANAN PPID",
		Title:     "Layanan (Menu Level 1)",
		OrderNo:   2,
		Children:  []*domain.Menu{child},
	}

	rootDashboard := &domain.Menu{
		BaseModel: domain.BaseModel{ID: "01HQ0000000000000000000004"},
		Category:  "UTAMA",
		Title:     "Dashboard",
		OrderNo:   1,
	}

	tree := []*domain.Menu{root, rootDashboard}
	sortMenuTree(tree)

	// Verify Level 1 ordering
	if tree[0].Title != "Dashboard" {
		t.Errorf("expected tree[0] to be Dashboard, got %s", tree[0].Title)
	}
	if tree[1].Title != "Layanan (Menu Level 1)" {
		t.Errorf("expected tree[1] to be Layanan, got %s", tree[1].Title)
	}

	// Verify Level 2 (Submenu)
	if len(tree[1].Children) != 1 || tree[1].Children[0].Title != "Permohonan Informasi (Submenu Level 2)" {
		t.Fatalf("expected 1 submenu, got %v", tree[1].Children)
	}

	// Verify Level 3 (Childmenu / Grandchild)
	subChildren := tree[1].Children[0].Children
	if len(subChildren) != 1 || subChildren[0].Title != "Permohonan Baru (Childmenu Level 3)" {
		t.Fatalf("expected 1 childmenu at level 3, got %v", subChildren)
	}

	// Test Grouping by Category
	categories := GroupDomainMenusByCategory(tree)
	if len(categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(categories))
	}

	if categories[0].Category != "UTAMA" {
		t.Errorf("expected category[0] to be UTAMA, got %s", categories[0].Category)
	}
	if categories[1].Category != "LAYANAN PPID" {
		t.Errorf("expected category[1] to be LAYANAN PPID, got %s", categories[1].Category)
	}
}
