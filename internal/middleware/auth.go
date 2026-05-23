package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/horlakz/jarakey.assessment/internal/utils"
)

const userIDContextKey = "userID"

func AuthRequired(jwtManager *utils.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return utils.NewAppError(fiber.StatusUnauthorized, "missing_authorization", "authorization header is required", nil)
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return utils.NewAppError(fiber.StatusUnauthorized, "invalid_authorization", "authorization header must use Bearer token", nil)
		}

		claims, err := jwtManager.ParseToken(parts[1])
		if err != nil {
			return utils.NewAppError(fiber.StatusUnauthorized, "invalid_token", "invalid or expired access token", err)
		}

		c.Locals(userIDContextKey, claims.UserID)
		return c.Next()
	}
}

func UserIDFromContext(c *fiber.Ctx) string {
	value, _ := c.Locals(userIDContextKey).(string)
	return value
}
