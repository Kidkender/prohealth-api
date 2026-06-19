package service

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	tmpl "github.com/kidkender/prohealth-api/internal/template"
	"github.com/resend/resend-go/v3"
)

type MailService struct {
	client *resend.Client
	from   string
}

func NewMailService(client *resend.Client, from string) *MailService {
	return &MailService{client: client, from: from}
}

type AppointmentCreatedData struct {
	PatientName     string
	ShiftDate       string
	AppointmentTime string
	Reason          string
}

func (s *MailService) SendAppointmentCreated(ctx context.Context, to string, data AppointmentCreatedData) error {
	t, err := template.ParseFS(tmpl.MailFS, "mail/appointment_created.html")
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	_, err = s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: "Xác nhận lịch hẹn - ProHealth",
		Html:    buf.String(),
	})
	return err
}

func (s *MailService) Send(ctx context.Context, to string, subject string, body string) error {
	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		Html:    body,
	})
	return err
}
