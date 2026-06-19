package service

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"

	tmpl "github.com/kidkender/prohealth-api/internal/template"
)

type MailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewMailService(host, port, username, password, from string) *MailService {
	return &MailService{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
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

	return s.send(to, "Xác nhận lịch hẹn - ProHealth", buf.String())
}

func (s *MailService) Send(ctx context.Context, to string, subject string, body string) error {
	return s.send(to, subject, body)
}

func (s *MailService) send(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, body,
	)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
}
