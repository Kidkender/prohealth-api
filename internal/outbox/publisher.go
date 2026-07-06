package outbox

import (
	"encoding/json"
	"fmt"

	"github.com/kidkender/prohealth-api/internal/event"
	"github.com/kidkender/prohealth-api/internal/model"
	"gorm.io/gorm"
)

type Publisher struct{}

func NewPublisher() *Publisher {
	return &Publisher{}
}

func (p *Publisher) Publish(tx *gorm.DB, e event.Event) error {
	jobs, err := buildJobs(e)
	if err != nil {
		return fmt.Errorf("Build outbox jobs for %s: %w", e.Name(), err)
	}

	for _, job := range jobs {
		if err := tx.Create(&job).Error; err != nil {
			return fmt.Errorf("Create outbox job for %s: %w", job.Type, err)
		}
	}
	return nil

}

func buildJobs(e event.Event) ([]model.OutboxJob, error) {
	switch e := e.(type) {
	case event.AppointmentCreated:
		return appointmentCreatedJobs(e)
	default:
		return nil, fmt.Errorf("Unsupported event type: %T", e)
	}
}

func appointmentCreatedJobs(ev event.AppointmentCreated) ([]model.OutboxJob, error) {
	emailPayload, err := json.Marshal(map[string]string{
		"to":     ev.PatientEmail,
		"name":   ev.PatientName,
		"date":   ev.Date,
		"time":   ev.Time,
		"reason": ev.Reason,
	})

	if err != nil {
		return nil, err
	}

	return []model.OutboxJob{
		{
			Type:    string(model.OutboxJobTypeAppointmentCreated),
			Status:  model.OutboxJobStatusPending,
			Payload: emailPayload,
		},
	}, nil
}
