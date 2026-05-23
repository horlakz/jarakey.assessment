package database

import (
	"fmt"

	"github.com/horlakz/jarakey.assessment/internal/config"
	"github.com/horlakz/jarakey.assessment/internal/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DatabaseDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.AutoMigrate(
		&entities.User{},
		&entities.Estate{},
		&entities.Role{},
		&entities.Permission{},
		&entities.RolePermission{},
		&entities.UserEstateMembership{},
		&entities.UserRoleAssignment{},
		&entities.UserPermissionOverride{},
		&entities.GateAccessAuditLog{},
	); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	return db, nil
}
