package outbox

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const maxRetry = 3

type OutboxWorker struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewOutboxWorker(db *gorm.DB, redis *redis.Client) *OutboxWorker {
	return &OutboxWorker{
		db:    db,
		redis: redis,
	}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			w.process(ctx)
		case <-ctx.Done():
			log.Println("outbox worker stopped")
			return
		}
	}
}

func (w *OutboxWorker) process(ctx context.Context) {
	slog.Info("outbox: tick - checking pending jobs")
	var jobs []model.OutboxJob
	if err := w.db.WithContext(ctx).
		Where("status = ? AND retry_count < ?", model.OutboxJobStatusPending, maxRetry).
		Order("created_at ASC").
		Limit(10).
		Find(&jobs).Error; err != nil {
		log.Printf("outbox: fetch jobs: %v", err)
		return
	}

	for _, job := range jobs {
		_, err := w.redis.XAdd(ctx, &redis.XAddArgs{
			Stream: model.OutboxJobType(job.Type).Stream(),
			Values: map[string]any{
				"type":    job.Type,
				"payload": job.Payload,
			},
		}).Result()

		if err != nil {
			w.db.WithContext(ctx).
				Model(&job).
				Updates(map[string]any{
					"retry_count": job.RetryCount + 1,
					"last_error":  err.Error(),
				})
			continue
		}

		slog.Info("outbox: job completed", "id", job.ID)
		w.db.WithContext(ctx).Model(&job).Update("status", model.OutboxJobStatusPublished)
	}
}
