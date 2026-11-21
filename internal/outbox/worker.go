package outbox

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	maxRetry  = 3
	batchSize = 10
)

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
	defer ticker.Stop()

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

// process claims a batch of pending jobs and publishes them. The batch is selected
// with FOR UPDATE SKIP LOCKED inside a transaction, so when several API instances
// run this worker, each row is claimed by one of them: the others skip it instead
// of publishing it a second time. The row locks are held until the status updates
// commit. If the process dies after XADD but before the commit, the job is sent
// again on the next tick, so consumers must stay idempotent.
func (w *OutboxWorker) process(ctx context.Context) {
	slog.Info("outbox: tick - checking pending jobs")
	err := w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var jobs []model.OutboxJob
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ? AND retry_count < ?", model.OutboxJobStatusPending, maxRetry).
			Order("created_at ASC").
			Limit(batchSize).
			Find(&jobs).Error; err != nil {
			return fmt.Errorf("fetch jobs: %w", err)
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
				if err := tx.Model(&job).Updates(map[string]any{
					"retry_count": job.RetryCount + 1,
					"last_error":  err.Error(),
				}).Error; err != nil {
					return fmt.Errorf("record failure for job %d: %w", job.ID, err)
				}
				continue
			}

			if err := tx.Model(&job).Update("status", model.OutboxJobStatusPublished).Error; err != nil {
				return fmt.Errorf("mark job %d published: %w", job.ID, err)
			}
			slog.Info("outbox: job completed", "id", job.ID)
		}
		return nil
	})
	if err != nil {
		log.Printf("outbox: %v", err)
	}
}
