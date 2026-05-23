package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        string `gorm:"primaryKey;type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (b *BaseModel) BeforeCreate(_ *gorm.DB) error {
	if b.ID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		b.ID = id.String()
	}
	return nil
}

type User struct {
	BaseModel
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
}

type Estate struct {
	BaseModel
	Name string `gorm:"not null"`
}

type Role struct {
	BaseModel
	Code string `gorm:"uniqueIndex;not null"`
	Name string `gorm:"not null"`
}

type Permission struct {
	BaseModel
	Code        string `gorm:"uniqueIndex;not null"`
	Description string `gorm:"not null"`
}

type RolePermission struct {
	RoleID       string `gorm:"primaryKey;type:text"`
	PermissionID string `gorm:"primaryKey;type:text"`
}

type UserEstateMembership struct {
	UserID    string `gorm:"primaryKey;type:text"`
	EstateID  string `gorm:"primaryKey;type:text;index"`
	CreatedAt time.Time
}

type UserRoleAssignment struct {
	UserID    string `gorm:"primaryKey;type:text"`
	EstateID  string `gorm:"primaryKey;type:text;index"`
	RoleID    string `gorm:"primaryKey;type:text"`
	CreatedAt time.Time
}

type OverrideEffect string

const (
	OverrideAllow OverrideEffect = "ALLOW"
	OverrideDeny  OverrideEffect = "DENY"
)

type UserPermissionOverride struct {
	UserID         string         `gorm:"primaryKey;type:text"`
	EstateID       string         `gorm:"primaryKey;type:text;index"`
	PermissionCode string         `gorm:"primaryKey;type:text"`
	Effect         OverrideEffect `gorm:"type:text;not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type GateAccessAuditLog struct {
	BaseModel
	UserID   string `gorm:"index;type:text"`
	EstateID string `gorm:"index;type:text"`
	Endpoint string `gorm:"not null"`
	Allowed  bool   `gorm:"not null"`
	Reason   string `gorm:"not null"`
}
