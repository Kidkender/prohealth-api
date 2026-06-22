package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"log/slog"
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
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: s.host})
	if err != nil {
		return fmt.Errorf("smtp dial tls: %w", err)
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Close()

	if err := client.Auth(smtp.PlainAuth("", s.username, s.password, s.host)); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, body,
	)
	if _, err := fmt.Fprint(w, msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	slog.Info("send mail success", "to", to)
	return w.Close()
}
