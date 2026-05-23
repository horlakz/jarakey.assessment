package authorization

import (
	"fmt"

	"github.com/horlakz/jarakey.assessment/internal/database"
	"github.com/horlakz/jarakey.assessment/internal/entities"
	"github.com/horlakz/jarakey.assessment/internal/repositories"
)

type Decision struct {
	Allowed bool
	Reason  string
}

type Service struct {
	repo repositories.AuthorizationRepository
}

func NewService(repo repositories.AuthorizationRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CheckPermission(userID, estateID, permissionCode string) (*Decision, error) {
	member, err := s.repo.IsMember(userID, estateID)
	if err != nil {
		return nil, fmt.Errorf("check estate membership: %w", err)
	}
	if !member {
		return &Decision{Allowed: false, Reason: "user is not a member of the active estate"}, nil
	}

	override, err := s.repo.FindOverride(userID, estateID, permissionCode)
	if err != nil {
		return nil, fmt.Errorf("find override: %w", err)
	}
	if override != nil && override.Effect == entities.OverrideDeny {
		return &Decision{Allowed: false, Reason: "permission explicitly denied by override"}, nil
	}
	if override != nil && override.Effect == entities.OverrideAllow {
		return &Decision{Allowed: true, Reason: "permission explicitly allowed by override"}, nil
	}

	allowed, err := s.repo.HasRolePermission(userID, estateID, permissionCode)
	if err != nil {
		return nil, fmt.Errorf("check role permission: %w", err)
	}
	if allowed {
		return &Decision{Allowed: true, Reason: "permission granted by estate role"}, nil
	}

	return &Decision{Allowed: false, Reason: "permission not granted in active estate"}, nil
}

func (s *Service) DowngradeToResident(userID, estateID string) error {
	return s.repo.ReplaceRoleAssignment(userID, estateID, database.ResidentRoleCode)
}
