package dto

import (
	"time"

	"github.com/kidkender/prohealth-api/internal/model"
)

type FilterAuditLog struct {
	PaginationQuery
	Action     *model.AuditActionType `form:"action"`
	EntityType *string                `form:"entity_type"`
	From       *time.Time             `form:"from" time_format:"2006-01-02"`
	To         *time.Time             `form:"to" time_format:"2006-01-02"`
}
