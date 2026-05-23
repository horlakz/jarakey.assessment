package database

import (
	"errors"
	"fmt"

	"github.com/horlakz/jarakey.assessment/internal/config"
	"github.com/horlakz/jarakey.assessment/internal/entities"
	"github.com/horlakz/jarakey.assessment/internal/utils"
	"gorm.io/gorm"
)

const (
	AdminRoleCode    = "admin"
	ResidentRoleCode = "resident"
	GateOpenCode     = "gate.open"
	DefaultEstate    = "Maple Residency"
)

type SeedResult struct {
	UserID   string
	EstateID string
}

func SeedDefaults(db *gorm.DB, cfg config.Config) (*SeedResult, error) {
	result := &SeedResult{}

	err := db.Transaction(func(tx *gorm.DB) error {
		adminRole := entities.Role{Code: AdminRoleCode, Name: "Admin"}
		if err := firstOrCreateByCode(tx, &adminRole); err != nil {
			return err
		}

		residentRole := entities.Role{Code: ResidentRoleCode, Name: "Resident"}
		if err := firstOrCreateByCode(tx, &residentRole); err != nil {
			return err
		}

		permission := entities.Permission{Code: GateOpenCode, Description: "Open estate gate"}
		if err := firstOrCreateByCode(tx, &permission); err != nil {
			return err
		}

		var count int64
		if err := tx.Model(&entities.RolePermission{}).
			Where("role_id = ? AND permission_id = ?", adminRole.ID, permission.ID).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := tx.Create(&entities.RolePermission{RoleID: adminRole.ID, PermissionID: permission.ID}).Error; err != nil {
				return err
			}
		}

		estate := entities.Estate{Name: DefaultEstate}
		if err := tx.Where("name = ?", estate.Name).FirstOrCreate(&estate).Error; err != nil {
			return err
		}

		passwordHash, err := utils.HashPassword(cfg.DefaultPassword)
		if err != nil {
			return err
		}

		user := entities.User{Email: cfg.DefaultEmail}
		err = tx.Where("email = ?", user.Email).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user.PasswordHash = passwordHash
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		if err := tx.FirstOrCreate(&entities.UserEstateMembership{}, entities.UserEstateMembership{
			UserID:   user.ID,
			EstateID: estate.ID,
		}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ? AND estate_id = ?", user.ID, estate.ID).
			Delete(&entities.UserRoleAssignment{}).Error; err != nil {
			return err
		}

		if err := tx.Create(&entities.UserRoleAssignment{
			UserID:   user.ID,
			EstateID: estate.ID,
			RoleID:   adminRole.ID,
		}).Error; err != nil {
			return err
		}

		result.UserID = user.ID
		result.EstateID = estate.ID
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("seed defaults: %w", err)
	}

	return result, nil
}

func firstOrCreateByCode(tx *gorm.DB, model interface{}) error {
	return tx.Where(model).FirstOrCreate(model).Error
}
