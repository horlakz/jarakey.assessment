package services

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/horlakz/jarakey.assessment/internal/dto"
	"github.com/horlakz/jarakey.assessment/internal/repositories"
	"github.com/horlakz/jarakey.assessment/internal/utils"
)

type AuthService struct {
	users repositories.UserRepository
	jwt   *utils.JWTManager
}

func NewAuthService(users repositories.UserRepository, jwt *utils.JWTManager) *AuthService {
	return &AuthService{users: users, jwt: jwt}
}

func (s *AuthService) Login(input dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.users.FindByEmail(input.Email)
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	if user == nil || !utils.VerifyPassword(input.Password, user.PasswordHash) {
		return nil, utils.NewAppError(fiber.StatusUnauthorized, "invalid_credentials", "invalid email or password", nil)
	}

	token, err := s.jwt.IssueToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("issue jwt: %w", err)
	}

	return &dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	}, nil
}
