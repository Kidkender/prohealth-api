package container

import (
	"github.com/kidkender/prohealth-api/internal/config"
	"github.com/kidkender/prohealth-api/internal/service"
	"gorm.io/gorm"
)

type Container struct {
	AuthService        *service.AuthService
	UserService        *service.UserService
	ScheduleService    *service.ScheduleService
	AppointmentService *service.AppointmentService
}

func New(db *gorm.DB, cfg *config.Config) *Container {
	mailService := service.NewMailService(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.MailFrom,
	)

	return &Container{
		AuthService:        service.NewAuthService(db),
		UserService:        service.NewUserService(db),
		ScheduleService:    service.NewScheduleService(db),
		AppointmentService: service.NewAppointmentService(db, mailService),
	}
}
