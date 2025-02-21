package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/savioruz/smrv2-api/pkg/helper"
	"github.com/savioruz/smrv2-api/pkg/jwt"
)

type AuthMiddleware struct {
	jwtService *jwt.JWTServiceImpl
}

func NewAuthMiddleware(jwtService *jwt.JWTServiceImpl) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

func (m *AuthMiddleware) ValidateJWT() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(helper.SingleError("authorization", "REQUIRED"))
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(helper.SingleError("authorization", "INVALID_FORMAT"))
		}

		claims, err := m.jwtService.ValidateToken(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(helper.SingleError("authorization", "INVALID_TOKEN"))
		}

		if claims != nil {
			c.Locals("userID", claims.UserID)
			c.Locals("email", claims.Email)
			c.Locals("level", claims.Level)
		}

		return c.Next()
	}
}
