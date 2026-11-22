package utils_test

import (
	"testing"

	"go-rbac-menu/pkg/utils"
)

func TestPasswordHashing(t *testing.T) {
	password := "secret123"
	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	if !utils.CheckPasswordHash(password, hash) {
		t.Errorf("expected valid password check to pass")
	}

	if utils.CheckPasswordHash("wrongpassword", hash) {
		t.Errorf("expected invalid password check to fail")
	}
}

