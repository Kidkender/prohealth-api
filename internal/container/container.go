package container

import (
	"log/slog"

	"github.com/kidkender/prohealth-api/internal/cache"
	"github.com/kidkender/prohealth-api/internal/config"
	"github.com/kidkender/prohealth-api/internal/lock"
	"github.com/kidkender/prohealth-api/internal/logger"
	"github.com/kidkender/prohealth-api/internal/outbox"
	"github.com/kidkender/prohealth-api/internal/service"
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
	OutboxWorker       *outbox.OutboxWorker
	RedisClient        *redis.Client
	Logger             *slog.Logger
	Publisher          *outbox.Publisher
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

	logger := logger.New()
	publisher := outbox.NewPublisher()

	return &Container{
		AuthService:        service.NewAuthService(db),
		UserService:        service.NewUserService(db),
		ScheduleService:    service.NewScheduleService(db, distLock),
		AppointmentService: service.NewAppointmentService(db, mailService, publisher),
		HospitalService:    service.NewHospitalService(db, cacheClient),
		AdminService:       service.NewAdminService(db),
		OutboxWorker:       outbox.NewOutboxWorker(db, redisClient),
		RedisClient:        redisClient,
		Logger:             logger,
	}, nil
}
