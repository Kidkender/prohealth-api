package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kidkender/prohealth-api/internal/cache"
	"github.com/kidkender/prohealth-api/internal/config"
	"github.com/kidkender/prohealth-api/internal/container"
	"github.com/kidkender/prohealth-api/internal/notification"
	"github.com/kidkender/prohealth-api/internal/service"
)

func main() {

	cfg := config.LoadEnv()

	redisClient, err := container.NewRedisClient(config.NewRedisConfig())
	if err != nil {
		log.Fatal(err)
	}

	mailService := service.NewMailService(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.MailFrom,
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	errClient := redisClient.XGroupCreateMkStream(
		ctx,
		string(cache.StreamNotifications),
		string(cache.NotificationGroup),
		"0",
	).Err()
	if errClient != nil && !strings.Contains(errClient.Error(), "BUSYGROUP") {
		log.Fatal(errClient)
	}

	consumer := notification.NewConsumer(
		redisClient,
		mailService,
		cfg.ConsumerName,
		5*time.Minute,
		10,
		30*time.Second,
	)
	consumer.Run(ctx)

}
