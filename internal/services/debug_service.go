package services

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/horlakz/jarakey.assessment/internal/authorization"
	"github.com/horlakz/jarakey.assessment/internal/dto"
	"github.com/horlakz/jarakey.assessment/internal/utils"
)

type DebugService struct {
	authorizer *authorization.Service
}

func NewDebugService(authorizer *authorization.Service) *DebugService {
	return &DebugService{authorizer: authorizer}
}

func (s *DebugService) DowngradeRole(userID, estateID string) (*dto.DebugDowngradeResponse, error) {
	if estateID == "" {
		return nil, utils.NewAppError(fiber.StatusBadRequest, "missing_estate_context", "X-Estate-ID header is required", nil)
	}

	decision, err := s.authorizer.CheckPermission(userID, estateID, "gate.open")
	if err != nil {
		return nil, fmt.Errorf("validate active estate: %w", err)
	}
	if !decision.Allowed && decision.Reason == "user is not a member of the active estate" {
		return nil, utils.NewAppError(fiber.StatusForbidden, "forbidden", decision.Reason, nil)
	}

	if err := s.authorizer.DowngradeToResident(userID, estateID); err != nil {
		return nil, fmt.Errorf("downgrade role: %w", err)
	}

	return &dto.DebugDowngradeResponse{Message: "role downgraded to resident"}, nil
}
