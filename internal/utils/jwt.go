package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type SessionClaims struct {
	UserID    string `json:"sub"`
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret []byte
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{secret: []byte(secret)}
}

func (m *JWTManager) IssueToken(userID string) (string, error) {
	now := time.Now().UTC()
	claims := SessionClaims{
		UserID:    userID,
		SessionID: fmt.Sprintf("sess_%d", now.UnixNano()),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) ParseToken(raw string) (*SessionClaims, error) {
	token, err := jwt.ParseWithClaims(raw, &SessionClaims{}, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*SessionClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
