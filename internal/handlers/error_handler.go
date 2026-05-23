package handlers

import (
	"errors"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/horlakz/jarakey.assessment/internal/utils"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	var appErr *utils.AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.StatusCode).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    appErr.Code,
				"message": appErr.Message,
			},
		})
	}

	var validationErr validator.ValidationErrors
	if errors.As(err, &validationErr) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "validation_failed",
				"message": validationErr.Error(),
			},
		})
	}

	slog.Error("unhandled error", "error", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    "internal_error",
			"message": "internal server error",
		},
	})
}
