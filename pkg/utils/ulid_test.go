package utils_test

import (
	"testing"

	"go-rbac-menu/pkg/utils"
)

func TestNewULID(t *testing.T) {
	id1 := utils.NewULID()
	id2 := utils.NewULID()

	if len(id1) != 26 {
		t.Errorf("expected ULID length to be 26, got %d", len(id1))
	}
	if len(id2) != 26 {
		t.Errorf("expected ULID length to be 26, got %d", len(id2))
	}
	if id1 == id2 {
		t.Errorf("expected unique ULIDs, got duplicate %s", id1)
	}
}
