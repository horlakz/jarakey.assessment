package repositories

import (
	"github.com/horlakz/jarakey.assessment/internal/entities"
	"gorm.io/gorm"
)

type AuditRepository interface {
	Create(log *entities.GateAccessAuditLog) error
}

type GormAuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *GormAuditRepository {
	return &GormAuditRepository{db: db}
}

func (r *GormAuditRepository) Create(log *entities.GateAccessAuditLog) error {
	return r.db.Create(log).Error
}
