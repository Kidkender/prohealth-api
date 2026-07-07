package dto

import (
	"time"

	"github.com/kidkender/prohealth-api/internal/model"
)

type FilterAuditLog struct {
	PaginationQuery
	ActorID    *uint                  `form:"actor_id"`
	HospitalID *uint                  `form:"hospital_id"`
	Action     *model.AuditActionType `form:"action"`
	EntityType *string                `form:"entity_type"`
	EntityID   *uint                  `form:"entity_id"`
	From       *time.Time             `form:"from" time_format:"2006-01-02"`
	To         *time.Time             `form:"to" time_format:"2006-01-02"`
}
