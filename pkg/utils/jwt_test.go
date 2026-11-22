package utils_test

import (
	"testing"

	"go-rbac-menu/pkg/utils"
)

func TestJWTTokens(t *testing.T) {
	secret := "test-secret-key"
	userID := utils.NewULID()
	username := "john_doe"
	email := "john@example.com"
	roles := []string{"admin", "user"}

	token, err := utils.GenerateToken(secret, 2, userID, username, email, roles)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := utils.ValidateToken(secret, token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected user_id %s, got %s", userID, claims.UserID)
	}
	if claims.Username != username {
		t.Errorf("expected username %s, got %s", username, claims.Username)
	}
	if len(claims.Roles) != 2 || claims.Roles[0] != "admin" {
		t.Errorf("expected roles %v, got %v", roles, claims.Roles)
	}
}

