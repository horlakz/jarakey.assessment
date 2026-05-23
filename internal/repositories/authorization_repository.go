package repositories

import (
	"errors"

	"github.com/horlakz/jarakey.assessment/internal/entities"
	"gorm.io/gorm"
)

type AuthorizationRepository interface {
	IsMember(userID, estateID string) (bool, error)
	FindOverride(userID, estateID, permissionCode string) (*entities.UserPermissionOverride, error)
	HasRolePermission(userID, estateID, permissionCode string) (bool, error)
	ReplaceRoleAssignment(userID, estateID, roleCode string) error
	UpsertOverride(override *entities.UserPermissionOverride) error
}

type GormAuthorizationRepository struct {
	db *gorm.DB
}

func NewAuthorizationRepository(db *gorm.DB) *GormAuthorizationRepository {
	return &GormAuthorizationRepository{db: db}
}

func (r *GormAuthorizationRepository) IsMember(userID, estateID string) (bool, error) {
	var count int64
	err := r.db.Model(&entities.UserEstateMembership{}).
		Where("user_id = ? AND estate_id = ?", userID, estateID).
		Count(&count).Error
	return count > 0, err
}

func (r *GormAuthorizationRepository) FindOverride(userID, estateID, permissionCode string) (*entities.UserPermissionOverride, error) {
	var override entities.UserPermissionOverride
	if err := r.db.Where(
		"user_id = ? AND estate_id = ? AND permission_code = ?",
		userID,
		estateID,
		permissionCode,
	).First(&override).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &override, nil
}

func (r *GormAuthorizationRepository) HasRolePermission(userID, estateID, permissionCode string) (bool, error) {
	var count int64
	err := r.db.Table("user_role_assignments AS ura").
		Joins("JOIN role_permissions AS rp ON rp.role_id = ura.role_id").
		Joins("JOIN permissions AS p ON p.id = rp.permission_id").
		Where("ura.user_id = ? AND ura.estate_id = ? AND p.code = ?", userID, estateID, permissionCode).
		Count(&count).Error
	return count > 0, err
}

func (r *GormAuthorizationRepository) ReplaceRoleAssignment(userID, estateID, roleCode string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var role entities.Role
		if err := tx.Where("code = ?", roleCode).First(&role).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ? AND estate_id = ?", userID, estateID).
			Delete(&entities.UserRoleAssignment{}).Error; err != nil {
			return err
		}

		return tx.Create(&entities.UserRoleAssignment{
			UserID:   userID,
			EstateID: estateID,
			RoleID:   role.ID,
		}).Error
	})
}

func (r *GormAuthorizationRepository) UpsertOverride(override *entities.UserPermissionOverride) error {
	return r.db.Save(override).Error
}
