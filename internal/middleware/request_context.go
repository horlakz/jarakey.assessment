package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func RequestID() fiber.Handler {
	return requestid.New()
}

func StructuredLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		slog.Info("http_request",
			"request_id", c.Locals("requestid"),
			"method", c.Method(),
			"path", c.OriginalURL(),
			"status", c.Response().StatusCode(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return err
	}
}
