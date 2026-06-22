package worker

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"time"

	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/internal/service"
	"gorm.io/gorm"
)

const maxRetry = 3

type OutboxWorker struct {
	db          *gorm.DB
	mailService *service.MailService
}

func NewOutboxWorker(db *gorm.DB, mailService *service.MailService) *OutboxWorker {
	return &OutboxWorker{
		db:          db,
		mailService: mailService,
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
		if err := w.handle(ctx, job); err != nil {
			log.Printf("outbox: job %d failed: %v", job.ID, err)
			w.db.WithContext(ctx).Model(&job).Updates(map[string]any{
				"status":      model.OutboxJobStatusFailed,
				"retry_count": job.RetryCount + 1,
				"last_error":  err.Error(),
			})
			continue
		}
		slog.Info("outbox: job completed", "id", job.ID)
		w.db.WithContext(ctx).Model(&job).Update("status", model.OutboxJobStatusCompleted)
	}
}

func (w *OutboxWorker) handle(ctx context.Context, job model.OutboxJob) error {

	switch model.OutboxJobType(job.Type) {
	case model.OutboxJobTypeAppointmentCreated:
		return w.handleAppointmentCreated(ctx, job.Payload)
	default:
		log.Printf("outbox: unknown job type %q, skipping", job.Type)
		return nil
	}
}

func (w *OutboxWorker) handleAppointmentCreated(ctx context.Context, payload []byte) error {
	var p struct {
		To     string `json:"to"`
		Name   string `json:"name"`
		Time   string `json:"time"`
		Date   string `json:"date"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	return w.mailService.SendAppointmentCreated(ctx, p.To, service.AppointmentCreatedData{
		PatientName:     p.Name,
		ShiftDate:       p.Date,
		AppointmentTime: p.Time,
		Reason:          p.Reason,
	})
}
