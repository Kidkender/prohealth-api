package audit

import (
	"github.com/kidkender/prohealth-api/internal/model"
	"gorm.io/gorm"
)

func Record(tx *gorm.DB, hospitalID uint, actorID *uint, action model.AuditActionType, entityType string, entityID uint) error {
	return tx.Create(&model.AuditLog{
		HospitalID: hospitalID,
		ActorID:    actorID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
	}).Error
}
