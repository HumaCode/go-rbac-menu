package middleware

import (
	"strings"

	"go-rbac-menu/config"
	"go-rbac-menu/internal/repository"
	"go-rbac-menu/pkg/response"
	"go-rbac-menu/pkg/utils"

	"github.com/nelthaarion/breeze/v2"
)

func AuthMiddleware(cfg *config.Config, userRepo repository.UserRepository) breeze.HandlerFunc {
	return func(ctx *breeze.Context) error {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			return response.Error(ctx, 401, "Authorization header required", nil)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return response.Error(ctx, 401, "Invalid authorization header format (Bearer <token>)", nil)
		}

		tokenStr := parts[1]
		claims, err := utils.ValidateToken(cfg.JWTSecret, tokenStr)
		if err != nil {
			return response.Error(ctx, 401, "Invalid or expired token", err)
		}

		// Strictly ensure user is still active in database (is_active = 1)
		user, err := userRepo.FindActiveByID(claims.UserID)
		if err != nil || user.IsActive != 1 {
			return response.Error(ctx, 401, "User account is inactive or not found", nil)
		}

		ctx.Set("user", claims)
		ctx.Set("user_id", claims.UserID)
		ctx.Set("user_roles", claims.Roles)

		return ctx.Next()
	}
}

