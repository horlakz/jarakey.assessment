package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/horlakz/jarakey.assessment/internal/dto"
	"github.com/horlakz/jarakey.assessment/internal/services"
	"github.com/horlakz/jarakey.assessment/internal/utils"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var request dto.LoginRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.NewAppError(fiber.StatusBadRequest, "invalid_json", "request body must be valid JSON", err)
	}
	if err := utils.Validate(request); err != nil {
		return err
	}

	response, err := h.service.Login(request)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(response)
}
