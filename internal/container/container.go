package container

import (
	"github.com/kidkender/prohealth-api/internal/cache"
	"github.com/kidkender/prohealth-api/internal/config"
	"github.com/kidkender/prohealth-api/internal/lock"
	"github.com/kidkender/prohealth-api/internal/service"
	"github.com/kidkender/prohealth-api/internal/worker"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Container struct {
	AuthService        *service.AuthService
	UserService        *service.UserService
	ScheduleService    *service.ScheduleService
	AppointmentService *service.AppointmentService
	HospitalService    *service.HospitalService
	AdminService       *service.AdminService
	OutboxWorker       *worker.OutboxWorker
	RedisClient        *redis.Client
}

func New(db *gorm.DB, cfg *config.Config) (*Container, error) {
	mailService := service.NewMailService(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.MailFrom,
	)

	redisClient, err := NewRedisClient(config.NewRedisConfig())
	if err != nil {
		return nil, err
	}
	cacheClient := cache.New(redisClient)
	distLock, err := lock.New(redisClient)
	if err != nil {
		return nil, err
	}

	return &Container{
		AuthService:        service.NewAuthService(db),
		UserService:        service.NewUserService(db),
		ScheduleService:    service.NewScheduleService(db, distLock),
		AppointmentService: service.NewAppointmentService(db, mailService),
		HospitalService:    service.NewHospitalService(db, cacheClient),
		AdminService:       service.NewAdminService(db),
		OutboxWorker:       worker.NewOutboxWorker(db, redisClient),
		RedisClient:        redisClient,
	}, nil
}
