package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/horlakz/jarakey.assessment/internal/middleware"
	"github.com/horlakz/jarakey.assessment/internal/services"
)

type DebugHandler struct {
	service *services.DebugService
}

func NewDebugHandler(service *services.DebugService) *DebugHandler {
	return &DebugHandler{service: service}
}

func (h *DebugHandler) DowngradeRole(c *fiber.Ctx) error {
	response, err := h.service.DowngradeRole(middleware.UserIDFromContext(c), c.Get("X-Estate-ID"))
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(response)
}
