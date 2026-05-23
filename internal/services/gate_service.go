package services

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/horlakz/jarakey.assessment/internal/authorization"
	"github.com/horlakz/jarakey.assessment/internal/dto"
	"github.com/horlakz/jarakey.assessment/internal/entities"
	"github.com/horlakz/jarakey.assessment/internal/repositories"
	"github.com/horlakz/jarakey.assessment/internal/utils"
)

type GateService struct {
	authorizer *authorization.Service
	audits     repositories.AuditRepository
}

func NewGateService(authorizer *authorization.Service, audits repositories.AuditRepository) *GateService {
	return &GateService{
		authorizer: authorizer,
		audits:     audits,
	}
}

func (s *GateService) Open(userID, estateID string) (*dto.GateOpenResponse, error) {
	if estateID == "" {
		return nil, utils.NewAppError(fiber.StatusBadRequest, "missing_estate_context", "X-Estate-ID header is required", nil)
	}

	decision, err := s.authorizer.CheckPermission(userID, estateID, "gate.open")
	if err != nil {
		return nil, fmt.Errorf("authorize gate open: %w", err)
	}

	if auditErr := s.audits.Create(&entities.GateAccessAuditLog{
		UserID:   userID,
		EstateID: estateID,
		Endpoint: "/gate/open",
		Allowed:  decision.Allowed,
		Reason:   decision.Reason,
	}); auditErr != nil {
		return nil, fmt.Errorf("write audit log: %w", auditErr)
	}

	if !decision.Allowed {
		return nil, utils.NewAppError(fiber.StatusForbidden, "forbidden", decision.Reason, nil)
	}

	return &dto.GateOpenResponse{Message: "gate opened"}, nil
}
