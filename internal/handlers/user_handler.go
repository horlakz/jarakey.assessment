package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/horlakz/jarakey.assessment/internal/middleware"
	"github.com/horlakz/jarakey.assessment/internal/services"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Me(c *fiber.Ctx) error {
	response, err := h.service.Me(middleware.UserIDFromContext(c))
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(response)
}
