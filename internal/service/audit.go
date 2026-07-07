package service

import (
	"context"

	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/model"
	"gorm.io/gorm"
)

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

func (s *AuditService) GetUserActivity(ctx context.Context, userID uint, filter dto.FilterAuditLog) ([]model.AuditLog, int64, error) {
	query := s.db.WithContext(ctx).Model(&model.AuditLog{}).Where("actor_id = ?", userID)

	if filter.Action != nil {
		query = query.Where("action = ?", *filter.Action)
	}
	if filter.EntityType != nil {
		query = query.Where("entity_type = ?", *filter.EntityType)
	}
	if filter.From != nil {
		query = query.Where("created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where("created_at <= ?", *filter.To)
	}

	var count int64
	if err := query.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	var logs []model.AuditLog
	if err := query.
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset()).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, count, nil
}
