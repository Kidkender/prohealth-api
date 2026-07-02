package notification

import (
	"encoding/json"

	"github.com/kidkender/prohealth-api/internal/service"
)

func handleAppointmentCreated(mailService *service.MailService, payload []byte) error {
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
	return mailService.SendAppointmentCreated(p.To, service.AppointmentCreatedData{
		PatientName:     p.Name,
		ShiftDate:       p.Date,
		AppointmentTime: p.Time,
		Reason:          p.Reason,
	})
}
