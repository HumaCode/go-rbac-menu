package rbacmenu_test

import (
	"testing"

	"github.com/HumaCode/go-rbac-menu/rbacmenu"
)

func TestNewULID(t *testing.T) {
	id := rbacmenu.NewULID()
	if len(id) != 26 {
		t.Errorf("expected ULID length 26, got %d", len(id))
	}
}

func TestGroupMenusByCategoryAnd3Levels(t *testing.T) {
	parentID := "01HQ0000000000000000000001"
	childID := "01HQ0000000000000000000002"
	grandchildID := "01HQ0000000000000000000003"

	grandchild := &rbacmenu.Menu{
		BaseModel: rbacmenu.BaseModel{ID: grandchildID},
		ParentID:  &childID,
		Title:     "Childmenu Level 3",
		OrderNo:   1,
	}

	submenu := &rbacmenu.Menu{
		BaseModel: rbacmenu.BaseModel{ID: childID},
		ParentID:  &parentID,
		Title:     "Submenu Level 2",
		OrderNo:   1,
		Children:  []*rbacmenu.Menu{grandchild},
	}

	rootMenu := &rbacmenu.Menu{
		BaseModel: rbacmenu.BaseModel{ID: parentID},
		Category:  "LAYANAN PPID",
		Title:     "Menu Level 1",
		OrderNo:   2,
		Children:  []*rbacmenu.Menu{submenu},
	}

	rootDashboard := &rbacmenu.Menu{
		BaseModel: rbacmenu.BaseModel{ID: "01HQ0000000000000000000004"},
		Category:  "UTAMA",
		Title:     "Dashboard",
		OrderNo:   1,
	}

	roots := []*rbacmenu.Menu{rootDashboard, rootMenu}
	categories := rbacmenu.GroupMenusByCategory(roots)

	if len(categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(categories))
	}

	// Verify Category 1: UTAMA -> Dashboard
	if categories[0].Category != "UTAMA" {
		t.Errorf("expected category[0] to be UTAMA, got %s", categories[0].Category)
	}
	if len(categories[0].Menus) != 1 || categories[0].Menus[0].Title != "Dashboard" {
		t.Errorf("expected Dashboard in UTAMA category")
	}

	// Verify Category 2: LAYANAN PPID -> Menu -> Submenu -> Childmenu
	if categories[1].Category != "LAYANAN PPID" {
		t.Errorf("expected category[1] to be LAYANAN PPID, got %s", categories[1].Category)
	}
	catMenus := categories[1].Menus
	if len(catMenus) != 1 || len(catMenus[0].Children) != 1 || len(catMenus[0].Children[0].Children) != 1 {
		t.Fatalf("expected 3 levels of nesting, got: %+v", catMenus)
	}

	if catMenus[0].Children[0].Children[0].Title != "Childmenu Level 3" {
		t.Errorf("expected Level 3 title to match, got %s", catMenus[0].Children[0].Children[0].Title)
	}
}
