package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/horlakz/jarakey.assessment/internal/middleware"
	"github.com/horlakz/jarakey.assessment/internal/services"
)

type GateHandler struct {
	service *services.GateService
}

func NewGateHandler(service *services.GateService) *GateHandler {
	return &GateHandler{service: service}
}

func (h *GateHandler) Open(c *fiber.Ctx) error {
	response, err := h.service.Open(middleware.UserIDFromContext(c), c.Get("X-Estate-ID"))
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(response)
}
