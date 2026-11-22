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

