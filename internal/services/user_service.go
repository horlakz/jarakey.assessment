package services

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/horlakz/jarakey.assessment/internal/dto"
	"github.com/horlakz/jarakey.assessment/internal/repositories"
	"github.com/horlakz/jarakey.assessment/internal/utils"
)

type UserService struct {
	users repositories.UserRepository
}

func NewUserService(users repositories.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Me(userID string) (*dto.MeResponse, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	if user == nil {
		return nil, utils.NewAppError(fiber.StatusUnauthorized, "unknown_user", "authenticated user no longer exists", nil)
	}
	return &dto.MeResponse{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}
