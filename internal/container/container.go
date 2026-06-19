package container

import (
	"github.com/kidkender/prohealth-api/internal/config"
	"github.com/kidkender/prohealth-api/internal/service"
	"github.com/resend/resend-go/v3"
	"gorm.io/gorm"
)

type Container struct {
	AuthService        *service.AuthService
	UserService        *service.UserService
	ScheduleService    *service.ScheduleService
	AppointmentService *service.AppointmentService
}

func New(db *gorm.DB, cfg *config.Config) *Container {
	resendClient := resend.NewClient(cfg.ResendAPIKey)
	mailService := service.NewMailService(resendClient, cfg.MailFrom)

	return &Container{
		AuthService:        service.NewAuthService(db),
		UserService:        service.NewUserService(db),
		ScheduleService:    service.NewScheduleService(db),
		AppointmentService: service.NewAppointmentService(db, mailService),
	}
}
